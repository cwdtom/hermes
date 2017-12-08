// author chenweidong

package entity

import (
	"os"
	"encoding/json"
	"fmt"
	"strings"
)

type config struct {
	Port       int
	Timeout    int64
	KeyLength  int
	BackupPath string
}

func InitConfig(args []string) {
	// 初始化配置
	var file *os.File
	if len(args) < 2 {
		file, _ = os.Open("configs/config.json")
	} else {
		file, _ = os.Open(fmt.Sprintf("configs/config-%s.json", args[1]))
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	tmp := config{}
	err := decoder.Decode(&tmp)
	if err != nil {
		fmt.Println("config file is not legal")
		os.Exit(1)
	}
	path := tmp.BackupPath
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	CONF = config{Port: tmp.Port, Timeout: tmp.Timeout, KeyLength: tmp.KeyLength, BackupPath: path}

	// 初始化服务表
	SERVERS = make([]Server, 0, 16)

	// 启动修改SERVERS协程
	go ModifyServers()
}
