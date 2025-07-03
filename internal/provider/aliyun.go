package provider

import (
	"fmt"

	alidns "github.com/alibabacloud-go/alidns-20150109/v4/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type Aliyun struct {
	AccessKeyID     string
	AccessKeySecret string
	Domain          string
	Subdomain       string
	Endpoint        string // 可选
}

func (a *Aliyun) UpdateRecord(ip string) error {
	// 构建 OpenAPI Client
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(a.AccessKeyID),
		AccessKeySecret: tea.String(a.AccessKeySecret),
	}
	if a.Endpoint != "" {
		cfg.Endpoint = tea.String(a.Endpoint)
	}
	client, err := alidns.NewClient(cfg)
	if err != nil {
		return err
	}
	fqdn := fmt.Sprintf("%s.%s", a.Subdomain, a.Domain)
	// 查询记录
	descReq := &alidns.DescribeSubDomainRecordsRequest{
		SubDomain: tea.String(fqdn),
		Type:      tea.String("A"),
	}
	descResp, err := client.DescribeSubDomainRecords(descReq)
	if err != nil {
		return err
	}
	records := descResp.Body.DomainRecords.Record
	var toUpdate *alidns.DescribeSubDomainRecordsResponseBodyDomainRecordsRecord
	for _, record := range records {
		if *record.RR == a.Subdomain && *record.Value == ip {
			if logInfo != nil {
				logInfo("Aliyun: record already up-to-date: %s => %s", fqdn, ip)
			}
			return nil // 完全一致，无需操作
		}
		if *record.RR == a.Subdomain {
			toUpdate = record
		}
	}
	if toUpdate != nil {
		// 有同RR，更新
		updateReq := &alidns.UpdateDomainRecordRequest{
			RecordId: tea.String(*toUpdate.RecordId),
			RR:       tea.String(a.Subdomain),
			Type:     tea.String("A"),
			Value:    tea.String(ip),
		}
		_, err = client.UpdateDomainRecord(updateReq)
		if err == nil && logInfo != nil {
			logInfo("Aliyun: record updated: %s => %s", fqdn, ip)
		}
		return err
	}
	// 没有同RR，自动添加
	if logWarn != nil {
		logWarn("Aliyun: record not found for %s, will try to add.", fqdn)
	}
	addReq := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(a.Domain),
		RR:         tea.String(a.Subdomain),
		Type:       tea.String("A"),
		Value:      tea.String(ip),
	}
	_, err = client.AddDomainRecord(addReq)
	if err != nil {
		if logError != nil {
			logError("Aliyun: add record failed for %s: %v", fqdn, err)
		}
		return fmt.Errorf("add record failed: %v", err)
	}
	if logInfo != nil {
		logInfo("Aliyun: record created: %s => %s", fqdn, ip)
	}
	return nil
}
