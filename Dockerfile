FROM golang:1.17-alpine3.13 as builder
WORKDIR /go/src/HttpServer
COPY . .
RUN apk add --no-cache make
RUN make go-build

FROM scratch
COPY --from=builder /go/src/HttpServer/bin/httpserver /usr/bin/httpserver
COPY --from=builder /go/src/HttpServer/microservice/certificates/. /home/certificates/
ENTRYPOINT [ "/usr/bin/httpserver" ]