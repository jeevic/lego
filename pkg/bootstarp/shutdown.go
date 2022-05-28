package bootstarp

import (
	"time"

	"github.com/jeevic/lego/pkg/app"
)

var shutdownFunc = []func(){
	ShutdownHttpServer,
	ShutdownGrpcServer,
	ShutdownApp,
}

func Shutdown() {
	t1 := time.Now()
	for _, f := range shutdownFunc {
		f()
	}
	cost := time.Since(t1)
	time.Sleep(5 * time.Second)
	app.App.GetLogger().Info("[shutdown] app all shutdown complete! time timeline:", cost)

}

func RegisterShutdown(f func()) {
	shutdownFunc = append(shutdownFunc, f)
}

func ShutdownHttpServer() {
	hs, _ := app.App.GetHttpServer()
	if hs != nil {
		hs.GracefulShutdown()
		app.App.GetLogger().Infof("[shutdown] shutdown httpserver  complete!")
	}
}

func ShutdownGrpcServer() {
	gs, _ := app.App.GetGrpcServer()
	if gs != nil {
		gs.GracefulShutdown()
		app.App.GetLogger().Infof("[shutdown] shutdown grpc server  complete!")
	}
}

func ShutdownApp() {
	app.App.Close()
	app.App.GetLogger().Infof("[shutdown] shutdown app complete!")
}
