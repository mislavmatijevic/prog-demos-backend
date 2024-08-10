FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ENV DOCKERVERSION=27.1.1
RUN curl -fsSLO https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKERVERSION}.tgz \
    && tar xzvf docker-${DOCKERVERSION}.tgz --strip 1 -C /usr/local/bin docker/docker \
    && rm docker-${DOCKERVERSION}.tgz

COPY ./ ./
COPY ./Docker/task-runner/task-runner.Dockerfile ./Docker/task-runner/entrypoint.sh ./

# g++ is used for syntax-checking C++ solutions before sending them to task-runner.
RUN apt-get install -y g++

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/prog-demos-backend/main.go

CMD ["./main"]
