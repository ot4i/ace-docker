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

log "Handling webusers configuration"

ADMINUSERSFILE=/home/aceuser/initial-config/webusers/admin-users.txt
OPERATORUSERSFILE=/home/aceuser/initial-config/webusers/operator-users.txt
EDITORUSERSFILE=/home/aceuser/initial-config/webusers/editor-users.txt
AUDITUSERSFILE=/home/aceuser/initial-config/webusers/audit-users.txt
VIEWERUSERSFILE=/home/aceuser/initial-config/webusers/viewer-users.txt

if [ -s $ADMINUSERSFILE ] || [ -s $OPERATORUSERSFILE ] || [ -s $EDITORUSERSFILE ] || [ -s $AUDITUSERSFILE ] || [ -s $VIEWERUSERSFILE ]; then
  OUTPUT=$(mqsichangeauthmode -w /home/aceuser/ace-server -s active -m file 2>&1)
  logAndExitIfError $? "${OUTPUT}"
fi

if [ -f $ADMINUSERSFILE ]; then
  if [ -s $ADMINUSERSFILE ]; then
    if [ -r $ADMINUSERSFILE ]; then
      OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r admin -p all+ 2>&1)
      logAndExitIfError $? "${OUTPUT}"

      OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r admin -o Data -p all+ 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    else
      log "ERROR: $ADMINUSERSFILE is not readable"
      exit 66
    fi
  else
    log "ERROR: $ADMINUSERSFILE is empty"
    exit 67
  fi
fi

if [ -f $OPERATORUSERSFILE ]; then
  if [ -s $OPERATORUSERSFILE ]; then
    if [ -r $OPERATORUSERSFILE ]; then
      OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r operator -p read+,write-,execute+ 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    else
      log "ERROR: $OPERATORUSERSFILE is not readable"
      exit 66
    fi
  else
    log "ERROR: $OPERATORUSERSFILE is empty"
    exit 67
  fi
fi

if [ -f $EDITORUSERSFILE ]; then
  if [ -s $EDITORUSERSFILE ]; then
    if [ -r $EDITORUSERSFILE ]; then
      OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r editor -p read+,write+,execute- 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    else
      log "ERROR: $EDITORUSERSFILE is not readable"
      exit 66
    fi
  else
    log "ERROR: $EDITORUSERSFILE is empty"
    exit 67
  fi
fi

if [ -f $AUDITUSERSFILE ]; then
  if [ -s $AUDITUSERSFILE ]; then
    if [ -r $AUDITUSERSFILE ]; then
      OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r audit -p read+,write-,execute- 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    else
      log "ERROR: $AUDITUSERSFILE is not readable"
      exit 66
    fi
  else
    log "ERROR: $AUDITUSERSFILE is empty"
    exit 67
  fi
fi

if [ -f $VIEWERUSERSFILE ]; then
  if [ -s $VIEWERUSERSFILE ]; then
    if [ -r $VIEWERUSERSFILE ]; then
      OUTPUT=$(mqsichangefileauth -w /home/aceuser/ace-server -r viewer -p read+,write-,execute- 2>&1)
      logAndExitIfError $? "${OUTPUT}"
    else
      log "ERROR: $VIEWERUSERSFILE is not readable"
      exit 66
    fi
  else
    log "ERROR: $VIEWERUSERSFILE is empty"
    exit 67
  fi
fi

if [ -s $ADMINUSERSFILE ] || [ -s $OPERATORUSERSFILE ] || [ -s $EDITORUSERSFILE ] || [ -s $AUDITUSERSFILE ] || [ -s $VIEWERUSERSFILE ]; then

  OLDIFS=${IFS}

  if [ -s $ADMINUSERSFILE ]; then
    IFS=$'\n'
    for line in $(cat $ADMINUSERSFILE | tr -d '\r'); do
      if [[ $line =~ ^\# ]]; then
        continue
      fi
      IFS=${OLDIFS}
      fields=($line)
      if [ $(mqsiwebuseradmin -w /home/aceuser/ace-server -l  | grep admin  | wc -l) -eq 0 ]
      then
        log "Creating admin user ${fields[0]}"

        OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r admin 2>&1)
        logAndExitIfError $? "${OUTPUT}"
      else
        log "admin user already exists"
      fi
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
      if [ $(mqsiwebuseradmin -w /home/aceuser/ace-server -l  | grep operator  | wc -l) -eq 0 ]
      then
        log "Creating operator user ${fields[0]}"

        OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r operator 2>&1)
        logAndExitIfError $? "${OUTPUT}"
      else
        log "operator user already exists"
      fi
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
      if [ $(mqsiwebuseradmin -w /home/aceuser/ace-server -l  | grep editor  | wc -l) -eq 0 ]
      then
        log "Creating editor user ${fields[0]}"

        OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r editor 2>&1)
        logAndExitIfError $? "${OUTPUT}"
      else
        log "editor user already exists"
      fi
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
      if [ $(mqsiwebuseradmin -w /home/aceuser/ace-server -l  | grep audit  | wc -l) -eq 0 ]
      then
        log "Creating audit user ${fields[0]}"

        OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r audit 2>&1)
        logAndExitIfError $? "${OUTPUT}"
      else
        log "audit user already exists"
      fi
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
      if [ $(mqsiwebuseradmin -w /home/aceuser/ace-server -l  | grep viewer  | wc -l) -eq 0 ]
      then
        log "Creating viewer user ${fields[0]}"

        OUTPUT=$(mqsiwebuseradmin -w /home/aceuser/ace-server -c -u ${fields[0]} -a ${fields[1]} -r viewer 2>&1)
        logAndExitIfError $? "${OUTPUT}"
      else
        log "viewer user already exists"
      fi
    done
  fi
fi
