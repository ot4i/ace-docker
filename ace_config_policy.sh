#!/bin/bash

# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
source ${SCRIPT_DIR}/ace_config_logging.sh

log "Handling policy configuration"

mkdir /home/aceuser/ace-server/overrides/DefaultPolicies

if ls /home/aceuser/initial-config/policy/*.policyxml >/dev/null 2>&1; then
  for policyfile in `ls /home/aceuser/initial-config/policy/*.policyxml`; do
    if [ -s "${policyfile}" ]; then
      cp "${policyfile}" /home/aceuser/ace-server/overrides/DefaultPolicies/.
    fi
  done
fi

if [ -s "/home/aceuser/initial-config/policy/policy.descriptor" ]; then
  cp /home/aceuser/initial-config/policy/policy.descriptor /home/aceuser/ace-server/overrides/DefaultPolicies/policy.descriptor
fi

log "Policy configuration complete"
