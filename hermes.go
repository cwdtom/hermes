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

var INFO_LIST []*Info

// 首页转发
func indexHandler(resp http.ResponseWriter, req *http.Request) interface{} {
	http.Redirect(resp, req, "/static/backend/index.html", 302)
	return nil
}

// 获取server列表
func serverListHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	// 权限校验
	login := loginHandler(nil, req).(Response)
	if login.Code != 0 {
		return login
	}

	return Response{Code: 0, Data: SERVERS}
}

// 获取运行信息
func infoHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	// 权限校验
	login := loginHandler(nil, req).(Response)
	if login.Code != 0 {
		return login
	}

	tmp := INFO_LIST
	INFO_LIST = make([]*Info, 0, 1024)
	return Response{Code: 0, Data: tmp}
}

// 登录
func loginHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	cookie, err := req.Cookie("sign")
	if err != nil {
		return Response{Code: 200}
	}
	if cookie.Value == CONF.Password {
		return Response{Code: 0}
	}
	return Response{Code: 200}
}

// 注册服务
func registerHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	paramMap := ParamToMap(req)
	code := 0
	id := paramMap["id"]
	sessionId := paramMap["sessionId"]
	INFO_LIST = append(INFO_LIST, NewInfo(2, time.Now().Unix(), id + "-" + sessionId))
	// 校验IP白名单
	ip := RemoteIp(req)
	if !CONF.CheckIpAuth(ip) {
		return Response{Code: 200}
	}
	puk, err := Register(id, sessionId, ip + ":" + paramMap["port"])
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
	sessionId := ParamToMap(req)["sessionId"]
	INFO_LIST = append(INFO_LIST, NewInfo(3, time.Now().Unix(), sessionId + " Heart Beat"))
	err := HeartBeat(sessionId)
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
	serverId := paramMap["serverId"]
	name := paramMap["name"]
	INFO_LIST = append(INFO_LIST, NewInfo(1, time.Now().Unix(), serverId + "-" + name))
	data, err := s.CallServer(serverId, name, paramMap["data"])
	if err != nil {
		INFO_LIST = append(INFO_LIST, NewInfo(4, time.Now().Unix(), serverId + "-" + name))
		return Response{Code: err.Code}
	}
	return Response{Code: 0, Data: data}
}

func main() {
	// 初始化配置
	InitConfig()
	INFO_LIST = make([]*Info, 0, 1024)
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
	hs.AddHandler("/serverList", serverListHandler)
	hs.AddHandler("/register", registerHandler)
	hs.AddHandler("/heartBeat", heartBeatHandler)
	hs.AddHandler("/server", serverHandler)
	hs.AddHandler("/favicon.ico", StaticFileHandler)
	hs.AddHandler("/info", infoHandler)
	hs.AddHandler("/login", loginHandler)
	hs.Start(CONF.Port)
}