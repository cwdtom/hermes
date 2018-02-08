// author chenweidong

package types

import (
	"os"
	"encoding/json"
	"fmt"
	"strings"
	"crypto/md5"
)

// 全局配置
var CONF Config

type Config struct {
	Port       int
	Timeout    int64
	KeyLength  int
	BackupPath string
	Password   string
	WhiteList  []string
}

func InitConfig() {
	// 初始化配置
	ss := strings.Split(os.Args[0], "/")
	confPath := strings.Join(ss[:len(ss)-1], "/") + "/hermes.json"
	file, err := os.Open(confPath)
	if err != nil {
		file, err = os.Open("hermes.json")
		if err != nil {
			fmt.Println("[" + os.Args[0] + "/hermes.json] config file not existed")
			fmt.Scanf("any press to exit")
			os.Exit(1)
		}
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	tmp := Config{}
	err = decoder.Decode(&tmp)
	if err != nil {
		fmt.Println("config file is not legal")
		fmt.Scanf("any press to exit")
		os.Exit(1)
	}
	path := tmp.BackupPath
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	has := md5.Sum([]byte(tmp.Password))
	password := fmt.Sprintf("%x", has)
	CONF = Config{
		Port:       tmp.Port,
		Timeout:    tmp.Timeout,
		KeyLength:  tmp.KeyLength,
		BackupPath: path,
		Password:   password,
		WhiteList:  tmp.WhiteList,
	}

	// 初始化服务表
	SERVERS = make([]Server, 0, 16)

	// 启动修改SERVERS协程
	go ModifyServers()
}

// 检查ip是否存在于白名单
func (conf *Config) CheckIpAuth(ip string) bool {
	// 不设置白名单全部IP允许注册
	if conf.WhiteList == nil || len(conf.WhiteList) == 0 {
		return true
	}
	for _, w := range conf.WhiteList {
		if w == ip {
			return true
		}
	}
	return false
}
