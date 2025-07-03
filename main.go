package main

import (
	"OpenDDNS/internal/config"
	ipfetcher "OpenDDNS/internal/ip_fetcher"
	"OpenDDNS/internal/logger"
	"OpenDDNS/internal/provider"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func checkUpdate(currentVersion string) {
	const repoAPI = "https://api.github.com/repos/GloryRedstoneUnion/OpenDDNS/releases/latest"
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(repoAPI)
	if err != nil {
		fmt.Println("[WARN] Failed to check update:", err)
		return
	}
	defer resp.Body.Close()
	var data struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("[WARN] Failed to parse update info:", err)
		return
	}
	if data.TagName != "" && data.TagName != currentVersion {
		fmt.Printf("\033[33m[UPDATE] New version available: %s → %s\nDownload: %s\033[0m\n", currentVersion, data.TagName, data.HTMLURL)
	} else {
		fmt.Println("[INFO] You are using the latest version.")
	}
}

func getMajorityIP(sources []config.IPSrc) string {
	ipResults := make(map[string]string)
	available := 0
	for _, src := range sources {
		ip, err := ipfetcher.FetchIP(src)
		if err == nil && ip != "" {
			logger.Debug("IP source %s returned: %s", src.Name, ip)
			ipResults[src.Name] = ip
			available++
		} else {
			logger.Warn("IP source %s failed: %v", src.Name, err)
		}
	}
	if available == 0 {
		logger.Error("No available IP sources.")
		return ""
	}
	if available < 2 {
		for _, ip := range ipResults {
			logger.Debug("Only one valid IP: %s", ip)
			return ip
		}
	}
	ipCounts := make(map[string]int)
	for _, ip := range ipResults {
		ipCounts[ip]++
	}
	var majorityIP string
	maxCount := 0
	for ip, count := range ipCounts {
		if count > maxCount {
			maxCount = count
			majorityIP = ip
		}
	}
	if maxCount >= 2 {
		logger.Debug("Majority IP: %s", majorityIP)
		return majorityIP
	}
	logger.Warn("IP conflict, using priority list.")
	for _, src := range sources {
		if ip, ok := ipResults[src.Name]; ok {
			logger.Debug("Priority IP: %s", ip)
			return ip
		}
	}
	return ""
}

func main() {
	// Parse config file flag, support -c and --config, and --no-check-update
	var configPath string
	var noCheckUpdate bool
	flag.StringVar(&configPath, "c", "config.yml", "Path to config file")
	flag.StringVar(&configPath, "config", "config.yml", "Path to config file")
	flag.BoolVar(&noCheckUpdate, "no-check-update", false, "Skip update check")
	flag.Parse()

	// Print startup info in English with color, show version
	blue := "\033[34m"
	green := "\033[32m"
	reset := "\033[0m"
	fmt.Printf("%s==============================%s\n", blue, reset)
	fmt.Printf("%s   OpenDDNS - Modern Multi-Cloud DDNS Tool%s\n", green, reset)
	fmt.Printf("%s   Version: %s   Build: %s%s\n", blue, Version, BuildTime, reset)
	fmt.Printf("%s   https://github.com/GloryRedstoneUnion/OpenDDNS%s\n", blue, reset)
	fmt.Printf("%s==============================%s\n", blue, reset)

	if !noCheckUpdate {
		checkUpdate(Version)
	}

	// 检查配置文件是否存在，不存在则自动生成并退出
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := `provider: "cloudflare"

domain: "example.com"
subdomain: "www"

log_level: "info"
log_file: ""

ip_sources:
  - name: "bilibili"
    url: "https://api.live.bilibili.com/xlive/web-room/v1/index/getIpInfo"
    type: "json"
    json_path: "data.addr"
  - name: "cloudflare"
    url: "https://www.cloudflare-cn.com/cdn-cgi/trace"
    type: "trace"

update_interval_minutes: 5

cloudflare:
  api_token: "YOUR_CLOUDFLARE_API_TOKEN"
  zone_id: ""
aliyun:
  access_key_id: "YOUR_ALIYUN_ACCESS_KEY_ID"
  access_key_secret: "YOUR_ALIYUN_ACCESS_KEY_SECRET"
  endpoint: "alidns.aliyuncs.com"
`
		err := os.WriteFile(configPath, []byte(defaultConfig), 0644)
		if err != nil {
			log.Fatalf("Failed to create default config: %v", err)
		}
		logger.Info("No config file found. A default config.yml has been created. Please edit it before running OpenDDNS.")
		fmt.Println("\033[33m[WARN] No config file found. A default config.yml has been created. Please edit it before running OpenDDNS.\033[0m")
		time.Sleep(5 * time.Second)
		return
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	fmt.Printf("%sConfig file:%s %s%s.%s%s\n", green, reset, blue, cfg.Subdomain, cfg.Domain, reset)
	fmt.Printf("%sDNS Provider:%s %s%s%s\n", green, reset, blue, cfg.Provider, reset)
	fmt.Printf("%sLog Level:%s %s%s%s\n", green, reset, blue, cfg.LogLevel, reset)
	if cfg.LogFile != "" {
		fmt.Printf("%sLog File:%s %s%s%s\n", green, reset, blue, cfg.LogFile, reset)
	} else {
		fmt.Printf("%sLog File:%s %sConsole only%s\n", green, reset, blue, reset)
	}
	fmt.Printf("%sIP Source Count:%s %s%d%s\n", green, reset, blue, len(cfg.IPSources), reset)
	fmt.Printf("%sSupported DNS Providers:%s %sCloudflare, Alicloud%s\n", green, reset, blue, reset)
	fmt.Printf("%s==============================%s\n", blue, reset)

	logger.SetLogLevel(cfg.LogLevel)
	logger.SetLogFile(cfg.LogFile)
	// Inject logger
	ipfetcher.SetLogger(logger.Debug, logger.Warn, logger.Error)
	provider.SetLogger(logger.Debug, logger.Info, logger.Warn, logger.Error)

	// Log program start
	logger.Info("Program started.")

	// Handle exit signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		logger.Info("Program exited.")
		os.Exit(0)
	}()

	var dnsProvider provider.DNSProvider
	switch cfg.Provider {
	case "cloudflare":
		dnsProvider = &provider.Cloudflare{
			APIToken:  cfg.Cloudflare.APIToken,
			ZoneID:    cfg.Cloudflare.ZoneID,
			Domain:    cfg.Domain,
			Subdomain: cfg.Subdomain,
		}
	case "aliyun":
		dnsProvider = &provider.Aliyun{
			AccessKeyID:     cfg.Aliyun.AccessKeyID,
			AccessKeySecret: cfg.Aliyun.AccessKeySecret,
			Domain:          cfg.Domain,
			Subdomain:       cfg.Subdomain,
			Endpoint:        cfg.Aliyun.Endpoint,
		}
	// case "tencentcloud":
	// 	... reserved for tencentcloud ...
	default:
		log.Fatalf("Unsupported provider: %s", cfg.Provider)
	}

	var lastIP string
	// 关键元素染色
	domainColor := "\033[36m"   // 青色
	providerColor := "\033[35m" // 紫色
	reset = "\033[0m"           // 修正为赋值，不再用 :=
	logger.Info("DDNS service started for %s%s.%s%s with provider %s%s%s", domainColor, cfg.Subdomain, cfg.Domain, reset, providerColor, cfg.Provider, reset)
	ticker := time.NewTicker(time.Duration(cfg.UpdateIntervalMinutes) * time.Minute)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		newIP := getMajorityIP(cfg.IPSources)
		if newIP == "" {
			logger.Warn("Failed to determine public IP.")
			continue
		}
		if newIP == lastIP {
			logger.Debug("IP not changed: %s", newIP)
			continue
		}
		logger.Info("Detected public IP: %s", newIP)
		err := dnsProvider.UpdateRecord(newIP)
		if err != nil {
			logger.Error("Error updating DNS record: %v", err)
		} else {
			logger.Info("DNS record updated successfully.")
			lastIP = newIP
		}
	}
}
