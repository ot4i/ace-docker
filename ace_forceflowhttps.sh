#!/bin/bash

# © Copyright IBM Corporation 2021.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

if [ -z "$MQSI_VERSION" ]; then
  source /opt/ibm/ace-12/server/bin/mqsiprofile
fi

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Creating force flows to be https keystore"

if [ -f /home/aceuser/ace-server/https-keystore.p12 ]; then
  OUTPUT=$(rm /home/aceuser/ace-server/https-keystore.p12 2>&1)
  logAndExitIfError $? "${OUTPUT}"
fi

IFS=$'\n'
KEYTOOL=/opt/ibm/ace-12/common/jdk/jre/bin/keytool
if [ ! -f "$KEYTOOL" ]; then
  KEYTOOL=/opt/ibm/ace-12/common/jre/bin/keytool
fi

if [ ! -f /home/aceuser/httpsNodeCerts/*.key ]; then
      log "No keystore files found at location /home/aceuser/httpsNodeCerts/*.key cannot create Force Flows HTTPS keystore"
      exit 1
fi

for keyfile in `ls /home/aceuser/httpsNodeCerts/*.key`; do
  if [ -s "${keyfile}" ]; then
    if [ -z "${1}" ]; then
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

    OUTPUT=$(openssl pkcs12 -export -in ${certfile} -inkey ${keyfile} -out /home/aceuser/ace-server/https-keystore.p12 -name ${alias} -password pass:${1} 2>&1)
    logAndExitIfError $? "${OUTPUT}"

    log "Setting https keystore password"
    cmd="mqsisetdbparms -w /home/aceuser/ace-server -n brokerHTTPSKeystore::password -u anything -p \"${1}\" 2>&1"
    OUTPUT=`eval "$cmd"`
    echo $OUTPUT

  fi
done

log "Force flows to be https keystore creation complete"
