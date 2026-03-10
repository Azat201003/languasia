FROM golang:alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY src ./src
RUN CGO_ENABLED=0 GOOS=linux go build -o main.out ./src/cmd

FROM alpine:latest
RUN apk --no-cache add curl
COPY --from=builder /build/main.out ./
COPY private_key.pem public_key.pem ./
CMD ./main.out | tee -a log


