package main

import (
	"Distributed_System/registry"
	"context"
	"net/http"
	"fmt"
	"log"
)

func main() {
	http.Handle("/Services",&registry.RegistryService{})

	ctx,cancer := context.WithCancel(context.Background())
	defer cancer()

	var srv http.Server
	srv.Addr = registry.ServerPort

	go func(){
		log.Println(srv.ListenAndServe())
		cancer()
	}()
	
	go func(){
		fmt.Println("Press any key to stop the process!!!!")
		var s string
		fmt.Scanln(&s)
		srv.Shutdown(ctx)
		cancer()
	}()
		
	<-ctx.Done()
	fmt.Println("Process finished!!!")
}