FROM golang:1.15.6 as builder
WORKDIR  /go/src/github.com/samkreter/givedirectly/
COPY . /go/src/github.com/samkreter/givedirectly/

# Ensure all tests pass to build
RUN go test ./... -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o givedirectly .

FROM alpine:3.8
RUN apk --update add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/samkreter/givedirectly/givedirectly .
ENTRYPOINT ["./givedirectly"]