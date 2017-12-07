// 所有接口只接受GET请求 author chenweidong

package main

import (
	"net/http"
	httpUtils "hermes/utils/http"
	"fmt"
	. "hermes/entity"
	"os"
	. "hermes/utils/io"
	"hermes/timer"
	. "hermes/utils/string"
	"encoding/json"
)

// 监控页面
func indexHandler(_ http.ResponseWriter, _ *http.Request) Response {
	return Response{Code: 0, Data: SERVERS}
}

// 注册服务
func registerHandler(_ http.ResponseWriter, req *http.Request) Response {
	paramMap := httpUtils.ParamToMap(req)
	code := 0
	puk, err := Register(paramMap["id"], paramMap["sessionId"], paramMap["host"])
	if err != nil {
		code = err.Code
	}
	return Response{Data: struct {
		PublicKey string
		Length    int
		Timeout   int64
	}{PublicKey: puk, Length: CONF.KeyLength, Timeout: CONF.Timeout}, Code: code}
}

// 心跳检测
func heartBeatHandler(_ http.ResponseWriter, req *http.Request) Response {
	paramMap := httpUtils.ParamToMap(req)
	err := HeartBeat(paramMap["sessionId"])
	if err != nil {
		return Response{Code: err.Code}
	}
	return Response{Code: 0}
}

// 调用服务
func serverHandler(_ http.ResponseWriter, req *http.Request) Response {
	paramMap := httpUtils.ParamToMap(req)
	data := paramMap["sessionId"] + paramMap["serverId"] + paramMap["name"] + paramMap["data"]
	return Response{Code: 0, Data: data}
}

func main() {
	// 初始化配置
	InitConfig(os.Args)
	// 恢复备份数据
	RestoreServers(CONF.BackupPath)
	// 定时器
	timer.MainTimer()
	// web服务
	ROUTER = map[string]func(http.ResponseWriter, *http.Request) Response{
		"/":          indexHandler,
		"/register":  registerHandler,
		"/heartBeat": heartBeatHandler,
		"/server":    serverHandler,
	}
	http.HandleFunc("/", router)
	PrintInfo("server is running on port: %d", CONF.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", CONF.Port), nil)
}

// 路由表
var ROUTER map[string]func(http.ResponseWriter, *http.Request) Response

// 路由器
func router(resp http.ResponseWriter, req *http.Request) {
	reqId := RandomString(6)
	PrintInfo("[%s] request %s", reqId, req.RequestURI)
	if req.Method != "GET" || ROUTER[req.URL.Path] == nil {
		PrintInfo("[%s] [false] response %s", reqId, 404)
		resp.Write([]byte("{\"code\": -1, \"errMsg\": \"404\"}"))
		return
	}
	data := ROUTER[req.URL.Path](resp, req)
	respData, err := json.Marshal(data)
	if err != nil {
		resp.Write([]byte("{\"code\": -1, \"errMsg\": \"" + err.Error() + "\"}"))
		PrintInfo("[%s] [false] response %s", reqId, data)
	} else {
		resp.Write(respData)
		PrintInfo("[%s] [%t] response %s", reqId, data.Code == 0, data)
	}
}
