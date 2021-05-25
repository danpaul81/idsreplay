# Reference for building PhotonOS Virtual Appliance (OVA) using Packer

Based on William Lams Packer PhotonOS Appliance (https://github.com/lamw/photonos-appliance)

## Requirements

* MacOS or Linux Desktop
* vCenter Server 
* [VMware OVFTool](https://www.vmware.com/support/developer/ovf/)
* [Packer](https://www.packer.io/intro/getting-started/install.html)
* [powershell] (https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-linux?view=powershell-7.1)
* [powercli] (https://developer.vmware.com/powercli)


> `packer` builds the OVA on a remote ESXi host via the [`vsphere-iso`](https://www.packer.io/docs/builders/vsphere-iso.html) builder. 
```
idsreplay specific modifications are done in files/setup.sh

```
./build.sh
````

If you wish to automatically deploy the PhotonOS appliance after successfully building the OVA. You can edit the `photon-dev.xml.template` file and change the `ovftool_deploy_*` variables and run `./build.sh dev` instead.
