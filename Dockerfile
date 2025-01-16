FROM golang:1.22-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY ./static /app/static

RUN go build -o main ./main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/.env /app/.env  

COPY --from=builder /app/static /app/static

EXPOSE 8080

CMD ["/app/main"]


