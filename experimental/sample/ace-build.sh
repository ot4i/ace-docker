#!/bin/bash

# Designed for openshift S2I builds; the target directory contains the data passed to the runtime container

echo "In ace-build.sh with target directory of $1"

. /opt/ibm/ace-11/server/bin/mqsiprofile

mqsicreateworkdir $1/ace-server

mqsibar -c -w $1/ace-server -a aceFunction.bar
