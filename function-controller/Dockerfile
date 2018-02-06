FROM golang:1.9 as builder

ARG PACKAGE=github.com/projectriff/function-controller
ARG COMMAND=function-controller

WORKDIR /go/src/${PACKAGE}
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 go build -v -a -installsuffix cgo cmd/${COMMAND}.go

###########

FROM scratch

ARG PACKAGE=github.com/projectriff/function-controller
ARG COMMAND=function-controller
COPY --from=builder /go/src/${PACKAGE}/${COMMAND} /${COMMAND}

ADD tmp/ tmp/
CMD ["/function-controller"]