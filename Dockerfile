FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/pperzyna/pactbroker_exporter/
COPY . .
RUN GO111MODULE=on go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/pactbroker_exporter

FROM scratch
COPY --from=builder /go/bin/pactbroker_exporter /go/bin/pactbroker_exporter
ENTRYPOINT ["/go/bin/pactbroker_exporter"]