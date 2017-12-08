// 定时器相关 author chenweidong

package timer

import (
	"hermes/entity"
	. "hermes/utils/io"
	"time"
)

var checkServerAliveTicker *time.Ticker
var removeFailureServerTicker *time.Ticker

func MainTimer() {
	checkServerAliveTicker = time.NewTicker(time.Duration(entity.CheckServersAliveDuration) * time.Second)
	PrintInfo("start checkServerAliveTimer")
	go checkServerAliveTimer()

	removeFailureServerTicker = time.NewTicker(time.Duration(entity.RemoveFailureServerDuration) * time.Second)
	PrintInfo("start removeFailureServerTimer")
	go removeFailureServerTimer()
}

// 定时检查服务是否存活
func checkServerAliveTimer() {
	for true {
		PrintInfo("start check server alive")
		entity.CheckServersStatus()
		PrintInfo("end check server alive")
		<-checkServerAliveTicker.C
	}
}

// 定时移除超时服务
func removeFailureServerTimer() {
	for true {
		PrintInfo("start remove failure server")
		entity.RemoveFailureServer()
		PrintInfo("end remove failure server")
		<-removeFailureServerTicker.C
	}
}
