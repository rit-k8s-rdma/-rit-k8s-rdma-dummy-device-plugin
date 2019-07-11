FROM golang:1.12 as builder

WORKDIR /go/src/github.com/swrap/rdma-dummy-device-plugin

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app

FROM scratch

WORKDIR /bin
COPY --from=builder /go/src/github.com/swrap/rdma-dummy-device-plugin .

CMD ["./app"]

# CMD "./app -logtostderr=true"
