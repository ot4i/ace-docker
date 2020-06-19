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

log "Handling odbcini configuration"

if [ -s "/home/aceuser/initial-config/odbcini/odbc.ini" ]; then
  ODBCINI=/home/aceuser/ace-server/odbc.ini
  cp /home/aceuser/initial-config/odbcini/odbc.ini ${ODBCINI}
fi

log "Odbcini configuration complete"
