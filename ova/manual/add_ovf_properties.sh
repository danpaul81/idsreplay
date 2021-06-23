#!/bin/bash

ORIGPATH=$(pwd)
cd ..
OUTPUT_PATH="$(pwd)/output-vsphere-iso"
cd $ORIGPATH

VAPP_OVF=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}_vapp.ovf
VAPP_MF=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}_vapp.mf

APP_OVF=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}_app.ovf
APP_MF=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}_app.mf

# copy OVF files from packer output
cp ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}.ovf ${VAPP_OVF}
cp ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}.ovf ${APP_OVF}

#####
## STEP 1:  Modify files for vapp
#####

# set outputfiles for virtual systems
VIRTUALSYSTEM1_TEMP=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/VirtualSystem1.xml.temp
VIRTUALSYSTEM2_TEMP=${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/VirtualSystem2.xml.temp

rm -f ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/${PHOTON_APPLIANCE_NAME}.mf
rm -f ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}/*.temp

#replace envelope with simple one -> xmlstarlet doesnt work with original...
sed -i 's/<Envelope.*/<Envelope>/g' $VAPP_OVF

#extract VirtualSystem definition
xmlstarlet sel -t -c '/Envelope/VirtualSystem' $VAPP_OVF > $VIRTUALSYSTEM1_TEMP 2>/dev/null
xmlstarlet -L  ed -d '/Envelope/VirtualSystem' $VAPP_OVF 2>/dev/null

#duplicate virtual system definition
cp $VIRTUALSYSTEM1_TEMP $VIRTUALSYSTEM2_TEMP

# save disksize from original OVF
DISKSIZE=$(grep "<File" $VAPP_OVF |cut -d\" -f6)

# replace packer-created OVF with template
cp $VAPP_OVF_TEMPLATE $VAPP_OVF

# replace version/name/disksize/network in template
sed -i "s/{{VERSION}}/${PHOTON_VERSION}/g" $VAPP_OVF
sed -i "s/{{APPLIANCENAME}}/${PHOTON_APPLIANCE_NAME}/g" $VAPP_OVF
sed -i "s/{{DISKSIZE}}/${DISKSIZE}/g" $VAPP_OVF
sed -i "s/{{NETWORK}}/${PHOTON_NETWORK}/g" $VAPP_OVF


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

cat $VIRTUALSYSTEM1_TEMP >>$VAPP_OVF
cat $VIRTUALSYSTEM2_TEMP >>$VAPP_OVF

cat >>$VAPP_OVF <<EOF
  </VirtualSystemCollection>
</Envelope>
EOF

# replace network used by packer with generic one
sed -i "s/${PHOTON_NETWORK}/VM_Network/g" $VAPP_OVF

sed -i 's/<VirtualHardwareSection>/<VirtualHardwareSection ovf:transport="com.vmware.guestInfo">/g' $VAPP_OVF
sed -i '/^      <vmw:ExtraConfig ovf:required="false" vmw:key="nvram".*$/d' $VAPP_OVF
sed -i "/^    <File ovf:href=\"${PHOTON_APPLIANCE_NAME}-file1.nvram\".*$/d" $VAPP_OVF

#####
## STEP 2:  Modify files for virtual appliance
#####

TEMPLATENETWORK=$(grep "Network ovf:name" $APP_OVF |cut -d\" -f2)
sed -i "s/${TEMPLATENETWORK}/VM_Network/g" $APP_OVF



sed -i 's/<VirtualHardwareSection>/<VirtualHardwareSection ovf:transport="com.vmware.guestInfo">/g' $APP_OVF
sed -i "/    <\/vmw:BootOrderSection>/ r ${APP_OVF_TEMPLATE}" $APP_OVF
sed -i "s/{{VERSION}}/${PHOTON_VERSION}/g" $APP_OVF
sed -i '/^      <vmw:ExtraConfig ovf:required="false" vmw:key="nvram".*$/d' $APP_OVF
sed -i "/^    <File ovf:href=\"${PHOTON_APPLIANCE_NAME}-file1.nvram\".*$/d" $APP_OVF


#####
## STEP 3:  Create vapp and appliance & cleanup
#####

#generate manifest with hash
cd ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}
openssl sha1 ${PHOTON_APPLIANCE_NAME}_vapp.ovf ${PHOTON_APPLIANCE_NAME}-disk-0.vmdk > ${VAPP_MF}
openssl sha1 ${PHOTON_APPLIANCE_NAME}_app.ovf  ${PHOTON_APPLIANCE_NAME}-disk-0.vmdk > ${APP_MF}
cd $ORIGPATH

ovftool ${VAPP_OVF} ${OUTPUT_PATH}/${FINAL_PHOTON_APPLIANCE_NAME}_vapp.ova
chmod a+r ${OUTPUT_PATH}/${FINAL_PHOTON_APPLIANCE_NAME}_vapp.ova

ovftool ${APP_OVF} ${OUTPUT_PATH}/${FINAL_PHOTON_APPLIANCE_NAME}_app.ova
chmod a+r ${OUTPUT_PATH}/${FINAL_PHOTON_APPLIANCE_NAME}_app.ova

rm -rf ${OUTPUT_PATH}/${PHOTON_APPLIANCE_NAME}

