FROM gcc:14.2.0

RUN chmod 100 /bin/sh /usr/bin/sh /bin/bash /usr/bin/bash /bin/rbash /usr/bin/rbash /bin/dash /usr/bin/dash

RUN useradd -M -s /bin/false tester
RUN mkdir /playground
RUN chown tester:tester /playground

COPY ./entrypoint.sh /entrypoint.sh
RUN chmod 500 /entrypoint.sh

ENTRYPOINT ["bash", "/entrypoint.sh"]