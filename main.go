// 所有接口只接受GET请求 author chenweidong

package main

import (
	"net/http"
	"time"
	hs "github.com/cwdtom/hermes/http_server"
	hst "github.com/cwdtom/hermes/http_server/types"
	hu "github.com/cwdtom/hermes/utils/http_utils"
	"github.com/cwdtom/hermes/types"
	"github.com/cwdtom/hermes/error"
)

var INFO_LIST []*types.Info

// 首页转发
func indexHandler(resp http.ResponseWriter, req *http.Request) interface{} {
	http.Redirect(resp, req, "/static/backend/index.html", 302)
	return nil
}

// 获取server列表
func serverListHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	// 权限校验
	login := loginHandler(nil, req).(hst.Response)
	if login.Code != 0 {
		return login
	}

	return hst.Response{Code: 0, Data: types.SERVERS}
}

// 获取运行信息
func infoHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	// 权限校验
	login := loginHandler(nil, req).(hst.Response)
	if login.Code != 0 {
		return login
	}

	tmp := INFO_LIST
	INFO_LIST = make([]*types.Info, 0, 1024)
	return hst.Response{Code: 0, Data: tmp}
}

// 登录
func loginHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	cookie, err := req.Cookie("sign")
	if err != nil {
		return hst.Response{Code: 200}
	}
	if cookie.Value == types.CONF.Password {
		return hst.Response{Code: 0}
	}
	return hst.Response{Code: 200}
}

// 注册服务
func registerHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	paramMap := hu.ParamToMap(req)
	code := 0
	id := paramMap["id"]
	sessionId := paramMap["sessionId"]
	INFO_LIST = append(INFO_LIST, types.NewInfo(2, time.Now().Unix(), id + "-" + sessionId))
	// 校验IP白名单
	ip := hu.RemoteIp(req)
	if !types.CONF.CheckIpAuth(ip) {
		return hst.Response{Code: 200}
	}
	puk, err := types.Register(id, sessionId, ip + ":" + paramMap["port"])
	if err != nil {
		code = err.Code
	}
	return hst.Response{Data: struct {
		PublicKey string
		Length    int
		Timeout   int64
	}{PublicKey: puk, Length: types.CONF.KeyLength, Timeout: types.CONF.Timeout}, Code: code}
}

// 心跳检测
func heartBeatHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	sessionId := hu.ParamToMap(req)["sessionId"]
	INFO_LIST = append(INFO_LIST, types.NewInfo(3, time.Now().Unix(), sessionId + " Heart Beat"))
	err := types.HeartBeat(sessionId)
	if err != nil {
		return hst.Response{Code: err.Code}
	}
	return hst.Response{Code: 0}
}

// 调用服务
func serverHandler(_ http.ResponseWriter, req *http.Request) interface{} {
	paramMap := hu.ParamToMap(req)
	_, s := types.GetServerBySessionId(paramMap["sessionId"])
	if s == nil || !s.Status {
		return hst.Response{Code: error.ServerNotExisted}
	}
	serverId := paramMap["serverId"]
	name := paramMap["name"]
	INFO_LIST = append(INFO_LIST, types.NewInfo(1, time.Now().Unix(), serverId + "-" + name))
	data, err := s.CallServer(serverId, name, paramMap["data"])
	if err != nil {
		INFO_LIST = append(INFO_LIST, types.NewInfo(4, time.Now().Unix(), serverId + "-" + name))
		return hst.Response{Code: err.Code}
	}
	return hst.Response{Code: 0, Data: data}
}

func main() {
	// 初始化配置
	types.InitConfig()
	INFO_LIST = make([]*types.Info, 0, 1024)
	// 恢复备份数据
	types.RestoreServers(types.CONF.BackupPath)
	// web服务
	s := hs.NewHttpServer()
	// 定时检查服务是否存活
	s.AddTimer(types.CheckServersAliveDuration * time.Second, types.CheckServersStatus, "CheckServersStatus")
	// 定时移除超时服务
	s.AddTimer(types.RemoveFailureServerDuration * time.Second, types.RemoveFailureServer, "RemoveFailureServer")
	// 设定日志级别
	s.SetLogLevel(hs.INFO)
	// 路由
	s.AddHandler("/", indexHandler)
	s.AddHandler("/serverList", serverListHandler)
	s.AddHandler("/register", registerHandler)
	s.AddHandler("/heartBeat", heartBeatHandler)
	s.AddHandler("/server", serverHandler)
	s.AddHandler("/favicon.ico", hs.StaticFileHandler)
	s.AddHandler("/info", infoHandler)
	s.AddHandler("/login", loginHandler)
	s.Start(types.CONF.Port)
}