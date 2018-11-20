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

log "Handling webusers configuration"

ADMINUSERSFILE=/home/aceuser/initial-config/webusers/admin-users.txt
VIEWERUSERSFILE=/home/aceuser/initial-config/webusers/viewer-users.txt

if [ ! -f $ADMINUSERSFILE ]; then
  log "File not found: ${ADMINUSERSFILE}"
  exit 1
fi

if [ ! -f $VIEWERUSERSFILE ]; then
  log "File not found: ${VIEWERUSERSFILE}"
  exit 1
fi

OUTPUT=$(mqsichangeauthmode -w /home/aceuser/ace-server -s active -m file 2>&1)
logAndExitIfError $? "${OUTPUT}"

OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r admin -p all+ 2>&1)
logAndExitIfError $? "${OUTPUT}"

OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r viewer -p read+ 2>&1)
logAndExitIfError $? "${OUTPUT}"

OLDIFS=${IFS}
IFS=$'\n'
for line in $(cat $ADMINUSERSFILE | tr -d '\r'); do
  if [[ $line =~ ^\# ]]; then
    continue
  fi
  IFS=${OLDIFS}
  fields=($line)
  log "Creating admin user ${fields[0]}"

  OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r admin 2>&1)
  logAndExitIfError $? "${OUTPUT}"
done

IFS=$'\n'
for line in $(cat $VIEWERUSERSFILE | tr -d '\r'); do
  if [[ $line =~ ^\# ]]; then
    continue
  fi
  IFS=${OLDIFS}
  fields=($line)
  log "Creating viewer user ${fields[0]}"

  OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r viewer 2>&1)
  logAndExitIfError $? "${OUTPUT}"
done
