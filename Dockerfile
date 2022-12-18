FROM golang AS builder
WORKDIR /nlib-app-logs
COPY go.mod /nlib-app-logs/go.mod
COPY go.sum /nlib-app-logs/go.sum
RUN go mod download
COPY . /nlib-app-logs
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine
WORKDIR /nlib-app-logs
COPY --from=builder /nlib-app-logs/nlib-app-logs /nlib-app-logs/nlib-app-logs
ENTRYPOINT ["/nlib-app-logs/nlib-app-logs"]
