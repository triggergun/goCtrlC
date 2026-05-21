// goCtrlC 包提供剪贴板监听功能
// 这个包可以帮助你实时获取系统剪贴板的内容变化

package goCtrlC

// 导入需要用到的依赖包

import (
	"context"                   // 用于控制程序生命周期和取消操作
	"fmt"                       // 用于格式化输出和错误处理
	"golang.design/x/clipboard" // 第三方剪贴板操作库
	"os"                        // 用于操作系统相关功能
	"os/signal"                 // 用于监听系统信号（如Ctrl+C）
	"syscall"                   // 提供系统调用接口
)

// GetClipboardChannel 函数用于监听剪贴板变化
//
// 作用：
//   实时监控系统剪贴板，当用户复制内容时，自动获取并返回
//
// 参数：
//   ctx - context.Context 类型，用于控制监听的生命周期
//         可以通过它来停止监听（比如超时或手动取消）
//
// 返回值：
//   <-chan string - 只读通道（管道），用于接收剪贴板文本内容
//                   每当剪贴板有新内容，就会发送到这个通道中
//   error - 错误信息，如果初始化失败会返回错误，成功则返回 nil
//
// 使用示例：
//   // 1. 创建一个可取消的 context
//   ctx, cancel := context.WithCancel(context.Background())
//   defer cancel() // 程序结束时自动停止监听
//
//   // 2. 调用函数获取剪贴板通道
//   clipboardChan, err := GetClipboardChannel(ctx)
//   if err != nil {
//       fmt.Println("初始化失败:", err)
//       return
//   }
//
//   // 3. 循环读取剪贴板内容
//   for content := range clipboardChan {
//       fmt.Println("复制的内容:", content)
//   }

func GetClipboardChannelByCtrlC(ctx context.Context) (<-chan string, error) {
	// 第一步：初始化剪贴板库
	// clipboard.Init() 会准备好系统剪贴板的访问能力
	if err := clipboard.Init(); err != nil {
		// 如果初始化失败，返回错误信息
		// fmt.Errorf 用于创建格式化的错误，%w 会保留原始错误信息
		return nil, fmt.Errorf("初始化剪贴板失败: %w", err)
	}

	// 第二步：创建一个带缓冲的通道
	// make(chan string, 16) 表示创建一个能容纳16个字符串的通道
	// 带缓冲可以避免快速复制时阻塞程序
	resultChan := make(chan string, 64)

	// 第三步：启动一个后台goroutine来监听剪贴板
	// go 关键字表示启动一个并发执行的函数
	go func() {
		// clipboard.Watch 开始监听剪贴板变化
		// ctx 用于控制何时停止监听
		// clipboard.FmtText 表示只监听文本类型的内容
		watchChan := clipboard.Watch(ctx, clipboard.FmtText)

		// 循环读取剪贴板的变化
		// 每当剪贴板有新内容，watchChan 就会收到数据
		for content := range watchChan {
			// select 语句用于同时监听多个通道操作
			select {
			// 情况1：将内容发送到 resultChan
			case resultChan <- string(content):
				// 发送成功，继续等待下一次复制

			// 情况2：收到取消信号（ctx被取消）
			case <-ctx.Done():
				// 关闭 resultChan，告诉接收方没有更多数据了
				close(resultChan)
				// 退出 goroutine
				return
			}
		}

		// 如果 watchChan 被关闭，也关闭 resultChan
		close(resultChan)
	}()

	// 返回通道和 nil（表示没有错误）
	return resultChan, nil
}

// GetClipboardContentByCtrlC 函数是一个完整的剪贴板监听程序
//
// 作用：
//   启动一个完整的剪贴板监听服务，用户按Ctrl+C复制内容时自动显示
//   再次按Ctrl+C（或Ctrl+Break）可以停止程序
//
// 参数：无
//
// 返回值：无
//
// 使用方式：
//   直接调用即可，程序会一直运行直到用户中断
//   GetClipboardContentByCtrlC()

func GetClipboardContentByCtrlC() {
	// 第一步：初始化剪贴板库
	if err := clipboard.Init(); err != nil {
		// 向标准错误输出打印错误信息
		fmt.Fprintf(os.Stderr, "初始化剪贴板失败: %v\n", err)
		// 退出程序，退出码为1表示有错误
		os.Exit(1)
	}

	// 第二步：设置优雅退出机制
	// 创建一个可取消的context，用于控制监听生命周期
	ctx, stop := context.WithCancel(context.Background())

	// 创建一个通道来接收系统信号
	sigChan := make(chan os.Signal, 1)

	// 告诉系统要监听哪些信号
	// syscall.SIGINT 对应 Ctrl+C
	// syscall.SIGTERM 对应系统终止命令
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动一个后台goroutine监听系统信号
	go func() {
		// 等待接收信号
		<-sigChan
		// 收到信号后调用 stop() 取消context
		stop()
		fmt.Println("\n停止监听")
	}()

	// 提示用户程序已开始运行
	fmt.Println("开始监听剪贴板，复制内容(Ctrl+C)自动获取，按 Ctrl+C 结束程序")

	// 第三步：开始监听剪贴板
	watchChan := clipboard.Watch(ctx, clipboard.FmtText)

	// 循环读取并打印剪贴板内容
	for content := range watchChan {
		fmt.Printf("监听到复制内容：%s\n", string(content))
	}
}
