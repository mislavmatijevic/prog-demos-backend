FROM gcc:14.2.0

RUN useradd -m -s /bin/false tester

WORKDIR /home/tester/
COPY ./entrypoint.sh /entrypoint.sh
RUN chmod 500 /entrypoint.sh

ENTRYPOINT ["bash", "/entrypoint.sh"]