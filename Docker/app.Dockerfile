FROM golang:1.22 AS builder

WORKDIR /app

COPY cmd cmd
COPY internal internal
COPY go.mod go.sum ./
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o prog-demos-backend ./cmd/prog-demos-backend/main.go

FROM python:3.9-slim

RUN pip install lizard
RUN apt-get update && apt-get install -y g++ curl && rm -rf /var/lib/apt/lists/*
ENV DOCKERVERSION=27.2.0
RUN curl -fsSLO https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKERVERSION}.tgz \
    && tar xzvf docker-${DOCKERVERSION}.tgz --strip 1 -C /usr/local/bin docker/docker \
    && rm docker-${DOCKERVERSION}.tgz

WORKDIR /app

COPY --from=builder /app/prog-demos-backend ./
COPY .env ./
COPY ./Docker/task-runner/task-runner.Dockerfile ./Docker/task-runner/entrypoint.sh ./
COPY ./internal/mailing/templates/ ./internal/mailing/templates/

CMD ["./prog-demos-backend"]
