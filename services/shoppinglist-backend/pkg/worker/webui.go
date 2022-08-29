package worker

import (
	"ShoppingList-Backend/pkg/application"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gocraft/work/webui"
	"go.uber.org/zap"
)

func NewWebUI(app *application.Application) *webui.Server {
	server := webui.NewServer(GetRedisNamespace(app.Cfg), app.Redis, app.Cfg.GetWorkerPort())
	return server
}

func StartWebUI(server *webui.Server, wg *sync.WaitGroup) {
	defer wg.Done()
	server.Start()
	zap.S().Infof("Worker web UI started")
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	server.Stop()
}
