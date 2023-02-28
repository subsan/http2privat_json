FROM golang:alpine AS builder
WORKDIR /build
ADD go.mod .
COPY . .
RUN go build -o main cmd/main.go

FROM alpine
WORKDIR /app
COPY --from=builder /build/main /app/
CMD ["./main"]