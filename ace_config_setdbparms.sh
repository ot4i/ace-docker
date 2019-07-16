#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

function argStrings {
  shlex() {
    python -c $'import sys, shlex\nfor arg in shlex.split(sys.stdin):\n\tsys.stdout.write(arg)\n\tsys.stdout.write(\"\\0\")'
  }
  args=()
  while IFS='' read -r -d ''; do
    args+=( "$REPLY" )
  done < <(shlex <<<$1)
  
  log "${args[0]}"
  log "${args[1]}"
  log "${args[2]}"

}

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
    if [[ $line == mqsisetdbparms* ]]; then 
      log "Running suppplied mqsisetdbparms command"
      OUTPUT=`eval "$line"`
    else
      shlex() {
        python -c $'import sys, shlex\nfor arg in shlex.split(sys.stdin):\n\tsys.stdout.write(arg)\n\tsys.stdout.write(\"\\0\")'
      }
      args=()
      while IFS='' read -r -d ''; do
        args+=( "$REPLY" )
      done < <(shlex <<<$line)
      log "Setting user and password for resource: ${args[0]}"
      cmd="mqsisetdbparms -w /home/aceuser/ace-server -n \"${args[0]}\" -u \"${args[1]}\" -p \"${args[2]}\" 2>&1"
      OUTPUT=`eval "$cmd"`
      echo $OUTPUT
    fi
    logAndExitIfError $? "${OUTPUT}"
  done
fi

log "setdbparms configuration complete"
