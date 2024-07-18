#!/bin/bash

# wait for the server to be available
sleep 10

# add server to known hosts
ssh-keyscan -H vm1 >> /root/.ssh/known_hosts
ssh-keyscan -H vm2 >> /root/.ssh/known_hosts

ssh root@vm1 "echo 'we made it into docker from docker (vm1)' > hello_docker.txt"
ssh root@vm2 "echo 'we made it into docker from docker (vm2)' > hello_docker.txt"
