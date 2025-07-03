package provider

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
)

var (
	logDebugFunc func(string, ...interface{})
	logWarnFunc  func(string, ...interface{})
	logErrorFunc func(string, ...interface{})
	logInfoFunc  func(string, ...interface{})
)

func SetLogger(debug, info, warn, err func(string, ...interface{})) {
	logDebugFunc = debug
	logInfoFunc = info
	logWarnFunc = warn
	logErrorFunc = err
}

func logDebug(format string, a ...interface{}) {
	if logDebugFunc != nil {
		logDebugFunc(format, a...)
	}
}
func logWarn(format string, a ...interface{}) {
	if logWarnFunc != nil {
		logWarnFunc(format, a...)
	}
}
func logError(format string, a ...interface{}) {
	if logErrorFunc != nil {
		logErrorFunc(format, a...)
	}
}
func logInfo(format string, a ...interface{}) {
	if logInfoFunc != nil {
		logInfoFunc(format, a...)
	}
}

type Cloudflare struct {
	APIToken  string
	ZoneID    string
	Domain    string
	Subdomain string
}

func (c *Cloudflare) getZoneID(api *cloudflare.API) (string, error) {
	ctx := context.Background()
	resp, err := api.ListZonesContext(ctx, cloudflare.WithZoneFilters(c.Domain, "", ""))
	if err != nil {
		logError("Cloudflare ListZonesContext error: %v", err)
		return "", err
	}
	for _, z := range resp.Result {
		if z.Name == c.Domain {
			logDebug("Cloudflare getZoneID success: %s => %s", c.Domain, z.ID)
			return z.ID, nil
		}
	}
	logWarn("Cloudflare zone not found for domain: %s", c.Domain)
	return "", fmt.Errorf("zone not found for domain: %s", c.Domain)
}

func (c *Cloudflare) UpdateRecord(ip string) error {
	api, err := cloudflare.NewWithAPIToken(c.APIToken)
	if err != nil {
		logError("Cloudflare API token error: %v", err)
		return err
	}
	ctx := context.Background()
	zoneID := c.ZoneID
	if zoneID == "" {
		zoneID, err = c.getZoneID(api)
		if err != nil {
			return fmt.Errorf("auto get zone_id failed: %v", err)
		}
	}
	rc := cloudflare.ZoneIdentifier(zoneID)
	fqdn := fmt.Sprintf("%s.%s", c.Subdomain, c.Domain)

	logDebug("Cloudflare update: fqdn=%s, ip=%s", fqdn, ip)
	records, _, err := api.ListDNSRecords(ctx, rc, cloudflare.ListDNSRecordsParams{
		Type: "A",
		Name: fqdn,
	})
	if err != nil {
		logError("Cloudflare ListDNSRecords error: %v", err)
		return err
	}

	var foundSame bool
	var toUpdate []cloudflare.DNSRecord
	for _, record := range records {
		if record.Type == "A" && record.Name == fqdn {
			if record.Content == ip {
				foundSame = true
				break // 有完全一致的，直接跳过
			}
			toUpdate = append(toUpdate, record)
		}
	}
	if foundSame {
		logInfo("Cloudflare: record already up-to-date: %s => %s", fqdn, ip)
		return nil
	}
	if len(toUpdate) > 0 {
		for _, record := range toUpdate {
			updateParams := cloudflare.UpdateDNSRecordParams{
				ID:      record.ID,
				Type:    "A",
				Name:    fqdn,
				Content: ip,
				TTL:     record.TTL,
				Proxied: record.Proxied,
			}
			_, err = api.UpdateDNSRecord(ctx, rc, updateParams)
			if err != nil {
				logError("Cloudflare update record failed: %v", err)
				return err
			}
			logInfo("Cloudflare updated record: %s => %s", fqdn, ip)
		}
		return nil
	}
	// 没有同名记录，自动添加
	proxied := false
	createParams := cloudflare.CreateDNSRecordParams{
		Type:    "A",
		Name:    fqdn,
		Content: ip,
		TTL:     60,
		Proxied: &proxied,
	}
	_, err = api.CreateDNSRecord(ctx, rc, createParams)
	if err != nil {
		logError("Cloudflare create record failed: %v", err)
		return fmt.Errorf("record not found and create failed: %v", err)
	}
	logInfo("Cloudflare created new record: %s => %s", fqdn, ip)
	return nil
}
