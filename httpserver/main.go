package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strings"

)


type helloHandler struct{}

func (*helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for k, v := range r.Header {
		//fmt.Println(k, strings.Join(v, ""))
		w.Header().Add(k, strings.Join(v, ""))
	}

	w.Header().Add("VERSION", os.Getenv("VERSION"))

	fmt.Fprintf(w, "Hello!")
}

type healthzHandler struct{}

func (*healthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "200")
}

// Request.RemoteAddress 包含了端口，我们需要把它删掉，比如: "[::1]:58292" => "[::1]"
func ipAddrWithoutPort(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

//
func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Access from", ipAddrWithoutPort(r.RemoteAddr))

		h.ServeHTTP(w, r)
	})
}



func main() {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = Logger(handler)

	mux.Handle("/hello", &helloHandler{})
	mux.Handle("/healthz", &healthzHandler{})

	server := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}

	// 创建系统信号接收器
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done

		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal("Shutdown server:", err)
		}
	}()

	log.Println("Starting HTTP server...")
	err := server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			log.Print("Server closed under request")
		} else {
			log.Fatal("Server closed unexpected")
		}
	}
}
