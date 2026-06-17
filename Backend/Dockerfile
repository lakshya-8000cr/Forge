FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY Backend/go.mod Backend/go.sum ./
RUN go mod download

COPY Backend/ ./

RUN go build -o forge-backend .

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache curl bash openssl ca-certificates

RUN curl -LO "https://dl.k8s.io/release/v1.35.5/bin/linux/amd64/kubectl" \
    && chmod +x kubectl \
    && mv kubectl /usr/local/bin/kubectl

RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

COPY --from=builder /app/forge-backend .

COPY charts ./charts

EXPOSE 8080

CMD ["./forge-backend"]