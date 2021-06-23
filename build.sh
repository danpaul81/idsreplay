#!/bin/bash

VERSION="0.2.0"

go build -o bin/idsreplay .
docker build . -t idsreplay

docker tag idsreplay danpaul81/idsreplay:$VERSION
docker tag idsreplay danpaul81/idsreplay:latest
docker tag idsreplay harbor-repo.vmware.com/dpaul/idsreplay:$VERSION
docker tag idsreplay harbor-repo.vmware.com/dpaul/idsreplay:latest

docker push harbor-repo.vmware.com/dpaul/idsreplay:$VERSION
docker push harbor-repo.vmware.com/dpaul/idsreplay:latest
docker push danpaul81/idsreplay:$VERSION
docker push danpaul81/idsreplay:latest