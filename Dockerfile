FROM golang:1.22-alpine AS gobuilder

ENV GO111MODULE="on"
ENV GOPROXY="https://goproxy.cn,direct"
ENV CGO_ENABLED=0

WORKDIR /go/src/app
COPY . .

RUN apk update && apk add --no-cache ca-certificates
RUN go mod tidy
RUN go build -o go-push

FROM scratch

WORKDIR /root

COPY --from=gobuilder /go/src/app/go-push .
COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

CMD ["./go-push"]
