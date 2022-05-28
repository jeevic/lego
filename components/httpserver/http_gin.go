package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	Engine  *gin.Engine
	Setting *Setting
	Server  *http.Server
}

type Setting struct {
	//默认 0.0.0.0
	Host string
	//默认 80
	Port    int
	IsHttps bool
}

func NewHttpServer(host string, port int, isHttps bool) *HttpServer {
	e := gin.New()
	//auto recover
	e.Use(gin.Recovery())

	c := Setting{Host: host, Port: port, IsHttps: isHttps}

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}
	return &HttpServer{Engine: e, Setting: &c, Server: srv}
}

func (h *HttpServer) SetServerModeRelease() {
	gin.SetMode(gin.ReleaseMode)
}

//add middleware
func (h *HttpServer) SetMiddleware(middleware ...gin.HandlerFunc) *HttpServer {
	h.Engine.Use(middleware...)
	return h
}

func (h *HttpServer) ServerRun() {

	isHttps := h.Setting.IsHttps

	//Coroutine start server
	go func() {
		if isHttps {
			if err := h.Server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("http server run https err:%s", err)
			}
		} else {
			if err := h.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("graceful server run http err:%s", err)
			}
		}
	}()
}

func (h *HttpServer) ServerRunAsync() {
	go h.ServerRun()
}

//graceful shutdown  http server wait 5 second
func (h *HttpServer) GracefulShutdown() {
	//shutdown signal
	//sds := make(chan os.Signal)
	//signal.Notify(sds, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSTOP)
	//<-sds
	log.Println("graceful shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.Server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown server error: %s", err)
	}
}
