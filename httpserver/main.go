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

	"github.com/sirupsen/logrus"

)


type helloHandler struct{}

func (*helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 将request中带的header写入response header
	for k, v := range r.Header {
		w.Header().Add(k, strings.Join(v, ""))
	}
    // 读取当前系统的环境变量中的VERSION配置，并写入response header
	w.Header().Add("VERSION", os.Getenv("VERSION"))

	fmt.Fprintf(w, "Hello!")
}

type healthzHandler struct{}

func (*healthzHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 当访问/healthz时，应返回200
	fmt.Fprintf(w, "200")
}

// 处理Request.RemoteAddress，只保留ip地址，比如: "[::1]:58292" => "[::1]"
func ipAddrWithoutPort(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

// 获取客户端真实IP
func getClientIP(r *http.Request) string {
	IPAddr := r.Header.Get("X-Real-Ip")
	if IPAddr =="" {
		IPAddr = r.Header.Get("X-Forwarded-For")
	}
	if IPAddr =="" {
		IPAddr = r.RemoteAddr
	}
	return ipAddrWithoutPort(IPAddr)
}

type (
	responseData struct {
		status int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData        *responseData
	}
)

// 获取response中的statuscode
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithHTTPLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{
			status: http.StatusOK,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		logrus.WithFields(logrus.Fields{
			"clientip":   getClientIP(r),
			"uri":      r.RequestURI,
			"method":   r.Method,
			"status":   responseData.status,
		}).Info()


	})
}



func main() {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = WithHTTPLogging(handler)

	mux.Handle("/hello", &helloHandler{})
	mux.Handle("/healthz", &healthzHandler{})

	server := &http.Server{
		Addr:    ":8000",
		Handler: handler,
	}

	// 创建系统信号接收器，捕捉并处理操作系统对进程产生的信号
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done

		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal("Shutdown Server:", err)
		}
	}()

	log.Println("Starting HTTP Server...")
	err := server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			log.Println("HTTP Server shutted down")
		} else {
			log.Fatal("HTTP Server shutted unexpected")
		}
	}

}
