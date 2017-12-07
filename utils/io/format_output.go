// io相关工具 author chenweidong

package io

import (
	"fmt"
	"time"
)

func PrintDebug(s string, a ...interface{}) {
	fmt.Printf("[" + time.Now().Format("2006-01-02 15:04:05") + "] DEBUG: [" + s + "]\n", a...)
}

func PrintInfo(s string, a ...interface{}) {
	fmt.Printf("[" + time.Now().Format("2006-01-02 15:04:05") + "] INFO: " + s + "\n", a...)
}

func PrintWarn(s string, a ...interface{}) {
	fmt.Printf("[" + time.Now().Format("2006-01-02 15:04:05") + "] WARN: " + s + "\n", a...)
}

func PrintError(s string, a ...interface{}) {
	fmt.Printf("[" + time.Now().Format("2006-01-02 15:04:05") + "] ERROR: " + s + "\n", a...)
}