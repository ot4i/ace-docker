#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling extensions configuration"

if [ -s "/home/aceuser/initial-config/extensions/extensions.zip" ]; then
  mkdir /home/aceuser/ace-server/extensions
  unzip /home/aceuser/initial-config/extensions/extensions.zip -d /home/aceuser/ace-server/extensions
fi

log "extensions configuration complete"
