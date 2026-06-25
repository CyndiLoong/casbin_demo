FROM golang:1.26-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/config-docker.yaml ./config.yaml
COPY --from=builder /app/pkg/casbin/model.conf ./pkg/casbin/

EXPOSE 8080

CMD ["./server"]
