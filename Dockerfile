FROM golang:1.18-alpine3.17 as builder

WORKDIR /app
COPY . .
RUN go build -o serve-json

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/serve-json .
COPY example.json config.yaml ./
ENTRYPOINT [ "./serve-json" ]