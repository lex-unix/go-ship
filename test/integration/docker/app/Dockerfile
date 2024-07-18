FROM ubuntu:22.04

WORKDIR /work

RUN apt-get update && apt-get install -y openssh-client

RUN mkdir -p /root/.ssh && ln -s /shared/ssh/id_rsa /root/.ssh/id_rsa

COPY boot.sh .

RUN chmod +x ./boot.sh

CMD [ "./boot.sh" ]
