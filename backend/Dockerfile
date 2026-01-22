FROM golang AS builder

WORKDIR build

COPY go.mod go.sum ./

RUN go mod download

COPY src ./src

RUN CGO_ENABLED=0 GOOS=linux go build -o /build/main.out ./src
RUN ls
RUN pwd

FROM alpine AS runner

COPY --from=builder /build/main.out ./
COPY private_key.pem public_key.pem ./

RUN apk --no-cache add curl

CMD ./main.out | tee -a log

