package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/pobearm/MsgBroker/broker"
	// "strconv"
)

type Pusher struct {
	broker *broker.Broker
}

func New_Pusher(broker *broker.Broker) *Pusher {
	return &Pusher{
		broker: broker,
	}
}

func (p *Pusher) decodeJson(w rest.ResponseWriter, r *rest.Request, req interface{}) bool {
	if err := r.DecodeJsonPayload(req); err != nil {
		// log.Error("api.decodeJson", err.Error())
		w.WriteJson(New_BaseResp(false, "请求json解析失败"))
		return false
	}
	return true
}

type PushReq struct {
	Token   string `json:"token"`
	Qos     byte   `json:"qos"`
	Payload string `json:"payload"`
}

type BaseResp struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

func New_BaseResp(r bool, m string) *BaseResp {
	return &BaseResp{
		Result:  r,
		Message: m,
	}
}

func (p *Pusher) Push(w rest.ResponseWriter, r *rest.Request) {
	req := &PushReq{}
	if !p.decodeJson(w, r, req) {
		return
	}
	// log.Infof("api.Push", "%v", req)
	p.broker.Push(req.Token, req.Qos, []byte(req.Payload))
	//todo:目前, 调用推送后, 并不返回任何信息.
}

//goroutine; heap; threadcreate; block; cpuprof; memprof; gc;
// func (p *Pusher) Profile(w rest.ResponseWriter, r *rest.Request) {
// 	cmd := r.FormValue("cmd")
// 	debug := r.FormValue("debug")
// 	d := 0
// 	if debug != "" {
// 		d, _ = strconv.Atoi(debug)
// 	}
// 	ProcessInput(cmd, d, w)
// }
