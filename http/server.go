package http

import (
	"context"
	"fmt"
	"log"
	net "net/http"
	"time"

	h "github.com/enorith/http"
	"github.com/valyala/fasthttp"
)

type Server struct {
	k *h.Kernel
}

func (s *Server) Serve(addr string, register h.RouterRegister, done chan struct{}) {
	register(s.k.Wrapper(), s.k)

	if s.k.Handler == h.HandlerFastHttp {
		s.serveFastHttp(addr, done)
	} else if s.k.Handler == h.HandlerNetHttp {
		s.serveNetHttp(addr, done)
	}
}

func (s *Server) serveFastHttp(addr string, done chan struct{}) {
	srv := s.GetFastHttpServer(s.k)

	go func() {
		if err := srv.ListenAndServe(addr); err != nil {
			log.Fatalf("listen %s error: %v", addr, err)
		}
	}()
	log.Printf("%s served at [%s]", logPrefix("fasthttp"), addr)
	<-done
	log.Printf("%s stoping...", logPrefix("fasthttp"))
	if e := srv.Shutdown(); e != nil {
		log.Fatalf("%s shutdown error: %v", logPrefix("fasthttp"), e)
	}
	log.Printf("%s stopped", logPrefix("fasthttp"))
}

func (s *Server) serveNetHttp(addr string, done chan struct{}) {
	srv := net.Server{
		Addr:         addr,
		Handler:      s.k,
		ReadTimeout:  h.ReadTimeout,
		WriteTimeout: h.WriteTimeout,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != net.ErrServerClosed {
			log.Fatalf("listen %s error: %v", addr, err)
		}
	}()
	log.Printf("%s served at [%s]", logPrefix("net/http"), addr)
	<-done
	log.Printf("%s stoping...", logPrefix("net/http"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if e := srv.Shutdown(ctx); e != nil {
		log.Fatalf("%s shutdown error: %v", logPrefix("net/http"), e)
	}
	log.Printf("%s stopped", logPrefix("net/http"))
}

func (s *Server) GetFastHttpServer(kernel *h.Kernel) *fasthttp.Server {

	return &fasthttp.Server{
		Handler:            kernel.FastHttpHandler,
		Concurrency:        kernel.RequestCurrency,
		TCPKeepalive:       kernel.IsKeepAlive(),
		MaxRequestBodySize: kernel.MaxRequestBodySize,
		ReadTimeout:        h.ReadTimeout,
		WriteTimeout:       h.WriteTimeout,
		IdleTimeout:        h.IdleTimeout,
	}
}

func NewServer(k *h.Kernel) *Server {
	return &Server{k: k}
}

func logPrefix(handler string) string {
	return fmt.Sprintf("enorith/%s (%s)", h.Version, handler)
}
