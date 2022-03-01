package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

//	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/spf13/cobra"
	"github.com/sirupsen/logrus"

	"github.com/CN/httpserver/config"
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
		responseData *responseData
	}
)

// WriteHeader: 获取response中的statuscode
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithHTTPLogging: 记录客户端访问日志，包括客户端IP，响应状态码等
func WithHTTPLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{
			status: http.StatusOK,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData: responseData,
		}
		h.ServeHTTP(&lw, r)

		logrus.WithFields(logrus.Fields{
			"clientIP": getClientIP(r),
			"uri":      r.RequestURI,
			"method":   r.Method,
			"status":   responseData.status,
		}).Info()

	})
}


func runServer() {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = WithHTTPLogging(handler)

	mux.Handle("/hello", &helloHandler{})
	mux.Handle("/healthz", &healthzHandler{})

	addr := viper.GetString("server.addr")
	log.Printf("HTTP Server listening on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// log.Println("HTTP Server starting...")
	go func() {
		// 开启一个goroutine启动服务
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP Server starting failed: %s\n", err)
		}
	}()

	// 等待中断信号来优雅地关闭服务器，为关闭服务器操作设置一个5秒的超时
	quit := make(chan os.Signal, 1)

	// signal.Notify把收到的 syscall.SIGINT或syscall.SIGTERM 信号转发给quit
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("HTTP Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("HTTP Server shutdown: ", err)
	}

	log.Println("HTTP Server shutdown successfully")
}

//
var rootCmd = &cobra.Command{
	Use:   "httpserver",
	Short: "httpserver is a simple restful api server",
	Long: `httpserver is a simple restful api server, use help get more info`,
	Run: func(cmd *cobra.Command, args []string) {
		runServer()
	},
}

var cfgFile string

// 初始化, 设置 flag 等
func init() {
	cobra.OnInitialize(initConfig)
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./conf/default.conf)")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c","", "config file (default: ./conf/default.conf)")
}

// 初始化配置
func initConfig() {
	c := config.Config{
		Name: cfgFile,
	}

	if err := c.InitConfig(); err != nil {
		panic(err)
	}
	log.Printf("config file loaded successful.")
	//c.WatchConfig(configChange)
}

// 包装了 rootCmd.Execute()
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}


