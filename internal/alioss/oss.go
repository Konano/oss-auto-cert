package alioss

import (
	"fmt"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/charmbracelet/log"
	"github.com/konano/oss-auto-cert/internal/config"
	"github.com/konano/oss-auto-cert/internal/types"
	"github.com/konano/oss-auto-cert/pkg/utils"
)

type AliYunOss struct {
	name   string
	Client *oss.Client
}

func NewAliYunOss(bucket config.Bucket, access oss.Credentials) (*AliYunOss, error) {
	client, err := oss.New(bucket.Endpoint, access.GetAccessKeyID(), access.GetAccessKeySecret())
	if err != nil {
		return nil, fmt.Errorf("创建 OSS Client 异常: %w", err)
	}

	return &AliYunOss{
		name:   bucket.Name,
		Client: client,
	}, nil
}

// GetCerts 获取 Bucket 下自定义域名证书 ID 信息
func (b *AliYunOss) GetCerts() ([]*types.CertInfo, error) {
	// 获取 Bucket 全部自定义域名列表
	result, err := b.Client.ListBucketCname(b.name)
	if err != nil {
		return nil, fmt.Errorf("获取 Bucket (%s) 下自定义域名列表异常: %w", b.name, err)
	}

	cnameArr := result.Cname
	if len(cnameArr) <= 0 {
		return nil, fmt.Errorf("Bucket (%s) 自定义域名为空，请检查 Bucket 配置", b.name)
	}

	infos := make([]*types.CertInfo, 0, len(cnameArr))
	for _, cname := range cnameArr {
		log.Debugf("处理 Bucket (%s) 自定义域名: %s", b.name, cname.Domain)
		log.Debugf("Status: %s", cname.Status)
		log.Debugf("Domain: %s", cname.Domain)
		log.Debugf("LastModified: %s", cname.LastModified)

		// 检查证书信息
		cert := cname.Certificate
		// 域名证书信息
		log.Debugf("证书信息: %s", cert)
		log.Debugf("Type: %s", cert.Type)
		log.Debugf("CertId: %s", cert.CertId)
		log.Debugf("Status: %s", cert.Status)
		log.Debugf("CreationDate: %s", cert.CreationDate)
		log.Debugf("Fingerprint: %s", cert.Fingerprint)
		log.Debugf("ValidStartDate: %s", cert.ValidStartDate)
		log.Debugf("ValidEndDate: %s", cert.ValidEndDate)

		certID := cert.CertId
		if certID == "" {
			return nil, fmt.Errorf("Bucket (%s) 域名 (%s) 证书信息 ID 为空", b.name, cname.Domain)
		}

		int64Str := utils.SplitFirst(certID, "-")
		int64ID, err := strconv.ParseInt(int64Str, 10, 64)
		if err != nil {
			return nil, err
		}

		infos = append(infos, &types.CertInfo{
			ID:     int64ID,
			Region: utils.SplitGetN(certID, "-", 2, 2),
			Domain: cname.Domain,
		})
	}

	return infos, nil
}

// UpgradeCert 更新域名绑定的证书
func (b *AliYunOss) UpgradeCert(domain string, certID string) error {
	log.Debugf("更新域名 (%s) 证书：%s", domain, certID)

	putCname := oss.PutBucketCname{
		Cname: domain,
		CertificateConfiguration: &oss.CertificateConfiguration{
			CertId:            certID,
			Force:             true,
			DeleteCertificate: false,
		},
	}
	err := b.Client.PutBucketCnameWithCertificate(b.name, putCname)
	if err != nil {
		return fmt.Errorf("Bucket (%s) 更新证书失败：%w", b.name, err)
	}

	log.Infof("OSS 自定义域名 (%s) 证书更新成功! 证书: %s", domain, certID)

	return nil
}
