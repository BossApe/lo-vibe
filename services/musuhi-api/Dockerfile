FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o musuhi-api .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates && \
    apk add --no-cache curl jq git bash && \
    git lfs install && \
    apk add jujutsu && \
    jj config set --user user.name  "musuhi" && \
    jj config set --user user.email "musuhi@example.com"

WORKDIR /
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /app
COPY --from=builder /app/musuhi-api .
RUN chmod +x ./musuhi-api
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
CMD ["./musuhi-api"]
