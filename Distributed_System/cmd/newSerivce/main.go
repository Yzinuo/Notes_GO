package main

import (
	"Distributed_System/log"
	"Distributed_System/registry"
	stlog "log"
)

func main() {
	if logProvider, err := registry.GetProvider(r.ServiceName); err == nil {
		stlog.Println("Logging service found at : %S\n",logProvider)
		log.SetClientLogger(logProvider,r.ServiceName)
	}
}