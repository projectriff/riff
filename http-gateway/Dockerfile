FROM golang:1.8

WORKDIR /go/src/github.com/sk8sio/http-gateway
COPY . .
RUN go build cmd/http-gateway.go
ENTRYPOINT ["./http-gateway"]