package middleware

import (
	"net/http"
	"strings"
)

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
