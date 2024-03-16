FROM golang:1.22-alpine as builder
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY ./cmd ./cmd
COPY ./internal ./internal
RUN go build -o /mount-service ./cmd/main.go

FROM alpine:latest
COPY --from=builder mount-service /bin/mount-service
ENTRYPOINT /bin/mount-service