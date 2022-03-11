package middleware

import (
	"github.com/sirupsen/logrus"
	"net/http"
)


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

		logrus.SetFormatter(&logrus.JSONFormatter{})

		logrus.WithFields(logrus.Fields{
			"clientIP": getClientIP(r),
			"uri":      r.RequestURI,
			"method":   r.Method,
			"status":   responseData.status,
		}).Info()

	})
}