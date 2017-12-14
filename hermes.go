// 所有接口只接受GET请求 author chenweidong

package main

import (
	"net/http"
	"time"
	. "hermes/http_server"
	. "hermes/utils/http"
	. "hermes/model"
	. "hermes/error"
)

// 监控页面
func indexHandler(_ http.ResponseWriter, _ *http.Request) interface{} {
	return Response{Code: 0, Data: SERVERS}
}

// 注册服务
func registerHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	paramMap := ParamToMap(req)
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
	paramMap := ParamToMap(req)
	err := HeartBeat(paramMap["sessionId"])
	if err != nil {
		return Response{Code: err.Code}
	}
	return Response{Code: 0}
}

// 调用服务
func serverHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	paramMap := ParamToMap(req)
	_, s := GetServerBySessionId(paramMap["sessionId"])
	if s == nil || !s.Status {
		return Response{Code: ServerNotExisted}
	}
	data, err := s.CallServer(paramMap["serverId"], paramMap["name"], paramMap["data"])
	if err != nil {
		return Response{Code: err.Code}
	}
	return Response{Code: 0, Data: data}
}

func main() {
	// 初始化配置
	InitConfig()
	// 恢复备份数据
	RestoreServers(CONF.BackupPath)
	// web服务
	hs := NewHttpServer()
	// 定时检查服务是否存活
	hs.AddTimer(CheckServersAliveDuration * time.Second, CheckServersStatus, "CheckServersStatus")
	// 定时移除超时服务
	hs.AddTimer(RemoveFailureServerDuration * time.Second, RemoveFailureServer, "RemoveFailureServer")
	// 设定日志级别
	hs.SetLogLevel(INFO)
	// 路由
	hs.AddHandler("/", indexHandler)
	hs.AddHandler("/register", registerHandler)
	hs.AddHandler("/heartBeat", heartBeatHandler)
	hs.AddHandler("/server", serverHandler)
	hs.AddHandler("/favicon.ico", StaticFileHandler)
	hs.Start(CONF.Port)
}