package service

import (
	"Distributed_System/registry"
	"context"
	"fmt"
	"log"
	"net/http"
)

func Start(ctx context.Context,host,port string,
	reg registry.Registration,registerHandler func())  (context.Context,error){
	registerHandler()
	ctx = StartService(ctx,host,port,string(reg.ServiceName))
	registry.RegisterService(reg)
	return ctx,nil
}

func StartService(ctx context.Context,host,port string,serviceName string) context.Context{
	// 用于创建一个可以被显式取消的上下文（context）。它在并发编程中非常有用，尤其是在需要控制多个 goroutine 的生命周期时。
	ctx,cancer := context.WithCancel(ctx)

	var server http.Server
	server.Addr = host+":"+port

	go func(){
		log.Println(server.ListenAndServe())
		cancer()
	}()
	
	go func(){
		fmt.Printf(" %v started, press any key to stop",serviceName)
		var s string
		fmt.Scanln(&s) 
		server.Shutdown(ctx)
		cancer()
	}()
	
	<- ctx.Done()
	return ctx
}