package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type IPSrc struct {
	Name     string `yaml:"name"`
	URL      string `yaml:"url"`
	Type     string `yaml:"type"` // json/trace
	JSONPath string `yaml:"json_path,omitempty"`
}

type CloudflareConfig struct {
	APIToken string `yaml:"api_token"`
	ZoneID   string `yaml:"zone_id"`
}

type AliyunConfig struct {
	AccessKeyID     string `yaml:"access_key_id"`
	AccessKeySecret string `yaml:"access_key_secret"`
	Endpoint        string `yaml:"endpoint"`
}

type TencentCloudConfig struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
}

type Config struct {
	Provider              string             `yaml:"provider"`
	Domain                string             `yaml:"domain"`
	Subdomain             string             `yaml:"subdomain"`
	IPSources             []IPSrc            `yaml:"ip_sources"`
	UpdateIntervalMinutes int                `yaml:"update_interval_minutes"`
	Cloudflare            CloudflareConfig   `yaml:"cloudflare"`
	Aliyun                AliyunConfig       `yaml:"aliyun"`
	TencentCloud          TencentCloudConfig `yaml:"tencentcloud"`
	LogLevel              string             `yaml:"log_level"`
	LogFile               string             `yaml:"log_file"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
