#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling agent configuration"

if [ -s "/home/aceuser/initial-config/agent/switch.json" ]; then
  mkdir -p /home/aceuser/ace-server/config/iibswitch/switch
  cp /home/aceuser/initial-config/agent/switch.json /home/aceuser/ace-server/config/iibswitch/switch/switch.json
fi

if [ -s "/home/aceuser/initial-config/agent/agentx.json" ]; then
  mkdir -p /home/aceuser/ace-server/config/iibswitch/agentx
  cp /home/aceuser/initial-config/agent/agentx.json /home/aceuser/ace-server/config/iibswitch/agentx/agentx.json
fi

if [ -s "/home/aceuser/initial-config/agent/agentp.json" ]; then
  mkdir -p /home/aceuser/ace-server/config/iibswitch/agentp
  cp /home/aceuser/initial-config/agent/agentp.json /home/aceuser/ace-server/config/iibswitch/agentp/agentp.json
fi

if [ -s "/home/aceuser/initial-config/agent/agentc.json" ]; then
  mkdir -p /home/aceuser/ace-server/config/iibswitch/agentc
  cp /home/aceuser/initial-config/agent/agentc.json /home/aceuser/ace-server/config/iibswitch/agentc/agentc.json
fi

log "agent configuration complete"
