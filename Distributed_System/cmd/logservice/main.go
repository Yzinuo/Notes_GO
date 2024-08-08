package main

import (
	"Distributed_System/log"
	"Distributed_System/registry"
	"Distributed_System/service"
	"context"
	"fmt"
	stlog "log"
)

func main(){
	log.Run("./distributed.log")
	host,port := "localhost","8080"

    Addr := fmt.Sprintf("http://%s:%s",host,port)

    reg  := registry.Registration{
		ServiceName: "Log Service",
		ServiceURL:  Addr,
	}

	ctx,err :=service.Start(
		context.Background(),
		host,
		port,
		reg,
		log.RegisterHandler, 
	)
	if err != nil{
		stlog.Fatalln(err)
	}
	<- ctx.Done()
	
	fmt.Println("Shutting Down the Service!!!!")
}