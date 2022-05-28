package bootstarp

import "github.com/jeevic/lego/pkg/app"

var StopChan = make(chan struct{})

func Start() {
	//启动httpserver
	hs, _ := app.App.GetHttpServer()
	if hs != nil {
		hs.ServerRunAsync()
	}

	//grpc server
	gs, _ := app.App.GetGrpcServer()
	if gs != nil {
		gs.RunAsync()
	}
}

//关闭服务
func Stop(stop bool) {
	Shutdown()
	if stop == true {
		StopChan <- struct{}{}
	}

}

func Restart() {
	Stop(false)
	Start()
}

func Run() {
	Start()
	<-StopChan
}
