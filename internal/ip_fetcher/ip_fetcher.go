package ip_fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"OpenDDNS/internal/config"
)

func FetchIP(src config.IPSrc) (string, error) {
	return FetchIPWithNetwork(src, "")
}

// FetchIPWithNetwork 获取IP地址，支持强制指定网络类型
// networkType: "ipv4", "ipv6" 或 "" (自动)
func FetchIPWithNetwork(src config.IPSrc, networkType string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 根据网络类型创建自定义 Transport
	if networkType != "" {
		transport := &http.Transport{}
		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}

		switch strings.ToLower(networkType) {
		case "ipv4":
			// 强制使用 IPv4
			dialer.FallbackDelay = -1 // 禁用 IPv6 fallback
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, "tcp4", addr)
			}
		case "ipv6":
			// 强制使用 IPv6
			transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, "tcp6", addr)
			}
		}

		client.Transport = transport
	}

	resp, err := client.Get(src.URL)
	if err != nil {
		return "", fmt.Errorf("fetch %s failed: %v", src.Name, err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	switch src.Type {
	case "json":
		var m map[string]interface{}
		if err := json.Unmarshal(body, &m); err != nil {
			return "", err
		}
		// 支持多级 json_path，如 data.addr
		path := strings.Split(src.JSONPath, ".")
		var v interface{} = m
		for _, p := range path {
			if mm, ok := v.(map[string]interface{}); ok {
				v = mm[p]
			} else {
				return "", fmt.Errorf("json path error")
			}
		}
		if ip, ok := v.(string); ok {
			return ip, nil
		}
		return "", fmt.Errorf("json path not string")
	case "trace":
		lines := strings.Split(string(body), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ip=") {
				return strings.TrimPrefix(line, "ip="), nil
			}
		}
		return "", fmt.Errorf("ip not found in trace")
	case "text":
		// 直接返回响应体内容（去除前后空白字符）
		ip := strings.TrimSpace(string(body))
		if ip == "" {
			return "", fmt.Errorf("empty response from %s", src.Name)
		}
		return ip, nil
	default:
		return "", fmt.Errorf("unknown type: %s", src.Type)
	}
}

// 日志等级集成
func LogDebug(format string, a ...interface{}) {
	if logDebugFunc != nil {
		logDebugFunc(format, a...)
	}
}
func LogWarn(format string, a ...interface{}) {
	if logWarnFunc != nil {
		logWarnFunc(format, a...)
	}
}
func LogError(format string, a ...interface{}) {
	if logErrorFunc != nil {
		logErrorFunc(format, a...)
	}
}

var (
	logDebugFunc func(string, ...interface{})
	logWarnFunc  func(string, ...interface{})
	logErrorFunc func(string, ...interface{})
)

func SetLogger(debug, warn, err func(string, ...interface{})) {
	logDebugFunc = debug
	logWarnFunc = warn
	logErrorFunc = err
}

// GetIPType 判断IP地址类型，返回A或AAAA
func GetIPType(ip string) string {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}
	if parsedIP.To4() != nil {
		return "A" // IPv4
	}
	return "AAAA" // IPv6
}

// DetermineRecordType 根据配置和IP地址确定DNS记录类型
func DetermineRecordType(ip string, configType string) string {
	switch strings.ToLower(configType) {
	case "a":
		return "A"
	case "aaaa":
		return "AAAA"
	case "auto", "":
		return GetIPType(ip)
	default:
		return GetIPType(ip) // 默认自动检测
	}
}
