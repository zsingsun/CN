# 

FROM golang:alpine AS build

ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE="on"

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download && go mod verify

COPY ./ /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o httpserver


#
 
FROM alpine

ENV VERSION=1.0

WORKDIR /app

COPY --from=build /app/httpserver /app/httpserver
COPY conf /app/conf

EXPOSE 8081

ENTRYPOINT ["/app/httpserver"]
CMD [ "--config", "./conf/server.conf" ]
