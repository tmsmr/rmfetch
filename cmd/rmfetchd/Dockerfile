FROM golang:1.19-alpine3.16 as builder
RUN apk add --no-cache git
COPY . /build
WORKDIR /build
RUN GOOS=linux GOARCH=amd64 go build

FROM alpine:3.16
RUN adduser -D -g '' rmfetchd
USER rmfetchd
ENV GIN_MODE=release
EXPOSE 8080
CMD ["/rmfetchd"]
COPY --from=builder /build/rmfetchd /rmfetchd
