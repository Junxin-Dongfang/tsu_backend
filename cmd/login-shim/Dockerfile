FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# 复制所有源代码
COPY . .

WORKDIR /app/cmd/login-shim

RUN CGO_ENABLED=0 go build -o /bin/login-shim .

FROM alpine:latest
COPY --from=builder /bin/login-shim /bin/login-shim

EXPOSE 8090
CMD [ "/bin/login-shim" ]