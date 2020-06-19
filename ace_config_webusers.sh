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
OPERATORUSERSFILE=/home/aceuser/initial-config/webusers/operator-users.txt
EDITORUSERSFILE=/home/aceuser/initial-config/webusers/editor-users.txt
AUDITUSERSFILE=/home/aceuser/initial-config/webusers/audit-users.txt
VIEWERUSERSFILE=/home/aceuser/initial-config/webusers/viewer-users.txt

if [ -s $ADMINUSERSFILE ] || [ -s $OPERATORUSERSFILE ] || [ -s $EDITORUSERSFILE ] || [ -s $AUDITUSERSFILE ] || [ -s $VIEWERUSERSFILE ]; then
  OUTPUT=$(mqsichangeauthmode -w /home/aceuser/ace-server -s active -m file 2>&1)
  logAndExitIfError $? "${OUTPUT}"

  OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r admin -p all+ 2>&1)
  logAndExitIfError $? "${OUTPUT}"

  OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r operator -p read+,write-,execute+ 2>&1)
  logAndExitIfError $? "${OUTPUT}"
  
  OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r editor -p read+,write+,execute- 2>&1)
  logAndExitIfError $? "${OUTPUT}"

  OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r audit -p read+,write-,execute- 2>&1)
  logAndExitIfError $? "${OUTPUT}"

  OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r viewer -p read+,write-,execute- 2>&1)
  logAndExitIfError $? "${OUTPUT}"

  OLDIFS=${IFS}

  if [ -s $ADMINUSERSFILE ]; then
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
  fi

  if [ -s $OPERATORUSERSFILE ]; then
    IFS=$'\n'
    for line in $(cat $OPERATORUSERSFILE | tr -d '\r'); do
      if [[ $line =~ ^\# ]]; then
        continue
      fi
      IFS=${OLDIFS}
      fields=($line)
      log "Creating operator user ${fields[0]}"

      OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r operator 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    done
  fi

  if [ -s $EDITORUSERSFILE ]; then
    IFS=$'\n'
    for line in $(cat $EDITORUSERSFILE | tr -d '\r'); do
      if [[ $line =~ ^\# ]]; then
        continue
      fi
      IFS=${OLDIFS}
      fields=($line)
      log "Creating editor user ${fields[0]}"

      OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r editor 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    done
  fi

  if [ -s $AUDITUSERSFILE ]; then
    IFS=$'\n'
    for line in $(cat $AUDITUSERSFILE | tr -d '\r'); do
      if [[ $line =~ ^\# ]]; then
        continue
      fi
      IFS=${OLDIFS}
      fields=($line)
      log "Creating audit user ${fields[0]}"

      OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r audit 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    done
  fi

  if [ -s $VIEWERUSERSFILE ]; then
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
  fi
fi
