package provider

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	esa "github.com/alibabacloud-go/esa-20240910/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliyunESA struct {
	AccessKeyID     string
	AccessKeySecret string
	SiteID          string
	Endpoint        string
	Domain          string
	Subdomain       string
}

func (e *AliyunESA) UpdateRecord(ip string, recordType string) error {
	// 构建端点URL
	endpoint := "esa.ap-southeast-1.aliyuncs.com" // 默认端点
	if e.Endpoint != "" {
		switch e.Endpoint {
		case "cn-hangzhou":
			endpoint = "esa.cn-hangzhou.aliyuncs.com"
		case "ap-southeast-1":
			endpoint = "esa.ap-southeast-1.aliyuncs.com"
		default:
			// 如果用户输入的是完整的域名，直接使用
			if strings.Contains(e.Endpoint, ".aliyuncs.com") {
				endpoint = e.Endpoint
			}
		}
	}

	// 构建 OpenAPI Client
	cfg := &openapi.Config{
		AccessKeyId:     tea.String(e.AccessKeyID),
		AccessKeySecret: tea.String(e.AccessKeySecret),
		Endpoint:        tea.String(endpoint),
	}

	client, err := esa.NewClient(cfg)
	if err != nil {
		return err
	}

	fqdn := fmt.Sprintf("%s.%s", e.Subdomain, e.Domain)

	// 转换记录类型为ESA格式
	esaRecordType := recordType
	if recordType == "A" || recordType == "AAAA" {
		esaRecordType = "A/AAAA"
	}

	// 获取SiteID
	var siteID int64
	if e.SiteID == "" {
		// 自动获取SiteID
		siteID, err = e.getSiteID(client)
		if err != nil {
			return fmt.Errorf("auto get site_id failed: %v", err)
		}
	} else {
		// 使用配置的SiteID
		siteID, err = strconv.ParseInt(e.SiteID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid site_id: %v", err)
		}
	}

	// 查询现有记录
	listReq := &esa.ListRecordsRequest{
		SiteId:     tea.Int64(siteID),
		PageSize:   tea.Int32(50),
		RecordName: tea.String(fqdn),
		Type:       tea.String(esaRecordType),
	}

	listResp, err := client.ListRecords(listReq)
	if err != nil {
		return fmt.Errorf("list records failed: %v", err)
	}

	var toUpdate *esa.ListRecordsResponseBodyRecords
	var foundSame bool
	var finalValue string

	// 检查现有记录
	if listResp.Body != nil && listResp.Body.Records != nil {
		targetIsIPv4 := e.isIPv4(ip)
		for _, record := range listResp.Body.Records {
			if record.RecordName != nil && record.RecordType != nil && record.Data != nil {
				// 比较记录名和记录类型（考虑ESA的A/AAAA格式）
				recordTypeMatch := false
				if *record.RecordType == recordType || *record.RecordType == esaRecordType {
					recordTypeMatch = true
				}

				if *record.RecordName == fqdn && recordTypeMatch {
					// 解析现有值
					oldValue := ""
					if record.Data.Value != nil {
						oldValue = *record.Data.Value
					}

					// 分割现有IP，保留不同类型的IP
					existingIPs := strings.Split(oldValue, ",")
					var newIPs []string

					for _, existingIP := range existingIPs {
						existingIP = strings.TrimSpace(existingIP)
						if existingIP == "" {
							continue
						}
						
						// 检查IP类型
						isV4 := e.isIPv4(existingIP)
						
						// 如果是不同类型的IP，保留
						if isV4 != targetIsIPv4 {
							newIPs = append(newIPs, existingIP)
						}
					}

					// 添加新的目标IP
					newIPs = append(newIPs, ip)
					finalValue = strings.Join(newIPs, ",")

					// 检查是否完全一致
					if e.compareIPLists(oldValue, finalValue) {
						foundSame = true
						logInfo("AliyunESA: record already up-to-date: %s => %s", fqdn, ip)
						break
					}
					// 找到需要更新的记录
					toUpdate = record
					break
				}
			}
		}
	}

	if foundSame {
		return nil // 记录已是最新，无需操作
	}

	if toUpdate != nil {
		// 更新现有记录
		updateReq := &esa.UpdateRecordRequest{
			RecordId: toUpdate.RecordId,
			Data: &esa.UpdateRecordRequestData{
				Value: tea.String(finalValue),
			},
		}

		_, err = client.UpdateRecord(updateReq)
		if err != nil {
			logError("AliyunESA: update record failed for %s: %v", fqdn, err)
			return fmt.Errorf("update record failed: %v", err)
		}

		logInfo("AliyunESA: record updated: %s => %s", fqdn, ip)
		return nil
	}

	// 创建新记录
	logWarn("AliyunESA: record not found for %s, will try to create.", fqdn)

	createReq := &esa.CreateRecordRequest{
		SiteId:     tea.Int64(siteID),
		RecordName: tea.String(fqdn),
		Type:       tea.String(esaRecordType),
		Data: &esa.CreateRecordRequestData{
			Value: tea.String(ip),
		},
		Ttl: tea.Int32(300), // 默认TTL 5分钟
	}

	_, err = client.CreateRecord(createReq)
	if err != nil {
		logError("AliyunESA: create record failed for %s: %v", fqdn, err)
		return fmt.Errorf("create record failed: %v", err)
	}

	logInfo("AliyunESA: record created: %s => %s", fqdn, ip)

	return nil
}

// compareRecordData 比较记录数据是否相同
func (e *AliyunESA) compareRecordData(recordData interface{}, ip string) bool {
	if data, ok := recordData.(*esa.ListRecordsResponseBodyRecordsData); ok {
		if data.Value != nil {
			return *data.Value == ip
		}
	}
	return false
}

// getSiteID 根据域名自动获取SiteID
func (e *AliyunESA) getSiteID(client *esa.Client) (int64, error) {
	listReq := &esa.ListSitesRequest{
		SiteName:       tea.String(e.Domain),
		SiteSearchType: tea.String("exact"), // 精确匹配域名
		PageSize:       tea.Int32(10),
	}

	listResp, err := client.ListSites(listReq)
	if err != nil {
		logError("AliyunESA ListSites error: %v", err)
		return 0, err
	}

	if listResp.Body == nil || listResp.Body.Sites == nil || len(listResp.Body.Sites) == 0 {
		return 0, fmt.Errorf("site not found for domain: %s", e.Domain)
	}

	// 查找匹配的站点
	for _, site := range listResp.Body.Sites {
		if site.SiteName != nil && site.SiteId != nil {
			if *site.SiteName == e.Domain {
				logDebug("AliyunESA getSiteID success: %s => %d", e.Domain, *site.SiteId)
				return *site.SiteId, nil
			}
		}
	}

	return 0, fmt.Errorf("site not found for domain: %s", e.Domain)
}

func (e *AliyunESA) isIPv4(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	return parsed.To4() != nil
}

func (e *AliyunESA) compareIPLists(list1, list2 string) bool {
	s1 := strings.Split(list1, ",")
	s2 := strings.Split(list2, ",")
	
	// Filter empty strings
	var cleanS1, cleanS2 []string
	for _, s := range s1 {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			cleanS1 = append(cleanS1, trimmed)
		}
	}
	for _, s := range s2 {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			cleanS2 = append(cleanS2, trimmed)
		}
	}

	if len(cleanS1) != len(cleanS2) {
		return false
	}

	m1 := make(map[string]bool)
	for _, ip := range cleanS1 {
		m1[ip] = true
	}

	for _, ip := range cleanS2 {
		if !m1[ip] {
			return false
		}
	}
	return true
}
