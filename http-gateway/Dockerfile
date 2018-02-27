FROM golang:1.9 as builder

ARG PACKAGE=github.com/projectriff/http-gateway
ARG COMMAND=cmd/http-gateway.go

WORKDIR /go/src/${PACKAGE}
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 go build -v -a -installsuffix cgo ${COMMAND}

###########

FROM scratch

ARG PACKAGE=github.com/projectriff/http-gateway
COPY --from=builder /go/src/${PACKAGE}/http-gateway /http-gateway

CMD ["/http-gateway"]