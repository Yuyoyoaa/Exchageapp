package main

import (
	"context"
	"exchangeapp/config"
	"exchangeapp/router"
	"exchangeapp/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// 1. 初始化配置
	config.InitConfig()

	// 2. 创建全局 Context 用于优雅控制后台任务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保 main 退出时 context 被取消

	// 3. 启动汇率定时任务 (传入 context)
	services.StartExchangeRateScheduler(ctx)

	// 4. 路由与服务器配置
	r := router.SetupRouter()
	port := resolvePort(config.AppConfig.App.Port)

	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// 5. 启动 HTTP 服务器
	go func() {
		log.Printf("Server started on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// 6. 监听退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 7. 优雅停机流程
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// 停止 HTTP 服务
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %s", err)
	}

	// 这里的 cancel() 会被 defer 调用，通知 scheduler 停止
	log.Println("Server exited")
}

// 辅助函数：处理端口格式
func resolvePort(port string) string {
	if port == "" {
		return ":8080"
	}
	if !strings.HasPrefix(port, ":") {
		return ":" + port
	}
	return port
}
