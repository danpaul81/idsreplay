#!/bin/bash
go build -o bin/idsreplay .
docker build . -t idsreplay
docker tag idsreplay danpaul81/idsreplay
docker tag idsreplay harbor-repo.vmware.com/dpaul/idsreplay

docker push harbor-repo.vmware.com/dpaul/idsreplay
docker push danpaul81/idsreplay