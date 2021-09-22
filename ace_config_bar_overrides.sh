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

if ls /home/aceuser/initial-config/bar_overrides/*.properties >/dev/null 2>&1; then
  for propertyFile in /home/aceuser/initial-config/bar_overrides/*.properties
  do
    for bar in /home/aceuser/initial-config/bars/*.bar
    do
      mqsiapplybaroverride -b $bar -p $propertyFile -r 
      echo $propertyFile >> /home/aceuser/initial-config/bar_overrides/logs.txt
    done
  done
fi