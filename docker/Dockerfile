# syntax=docker/dockerfile:1
FROM alpine:3.19 AS builder

# Install Go 1.24
RUN apk add --no-cache wget tar
RUN wget -O go.tar.gz https://go.dev/dl/go1.24.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go.tar.gz && \
    rm go.tar.gz
ENV PATH="/usr/local/go/bin:$PATH"

# Install CGO dependencies and libxml2 for XSD validation
RUN apk add --no-cache \
    gcc \
    musl-dev \
    libxml2-dev \
    pkgconfig \
    git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o plugnfce-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/plugnfce-api .
EXPOSE 8080
CMD ["./plugnfce-api"] 