// http相关工具 author chenweidong

package http_utils

import (
	"net/http"
	"strings"
)

// 请求体转为字典
func ParamToMap(req *http.Request) map[string]string {
	paramStr := strings.Trim(req.URL.RawQuery, " ")
	paramMap := map[string]string{}
	params := strings.Split(paramStr, "&")
	for _, ele := range params {
		temp := strings.Split(ele, "=")
		if len(temp) > 1 {
			paramMap[temp[0]] = temp[1]
		}
	}
	return paramMap
}

// 获取请求IP地址
func RemoteIp(req *http.Request) string {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get("X-Forwarded-For"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr = strings.Split(remoteAddr, ":")[0]
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}