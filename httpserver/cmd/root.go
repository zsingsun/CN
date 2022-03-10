package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"github.com/spf13/cobra"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/CN/httpserver/config"
	"github.com/CN/httpserver/middleware"
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
	// 添加0-2s 随机时延，并计算时延

	timeStart := time.Now()
	defer func() {
		//lantency := time.Since(timeStart)
		//log.Println(lantency)
		http_request_duration_seconds.Observe(time.Since(timeStart).Seconds())
	}()

	delay := rand.Intn(2000)
	time.Sleep(time.Millisecond*time.Duration(delay))

	// 当访问/healthz时，应返回200
	fmt.Fprintf(w, "200")
}

var (
	http_request_duration_seconds = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:		"http_request_duration_seconds",
            Help:		"Histogram of lantencies for HTTP requests",
         // Buckets:	[]float64{.1, .2, .4, 1, 3, 8, 20, 60, 120},
        },
    )
)

func runServer() {
	mux := http.NewServeMux()

	var handler http.Handler = mux
	handler = middleware.WithHTTPLogging(handler)

	mux.Handle("/hello", &helloHandler{})
	mux.Handle("/healthz", &healthzHandler{})

	mux.Handle("/metrics", promhttp.Handler())

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


