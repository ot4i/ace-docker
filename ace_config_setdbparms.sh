#!/bin/bash

# © Copyright IBM Corporation 2018.
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
    if [[ $line == mqsisetdbparms* ]]; then
      log "Running suppplied mqsisetdbparms command"
      OUTPUT=`eval "$line"`
    else

      printf "%s" "$line" | xargs -n 1 printf "%s\n" > /tmp/creds
      IFS=$'\n' read -d '' -r -a lines <  /tmp/creds


      log "Setting user and password for resource: ${lines[0]}"
      cmd="mqsisetdbparms -w /home/aceuser/ace-server -n \"${lines[0]}\" -u \"${lines[1]}\" -p \"${lines[2]}\" 2>&1"
      OUTPUT=`eval "$cmd"`
      echo $OUTPUT
    fi
    logAndExitIfError $? "${OUTPUT}"
  done
fi

log "setdbparms configuration complete"
