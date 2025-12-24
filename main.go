package main

import (
	"bm_binus/internal/config"
	"bm_binus/internal/factory"
	httpbm_binus "bm_binus/internal/http"
	middlewareEcho "bm_binus/internal/middleware"
	db "bm_binus/pkg/database"
	"bm_binus/pkg/log"
	"bm_binus/pkg/ngrok"
	"bm_binus/pkg/ws"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// @title bm_binus
// @version 1.0.1
// @description This is a doc for bm_binus.

func main() {
	config.Init()

	log.Init()

	db.Init()

	e := echo.New()

	f := factory.NewFactory()

	middlewareEcho.Init(e, f.DbRedis)

	httpbm_binus.Init(e, f)

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ws.InitCentrifugal(ctx, e, f)

	go func() {
		runNgrok := false
		addr := ""
		if runNgrok {
			listener := ngrok.Run()
			e.Listener = listener
			addr = "/"
		} else {
			addr = ":" + config.Get().App.Port
		}
		err := e.Start(addr)
		if err != nil {
			if err != http.ErrServerClosed {
				logrus.Fatal(err)
			}
		}
	}()

	<-ch

	logrus.Println("Shutting down server...")
	cancel()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	e.Shutdown(ctx2)
	logrus.Println("Server gracefully stopped")
}
