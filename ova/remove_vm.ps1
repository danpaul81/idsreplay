Param(
    [Parameter(Position=1)]
    [string]$VISERVER,

    [Parameter(Position=2)]
    [string]$VIUSERNAME,

    [Parameter(Position=3)]
    [string]$VIPASSWORD,

    [Parameter(Position=4)]
    [string]$VMNAME
)

Connect-VIServer -Force -Server "$VISERVER" -User "$VIUSERNAME" -Password "$VIPASSWORD"
$vm = Get-VM "$VMNAME"
Remove-VM -VM $vm -DeleteFromDisk -Confirm:$false -RunAsync
Disconnect-VIServer * -Confirm:$false