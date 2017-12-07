// http相关工具 author chenweidong

package http

import (
	"net/http"
	"strings"
)

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