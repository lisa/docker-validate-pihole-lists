FROM golang:1.12

ARG GOARCH=amd64

RUN mkdir -p /go/src/github.com/lisa/docker-validate-pihole-lists
WORKDIR /go/src/github.com/lisa/docker-validate-pihole-lists
COPY validate.go .
RUN CGO_ENABLED=0 GOARCH=${GOARCH} go build -ldflags '-extldflags "-static"' -a validate.go

FROM scratch
COPY --from=0 /go/src/github.com/lisa/docker-validate-pihole-lists/validate /validate
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT [ "/validate" ]