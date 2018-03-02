FROM golang:1.9 as builder

ARG COMPONENT

WORKDIR /go/src/github.com/projectriff/riff
COPY vendor/ vendor/
COPY kubernetes-crds/ kubernetes-crds/
COPY message-transport/ message-transport/

COPY ${COMPONENT}/cmd/ ${COMPONENT}/cmd/
COPY ${COMPONENT}/pkg/ ${COMPONENT}/pkg/

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /riff-entrypoint ${COMPONENT}/cmd/${COMPONENT}.go

###########

FROM scratch

# The following line forces the creation of a /tmp directory
WORKDIR /tmp

WORKDIR /

COPY --from=builder /riff-entrypoint /riff-entrypoint

ENTRYPOINT ["/riff-entrypoint"]
