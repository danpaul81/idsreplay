#!/bin/bash -eux

##
## Misc configuration
##

echo '> Disable IPv6'
echo "net.ipv6.conf.all.disable_ipv6 = 1" >> /etc/sysctl.conf

echo '> Enable Docker Daemon'
systemctl enable docker
echo '> Start Docker'
systemctl start docker

echo '> Loading Docker Images'
docker pull harbor-repo.vmware.com/dpaul/idsreplay
docker tag harbor-repo.vmware.com/dpaul/idsreplay idsreplay
docker pull harbor-repo.vmware.com/dpaul/nsx-demo
docker tag harbor-repo.vmware.com/dpaul/nsx-demo nsx-demo

#echo '> Applying latest Updates...'
tdnf -y update || true

echo '> Installing Additional Packages...'
tdnf install -y \
  less \
  logrotate \
  curl \
  wget \
  unzip \
  awk \
  tar

echo '> Done'
