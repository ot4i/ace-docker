#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling keystore configuration"

if ls /home/aceuser/initial-config/keystore/*.key >/dev/null 2>&1; then

  if [ $(cat /home/aceuser/initial-config/keystore/*.key | wc -l) -gt 0 ]; then
    if [ -f /home/aceuser/ace-server/keystore.jks ]; then
      OUTPUT=$(rm /home/aceuser/ace-server/keystore.jks 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    fi
  fi

  IFS=$'\n'
  for keyfile in `ls /home/aceuser/initial-config/keystore/*.key`; do
    if [ -s "${keyfile}" ]; then
      if [ -z "${ACE_KEYSTORE_PASSWORD}" ]; then
        log "No keystore password defined"
        exit 1
      fi

      filename=$(basename ${keyfile})
      dirname=$(dirname ${keyfile})
      alias=$(echo ${filename} | sed -e 's/\.key$'//)
      certfile=${dirname}/${alias}.crt
      passphrasefile=${dirname}/${alias}.pass

      if [ ! -f ${certfile} ]; then
        log "Certificate file ${certfile} not found."
        exit 1
      fi

      if [ -f ${passphrasefile} ];then
        ACE_PRI_KEY_PASS=$(cat ${passphrasefile})
        OUTPUT=$(openssl pkcs12 -export -in ${certfile} -inkey ${keyfile} -passin pass:${ACE_PRI_KEY_PASS} -out /home/aceuser/ace-server/keystore.p12 -name ${alias} -password pass:${ACE_KEYSTORE_PASSWORD} 2>&1)
      else
        OUTPUT=$(openssl pkcs12 -export -in ${certfile} -inkey ${keyfile} -out /home/aceuser/ace-server/keystore.p12 -name ${alias} -password pass:${ACE_KEYSTORE_PASSWORD} 2>&1)
      fi
      logAndExitIfError $? "${OUTPUT}"

      OUTPUT=$(/opt/ibm/ace-11/common/jdk/jre/bin/keytool -importkeystore -srckeystore /home/aceuser/ace-server/keystore.p12 -destkeystore /home/aceuser/ace-server/keystore.jks -srcstorepass ${ACE_KEYSTORE_PASSWORD} -deststorepass ${ACE_KEYSTORE_PASSWORD} -srcalias ${alias} -destalias ${alias} -srcstoretype PKCS12 -noprompt 2>&1)
      logAndExitIfError $? "${OUTPUT}"

      OUTPUT=$(rm /home/aceuser/ace-server/keystore.p12 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    fi
  done
fi

log "Keystore configuration complete"
