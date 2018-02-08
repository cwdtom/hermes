// 封装http服务 author chenweidong

package http_server

import (
	"net/http"
	"strings"
	"io/ioutil"
	"encoding/json"
	"time"
	"strconv"
	"math/rand"
	"os"

	"github.com/cwdtom/hermes/http_server/types"
)

// 全局LOG
var LOG *log
// root path
var ROOT_PATH string

func NewHttpServer() *HttpServer {
	ss := strings.Split(os.Args[0], "/")
	ROOT_PATH = strings.Join(ss[:len(ss)-1], "/") + "/"
	h := HttpServer{
		routerMap:  map[string]func(http.ResponseWriter, *http.Request) interface{}{},
		filterList: make([]func(http.ResponseWriter, *http.Request) bool, 0, 16),
		timerList:  make([]*Timer, 0, 16),
		log:        NewLog(WARN),
	}
	LOG = h.log
	http.HandleFunc("/", h.routerServer)
	return &h
}

type HttpServer struct {
	// 路由表
	routerMap map[string]func(http.ResponseWriter, *http.Request) interface{}
	// 拦截器表，返回是否继续处理
	filterList []func(http.ResponseWriter, *http.Request) bool
	// 定时器
	timerList []*Timer
	// 日志
	log *log
}

func (h *HttpServer) AddHandler(path string, handler func(http.ResponseWriter, *http.Request) interface{}) {
	h.routerMap[path] = handler
}

func (h *HttpServer) AddFilter(filter func(http.ResponseWriter, *http.Request) bool) {
	h.filterList = append(h.filterList, filter)
}

func (h *HttpServer) AddTimer(duration time.Duration, f func(), name string) {
	h.timerList = append(h.timerList, NewTimer(duration, f, name))
}

func (h *HttpServer) StopTimer(name string) {
	for _, t := range h.timerList {
		if t.Name == name {
			t.Stop()
		}
	}
}

func (h *HttpServer) SetLogLevel(level int) {
	h.log.SetLevel(level)
}

func (h *HttpServer) GetLog() *log {
	return h.log
}

func (h *HttpServer) Start(port int) {
	// 启动定时器
	for _, t := range h.timerList {
		t.Start()
	}
	h.log.Info("server is running on port: %d", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

// 路由器
func (h *HttpServer) routerServer(resp http.ResponseWriter, req *http.Request) {
	// 拦截器
	for _, f := range h.filterList {
		isContinue := f(resp, req)
		if !isContinue {
			return
		}
	}
	// 判断是否访问静态文件
	args := strings.Split(req.URL.Path, "/")
	if len(args) >= 3 && args[1] == "static" {
		StaticFileHandler(resp, req)
		return
	}
	reqId := randomString(6)
	h.log.Info("[%s] request %s", reqId, req.RequestURI)
	if h.routerMap[req.URL.Path] == nil {
		h.log.Warn("[%s] [false] response %d", reqId, 404)
		resp.Write([]byte("{\"code\": -1, \"errMsg\": \"404\"}"))
		return
	}
	data := h.routerMap[req.URL.Path](resp, req)
	if data == nil {
		return
	}
	respData, err := json.Marshal(data)
	if err != nil {
		resp.Write([]byte("{\"code\": -1, \"errMsg\": \"" + err.Error() + "\"}"))
		h.log.Error("[%s] [false] response %s", reqId, data)
	} else {
		resp.Write(respData)
		h.log.Info("[%s] [%t] response %s", reqId, data.(types.Response).Code == 0, data)
	}
}

// 处理静态文件
func StaticFileHandler(resp http.ResponseWriter, req *http.Request) interface{} {
	filePath := req.URL.Path[1:]
	if strings.Split(filePath, "/")[0] != "static" {
		filePath = "static/" + filePath
	}
	filePath = ROOT_PATH + filePath
	data, err := ioutil.ReadFile(filePath)
	if strings.HasSuffix(filePath, ".css") || strings.HasSuffix(filePath, ".css.map") {
		resp.Header().Add("Content-Type", "text/css")
	} else if strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".js.map") {
		resp.Header().Add("Content-Type", "application/javascript")
	}
	if err != nil {
		resp.Write([]byte("404"))
		return nil
	}
	resp.Write(data)
	return nil
}

// 生成随机字符串
func randomString(strLen int) string {
	rand.Seed(time.Now().UnixNano())
	data := make([]byte, strLen)
	var num int
	for i := 0; i < strLen; i++ {
		num = rand.Intn(57) + 65
		for {
			if num > 90 && num < 97 {
				num = rand.Intn(57) + 65
			} else {
				break
			}
		}
		data[i] = byte(num)
	}
	return string(data)
}
