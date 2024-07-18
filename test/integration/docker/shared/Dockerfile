FROM ubuntu:22.04

WORKDIR /work

RUN apt-get update && apt-get -y install openssh-client

RUN mkdir ssh & \
    ssh-keygen -t rsa -f ssh/id_rsa -N ""

COPY ./boot.sh .

RUN chmod +x ./boot.sh

HEALTHCHECK --interval=10s CMD pgrep sleep

CMD [ "./boot.sh" ]
