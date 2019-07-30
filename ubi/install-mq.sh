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

mqm_uid=${1:-888}

test -f /usr/bin/microdnf && MICRODNF=true || MICRODNF=false
test -f /usr/bin/rpm && RPM=true || RPM=false


# Download and extract the MQ installation files
DIR_EXTRACT=/tmp/mq
mkdir -p ${DIR_EXTRACT}
cd ${DIR_EXTRACT}
if [ -z "$MQ_URL_USER" ]
then
      curl -LO $MQ_URL
else
      curl -LO -u  ${MQ_URL_USER}:${MQ_URL_PASS} $MQ_URL
fi
tar -zxf ./*.tar.gz

# Recommended: Create the mqm user ID with a fixed UID and group, so that the file permissions work between different images
groupadd --system --gid ${mqm_uid} mqm
useradd --system --uid ${mqm_uid} --gid mqm --groups 0 mqm
usermod -a -G mqm aceuser

# Find directory containing .rpm files
$RPM && DIR_RPM=$(find ${DIR_EXTRACT} -name "*.rpm" -printf "%h\n" | sort -u | head -1)
# Find location of mqlicense.sh
MQLICENSE=$(find ${DIR_EXTRACT} -name "mqlicense.sh")

# Accept the MQ license
${MQLICENSE} -text_only -accept

# Install MQ using the rpm packages
$RPM && cd $DIR_RPM && rpm -ivh $MQ_PACKAGES

# Remove tar.gz files unpacked by RPM postinst scripts
find /opt/mqm -name '*.tar.gz' -delete

# Clean up all the downloaded files
rm -rf ${DIR_EXTRACT}
