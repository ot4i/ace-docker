#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling truststore configuration"

if ls /home/aceuser/initial-config/truststore/*.crt >/dev/null 2>&1; then

  if [ $(cat /home/aceuser/initial-config/truststore/*.crt | wc -l) -gt 0 ]; then
    if [ -f /home/aceuser/ace-server/truststore.jks ]; then
      OUTPUT=$(rm /home/aceuser/ace-server/truststore.jks 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    fi
  fi

  IFS=$'\n'
  for file in `ls /home/aceuser/initial-config/truststore/*.crt`; do
    if [ -s "${file}" ]; then
      if [ -z "${ACE_TRUSTSTORE_PASSWORD}" ]; then
        log "No truststore password defined"
        exit 1
      fi

      filename=$(basename $file)
      alias=$(echo $filename | sed -e 's/\.crt$'//)
      OUTPUT=$(/opt/ibm/ace-11/common/jdk/jre/bin/keytool -importcert -trustcacerts -alias ${filename} -file ${file} -keystore /home/aceuser/ace-server/truststore.jks -storepass ${ACE_TRUSTSTORE_PASSWORD} -noprompt 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    fi
  done
fi

log "Truststore configuration complete"
