#!/bin/bash
# © Copyright IBM Corporation 2015.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v1.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v10.html

if [ -z "$MQSI_VERSION" ]; then
  echo "Sourcing profile"
  source /opt/ibm/ace-11.0.0.0/server/bin/mqsiprofile
fi
