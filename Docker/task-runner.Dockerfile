FROM node:21-alpine

RUN apk add python3 make g++

WORKDIR /app

COPY task-runner/task-runner.js task-runner/package*.json ./
RUN npm install --production

CMD ["node", "task-runner.js"]