// 所有接口只接受GET请求 author chenweidong

package main

import (
	"net/http"
	httpUtils "hermes/utils/http"
	. "hermes/entity"
	"os"
	"hermes/timer"
	"hermes/router"
)

// 监控页面
func indexHandler(_ http.ResponseWriter, _ *http.Request) interface{} {
	return Response{Code: 0, Data: SERVERS}
}

// 注册服务
func registerHandler(_ http.ResponseWriter, req *http.Request) interface{} {
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
func heartBeatHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	paramMap := httpUtils.ParamToMap(req)
	err := HeartBeat(paramMap["sessionId"])
	if err != nil {
		return Response{Code: err.Code}
	}
	return Response{Code: 0}
}

// 调用服务
func serverHandler(_ http.ResponseWriter, req *http.Request) interface{} {
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
	r := router.NewRouter()
	r.AddHandler("/", indexHandler)
	r.AddHandler("/register", registerHandler)
	r.AddHandler("/heartBeat", heartBeatHandler)
	r.AddHandler("/server", serverHandler)
	r.AddHandler("/favicon.ico", router.StaticFileHandler)
	r.Start()
}