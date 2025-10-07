# Prog Demos Backend

## How to run?

### 1. build task-runner
Special docker image used for running C++ tasks needs to be built before running the backend.
The backend depends on it for correct running of the tasks.
`sudo docker build -t task-runner:latest -f ./Docker/task-runner/task-runner.Dockerfile ./Docker/task-runner/`

### 2. build and run backend
Run the command to build the backend: `docker compose up --build`

#### To run locally
- use UNIX system (Linux/Mac) or proceed with Windows if you're not going to run tasks
- install `lizard` for solution code analysis (`pip install lizard`)
- ensure to explicitly expose port 5432 of the DB container
- run `docker compose up --build` with now exposed database on port 5432
- delete the backend container
- start the backend manually by running a VSCode task "Rock again!" (take a look at tasks.json what it does)

## How to access running services?

The backend serves `localhost:8080`.
To access the database, pgAdmin is available at `localhost:8888`.
