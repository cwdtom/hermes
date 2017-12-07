// 定时器相关 author chenweidong

package timer

import (
	"hermes/entity"
	. "hermes/utils/io"
	"github.com/robfig/cron"
	"fmt"
)

func MainTimer() {
	c := cron.New()
	// 检查服务状态定时器
	PrintInfo("start checkServerAliveTimer")
	checkServerAliveTimer()
	spec := fmt.Sprintf("@every %ds", entity.CheckServersAliveDuration)
	c.AddFunc(spec, checkServerAliveTimer)
	// 移除超时的失效定时器
	PrintInfo("start removeFailureServerTimer")
	removeFailureServerTimer()
	spec = fmt.Sprintf("@every %ds", entity.RemoveFailureServerDuration)
	c.AddFunc(spec, removeFailureServerTimer)
	// 启动定时器
	c.Start()
}

func checkServerAliveTimer() {
	PrintInfo("start check server alive")
	entity.CheckServersStatus()
	PrintInfo("end check server alive")
}

func removeFailureServerTimer() {
	PrintInfo("start remove failure server")
	entity.RemoveFailureServer()
	PrintInfo("end remove failure server")
}