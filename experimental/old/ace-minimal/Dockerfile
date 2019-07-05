FROM ubuntu:18.04

# Prevent errors about having no terminal when using apt-get
ENV DEBIAN_FRONTEND noninteractive


RUN apt-get update && apt-get install -y --no-install-recommends curl

# Set up exclusions
COPY excludes*txt /tmp/

RUN cat /tmp/excludes-base.txt >> /tmp/all-excludes.txt && \
# Adapters are optional
    cat /tmp/excludes-adapters.txt >> /tmp/all-excludes.txt && \
# CICS is optional
    cat /tmp/excludes-cics.txt >> /tmp/all-excludes.txt && \
# FTE is optional
    cat /tmp/excludes-fte.txt >> /tmp/all-excludes.txt && \
# NPM install support is optional
    cat /tmp/excludes-npm.txt >> /tmp/all-excludes.txt && \
# ODBC drivers are optional
    cat /tmp/excludes-odbc.txt >> /tmp/all-excludes.txt && \
# REST Web UI support optional; use --admin-rest-api -1 on server start line
    cat /tmp/excludes-webui.txt >> /tmp/all-excludes.txt && \
# XSLT nodes are optional
    cat /tmp/excludes-xmlt.txt >> /tmp/all-excludes.txt && \
# Toolkit is optional
    cat /tmp/excludes-tools.txt >> /tmp/all-excludes.txt

# Install ACE v11.0.0.4 and accept the license
RUN apt-get install -y zip binutils && \
    mkdir /opt/ibm && \
    curl http://kenya.hursley.uk.ibm.com:52367/ace-11.0.16836.5.tar.gz \
    | tar zx --exclude-from=/tmp/all-excludes.txt --directory /opt/ibm && \
    mv /opt/ibm/ace-11.0.16836.5 /opt/ibm/ace-11 && \
    find /opt/ibm -name "*.so*" -exec strip {} ";" && \
    find /opt/ibm -name "*.wrk" -exec strip {} ";" && \
    find /opt/ibm -name "*.lil" -exec strip {} ";" && \
    ( strip /opt/ibm/ace-11/server/bin/* 2>/dev/null || /bin/true ) && \
    zip -d /opt/ibm/ace-11/common/classes/IntegrationAPI.jar BIPmsgs_de.properties BIPmsgs_es.properties BIPmsgs_fr.properties BIPmsgs_it.properties BIPmsgs_ja.properties BIPmsgs_ko.properties BIPmsgs_pl.properties BIPmsgs_pt_BR.properties BIPmsgs_ru.properties BIPmsgs_tr.properties BIPmsgs_zh.properties BIPmsgs_zh_HK.properties BIPmsgs_zh_TW.properties && \
    apt-get remove -y zip binutils binutils-common binutils-x86-64-linux-gnu libbinutils && \
    /opt/ibm/ace-11/ace make registry global accept license silently
 
# The product is now installed, and everything below is ease-of-use enhancements.


# Copy in script files - this is used to prevent the "mqsiprofile 
# repetition disallowed" message from being printed on shell start.
COPY ace_env.sh /usr/local/bin/

# Configure the system and increase security
RUN echo "ACE_11:" > /etc/debian_chroot \
  && sed -i 's/sha512/sha512 minlen=8/' /etc/pam.d/common-password \
  && sed -i 's/PASS_MIN_DAYS\t0/PASS_MIN_DAYS\t1/' /etc/login.defs \
  && sed -i 's/PASS_MAX_DAYS\t99999/PASS_MAX_DAYS\t90/' /etc/login.defs \
  && echo ". /usr/local/bin/ace_env.sh" >> /root/.bashrc

# Set BASH_ENV to source mqsiprofile when using docker exec bash -c
ENV BASH_ENV=/usr/local/bin/ace_env.sh

# Create a user to run as, create the ace workdir, and chmod script files
RUN useradd --create-home --home-dir /home/aceuser --shell /bin/bash -G mqbrkrs,sudo aceuser \
  && su - aceuser -c '. /opt/ibm/ace-11/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server' \
  && echo ". /usr/local/bin/ace_env.sh" >> /home/aceuser/.bashrc \
  && chmod 755 /usr/local/bin/*


USER aceuser
ENTRYPOINT ["bash"]
