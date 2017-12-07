// 实体相关 author chenweidong

package entity

import (
	"time"
	"hermes/utils/encipher"
	"encoding/json"
	"os"
	"bufio"
	. "hermes/utils/io"
)

type Server struct {
	Id         string
	SessionId  string
	Expire     int64
	PrivateKey string
	Host       string
	Status     bool
}

func Register(id, sessionId, host string) (string, *Error) {
	// 生成公私钥
	key, err := encipher.GenRsaKey(CONF.KeyLength)
	if err != nil {
		return "", NewError(ServerError, err.Error())
	}

	newServer := Server{Id: id, SessionId: sessionId, Host: host,
		Status: true, Expire: time.Now().Unix() + CONF.Timeout, PrivateKey: key.PrivateKey}
	// 检查是否已注册
	index, status := newServer.IsExisted()
	if index >= 0 {
		if status {
			return "", NewError(RepeatRegister, "server is already existed")
		}
		// 通知更新服务信息
		modifyServerChannel <- serverChannel{operate: 1, server: newServer, index: index}
	} else {
		// 检查sessionId是否重复
		if isSessionIdRepeat(newServer.SessionId) {
			return "", NewError(SessionIdRepeat, "sessionId is already existed")
		}
		// 通知添加服务
		modifyServerChannel <- serverChannel{operate: 2, server: newServer}
	}
	// 备份数据
	go BackUpServers(CONF.BackupPath)
	return key.PublicKey, nil
}

func HeartBeat(sessionId string) *Error {
	// 通知修改服务状态
	server := GetServerBySessionId(sessionId)
	if server == nil {
		return NewError(ServerNotExisted, "server not existed")
	}
	modifyServerChannel <- serverChannel{operate: 5, sessionId: sessionId}
	return nil
}

// @return 位置 int 是否可用 bool
func (s *Server) IsExisted() (int, bool) {
	if len(SERVERS) < 1 {
		return -1, false
	}
	for index, ele := range SERVERS {
		if s.Host == ele.Host {
			return index, ele.Status
		}
	}
	return -1, false
}

func isSessionIdRepeat(sessionId string) bool {
	return GetServerBySessionId(sessionId) != nil
}

func GetServerBySessionId(sessionId string) *Server {
	for _, s := range SERVERS {
		if s.SessionId == sessionId {
			return &s
		}
	}
	return nil
}

func BackUpServers(path string) {
	data, err := json.Marshal(SERVERS)
	if err != nil {
		PrintError("back up error: \nerror code: %d\nerror message: %s", ServerError, err.Error())
	}
	filePath := path + BackupFileName
	out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		PrintError("back up error: \nerror code: %d\nerror message: %s", ServerError, err.Error())
	}
	defer out.Close()
	outWriter := bufio.NewWriter(out)
	_, err = outWriter.Write(data)
	if err != nil {
		PrintError("back up error: \nerror code: %d\nerror message: %s", ServerError, err.Error())
	}
	outWriter.Flush()
	PrintInfo("servers backup success")
}

func RestoreServers(path string) {
	filePath := path + BackupFileName
	in, err := os.Open(filePath)
	if err != nil {
		PrintWarn("backup file not existed")
		return
	}
	defer in.Close()
	decoder := json.NewDecoder(in)
	var tmp []Server
	err = decoder.Decode(&tmp)
	SERVERS = copyServers(tmp)
}

// 检查服务状态，超时的置为不可用
func CheckServersStatus() {
	// 通知修改状态
	modifyServerChannel <- serverChannel{operate: 3}
}

// 移除失效并超出保留时间的服务
func RemoveFailureServer() {
	// 通知操作删除
	modifyServerChannel <- serverChannel{operate: 4}
}

func copyServers(tmp []Server) []Server {
	servers := make([]Server, 0)
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
		case 1:    // 更新已有服务
			if SERVERS[sc.index].Host == sc.server.Host {
				SERVERS[sc.index] = sc.server
				PrintInfo("serverId: %s host: %s sessionId: %s update success",
					sc.server.Id, sc.server.Host, sc.server.SessionId)
			}
			break
		case 2:    // 注册新服务
			SERVERS = append(SERVERS, sc.server)
			PrintInfo("serverId: %s host: %s sessionId: %s register success",
				sc.server.Id, sc.server.Host, sc.server.SessionId)
			break
		case 3:    // 检查服务状态，设置超时服务失效
			now := time.Now().Unix()
			for index, s := range SERVERS {
				if s.Status && now > s.Expire {
					PrintWarn("serverId: %s host: %s sessionId: %s already failure",
						s.Id, s.Host, s.SessionId)
					SERVERS[index].Status = false
				}
			}
			break
		case 4:    // 移除失效并超出保留时间的服务
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
					PrintWarn("serverId: %s host: %s sessionId: %s already removed",
						s.Id, s.Host, s.SessionId)
				}
			}
			// 备份数据
			if isRemove {
				go BackUpServers(CONF.BackupPath)
			}
			break
		case 5:    // 收到服务心跳，更新服务状态
			for index, s := range SERVERS {
				if s.SessionId == sc.sessionId {
					SERVERS[index].Status = true
					SERVERS[index].Expire = time.Now().Unix() + CONF.Timeout
				}
			}
			break
		default:
			PrintError("ModifyServers fail data: ", sc)
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
