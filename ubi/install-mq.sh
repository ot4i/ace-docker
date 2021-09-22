#!/bin/bash
# -*- mode: sh -*-
# © Copyright IBM Corporation 2015, 2019
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

# Download and extract the MQ unzippable files
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
ls -la ${DIR_TMP}

# Generate MQ package in INSTALLATION_DIR
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
${DIR_TMP}/bin/genmqpkg.sh -b ${INSTALLATION_DIR}
ls -la ${INSTALLATION_DIR}
rm -rf ${DIR_TMP}

# Accept the MQ license
${INSTALLATION_DIR}/bin/mqlicense -accept

# Create the directory for MQ configuration files
install --directory --mode 2775 --owner 1001 --group root /etc/mqm

