#!/bin/bash

OUTPUT_PATH="../output-vsphere-iso"

PACKER_OVF=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}.ovf

# set outputfiles for virtual systems
VIRTUALSYSTEM1_TEMP=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/VirtualSystem1.xml.temp
VIRTUALSYSTEM2_TEMP=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/VirtualSystem2.xml.temp

rm -f ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}.mf
rm -f ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/*.temp

# backup original ovf file for debug
cp ${PACKER_OVF} ${PACKER_OVF}.bak

#replace envelope with simple one -> xmlstarlet doesnt work with original...
sed -i 's/<Envelope.*/<Envelope>/g' $PACKER_OVF

#extract VirtualSystem definition
xmlstarlet sel -t -c '/Envelope/VirtualSystem' $PACKER_OVF > $VIRTUALSYSTEM1_TEMP 2>/dev/null
xmlstarlet -L  ed -d '/Envelope/VirtualSystem' $PACKER_OVF 2>/dev/null

#duplicate virtual system definition
cp $VIRTUALSYSTEM1_TEMP $VIRTUALSYSTEM2_TEMP

# save disksize from original OVF
DISKSIZE=$(grep "<File" $PACKER_OVF |cut -d\" -f6)

# replace packer-created OVF with template
cp $PHOTON_OVF_TEMPLATE $PACKER_OVF

# replace version/name/disksize/network in template
sed -i "s/{{VERSION}}/${PHOTON_VERSION}/g" $PACKER_OVF
sed -i "s/{{APPLIANCENAME}}/${PHOTON_APPLIANCE_NAME}/g" $PACKER_OVF
sed -i "s/{{DISKSIZE}}/${DISKSIZE}/g" $PACKER_OVF
sed -i "s/{{NETWORK}}/${PHOTON_NETWORK}/g" $PACKER_OVF


#setup Virtual System for idsreplay source
    #modify name
    sed -i "s/<VirtualSystem.*/<VirtualSystem ovf:id=\"${PHOTON_APPLIANCE_NAME}_${PHOTON_VERSION}_src\">/g" $VIRTUALSYSTEM1_TEMP
    sed -i "s/<Name>.*<\/Name>/<Name>${PHOTON_APPLIANCE_NAME}_${PHOTON_VERSION}_src<\/Name>/g" $VIRTUALSYSTEM1_TEMP
    #remove last tag and add new footer
    sed -i "/  <\/VirtualSystem>/d" $VIRTUALSYSTEM1_TEMP
    cat >>$VIRTUALSYSTEM1_TEMP <<EOF
    <ProductSection>
     <Info>Information about the installed software</Info>
      <Property ovf:userConfigurable="false" ovf:value="source" ovf:type="string" ovf:key="guestinfo.idsreplayrole">
	<Label>IDSreplay Source</Label>
        <Description>is this idsreplay source? If yes, also provide target IP</Description>
      </Property>
    </ProductSection>
  </VirtualSystem>
EOF
#setup Virtual System for idsreplay target
    #modify name
    sed -i "s/<VirtualSystem.*/<VirtualSystem ovf:id=\"${PHOTON_APPLIANCE_NAME}_${PHOTON_VERSION}_dst\">/g" $VIRTUALSYSTEM2_TEMP
    sed -i "s/<Name>.*<\/Name>/<Name>${PHOTON_APPLIANCE_NAME}_${PHOTON_VERSION}_dst<\/Name>/g" $VIRTUALSYSTEM2_TEMP
    #remove last tag and add new footer
    sed -i "/  <\/VirtualSystem>/d" $VIRTUALSYSTEM2_TEMP
    #modify disk to vmdisk2
    sed -i "s/vmdisk1/vmdisk2/g" $VIRTUALSYSTEM2_TEMP
    cat >>$VIRTUALSYSTEM2_TEMP <<EOF
    <ProductSection>
      <Info>Information about the installed software</Info>
      <Property ovf:userConfigurable="false" ovf:value="target" ovf:type="string" ovf:key="guestinfo.idsreplayrole">
       <Label>IDSreplay Role</Label>
       <Description>is this VM idsreplay target or source?</Description>
      </Property>
    </ProductSection>
  </VirtualSystem>
EOF

cat $VIRTUALSYSTEM1_TEMP >>$PACKER_OVF
cat $VIRTUALSYSTEM2_TEMP >>$PACKER_OVF

cat >>$PACKER_OVF <<EOF
  </VirtualSystemCollection>
</Envelope>
EOF

# replace network used by packer with generic one
sed -i "s/${PHOTON_NETWORK}/VM_Network/g" $PACKER_OVF

sed -i 's/<VirtualHardwareSection>/<VirtualHardwareSection ovf:transport="com.vmware.guestInfo">/g' $PACKER_OVF
sed -i '/^      <vmw:ExtraConfig ovf:required="false" vmw:key="nvram".*$/d' $PACKER_OVF
sed -i "/^    <File ovf:href=\"${PHOTON_APPLIANCE_NAME}-file1.nvram\".*$/d" $PACKER_OVF


ovftool ${PACKER_OVF} ${OUTPUT_PATH}/${FINAL_PHOTON_APPLIANCE_NAME}_vapp.ova
chmod a+r ${OUTPUT_PATH}/${FINAL_PHOTON_APPLIANCE_NAME}_vapp.ova

rm -rf ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}

