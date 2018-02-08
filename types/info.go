// author chenweidong

package types

/**
InfoType 枚举
1. 新请求
2. 服务注册
3. 服务心跳
4. 服务未响应
 */
type Info struct {
	InfoType  int
	TimeStamp int64
	Content   string
}

func NewInfo(infoType int, timeStamp int64, content string) *Info {
	return &Info{
		InfoType:  infoType,
		TimeStamp: timeStamp,
		Content:   content,
	}
}