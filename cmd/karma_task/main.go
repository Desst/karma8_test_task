package main

import (
	"context"
	"fmt"
	service "karma/internal/service"
	"karma/internal/webhost"
	"karma/internal/webhost/handlers"
	"log"
	"time"
)

func main() {
	srv, err := service.NewStorageService(service.DefaultConfig())
	if err != nil {
		log.Fatalf(err.Error())
	}

	serviceCtrl := handlers.NewStorageService(srv)
	webServer := webhost.NewServer("127.0.0.1:8080", serviceCtrl)

	go func() {
		<-time.After(1 * time.Minute)
		if err := srv.AddNode(); err != nil {
			log.Println(fmt.Errorf("failed to add node: %w", err))
		}
	}()

	if err = webServer.Run(context.Background()); err != nil {
		log.Fatalf(err.Error())
	}
}
