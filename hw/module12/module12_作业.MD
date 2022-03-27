### 生成证书并创建secret
```sh
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=cncamp.example.io/O=cncamp" -addext "subjectAltName = DNS:cncamp.example.io"

kubectl create -n istio-system secret tls cncamp-credential --key=tls.key --cert=tls.crt
```

### 修改httpserver应用，添加service0和service1，将header透传，并部署到k8s
```sh
kubectl create ns tracing
kubectl label ns tracing istio-injection=enabled
kubectl -n tracing apply -f service0.yaml
kubectl -n tracing apply -f service1.yaml
kubectl -n tracing apply -f service2.yaml
```

### istio配置 /service0 和 /nginx 路由及https
```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: service0
spec:
  gateways:
    - service0
  hosts:
    - 'httpsserver.cncamp.example.io'
  http:
  - match:
      - port:
          443
      - uri:
          exact: /service0
    route:
      - destination:
          host: service0
          port:
            number: 80
  - match:
      - port:
          443
      - uri:
          prefix: "/nginx"
    rewrite:
      uri: "/"
    route:
      - destination:
          host: nginx
          port:
            number: 80
---
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: service0
spec:
  selector:
    istio: ingressgateway
  servers:
    - hosts:
        - 'httpsserver.cncamp.example.io'
      port:
        name: https-service0
        number: 443
        protocol: HTTPS
      tls:
        mode: SIMPLE
        credentialName: cncamp-credential
``` 

### 测试https及路由配置
```sh

t@node1:~/ws/CN/hw/module12# curl --resolve httpsserver.cncamp.example.io:443:172.16.184.193 https://httpsserver.cncamp.example.io/nginx -k
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
html { color-scheme: light dark; }
body { width: 35em; margin: 0 auto;
font-family: Tahoma, Verdana, Arial, sans-serif; }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
root@node1:~/ws/CN/hw/module12# curl --resolve httpsserver.cncamp.example.io:443:172.16.184.193 https://httpsserver.cncamp.example.io/service0 -k
===================Details of the http request header:============
HTTP/1.1 200 OK
Transfer-Encoding: chunked
Content-Type: text/plain; charset=utf-8
Date: Sun, 27 Mar 2022 19:18:17 GMT
Server: envoy
X-Envoy-Upstream-Service-Time: 107

348
===================Details of the http request header:============
HTTP/1.1 200 OK
Content-Length: 605
Content-Type: text/plain; charset=utf-8
Date: Sun, 27 Mar 2022 19:18:18 GMT
Server: envoy
X-Envoy-Upstream-Service-Time: 89

X-Forwarded-Proto=[https]
X-Envoy-Internal=[true]
X-Forwarded-Client-Cert=[By=spiffe://cluster.local/ns/tracing/sa/default;Hash=ac26b5a1d0182efdf20078a41b5b05954e6dbba1b654db0adc034a4816c65508;Subject="";URI=spiffe://cluster.local/ns/tracing/sa/default]
X-B3-Parentspanid=[368a5a9f597e303e]
Accept=[*/*]
Accept-Encoding=[gzip,gzip]
X-Request-Id=[9b9c9953-3f0e-9865-a088-173b08c2c764]
X-B3-Spanid=[3aea1261289092a1]
User-Agent=[Go-http-client/1.1,Go-http-client/1.1,curl/7.74.0]
X-Envoy-Attempt-Count=[1]
X-Forwarded-For=[172.17.166.128]
X-B3-Traceid=[18e561cc8c577c162bac2cef4c81e926]
X-B3-Sampled=[1]
ok

0

root@node1:~/ws/CN/hw/module12#
```

### 安装jaeger并配置tracing

```sh
kubectl apply -f jaeger.yaml
kubectl edit configmap istio -n istio-system

apiVersion: v1
data:
  mesh: |-
    accessLogFile: /dev/stdout
    defaultConfig:
      discoveryAddress: istiod.istio-system.svc:15012
      proxyMetadata: {}
      tracing:
        zipkin:
          address: jaeger-collector:9411
        sampling: 100
```

### 通过istioingress访问100次
```sh
kubectl get svc -nistio-system
NAME                   TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)                                                                      AGE
istio-egressgateway    ClusterIP      172.16.65.84     <none>        80/TCP,443/TCP                                                               28h
istio-ingressgateway   LoadBalancer   172.16.184.193   <pending>     15021:30916/TCP,80:32102/TCP,443:32497/TCP,31400:31407/TCP,15443:31608/TCP 

for i in `seq 100`; do curl --resolve httpsserver.cncamp.example.io:443:172.16.184.193 https://httpsserver.cncamp.example.io/service0 -v -k; done

```

### 通过jaeger dashboard查看tracing
```sh
istioctl dashboard jaeger
```

![image](https://github.com/zsingsun/CN/blob/master/hw/module12/jaeger_httpserver.jpg)
