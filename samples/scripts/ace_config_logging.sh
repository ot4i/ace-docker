#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

log() {
  MSG=$1
  TIMESTAMP=$(date +%Y-%m-%dT%H:%M:%S.%3NZ%:z)
  echo "${TIMESTAMP} ${MSG}"
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
