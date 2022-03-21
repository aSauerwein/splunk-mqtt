#build stage
FROM golang:alpine AS builder
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin/splunk-mqtt -v ./...

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/splunk-mqtt /splunk-mqtt
ENTRYPOINT /splunk-mqtt
LABEL Name=splunkmqtt Version=0.0.1
