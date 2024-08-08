package log

import (
	"Distributed_System/registry"
	"bytes"
	"fmt"
	stlog "log"
	"net/http"
)

// Server.go是指服务端的log service，客户端调用它比较麻烦，我们需要编写一套逻辑能够要我们方便调用log service

// 重新定义logger
func SetClientLogger(serviceURL string, ClientServiceName registry.ServiceName){
	stlog.SetPrefix(fmt.Sprintf("[%v] -",ClientServiceName))
	stlog.SetFlags(0)
	// 自定义输出的时候出发Write逻辑，向logservice的url发送post请求
	stlog.SetOutput(&clientLogger{url : serviceURL})
}

type clientLogger struct{
	url string
}

func ( c *clientLogger) Write(data []byte)(int ,error){
	b := bytes.NewBuffer([]byte(data))
	res,err := http.Post(c.url+"/log","text/plain",b)
	if err != nil{
		return 0,err
	}
	if res.StatusCode != http.StatusOK{
		return 0,fmt.Errorf("Post log Service Unsuccess!")
	}

	return (len(data)),nil
}
