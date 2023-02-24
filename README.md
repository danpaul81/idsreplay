# idsreplay

idsreplay reads IDS/IPS signature rule files (suricata format) and replays selected rules against a given target (just an open TCP port needed). 

I'm using it to demo VMware NSX IDS/IPS without the need to install tools like metasploit or known vulnerable software versions.

Stephan Wolf created a YouTube Video: [See it in action](https://www.youtube.com/watch?v=iMnIwOu5QhY)


The current version parses rules from the [Open Emerging Threats Ruleset](https://rules.emergingthreats.net/open/suricata-4.0/) and replays randomly some of the basic http rules (~300)

If you want to replay a pre-defined set of SID you can pass them using the --sidlist parameter 
(defined within the IDSREPLAYOPTS environment variable when using the container images or define them in the deployment options when using the ova files)



## 1) How to run?

idsreplay source and binary is available [here](https://github.com/danpaul81/idsreplay). 

I've also created container images, a k8s deployment and OVA image.

### a) Container Host
#### Run a Container image based demo against IP 172.16.10.20 TCP Port 80

```bash
docker run --name=idsreplay -e IDSREPLAYOPTS='--dest 172.16.10.20 --dport 80' danpaul81/idsreplay:0.2.2
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
Download from [my repo](https://github.com/danpaul81/idsreplay/releases)

There is a *_vapp.ova which automatically creates source and target VM within a vApp.

If you cannot create a vApp in your vCenter you can also deploy the *_app.ova two times with the same (!) settings except:
#### Deploy first time with option "Rolename" "target"
#### Deploy second time with option "Rolename" "source"

## 2) How to Demo NSX IPS mode?
Most of the replayed rules will match NSX IDS Signature 2024897 which matches the http user agent "go http client user-agent". 

Setting this signature action to "drop" and creating a prevent rule should work fine.

## 3) How to Demo AVI WAF mode?
create a AVI application and enable WAF mode

you can use idsreplay virtual appliance in "target" role as backend-server(s)

configure source:

when running the go source or docker image add the --sidlist parameter with [these sids](https://github.com/danpaul81/idsreplay/blob/main/avi_waf_sid.txt)

when using idsreplay appliance (not vapp) in "source" role configure AVI VIF as destination IP and select "AVI WAF Demo Mode"

## 4) idsreplay command line options
You can pass command line options to container based workloads using the IDSREPLAYOPTS environment variable. 

Valid options are:
```
--count [num of replay attempts] default 0 -> unlimited, counts only successful TCP connects
--debug debug mode, show details when parsing rules
--dest [target ip or fqdn], default 127.0.0.1
--dport [target tcp port], default 80
--rulefile [path to ids signatures, suricata 4 format] default /idsreplay/emerging-all.rules
--sidlist [comma separated list of rule SID] replay a set of pre-defined rules
--waitsec [seconds to wait between replay attempts], default 5

```

