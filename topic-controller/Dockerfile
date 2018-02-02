FROM golang:1.9 as builder

ARG PACKAGE=github.com/projectriff/topic-controller
ARG COMMAND=topic-controller

WORKDIR /go/src/${PACKAGE}
COPY vendor/ vendor/
COPY cmd/ cmd/
COPY pkg/ pkg/

RUN CGO_ENABLED=0 go build -v -a -installsuffix cgo cmd/${COMMAND}.go

###########

FROM scratch

ARG PACKAGE=github.com/projectriff/topic-controller
COPY --from=builder /go/src/${PACKAGE}/${COMMAND} /${COMMAND}

ADD tmp /tmp
CMD ["/topic-controller"]
