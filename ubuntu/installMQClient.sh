#!/bin/bash -ex

MQ_PACKAGES="ibmmq-runtime ibmmq-client"
export DEBIAN_FRONTEND=noninteractive

 APT_URL="http://archive.ubuntu.com/ubuntu/"
 UBUNTU_CODENAME="xenial"
 # Use a reduced set of apt repositories.
# This ensures no unsupported code gets installed, and makes the build faster
echo "deb ${APT_URL} ${UBUNTU_CODENAME} main restricted" > /etc/apt/sources.list
echo "deb ${APT_URL} ${UBUNTU_CODENAME}-updates main restricted" >> /etc/apt/sources.list
echo "deb ${APT_URL} ${UBUNTU_CODENAME}-security main restricted" >> /etc/apt/sources.list
# Install additional packages required by MQ, this install process and the runtime scripts
apt-get update
apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    tar \

DIR_EXTRACT=/tmp/mq
cd ${DIR_EXTRACT}
curl -LO $MQ_URL
tar -zxvf ./*.tar.gz

# Remove packages only needed by this script
apt-get purge -y \
    ca-certificates \
    curl 

# Remove any orphaned packages
apt-get autoremove -y

MQLICENSE=$(find ${DIR_EXTRACT} -name "mqlicense.sh")
${MQLICENSE} -text_only -accept
DIR_DEB=$(find ${DIR_EXTRACT} -name "*.deb" -printf "%h\n" | sort -u | head -1)
echo "deb [trusted=yes] file:${DIR_DEB} ./" > /etc/apt/sources.list.d/IBM_MQ.list
# Accept the MQ license 
# Install MQ using the DEB packages
apt-get update
apt-get install -y $MQ_PACKAGES
# Remove 32-bit libraries from 64-bit container
find /opt/mqm /var/mqm -type f -exec file {} \; | awk -F: '/ELF 32-bit/{print $1}' | xargs --no-run-if-empty rm -f
# Remove tar.gz files unpacked by RPM postinst scripts
find /opt/mqm -name '*.tar.gz' -delete

# Clean up all the downloaded files
rm -f /etc/apt/sources.list.d/IBM_MQ.list
rm -rf ${DIR_EXTRACT}