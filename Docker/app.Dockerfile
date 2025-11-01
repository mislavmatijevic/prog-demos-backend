FROM golang:1.25.1 AS builder

WORKDIR /app

COPY cmd cmd
COPY internal internal
COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o prog-demos-backend ./cmd/prog-demos-backend/main.go

FROM python:3.11-slim

RUN apt-get update && apt-get upgrade -y && apt-get install -y g++ curl && rm -rf /var/lib/apt/lists/*
RUN pip install --upgrade pip && pip install lizard
ENV DOCKERVERSION=27.2.0
RUN curl -fsSLO https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKERVERSION}.tgz \
    && tar xzvf docker-${DOCKERVERSION}.tgz --strip 1 -C /usr/local/bin docker/docker \
    && rm docker-${DOCKERVERSION}.tgz

WORKDIR /app

COPY --from=builder /app/prog-demos-backend ./
COPY .env ./
COPY prog-demos-bug-reporter.pem ./
COPY ./internal/mailing/templates/ ./internal/mailing/templates/

CMD ["./prog-demos-backend"]
