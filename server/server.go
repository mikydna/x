package server

import (
	"log"
	"net/http"
	"time"
)

import (
	"github.com/facebookgo/httpdown"
)

type Server struct {
	instance httpdown.Server

	*http.ServeMux
}

func NewServer() *Server {
	server := &Server{
		ServeMux: http.NewServeMux(),
	}

	return server
}

func (s *Server) Stop() {
	if s.instance != nil {
		s.instance.Stop()

		log.Printf("Server instance stopped")

	} else {
		log.Println("Nothing to stop")

	}

	s.instance = nil
}

func (s *Server) ListenAndServe(addr string) error {
	conf := httpdown.HTTP{
		StopTimeout: 2 * time.Second,
		KillTimeout: 2 * time.Second,
	}

	server := &http.Server{
		Addr:    addr,
		Handler: s,
	}

	log.Printf("Starting control server on '%s'", addr)
	if inst, err := conf.ListenAndServe(server); err == nil {
		s.instance = inst
		return s.instance.Wait()

	} else {
		return err
	}
}
