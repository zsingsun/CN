# 

FROM golang:alpine AS build

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE="on"

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY httpserver/main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o httpserver


#
 
FROM scratch

ENV VERSION=1.0.0

COPY --from=build /app/httpserver .
CMD ["./httpserver"]
