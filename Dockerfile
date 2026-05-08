FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/grafana-alert-webhook .

FROM alpine:3.22

ARG VERSION=dev
LABEL org.opencontainers.image.version=$VERSION

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/grafana-alert-webhook /usr/local/bin/grafana-alert-webhook

EXPOSE 1111

CMD ["grafana-alert-webhook", "-c", "/app/config.json"]
