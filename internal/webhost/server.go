package webhost

import (
	"context"
	"karma/internal/webhost/handlers"
	"log"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
}

func NewServer(addr string, storageServiceHandler *handlers.StorageService) *Server {
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: handlers.CreateRouter(storageServiceHandler),
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	listenCtx, listenCancel := context.WithCancel(context.Background())
	go func() {
		log.Printf("Server started at %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				log.Println("Stopped serving")
				return
			}
			log.Printf("Serving error: %s", err.Error())
		}

		// if we've stopped listening, everything gotta stop
		listenCancel()
	}()

	select {
	case <-ctx.Done():
	case <-listenCtx.Done():
	}

	const gracefulShutdownTimeout = 1 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	err := s.server.Shutdown(ctx)
	cancel()
	return err
}
