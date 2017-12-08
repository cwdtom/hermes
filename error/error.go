package error

type Error struct {
	Code   int
	ErrMsg string
}

func NewError(code int, errMsg string) *Error {
	return &Error{Code: code, ErrMsg: errMsg}
}

// 全局错误
const ServerError = -1
// 重复注册错误
const RepeatRegister = 1
// SessionId重复
const SessionIdRepeat = 2
// 服务不存在
const ServerNotExisted = 3