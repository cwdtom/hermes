// author chenweidong

package service

// 全局常量
// 备份文件名字
const BackupFileName = "hermes.backup"
// 检查存活服务时间间隔，单位秒
const CheckServersAliveDuration = 60
// 移除失效服务时间间隔，单位秒
const RemoveFailureServerDuration = 120
// 失效服务超时时移除，单位秒
const FailServerTimeout = 300
