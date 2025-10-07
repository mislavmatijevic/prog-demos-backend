# Prog Demos Backend

## How to run?

### 1. build task-runner
Special docker image used for running C++ tasks needs to be built before running the backend.
The backend depends on it for correct running of the tasks.
`sudo docker build -t task-runner:latest -f ./Docker/task-runner/task-runner.Dockerfile ./Docker/task-runner/`

### 2. build and run backend
Run the command to build the backend: `docker compose up --build`

## How to access?

The backend serves `localhost:8080`.
To access the database, pgAdmin is available at `localhost:8888`.
