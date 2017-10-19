FROM golang:1.8

WORKDIR /go/src/github.com/sk8sio/function-sidecar
COPY . .
RUN go build cmd/function-sidecar.go
ENTRYPOINT ["./function-sidecar"]