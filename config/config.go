package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

//加载配置文件
func LoadConfig() *TConfig {
	cf := "ps.conf"
	file, err := os.Open(cf)
	if err != nil {
		panic("not found ps.conf")
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)
	cfg := &TConfig{}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		panic("configure file error")
	}
	return cfg
}

type TConfig struct {
	MaxQueueDepth int    //每个客户端最大的内存缓存消息数
	Tcp           string //tcp端口
	Websocket     string //目前没有实现websocket
	UserName      string //验证的用户名
	Password      string //验证的密码
	HttpAddr      string //发送方的调用端口
}
