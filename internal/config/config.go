package config

import (
	"io"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
)

const (
	DefaultExpiredEarly = 15
	DefaultConfigPath   = "./config.yaml"
)

var (
	// expiredEarlyDay 提前过期时间点 默认 15 天
	expiredEarlyDay = DefaultExpiredEarly
	// expiredEarlyTime 提前过期时间
	expiredEarlyTime = time.Hour * 24 * DefaultExpiredEarly
)

type Config struct {
	// 配置文件路径
	// 默认路径: DefaultConfigPath
	Path string
	// 通知地址
	Webhook string `yaml:"webhook"`
	// 通知消息模版
	WebhookTpl string `yaml:"webhook-tpl"`
	// Acme 配置
	Acme Acme `yaml:"acme"`
	// Bucket 配置
	Buckets []Bucket `yaml:"buckets"`
}

type Acme struct {
	// 证书申请邮箱
	Email string `yaml:"email"`

	// 证书保存位置
	DataDir string `yaml:"data-dir"`

	// 证书提前 renew 时间
	ExpiredEarly int `yaml:"expired-early"`
}

// Bucket OSS 存储 Bucket 配置
type Bucket struct {
	// bucket 名称
	Name string `yaml:"name"`
	// Endpoint
	Endpoint string `yaml:"endpoint"`
}

// LoadOptions 加载配置
func (conf *Config) LoadOptions() {
	if conf.Path == "" {
		conf.Path = DefaultConfigPath
	}

	f, err := os.Open(conf.Path)
	if err != nil {
		log.Fatalf("读取配置文件 %s 出错: %s", conf.Path, err.Error())
	}
	defer f.Close()

	bts, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("读取配置文件 %s 出错: %s", conf.Path, err.Error())
	}

	err = yaml.Unmarshal(bts, conf)
	if err != nil {
		log.Fatalf("读取配置文件 %s 出错: %s", conf.Path, err.Error())
	}

	conf.setExpiredEarlyTime()

	log.Debugf("配置文件: %v", conf)
}

func (conf *Config) LoadOptionsFromEnv() {
	value := os.Getenv("ACME_EMAIL")
	if value != "" {
		conf.Acme.Email = value
		log.Debugf("set acme (email) from env: %s", value)
	}

	value = os.Getenv("ACME_DATA_DIR")
	if value != "" {
		conf.Acme.DataDir = value
		log.Debugf("set acme (data dir) from env: %s", value)
	}

	value = os.Getenv("ACME_EXPIRED_EARLY")
	if value != "" {
		if valueInt, err := strconv.Atoi(value); err != nil {
			log.Warnf("%v", err)
		} else {
			conf.Acme.ExpiredEarly = valueInt
			log.Debugf("set acme (expired early) from env: %d", valueInt)
			conf.setExpiredEarlyTime()
		}
	}

	log.Debugf("配置文件: %v", conf)
}

func (conf *Config) setExpiredEarlyTime() {
	// 根据配置更新证书更新提前过期时间
	expiredEarlyDay = max(DefaultExpiredEarly, conf.Acme.ExpiredEarly)
	expiredEarlyTime = time.Hour * 24 * time.Duration(expiredEarlyDay)
}

func GetExpiredEarlyDay() int {
	return expiredEarlyDay
}

func GetExpiredEarlyTime() time.Duration {
	return expiredEarlyTime
}
