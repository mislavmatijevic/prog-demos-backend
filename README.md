# Prog Demos Backend

## How to run?

### 1. build task-runner
Special docker image used for running C++ tasks needs to be built before running the backend.
The backend depends on it for correct running of the tasks.
`sudo docker build -t task-runner:latest -f ./Docker/task-runner/task-runner.Dockerfile ./Docker/task-runner/`

### 2. build and run backend
If you want to host tasks repository, you need to put some content in `Docker/prog-demos-repository/files` or another location specified in .env's `TASKS_REPOSITORY_LOCATION`.
Run the command to build and start the backend: `docker compose up --build`

#### To run locally
- use UNIX system (Linux/Mac) or proceed with Windows if you're not going to run tasks
- install `lizard` for solution code analysis (`pip install lizard`)
- run command `cd ./prog-demos-backend && docker compose --project-name prog-demos-backend-dev -f dev.docker-compose.yml --env-file .env up -d && nodemon --watch './**/*.go' --signal SIGTERM --exec 'go' run ./cmd/prog-demos-backend/main.go`

## How to access running services?

The backend serves `localhost:8080`.
To access the database, pgAdmin is available at `localhost:8888`.
