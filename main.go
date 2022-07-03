package main

import (
	"context"
	"github.com/jinzhu/configor"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Config struct {
	GinPort string `default:":7722"`
}

func (c *Config) GetConfStr() (Config Config) {
	err := configor.Load(&Config, "config.yml")
	if err != nil {
		panic(err)
	}
	return
}

func main() {
	//f, _ := os.Create("gin.log")
	//gin.DefaultWriter = io.MultiWriter(f)

	config := Config{}
	conf := config.GetConfStr()
	var appPort = conf.GinPort
	router := InitRouter()

	srv := &http.Server{
		Addr:           appPort,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Println("start :", appPort)
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	// 等待中断信号以优雅地关闭服务器（设置 5 秒的超时时间）
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")

}
