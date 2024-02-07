# Start by building the application.
FROM golang:1.22 as builder

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/splunk-mqtt

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/splunk-mqtt /splunk-mqtt
CMD ["/splunk-mqtt"]