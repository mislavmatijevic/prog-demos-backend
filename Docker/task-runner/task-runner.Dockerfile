FROM gcc:14.2.0

RUN chmod 100 /bin/sh /usr/bin/sh /bin/bash /usr/bin/bash /bin/rbash /usr/bin/rbash /bin/dash /usr/bin/dash
RUN useradd -m -s /bin/false tester

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod 500 /entrypoint.sh

ENTRYPOINT ["bash", "/entrypoint.sh"]