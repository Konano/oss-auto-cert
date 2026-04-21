package alioss

import (
	"bytes"
	"fmt"

	cas20200407 "github.com/alibabacloud-go/cas-20200407/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/charmbracelet/log"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/konano/oss-auto-cert/internal/config"
	"github.com/konano/oss-auto-cert/internal/types"
	"github.com/konano/oss-auto-cert/pkg/utils"
)

type CasService struct {
	client *cas20200407.Client
}

func NewCasService(access oss.Credentials) *CasService {
	client, err := cas20200407.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(access.GetAccessKeyID()),
		AccessKeySecret: tea.String(access.GetAccessKeySecret()),

		// endpoint 参考：https://api.aliyun.com/product/cas
		Endpoint: tea.String("cas.aliyuncs.com"),
	})
	if err != nil {
		log.Fatalf("%v", err)
	}

	return &CasService{
		client: client,
	}
}

// GetDetail 根据 ID 获取证书详情
func (s *CasService) GetDetail(certID int64) (*cas20200407.GetUserCertificateDetailResponseBody, error) {
	request := new(cas20200407.GetUserCertificateDetailRequest)
	request.SetCertId(certID)
	resp, err := s.client.GetUserCertificateDetail(request)
	if err != nil {
		return nil, fmt.Errorf("获取证书 (%d) 详情异常: %w", certID, err)
	}

	if *resp.StatusCode != 200 {
		return nil, fmt.Errorf("获取证书 (%d) 详情请求响应异常: 状态码 -> %d；响应: %s", certID, resp.StatusCode, resp)
	}

	return resp.Body, nil
}

// IsExpired 检查证书是否过期
func (s *CasService) IsExpired(certID int64) (bool, error) {
	detail, err := s.GetDetail(certID)
	if err != nil {
		return false, err
	}

	if *detail.Expired || utils.DateIsExpire(*detail.EndDate, config.GetExpiredEarlyTime()) {
		log.Warnf("证书 (%s, %d) 过期，需要更换新证书", *detail.Name, certID)
		return true, nil
	} else {
		log.Infof("证书 (%s, %d) 未过期，过期日期: %s, 还剩 %d 天, 将提前 %d 天过期", *detail.Name, certID, *detail.EndDate,
			utils.TimeDiffDay(*detail.EndDate), config.GetExpiredEarlyDay())
		return false, nil
	}
}

// Upload 上传证书到 证书管理服务
func (s *CasService) Upload(cert *certificate.Resource) (*types.CertInfo, error) {
	name := utils.ShortDomain(cert.Domain) + "-" + utils.SplitFirst(utils.UUID(), "-")

	req := new(cas20200407.UploadUserCertificateRequest)
	// 证书名称
	req.Name = tea.String(name)
	// 证书私钥
	req.Key = tea.String(bytes.NewBuffer(cert.PrivateKey).String())
	// 证书内容
	req.Cert = tea.String(bytes.NewBuffer(cert.Certificate).String())
	// 上传证书到证书管理服务
	resp, err := s.client.UploadUserCertificate(req)
	if err != nil {
		return nil, fmt.Errorf("上传证书失败：%w", err)
	}

	if *resp.StatusCode != 200 {
		return nil, fmt.Errorf("上传证书请求响应异常: 状态码 -> %d；响应: %s", resp.StatusCode, resp)
	}

	upload := resp.Body
	log.Infof("上传证书成功响应：%s", upload)

	return &types.CertInfo{
		ID:     *upload.CertId,
		Name:   name,
		Domain: cert.Domain,
	}, nil
}
