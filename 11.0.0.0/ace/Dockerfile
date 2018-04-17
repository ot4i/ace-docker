# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v1.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

FROM ubuntu:16.04

LABEL maintainer "Ben Thompson <bthomps@uk.ibm.com>, Trevor Dolby <tdolby@uk.ibm.com>, Dan Robinson <dan.robinson@uk.ibm.com>, Sam Rogers <srogers@uk.ibm.com>"

LABEL "ProductID"="447aefb5fd1342d5b893f3934dfded73" \
      "ProductName"="IBM App Connect Enterprise" \
      "ProductVersion"="11.0.0.0"

# Install curl
RUN apt-get update && \
    apt-get install -y curl rsyslog sudo && \
rm -rf /var/lib/apt/lists/*

# Install the product
RUN cd /opt && \
    mkdir /opt/ibm && \
    curl http://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/integration/11.0.0.0-ACE-LINUX64-DEVELOPER.tar.gz \
    | tar zx --exclude ace-11.0.0.0/tools --directory /opt/ibm && \
    /opt/ibm/ace-11.0.0.0/ace make registry global accept license silently

# Deploy a BAR file; Use bash to make the profile work
COPY Transformation_Map.bar /tmp
RUN cd /root && bash -c '. /opt/ace-11.0.0.0/server/bin/mqsiprofile && mqsicreateworkdir /root/ace-server && mqsibar -w /root/ace-server -a /tmp/Mandalay.bar -c'

EXPOSE 7600 7800

# Default command to run
CMD ["/bin/bash", "-c", ". /opt/ace-11.0.0.0/server/bin/mqsiprofile && IntegrationServer -w /root/ace-server --console-log --admin-rest-api 7600"]
