#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling SSL files"

mkdir /home/aceuser/ace-server/ssl/

if ls /home/aceuser/initial-config/ssl/* >/dev/null 2>&1; then
  for sslfile in `ls /home/aceuser/initial-config/ssl/*`; do
    if [ -s "${sslfile}" ]; then
      cp "${sslfile}" /home/aceuser/ace-server/ssl/.
    fi
  done
fi

log "SSL configuration complete"
