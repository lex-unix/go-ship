#!/bin/bash

install_goship() {
    cd /go-ship && go build -o /usr/local/bin/goship ./cmd/goship/
}

install_goship

ssh-keyscan -H vm1 >> /root/.ssh/known_hosts
ssh-keyscan -H vm2 >> /root/.ssh/known_hosts
