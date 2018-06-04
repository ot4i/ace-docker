# © Copyright IBM Corporation 2018.
#
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html

FROM ubuntu:16.04

LABEL "maintainer"="Dan Robinson <dan.robinson@uk.ibm.com>" \
      "product.id"="447aefb5fd1342d5b893f3934dfded73" \
      "product.name"="IBM App Connect Enterprise" \
      "product.version"="11.0.0.0"

WORKDIR /opt/ibm

# Install ACE V11 Developer Edition
RUN apt update && apt -y install --no-install-recommends curl rsyslog sudo \
  && curl http://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/integration/ACE-LINUX64-DEVELOPER.tar.gz \
   | tar xz --exclude ace-11.0.0.0/tools --directory /opt/ibm/ \
  && /opt/ibm/ace-11.0.0.0/ace make registry global accept license silently \
  && apt remove -y curl \
  && rm -rf /var/lib/apt/lists/*

# Configure the system
RUN echo "ACE_11:" > /etc/debian_chroot \
  && touch /var/log/syslog \
  && chown syslog:adm /var/log/syslog \
# Increase security
  && sed -i 's/sha512/sha512 minlen=8/'  /etc/pam.d/common-password \
  && sed -i 's/PASS_MIN_DAYS\t0/PASS_MIN_DAYS\t1/'  /etc/login.defs \
  && sed -i 's/PASS_MAX_DAYS\t99999/PASS_MAX_DAYS\t90/' /etc/login.defs

# Copy in script files
COPY *.sh /usr/local/bin/

# Create a user to run as, create the ace workdir, and chmod script files
RUN useradd --create-home --home-dir /home/aceuser -G mqbrkrs,sudo aceuser \
  && sed -e 's/^%sudo	.*/%sudo	ALL=NOPASSWD:ALL/g' -i /etc/sudoers \
  && su - aceuser -c '. /opt/ibm/ace-11.0.0.0/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server' \
  && chmod 755 /usr/local/bin/*

# Set BASH_ENV to source mqsiprofile when using docker exec bash -c
ENV BASH_ENV=/usr/local/bin/ace_env.sh

# Expose ports
EXPOSE 7800 7600

USER aceuser

WORKDIR /home/aceuser

# Set entrypoint to run management script
CMD ["/bin/bash", "-c", "/usr/local/bin/ace_license_check.sh && IntegrationServer -w /home/aceuser/ace-server --console-log"]
