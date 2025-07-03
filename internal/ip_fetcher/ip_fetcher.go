package ip_fetcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"OpenDDNS/internal/config"
)

func FetchIP(src config.IPSrc) (string, error) {
	resp, err := http.Get(src.URL)
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
