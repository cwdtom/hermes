// 路由相关 author chenweidong

package router

import (
	"net/http"
	"strings"
	"io/ioutil"
	"encoding/json"
	. "hermes/utils/string"
	. "hermes/utils/io"
	. "hermes/entity"
	"fmt"
)

func NewRouter() *Router {
	r := Router{
		routerMap:  map[string]func(http.ResponseWriter, *http.Request) interface{}{},
		filterList: make([]func(http.ResponseWriter, *http.Request) bool, 0, 16),
	}
	http.HandleFunc("/", r.router)
	return &r
}

type Router struct {
	// 路由表
	routerMap map[string]func(http.ResponseWriter, *http.Request) interface{}
	// 拦截器表，返回是否继续处理
	filterList []func(http.ResponseWriter, *http.Request) bool
}

func (r *Router) AddHandler(path string, handler func(http.ResponseWriter, *http.Request) interface{}) {
	r.routerMap[path] = handler
}

func (r *Router) AddFilter(filter func(http.ResponseWriter, *http.Request) bool) {
	r.filterList = append(r.filterList, filter)
}

func (r *Router) Start() {
	PrintInfo("server is running on port: %d", CONF.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", CONF.Port), nil)
}

// 路由器
func (r *Router) router(resp http.ResponseWriter, req *http.Request) {
	// 拦截器
	for _, f := range r.filterList {
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

	reqId := RandomString(6)
	PrintInfo("[%s] request %s", reqId, req.RequestURI)
	if r.routerMap[req.URL.Path] == nil {
		PrintWarn("[%s] [false] response %d", reqId, 404)
		resp.Write([]byte("{\"code\": -1, \"errMsg\": \"404\"}"))
		return
	}
	data := r.routerMap[req.URL.Path](resp, req)
	if data == nil {
		return
	}
	respData, err := json.Marshal(data)
	if err != nil {
		resp.Write([]byte("{\"code\": -1, \"errMsg\": \"" + err.Error() + "\"}"))
		PrintInfo("[%s] [false] response %s", reqId, data)
	} else {
		resp.Write(respData)
		PrintInfo("[%s] [%t] response %s", reqId, data.(Response).Code == 0, data)
	}
}

// 处理静态文件
func StaticFileHandler(resp http.ResponseWriter, req *http.Request) interface{} {
	filePath := req.URL.Path[1:]
	if strings.Split(filePath, "/")[0] != "static" {
		filePath = "static/" + filePath
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		resp.Write([]byte("404"))
		return nil
	}
	resp.Write(data)
	return nil
}