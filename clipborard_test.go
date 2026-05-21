package goCtrlC

import (
	"context"
	"fmt"
	"log"
	"testing"
)

// 单元测试
func TestCopy(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	clipboardChan, err := GetClipboardChannel(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for content := range clipboardChan {
		fmt.Println("获取到剪贴板内容:", content)
	}
}
