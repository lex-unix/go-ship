#!/bin/bash

mkdir -p /usr/local/bin/
cd /faino && go build -buildvcs=false -o /usr/local/bin/faino ./cmd/faino/
chmod +x /usr/local/bin/faino

push_to_registry() {
    if ! stat /registry/docker/registry/v2/repositories/$1/_manifests/tags/$2/current/link > /dev/null; then
        hub_tag=$1:$2
        registry_4443_tag=registry:4443/$1:$2
        docker pull $hub_tag
        docker tag $hub_tag $registry_4443_tag
        docker push $registry_4443_tag
    fi
}

push_to_registry node 20.15.1-slim
push_to_registry traefik v3.1

rm -f /root/.ssh/known_hosts
ssh-keyscan -H vm1 >> /root/.ssh/known_hosts
ssh-keyscan -H vm2 >> /root/.ssh/known_hosts
