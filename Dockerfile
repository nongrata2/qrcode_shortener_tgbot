FROM golang:1.23 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o /tg_bot ./cmd/main.go

FROM alpine:latest

COPY --from=build /tg_bot /tg_bot

ENTRYPOINT ["/tg_bot"]