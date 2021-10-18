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

# Enable TLS on both MQ and DB2
if [ -d /opt/mqm/gskit8/lib64 ]; then
  export LD_LIBRARY_PATH=/opt/mqm/gskit8/lib64:$LD_LIBRARY_PATH
fi

if [ -s /home/aceuser/ace-server/odbc.ini ]; then
  export ODBCINI=/home/aceuser/ace-server/odbc.ini
fi

# For customers running pod environment
if [ -s /home/aceuser/generic/odbcinst.ini ]; then
  export ODBCSYSINI=/home/aceuser/generic
fi

# For customers running ace in docker themselves
if [ -s /home/aceuser/ace-server/extensions/odbcinst.ini ]; then
  export ODBCSYSINI=/home/aceuser/ace-server/extensions
fi

# We need to keep the kubernetes port overrides as customers could be running ace in docker themselves
# but we need to allow the ports to be overwritten in the pod environment if set by the operator
if ! [[ -z "${KUBERNETES_PORT}" ]] && ! [[ -z "${SERVICE_NAME}" ]] && ! [[ -z "${MQSI_OVERRIDE_HTTP_PORT}" ]] && ! [[ -z "${MQSI_OVERRIDE_HTTPS_PORT}" ]] ; then
  . /home/aceuser/portOverrides
fi

exec IntegrationServer $*
