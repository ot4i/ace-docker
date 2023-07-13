#!/bin/bash
# -*- mode: sh -*-
# © Copyright IBM Corporation 2022
#
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Fail on any non-zero return code
set -ex

# Download and extract the MQ files
DIR_TMP=/tmp/mq
mkdir -p ${DIR_TMP}
cd ${DIR_TMP}
if [ -z "$MQ_URL_USER" ]
then
      curl -LO $MQ_URL
else
      curl -LO -u  ${MQ_URL_USER}:${MQ_URL_PASS} $MQ_URL
fi
tar -xzf ./*.tar.gz
rm -f ./*.tar.gz

# Check what sort of MQ package was downloaded
if [ -f "${DIR_TMP}/bin/genmqpkg.sh" ]
then 
  # Generate MQ package in INSTALLATION_DIR
  # 
  # Used if the downloaded package is the MQ redistributable client package from FixCentral. Example URL:
  #
  # https://ak-delivery04-mul.dhe.ibm.com/sdfdl/v2/sar/CM/WS/0b41k/0/Xa.2/Xb.juSYLTSp44S03gmhuuKNya5CV2_SpQJ-2LPjgK-iHtqctHmMpcuW1gX5kas/Xc.CM/WS/0b41k/0/9.3.1.1-IBM-MQC-Redist-LinuxX64.tar.gz/Xd./Xf.LPR.D1VK/Xg.12236559/Xi.habanero/XY.habanero/XZ.0CQLlyfoENy06FxypR3Q8j0XiSzkn_t0/9.3.1.1-IBM-MQC-Redist-LinuxX64.tar.gz  
  #
  # or directly from
  #
  # https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.3.2.0-IBM-MQC-Redist-LinuxX64.tar.gz
  #
  echo "Detected genmqpkg.sh; installing MQ client components"
  export genmqpkg_inc32=0
  export genmqpkg_incadm=1
  export genmqpkg_incamqp=0
  export genmqpkg_incams=0
  export genmqpkg_inccbl=0
  export genmqpkg_inccics=0
  export genmqpkg_inccpp=1
  export genmqpkg_incdnet=0
  export genmqpkg_incjava=1
  export genmqpkg_incjre=${INSTALL_JRE}
  export genmqpkg_incman=0
  export genmqpkg_incmqbc=0
  export genmqpkg_incmqft=0
  export genmqpkg_incmqsf=0
  export genmqpkg_incmqxr=0
  export genmqpkg_incnls=0
  export genmqpkg_incras=1
  export genmqpkg_incsamp=0
  export genmqpkg_incsdk=0
  export genmqpkg_incserver=0
  export genmqpkg_inctls=1
  export genmqpkg_incunthrd=0
  export genmqpkg_incweb=0
  export INSTALLATION_DIR=/opt/mqm
  
  # Install requested parts
  ${DIR_TMP}/bin/genmqpkg.sh -b ${INSTALLATION_DIR}

  # Accept the MQ license
  ${INSTALLATION_DIR}/bin/mqlicense -accept
else
  # Check if should try install using RPM
  test -f /usr/bin/rpm && RPM=true || RPM=false
  if [ ! $RPM ]; then
    echo "Did not find the rpm command; cannot continue MQ client install without rpm"
    exit 9
  fi
  # 
  # Used if the downloaded package is the MQ client package from FixCentral. Example URL:
  #
  # https://ak-delivery04-mul.dhe.ibm.com/sdfdl/v2/sar/CM/WS/0b41a/0/Xa.2/Xb.jusyLTSp44S03gmhUUKWLqYocJ4a4bPB68Y8a4ir1VLbUn2OCPlMcoPXCfk/Xc.CM/WS/0b41a/0/9.3.1.1-IBM-MQC-LinuxX64.tar.gz/Xd./Xf.LPR.D1VK/Xg.12236562/Xi.habanero/XY.habanero/XZ.vr1ZpkKu0T03LhrAVMoMz9ec8315viqv/9.3.1.1-IBM-MQC-LinuxX64.tar.gz
  # 
  # Also used if the downloaded package is the full MQ developer package. Example URL:
  #
  # https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqadv/mqadv_dev932_linux_x86-64.tar.gz
  #
  echo "Did not find genmqpkg.sh; installing MQ client components using rpm"
  $RPM && DIR_RPM=$(find ${DIR_TMP} -name "*.rpm" -printf "%h\n" | sort -u | head -1)

  # Find location of mqlicense.sh
  MQLICENSE=$(find ${DIR_TMP} -name "mqlicense.sh")
  
  # Accept the MQ license
  ${MQLICENSE} -text_only -accept

  # Install MQ using the rpm packages
  $RPM && cd $DIR_RPM && rpm -ivh $MQ_PACKAGES

  # Remove tar.gz files unpacked by RPM postinst scripts
  find /opt/mqm -name '*.tar.gz' -delete
fi

rm -rf ${DIR_TMP}

# Create the directory for MQ configuration files
install --directory --mode 2775 --owner 1001 --group root /etc/mqm
