FROM gcc:14.2.0

RUN useradd -m tester

WORKDIR /home/tester/
COPY ./entrypoint.sh /home/tester/entrypoint.sh
RUN chmod +x /home/tester/entrypoint.sh
RUN chmod +x /home/tester/entrypoint.sh

ENTRYPOINT ["bash", "/home/tester/entrypoint.sh"]