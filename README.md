# idsreplay

idsreplay reads IDS/IPS signature rule files (suricata format) and replays selected rules against a given target (just an open TCP port needed). 

I'm using it to demo VMware NSX IDS/IPS without the need to install tools like metasploit or known vulnerable software versions.


The current version parses rules from the [Open Emerging Threats Ruleset](https://rules.emergingthreats.net/open/suricata-4.0/) and replays randomly some of the basic http rules (~300)



## How to run?

idsreplay source and binary is available [here](https://github.com/danpaul81/idsreplay). 

I've also created container images, a k8s deployment and OVA image.

### a) Container Host
#### Run a Container image based demo against IP 172.16.10.20 TCP Port 80

```bash
docker run --name=idsreplay -e IDSREPLAYOPTS='--dest 172.16.10.20 --dport 80' danpaul81/idsreplay
```

#### To setup a  possible target you can use the simple golang webserver on host 172.16.10.20
```bash
docker run --name=nsx-demo -p 80:5000  danpaul81/nsx-demo
```

### b) k8s deployment when using NSX CNI
This is rolling out a "target", a k8s service and the idsreplay "source" deployment within a new namespace.
```bash
kubectl apply -f https://raw.githubusercontent.com/danpaul81/idsreplay/main/k8s-idsreplay.yaml
````
When using in non-vmware corp network  change the image source to your own registry / dockerhub

### c) OVA Image
Download from [my repo](https://github.com/danpaul81/idsparser/ova/output-vsphere-iso)

Target and Source are combined in this ova.

#### Deploy first time with option "IDS Replay Source" = True, the Target Port and the Target IP

#### Deploy second time with option "IDS Replay Source" = False and the Target Port

## How to Demo NSX IPS mode?
Most of the replayed rules will match NSX IDS Signature 2024897 which matches the http user agent "go http client user-agent". 

Setting this signature action to "drop" and creating a prevent rule should work fine.

## idsreplay command line options
You can pass command line options to container based workloads using the IDSREPLAYOPTS environment variable. 

Valid options are:
```
--dest [target ip or fqdn], default 127.0.0.1
--dport [target tcp port], default 80
--count [num of replay attempts] default 0 -> unlimited, counts only successful TCP connects
--waitsec [seconds to wait between replay attempts], default 5
--rulefile [path to ids signatures, suricata 4 format] default /idsreplay/emerging-all.rules
```

