FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY cmd/admin/main.go .
RUN go mod init admin || true
RUN go build -o /admin main.go

EXPOSE 8080
CMD [ "/admin" ]