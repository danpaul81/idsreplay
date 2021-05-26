# Reference for building PhotonOS Virtual Appliance (OVA) using Packer

Based on William Lams Packer PhotonOS Appliance (https://github.com/lamw/photonos-appliance)

## Requirements

* MacOS or Linux Desktop
* vCenter Server 
* [VMware OVFTool](https://www.vmware.com/support/developer/ovf/)
* [Packer](https://www.packer.io/intro/getting-started/install.html)
* [PowerShell](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell-core-on-linux?view=powershell-7.1)
* [PowerCli](https://developer.vmware.com/powercli)


> `packer` builds the OVA on a remote ESXi host via the [`vsphere-iso`](https://www.packer.io/docs/builders/vsphere-iso.html) builder. 


`photon-builder.json` vCenter Server settings where build VM will be created
`photon-version.json` PhotonOS ISO Image / Source and VM Settings 
`photon.json` packer main config file
`photon-dev.json` identical to `photon.json` but additionally deploys newly created OVA appliance(s) to a vCenter Server

idsreplay specific modifications are done in files/setup.sh

```
./build.sh
````

If you wish to automatically deploy the PhotonOS appliance after successfully building the OVA run `./build.sh dev` instead.
