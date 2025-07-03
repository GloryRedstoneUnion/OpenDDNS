package provider

// DNSProvider 统一接口
type DNSProvider interface {
	UpdateRecord(ip string) error
}
