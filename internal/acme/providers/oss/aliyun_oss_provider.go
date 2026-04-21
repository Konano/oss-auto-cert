package oss

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/charmbracelet/log"
	"github.com/go-acme/lego/v4/challenge/http01"
)

type AliYunOssHTTPProvider struct {
	bucket    string
	ossClient *oss.Client
}

func NewAliYunOssHTTPProvider(bucket string, ossClient *oss.Client) (*AliYunOssHTTPProvider, error) {
	return &AliYunOssHTTPProvider{
		bucket:    bucket,
		ossClient: ossClient,
	}, nil
}

func (s *AliYunOssHTTPProvider) Present(domain, token, keyAuth string) error {
	bucket, err := s.ossClient.Bucket(s.bucket)
	if err != nil {
		return fmt.Errorf("OSS: Bucket 异常: %w", err)
	}

	// 设置访问权限为只读
	acl := oss.ObjectACL(oss.ACLPublicRead)
	// 上传验证文件到 OSS 存储 bucket
	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	log.Infof("上传 HTTP 域名 (%s) 验证文件: key -> %s", domain, objectKey)
	err = bucket.PutObject(objectKey, bytes.NewReader([]byte(keyAuth)), acl)
	if err != nil {
		return fmt.Errorf("OSS: Failed to upload token to OSS: %w", err)
	}

	return nil
}

func (s *AliYunOssHTTPProvider) CleanUp(domain, token, keyAuth string) error {
	bucket, err := s.ossClient.Bucket(s.bucket)
	if err != nil {
		return fmt.Errorf("OSS: Bucket 异常: %w", err)
	}

	objectKey := strings.Trim(http01.ChallengePath(token), "/")
	log.Infof("删除 HTTP 域名 (%s) 验证文件: key -> %s", domain, objectKey)
	err = bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("OSS: could not remove file in OSS Bucket after HTTP challenge: %w", err)
	}

	return nil
}
