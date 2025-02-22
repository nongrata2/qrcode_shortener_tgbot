FROM golang:1.23.6 AS builder
    
WORKDIR /app

COPY main.go go.mod go.sum ./

RUN go mod tidy

COPY . .

COPY .env .env

RUN cat .env

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o server .

FROM alpine:latest

COPY --from=builder /app/server ./

EXPOSE 8080

CMD ["./server", "--port", "8080"]