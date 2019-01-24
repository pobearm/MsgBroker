package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	pushapi "github.com/pobearm/MsgBroker/api"
	"github.com/pobearm/MsgBroker/broker"
	"github.com/pobearm/MsgBroker/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

//env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

func main() {
	//配置文件读
	tconfig := config.LoadConfig()
	//设置日志
	// log.InitLog(tconfig.LogConfig)
	//服务
	b := broker.NewBroker(tconfig.MaxQueueDepth, tconfig.UserName, tconfig.Password)
	err := b.Start(tconfig.Tcp, tconfig.Websocket)
	if err != nil {
		log.Fatal("main", err.Error())
		panic(err)
	}
	//验证器
	valid := pushapi.New_Valid(tconfig.UserName, tconfig.Password)
	//web api
	api := rest.NewApi()
	api.Use(rest.DefaultCommonStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm:         "pushservice",
		Authenticator: valid.ValidateUser,
	})
	//设置路由
	pusher := pushapi.New_Pusher(b)
	router, err := rest.MakeRouter(
		&rest.Route{HttpMethod: "POST", PathExp: "/api/push", Func: pusher.Push},
	)
	if err != nil {
		log.Fatal("main", err.Error())
	}
	api.SetApp(router)

	log.Printf("start listen http: %s", tconfig.HttpAddr)
	//启动web端口
	log.Fatal("main", http.ListenAndServe(tconfig.HttpAddr, api.MakeHandler()).Error())
	//信号
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Print("main", "get signal stop")
	b.Stop()
}
