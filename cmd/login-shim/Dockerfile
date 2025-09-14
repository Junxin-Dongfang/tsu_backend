FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY cmd/login-shim/main.go .
RUN go mod init login-shim || true
RUN go build -o /login-shim main.go

EXPOSE 8090
CMD [ "/login-shim" ]