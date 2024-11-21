FROM alpine:latest

RUN apk update && \
    apk add --no-cache clang libstdc++ bash util-linux

RUN chmod 100 /bin/sh /bin/bash

RUN adduser -D -s /bin/false tester
RUN mkdir /playground
RUN chown tester:tester /playground

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod 500 /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
