// 字符串相关工具 author chenweidong

package string

import (
	"time"
	"math/rand"
)

// 生成随机字符串
func RandomString(strLen int) string {
	rand.Seed(time.Now().UnixNano())
	data := make([]byte, strLen)
	var num int
	for i := 0; i < strLen; i++ {
		num = rand.Intn(57) + 65
		for {
			if num > 90 && num < 97 {
				num = rand.Intn(57) + 65
			} else {
				break
			}
		}
		data[i] = byte(num)
	}
	return string(data)
}
