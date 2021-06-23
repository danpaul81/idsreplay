#!/bin/bash

# Bootstrap script

set -euo pipefail

if [ -e /root/ran_customization ]; then
    exit
else
    NETWORK_CONFIG_FILE=$(ls /etc/systemd/network | grep .network)

    DEBUG_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.debug")
    DEBUG=$(echo "${DEBUG_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    LOG_FILE=/var/log/bootstrap.log
    if [ ${DEBUG} == "True" ]; then
        LOG_FILE=/var/log/photon-customization-debug.log
        set -x
        exec 2> ${LOG_FILE}
        echo
        echo "### WARNING -- DEBUG LOG CONTAINS ALL EXECUTED COMMANDS WHICH INCLUDES CREDENTIALS -- WARNING ###"
        echo "### WARNING --             PLEASE REMOVE CREDENTIALS BEFORE SHARING LOG            -- WARNING ###"
        echo
    fi

    SRC_HOSTNAME_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.src_hostname")
    DST_HOSTNAME_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.dst_hostname")
    SRC_IP_ADDRESS_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.src_ip")
    DST_IP_ADDRESS_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.dst_ip")
    NETMASK_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.netmask")
    GATEWAY_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.gateway")
    DNS_SERVER_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.dns")
    DNS_DOMAIN_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.domain")
    ROOT_PASSWORD_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.root_password")
    IDSREPLAY_ROLE_PROPERTY==$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.idsreplayrole")
    IDSREPLAY_PORT_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.idsreplayport")
    IDSREPLAY_SIDLIST_PROPERTY=$(vmtoolsd --cmd "info-get guestinfo.ovfEnv" | grep -m1 "guestinfo.sidlist")

    ROLE=$(echo "${IDSREPLAY_ROLE_PROPERTY}" | cut -d'"' -f4)
    IDSREPLAY_PORT=$(echo "${IDSREPLAY_PORT_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    IDSREPLAY_SIDLIST=$(echo "${IDSREPLAY_SIDLIST_PROPERTY}" | cut -d'"' -f4)
    
    if ! [ -z "${IDSREPLAY_SIDLIST}" ]; then
	IDSREPLAY_SIDLIST="--sidlist '${IDSREPLAY_SIDLIST}'"
    fi


    if [ ${ROLE} == "source" ]; then
        DST_IP_ADDRESS=$(echo "${DST_IP_ADDRESS_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
	IP_ADDRESS_PROPERTY=$SRC_IP_ADDRESS_PROPERTY
	HOSTNAME_PROPERTY=$SRC_HOSTNAME_PROPERTY
    else
	IP_ADDRESS_PROPERTY=$DST_IP_ADDRESS_PROPERTY
	HOSTNAME_PROPERTY=$DST_HOSTNAME_PROPERTY
    fi

    ##################################
    ### No User Input, assume DHCP ###
    ##################################
    HOSTNAME=$(echo "${HOSTNAME_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
    if [ -z "${HOSTNAME}" ]; then
        cat > /etc/systemd/network/${NETWORK_CONFIG_FILE} << __CUSTOMIZE_PHOTON__
[Match]
Name=e*

[Network]
DHCP=yes
IPv6AcceptRA=no
__CUSTOMIZE_PHOTON__
    #########################
    ### Static IP Address ###
    #########################
    else

        IP_ADDRESS=$(echo "${IP_ADDRESS_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        NETMASK=$(echo "${NETMASK_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        GATEWAY=$(echo "${GATEWAY_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        DNS_SERVER=$(echo "${DNS_SERVER_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')
        DNS_DOMAIN=$(echo "${DNS_DOMAIN_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')

        echo -e "\e[92mConfiguring Static IP Address ..." > /dev/console
        cat > /etc/systemd/network/${NETWORK_CONFIG_FILE} << __CUSTOMIZE_PHOTON__
[Match]
Name=e*

[Network]
Address=${IP_ADDRESS}/${NETMASK}
Gateway=${GATEWAY}
DNS=${DNS_SERVER}
Domain=${DNS_DOMAIN}
__CUSTOMIZE_PHOTON__

    echo -e "\e[92mConfiguring hostname ..." > /dev/console
    hostnamectl set-hostname ${HOSTNAME}
    echo "${IP_ADDRESS} ${HOSTNAME}" >> /etc/hosts
    echo -e "\e[92mRestarting Network ..." > /dev/console
    systemctl restart systemd-networkd
    fi
    

    echo -e "\e[92mConfiguring root password ..." > /dev/console
    ROOT_PASSWORD=$(echo "${ROOT_PASSWORD_PROPERTY}" | awk -F 'oe:value="' '{print $2}' | awk -F '"' '{print $1}')

    if [ -z "${ROOT_PASSWORD}" ]; then
	echo "Empty password setting. No Change"
    else
        echo "root:${ROOT_PASSWORD}" | /usr/sbin/chpasswd
    fi

# idsreplay section  remove for other projects
# depending on appliance role (idsreplay source or target) prepare systemd service which starts the right container image with properties

    if [ ${ROLE} == "source" ]; then
	IDSSTARTCMD="/usr/bin/docker run --name idsreplay-src  -e IDSREPLAYOPTS='--dest ${DST_IP_ADDRESS}  --dport ${IDSREPLAY_PORT} ${IDSREPLAY_SIDLIST}' idsreplay"
	IDSSTOPCMD="/usr/bin/docker rm -f idsreplay-src"
    else
	IDSSTARTCMD="/usr/bin/docker run --name idsreplay-tgt -p ${IDSREPLAY_PORT}:5000 nsx-demo"
	IDSSTOPCMD="/usr/bin/docker rm -f idsreplay-tgt"
    fi

    cat > /etc/systemd/system/idsreplay.service << __CUSTOMIZE_SERVICE__
[Unit]
Description = idsreplay systemd service
After=docker.service

[Service]
StandardOutput=journal+console
ExecStart=${IDSSTARTCMD}
ExecStop=${IDSSTOPCMD}

[Install]
WantedBy=default.target
__CUSTOMIZE_SERVICE__
    systemctl enable idsreplay.service
    systemctl daemon-reload
    systemctl start idsreplay.service
# finished idsreplay specific stuff


    # Ensure we don't run customization again
    touch /root/ran_customization
fi
