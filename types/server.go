// 实体相关 author chenweidong

package types

import (
	"time"
	"encoding/json"
	"os"
	"bufio"
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/hex"

	"github.com/cwdtom/hermes/error"
	"github.com/cwdtom/hermes/utils/encipher"

	hs "github.com/cwdtom/hermes/http_server"
)

// 全局服务
var SERVERS []Server

type Server struct {
	Id           string
	SessionId    string
	Expire       int64
	PrivateKey   string
	Host         string
	Status       bool
	CallCount    int64
	SuccessCount int64
}

func Register(id, sessionId, host string) (string, *error.Error) {
	// 生成公私钥
	key, err := encipher.GenRsaKey(CONF.KeyLength)
	if err != nil {
		return "", error.NewError(error.ServerError, err.Error())
	}

	newServer := Server{Id: id, SessionId: sessionId, Host: host, Status: true,
						Expire: time.Now().Unix() + CONF.Timeout, PrivateKey: key.PrivateKey,
						CallCount: 0, SuccessCount: 0}
	// 检查是否已注册
	index := newServer.IsExisted()
	if index >= 0 {
		// 通知更新服务信息
		modifyServerChannel <- serverChannel{operate: 1, server: newServer, index: index}
	} else {
		// 检查sessionId是否重复
		if isSessionIdRepeat(newServer.SessionId) {
			return "", error.NewError(error.SessionIdRepeat, "sessionId is already existed")
		}
		// 通知添加服务
		modifyServerChannel <- serverChannel{operate: 2, server: newServer}
	}
	// 备份数据
	go BackUpServers(CONF.BackupPath)
	return key.PublicKey, nil
}

func HeartBeat(sessionId string) *error.Error {
	// 通知修改服务状态
	index, _ := GetServerBySessionId(sessionId)
	if index < 0 {
		return error.NewError(error.ServerNotExisted, "server not existed")
	}
	modifyServerChannel <- serverChannel{operate: 5, sessionId: sessionId}
	return nil
}

// @return 位置 int 是否可用 bool
func (s *Server) IsExisted() int {
	if len(SERVERS) < 1 {
		return -1
	}
	for index, ele := range SERVERS {
		if s.Host == ele.Host {
			return index
		}
	}
	return -1
}

func isSessionIdRepeat(sessionId string) bool {
	index, _ := GetServerBySessionId(sessionId)
	return  index > 0
}

func GetServerBySessionId(sessionId string) (int, *Server) {
	for index, s := range SERVERS {
		if s.SessionId == sessionId {
			return index, &s
		}
	}
	return -1, nil
}

func BackUpServers(path string) {
	data, err := json.Marshal(SERVERS)
	if err != nil {
		hs.LOG.Error("back up error: %s", err.Error())
	}
	filePath := path + BackupFileName
	out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		hs.LOG.Error("back up error: %s", err.Error())
	}
	defer out.Close()
	outWriter := bufio.NewWriter(out)
	_, err = outWriter.Write(data)
	if err != nil {
		hs.LOG.Error("back up error: %s", err.Error())
	}
	outWriter.Flush()
	hs.LOG.Info("servers backup success")
}

func RestoreServers(path string) {
	filePath := path + BackupFileName
	in, err := os.Open(filePath)
	if err != nil {
		hs.LOG.Warn("backup file not existed")
		return
	}
	defer in.Close()
	decoder := json.NewDecoder(in)
	var tmp []Server
	decoder.Decode(&tmp)
	SERVERS = copyServers(tmp)
}

// 检查服务状态，超时的置为不可用
func CheckServersStatus() {
	// 通知检查服务状态，超时的置为不可用
	hs.LOG.Info("notice check servers status")
	modifyServerChannel <- serverChannel{operate: 3}
}

// 移除失效并超出保留时间的服务
func RemoveFailureServer() {
	// 通知移除失效并超出保留时间的服务
	hs.LOG.Info("notice remove failure server")
	modifyServerChannel <- serverChannel{operate: 4}
}

func copyServers(tmp []Server) []Server {
	servers := make([]Server, 0, 16)
	for _, ele := range tmp {
		if ele.Status {
			servers = append(servers, ele.DeepCopy())
		}
	}
	return servers
}

func (s *Server) DeepCopy() Server {
	target := Server{
		Id:         s.Id,
		SessionId:  s.SessionId,
		Host:       s.Host,
		Status:     s.Status,
		Expire:     s.Expire,
		PrivateKey: s.PrivateKey,
	}
	return target
}

func (s *Server) CallServer(id, name, data string) ([]byte, *error.Error) {
	// hex to bytes
	bytes, _ := hex.DecodeString(data)
	// 解密
	text, err := encipher.RsaDecryptByPrk(bytes, s.PrivateKey)
	if err != nil {
		hs.LOG.Warn("request call decrypt fail：%s", err.Error())
		return nil, &error.Error{Code: error.RsaError}
	}
	// 提取目标server
	tmp := make([]*Server, 0, len(SERVERS))
	for _, s := range SERVERS {
		if s.Id == id && s.Status {
			tmp = append(tmp, &s)
		}
	}
	if len(tmp) == 0 {
		return nil, &error.Error{Code: error.ServerNotExisted}
	}
	target := tmp[int(time.Now().UnixNano() / 1000000) % len(tmp)]
	index, _ := GetServerBySessionId(target.SessionId)
	// 增加调用次数
	SERVERS[index].CallCount += 1
	// 明文加密
	sendData, err := encipher.RsaEncryptByPrk(text, target.PrivateKey)
	if err != nil {
		hs.LOG.Error("request call encrypt fail：%s", err.Error())
		return nil, &error.Error{Code: error.RsaError}
	}
	// bytes to hex
	send := hex.EncodeToString(sendData)
	// 发送请求
	address := fmt.Sprintf("http_utils://%s/hermes?sessionId=%s&name=%s&data=%s",
		target.Host, target.SessionId, name, send)
	hs.LOG.Info("call server: %s", address)
	resp, err := http.Get(address)
	if err != nil {
		return nil, &error.Error{Code: error.RequestError}
	}
	// 响应解密
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		hs.LOG.Error("request fail：%s | sessionId: %s", err.Error(), target.SessionId)
		return nil, &error.Error{Code: error.ReadResponseError}
	}
	respData, err = encipher.RsaDecryptByPrk(respData, target.PrivateKey)
	if err != nil {
		hs.LOG.Error("response decrypt fail：%s", err.Error())
		return nil, &error.Error{Code: error.RsaError}
	}
	// 响应加密
	respData, err = encipher.RsaEncryptByPrk(respData, s.PrivateKey)
	if err != nil {
		hs.LOG.Error("response encrypt fail：%s", err.Error())
		return nil, &error.Error{Code: error.RsaError}
	}
	// 添加成功调用次数
	SERVERS[index].SuccessCount += 1
	return respData, nil
}

// 服务修改协程通道
var modifyServerChannel chan serverChannel

// 修改环境SERVERS，协程
func ModifyServers() {
	modifyServerChannel = make(chan serverChannel, 0)
	for true {
		sc, ok := <-modifyServerChannel
		if !ok {
			return
		}
		switch sc.operate {
		case 1: // 更新已有服务
			if SERVERS[sc.index].Host == sc.server.Host {
				SERVERS[sc.index] = sc.server
				hs.LOG.Info("serverId: %s host: %s sessionId: %s update success",
					sc.server.Id, sc.server.Host, sc.server.SessionId)
			}
			break
		case 2: // 注册新服务
			SERVERS = append(SERVERS, sc.server)
			hs.LOG.Info("serverId: %s host: %s sessionId: %s register success",
				sc.server.Id, sc.server.Host, sc.server.SessionId)
			break
		case 3: // 检查服务状态，设置超时服务失效
			now := time.Now().Unix()
			for index, s := range SERVERS {
				if s.Status && now > s.Expire {
					hs.LOG.Warn("serverId: %s host: %s sessionId: %s already failure",
						s.Id, s.Host, s.SessionId)
					SERVERS[index].Status = false
				}
			}
			break
		case 4: // 移除失效并超出保留时间的服务
			length := len(SERVERS)
			now := time.Now().Unix()
			isRemove := false
			for i := 0; i < length; i++ {
				s := SERVERS[i]
				if !s.Status && now-s.Expire > FailServerTimeout {
					SERVERS = append(SERVERS[:i], SERVERS[i+1:] ...)
					i--
					length--
					isRemove = true
					hs.LOG.Warn("serverId: %s host: %s sessionId: %s already removed",
						s.Id, s.Host, s.SessionId)
				}
			}
			// 备份数据
			if isRemove {
				go BackUpServers(CONF.BackupPath)
			}
			break
		case 5: // 收到服务心跳，更新服务状态
			for index, s := range SERVERS {
				if s.SessionId == sc.sessionId {
					SERVERS[index].Status = true
					SERVERS[index].Expire = time.Now().Unix() + CONF.Timeout
				}
			}
			break
		default:
			hs.LOG.Error("ModifyServers fail: operate: %d, sessionId: %s", sc.operate, sc.sessionId)
		}
	}
}

/**
operate 枚举
1 = 通知更新服务信息
2 = 通知添加服务
3 = 通知修改状态
4 = 通知操作删除
5 = 通知修改服务心跳状态
 */
type serverChannel struct {
	operate   int
	sessionId string
	server    Server
	index     int
}
