#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling server.conf configuration"

if [ -s "/home/aceuser/initial-config/serverconf/server.conf.yaml" ]; then
  cp /home/aceuser/initial-config/serverconf/server.conf.yaml /home/aceuser/ace-server/overrides/server.conf.yaml
fi

log "server.conf configuration complete"
