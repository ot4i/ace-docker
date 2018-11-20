#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

if [ -z "${ACE_SERVER_NAME}" ]; then
  export ACE_SERVER_NAME=$(hostname | sed -e 's/[^a-zA-Z0-9._%/]//g')
fi

log() {
  MSG=$1
  TIMESTAMP=$(date +%Y-%m-%dT%H:%M:%S.%3NZ%:z)

  if [ "${LOG_FORMAT}" == "json" ]; then
    HOST=$(hostname)
    PID=$$
    PROCESSNAME=$(basename $0)
    USERNAME=$(id -un)
    ESCAPEDMSG=$(echo $MSG | sed -e 's/[\"]/\\&/g')
    # TODO: loglevel
    echo "{\"host\":\"${HOST}\",\"ibm_datetime\":\"${TIMESTAMP}\",\"ibm_processId\":\"${PID}\",\"ibm_processName\":\"${PROCESSNAME}\",\"ibm_serverName\":\"${ACE_SERVER_NAME}\",\"ibm_userName\":\"${USERNAME}\",\"message\":\"${ESCAPEDMSG}\",\"ibm_type\":\"ace_containerlog\"}"
  else
    echo "${TIMESTAMP} ${MSG}"
  fi
}

# logAndExitIfError - if the return code given is 0 and the command outputed text, log it to stdout
# If the return code is not 0, log command output text to stderr and exit with command's return code
# $1 - Return code from executed command
# $2 - Command output
logAndExitIfError() {
  RC=$1
  if [ "${RC}" -eq "0" ]; then
    if [ ! -z "$2" ]; then
      # For success, print as a log message
      log "$2"
    fi
  else
    # For errors, print the outhput to stderr
    >&2 echo -n $2
    exit ${RC}
  fi
}
