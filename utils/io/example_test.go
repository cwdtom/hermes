// io包测试 author chenweidong

package io

import (
	"testing"
)

func TestPrintDebug(t *testing.T) {
	PrintDebug("debug %s", "test")
}

func TestPrintInfo(t *testing.T) {
	PrintInfo("info %s", "test")
}

func TestPrintWarn(t *testing.T) {
	PrintWarn("warn %s", "test")
}

func TestPrintError(t *testing.T) {
	PrintError("error %s", "test")
}