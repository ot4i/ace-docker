FROM registry.access.redhat.com/rhel7

ENV SUMMARY="Integration Server for App Connect Enterprise on ICP" \
    DESCRIPTION="Integration Server for App Connect Enterprise on ICP" \
    PRODNAME="AppConnectEnterprise" \
    COMPNAME="IntegrationServer"

LABEL summary="$SUMMARY" \
      description="$DESCRIPTION" \
      io.k8s.description="$DESCRIPTION" \
      io.k8s.display-name="Integration Server for App Connect Enterprise on ICP" \
      io.openshift.tags="$PRODNAME,$COMPNAME" \
      com.redhat.component="$PRODNAME-$COMPNAME" \
      name="$PRODNAME/$COMPNAME" \
      vendor="IBM" \
      version="11.0.0.3" \
      release="1" \
      license="IBM" \
      maintainer="Hybrid Integration Platform Cloud" \
      io.openshift.expose-services="" \
      usage=""

# Increase security
RUN sed -i 's/# minlen = 9/minlen = 9/'  /etc/security/pwquality.conf && \
    sed -i 's/PASS_MIN_DAYS\t0/PASS_MIN_DAYS\t1/'  /etc/login.defs && \
    sed -i 's/PASS_MAX_DAYS\t99999/PASS_MAX_DAYS\t90/'  /etc/login.defs

# Add required license as text file in Liceses directory (GPL, MIT, APACHE, Partner End User Agreement, etc)
RUN mkdir /licenses
COPY LICENSE /licenses/licensing.txt

WORKDIR /opt/ibm

# Install ACE V11
RUN  yum update -y && \
     yum upgrade -y && \
     yum install sudo openssl -y && \
     rm -rf /var/lib/apt/lists/*

ADD ./rhel/ace-11  /opt/ibm/ace-11
RUN /opt/ibm/ace-11/ace make registry global accept license silently

# Copy in PID1 process
COPY ./rhel/runaceserver /usr/local/bin/
COPY ./rhel/chkaceready /usr/local/bin/
COPY ./rhel/chkacehealthy /usr/local/bin/

# Configure the system and Increase security
RUN echo "ACE_11:" > /etc/debian_chroot \
  && sed -i 's/# minlen = 9/minlen = 8/' /etc/security/pwquality.conf \
  && sed -i 's/PASS_MIN_DAYS\t0/PASS_MIN_DAYS\t1/' /etc/login.defs \
  && sed -i 's/PASS_MAX_DAYS\t99999/PASS_MAX_DAYS\t90/' /etc/login.defs

# Copy in script files
COPY *.sh /usr/local/bin/

# Create a user to run as, create the ace workdir, and chmod script files
RUN useradd -d /home/aceuser -G mqbrkrs,wheel aceuser \
  && sed -e 's/^%sudo   .*/%sudo        ALL=NOPASSWD:ALL/g' -i /etc/sudoers \
  && su - aceuser -c '. /opt/ibm/ace-11/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server' \
  && chmod 755 /usr/local/bin/*

# Set BASH_ENV to source mqsiprofile when using docker exec bash -c
ENV BASH_ENV=/usr/local/bin/ace_env.sh

# Expose ports.  7600, 7800, 7843 for ACE; 9483 for ACE metrics
EXPOSE 7600 7800 7843 9483

USER aceuser

WORKDIR /home/aceuser
RUN mkdir /home/aceuser/initial-config && chown aceuser:aceuser /home/aceuser/initial-config

ENV LOG_FORMAT=basic

# Set entrypoint to run management script
ENTRYPOINT ["runaceserver"]
