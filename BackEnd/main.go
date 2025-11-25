package main

import (
	"context"
	"exchangeapp/config"
	"exchangeapp/router"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strings"
)

func main() {
	config.InitConfig()

	r := router.SetupRouter()

	// 确保端口号以冒号开头
	port := config.AppConfig.App.Port

	if port == "" {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	r.Run(port)

	srv := &http.Server{ // http服务器实例
		Addr:    port,
		Handler: r,
	}

	// 启动服务器并且进行监听
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()
	log.Printf("Server started on port %v", port)

	// 创建通道监听系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // 中断信号发送给quit
	<-quit                                             // 阻塞主 goroutine，直到收到信号，程序才会继续执行后面的退出逻辑

	log.Println("Shutting down server...") // 程序接收到中断信号后的处理

	// 创建 5 秒超时上下文，正常情况下等待正在处理的请求完成，如果超过 5 秒仍未完成，则强制关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 函数退出时释放上下文资源

	// 会停止接收新的请求
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	log.Println("Server exiting")
}
