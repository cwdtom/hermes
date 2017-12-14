// author chenweidong

package model

import (
	"os"
	"encoding/json"
	"fmt"
	"strings"
)

// 全局配置
var CONF Config

type Config struct {
	Port       int
	Timeout    int64
	KeyLength  int
	BackupPath string
}

func InitConfig() {
	// 初始化配置
	ss := strings.Split(os.Args[0], "/")
	confPath := strings.Join(ss[:len(ss) - 1], "/") + "/hermes.json"
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
	CONF = Config{Port: tmp.Port, Timeout: tmp.Timeout, KeyLength: tmp.KeyLength, BackupPath: path}

	// 初始化服务表
	SERVERS = make([]Server, 0, 16)

	// 启动修改SERVERS协程
	go ModifyServers()
}
