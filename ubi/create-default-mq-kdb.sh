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


if [ -f "/opt/mqm/bin/runmqckm" ]
then
  # 
  # Used if the downloaded package is the MQ client package from FixCentral. Example URL:
  #  
  # https://ak-delivery04-mul.dhe.ibm.com/sdfdl/v2/sar/CM/WS/0a3ih/0/Xa.2/Xb.jusyLTSp44S0BnrSUlhcQXsmOX33PXiMu_opTWF4XkF7jFZV8UxrP0RFSE0/Xc.CM/WS/0a3ih/0/9.2.0.4-IBM-MQC-LinuxX64.tar.gz/Xd./Xf.LPR.D1VK/Xg.11634360/Xi.habanero/XY.habanero/XZ.m7uIgNXpo_VTCGzC-hylOC79m0eKS5pi/9.2.0.4-IBM-MQC-LinuxX64.tar.gz
  # 
  # Also used if the downloaded package is the full MQ developer package. Example URL:
  #
  # https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqadv/mqadv_dev924_linux_x86-64.tar.gz
  #
  echo "Using runmqckm to create default MQ kdb from Java cacerts"
  /opt/mqm/bin/runmqckm -keydb -convert -db $MQSI_JREPATH/lib/security/cacerts -old_format jks -new_format kdb -pw changeit -target /tmp/mqcacerts.kdb -stash
else
  # 
  # Used if the downloaded package is the MQ redistributable client. Example URL:
  # 
  # https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.2.0.4-IBM-MQC-Redist-LinuxX64.tar.gz
  #
  echo "Did not find runmqckm; using keytool and runmqakm to create default MQ kdb from Java cacerts"
  $MQSI_JREPATH/bin/keytool -importkeystore -srckeystore $MQSI_JREPATH/lib/security/cacerts -srcstorepass changeit -destkeystore /tmp/java-cacerts.p12 -deststoretype pkcs12 -deststorepass changeit 
  /opt/mqm/bin/runmqakm -keydb -convert -db /tmp/java-cacerts.p12 -old_format p12 -new_format kdb -pw changeit -target /tmp/mqcacerts.kdb -stash
fi

