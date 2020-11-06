FROM golang:1.15 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server ./cmd/inventory-notification

FROM alpine:3
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server /server

CMD ["/server"]