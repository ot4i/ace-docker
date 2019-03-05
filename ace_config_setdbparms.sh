#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

if [ -z "$MQSI_VERSION" ]; then
  source /opt/ibm/ace-11/server/bin/mqsiprofile
fi

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling setdbparms configuration"

if [ -s "/home/aceuser/initial-config/setdbparms/setdbparms.txt" ]; then
  FILE=/home/aceuser/initial-config/setdbparms/setdbparms.txt

  OLDIFS=${IFS}
  IFS=$'\n'
  for line in $(cat $FILE | tr -d '\r'); do
    if [[ $line =~ ^\# ]]; then
      continue
    fi
    IFS=${OLDIFS}
    fields=($line)
    log "Setting user and password for resource ${fields[0]}"

    OUTPUT=$(mqsisetdbparms -w /home/aceuser/ace-server -n ${fields[0]} -u ${fields[1]} -p ${fields[2]} 2>&1)
    logAndExitIfError $? "${OUTPUT}"
  done
fi

log "setdbparms configuration complete"
