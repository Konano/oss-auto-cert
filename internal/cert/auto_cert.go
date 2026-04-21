package cert

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/charmbracelet/log"
	"github.com/nekoimi/oss-auto-cert/internal/acme"
	"github.com/nekoimi/oss-auto-cert/internal/alioss"
	"github.com/nekoimi/oss-auto-cert/internal/config"
	"github.com/nekoimi/oss-auto-cert/internal/types"
	"github.com/nekoimi/oss-auto-cert/pkg/webhook"
)

type AutoCert struct {
	ctx           context.Context
	running       atomic.Bool
	buckets       []config.Bucket
	access        oss.Credentials
	cas           *alioss.CasService
	cdn           *alioss.CDNService
	acme          *acme.LegoAcme
	messageCh     chan string
	messageHandle func(message string)
}

func NewAutoCert(ctx context.Context, conf *config.Config) (*AutoCert, error) {
	credentialsProvider, err := oss.NewEnvironmentVariableCredentialsProvider()
	if err != nil {
		log.Errorf("缺少 OSS 访问 AccessKey 环境变量配置: %s", err.Error())
		return nil, err
	}

	access := credentialsProvider.GetCredentials()
	c := &AutoCert{
		ctx:       ctx,
		buckets:   conf.Buckets,
		access:    access,
		cas:       alioss.NewCasService(access),
		cdn:       alioss.NewCDNService(access),
		acme:      acme.NewLegoAcme(conf.Acme),
		messageCh: make(chan string, 32),
	}
	c.running.Store(false)
	if len(c.buckets) <= 0 {
		log.Warnf("OSS 存储 Bucket 配置为空!")
	} else {
		for _, b := range c.buckets {
			log.Debugf("Bucket 开启监测: %s => %s", b.Name, b.Endpoint)
		}
	}

	if conf.Webhook != "" {
		tplHook := webhook.NewTplWebHook(conf.Webhook, conf.WebhookTpl)
		c.withMessageHandle(func(message string) {
			tplHook.SendHook(message)
		})
	} else {
		c.withMessageHandle(func(message string) {
			// 打印日志
			log.Info(message)
		})
	}

	return c, nil
}

func (c *AutoCert) withMessageHandle(messageHandle func(message string)) {
	c.messageHandle = messageHandle
}

func (c *AutoCert) sendMessage(message string) {
	select {
	case <-c.ctx.Done():
		return
	case c.messageCh <- message:
	}
}

func (c *AutoCert) ScheduleRun() {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case message := <-c.messageCh:
				if c.messageHandle != nil {
					c.messageHandle(message)
				}
			}
		}
	}()

	go c.run()

	go func() {
		tick := time.NewTicker(6 * time.Hour)
		defer tick.Stop()
		for {
			select {
			case <-c.ctx.Done():
				return
			case <-tick.C:
				go c.run()
			}
		}
	}()
}

func (c *AutoCert) Stop() {
	c.acme.Stop()
}

func (c *AutoCert) run() {
	if !c.running.CompareAndSwap(false, true) {
		return
	}
	defer func() {
		c.running.Store(false)
	}()

	for _, bucket := range c.buckets {
		log.Debugf("开始检测 Bucket: %s ...", bucket.Name)

		b, err := alioss.NewAliYunOss(bucket, c.access)
		if err != nil {
			log.Errorf(err.Error())
			continue
		}

		infos, err := b.GetCerts()
		if err != nil {
			log.Errorf(err.Error())
			continue
		}

		for _, info := range infos {
			expired, err := c.cas.IsExpired(info.ID)
			if err != nil {
				log.Errorf(err.Error())
				continue
			}

			if expired {
				bucketName := bucket.Name
				domain := info.Domain
				region := info.Region
				messagePrefix := fmt.Sprintf("[oss-auto-cert] Bucket <%s> 域名: %s\n", bucketName, domain)

				c.sendMessage(fmt.Sprintf("%s 证书过期，需要更换新证书", messagePrefix))

				// 过期，申请新证书
				cert, err := c.acme.Obtain(bucketName, domain, b.Client)
				if err != nil {
					log.Errorf(err.Error())
					continue
				}

				// 上传证书文件到阿里云数字证书管理服务
				certInfo, err := c.cas.Upload(cert)
				if err != nil {
					log.Errorf(err.Error())
					c.sendMessage(fmt.Sprintf("%s 上传证书到数字证书管理异常: %s", messagePrefix, err.Error()))
					continue
				}

				certInfo.Region = region

				log.Infof("证书上传信息: %s", certInfo)

				go func(bucketName string, domain string, region string, certInfo *types.CertInfo, messagePrefix string) {
					// 更新 OSS 域名关联的证书
					err := b.UpgradeCert(domain, fmt.Sprintf("%d-%s", certInfo.ID, region))
					if err != nil {
						log.Errorf(err.Error())
						c.sendMessage(fmt.Sprintf("%s 更新 OSS 域名证书失败: %s", messagePrefix, err.Error()))
					} else {
						c.sendMessage(fmt.Sprintf("%s 更新 OSS 域名证书成功，请及时检查证书生效", messagePrefix))
					}
				}(bucketName, domain, region, certInfo, messagePrefix)

				go func(domain string, certInfo *types.CertInfo, messagePrefix string) {
					// 更新 CDN 关联的域名证书
					err := c.cdn.UpgradeCert(domain, certInfo)
					if err != nil {
						log.Errorf(err.Error())
						c.sendMessage(fmt.Sprintf("%s 更新 CDN 加速域名证书失败: %s", messagePrefix, err.Error()))
					} else {
						c.sendMessage(fmt.Sprintf("%s 更新 CDN 加速域名证书成功，请及时检查证书生效", messagePrefix))
					}
				}(domain, certInfo, messagePrefix)
			}
		}
	}
}
