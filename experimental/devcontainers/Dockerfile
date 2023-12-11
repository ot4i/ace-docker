FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

MAINTAINER Trevor Dolby <trevor.dolby@ibm.com> (@trevor-dolby-at-ibm-com)

# Build:
#
# docker build -t ace-devcontainer:12.0.10.0 -f Dockerfile .

# Later versions from the same site, or else via the Developer edition download site linked from
# https://www.ibm.com/docs/en/app-connect/12.0?topic=enterprise-download-ace-developer-edition-get-started
ARG DOWNLOAD_URL=https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/integration/12.0.10.0-ACE-LINUX64-DEVELOPER.tar.gz
# Can be set higher if needed
ARG DOWNLOAD_CONNECTION_COUNT=1

# Prevent errors about having no terminal when using apt-get
ENV DEBIAN_FRONTEND noninteractive

# Set up exclusions
COPY excludes*txt /tmp/

RUN cat /tmp/excludes-base.txt >> /tmp/all-excludes.txt && \
# Java11 is not required
    cat /tmp/excludes-java11.txt >> /tmp/all-excludes.txt && \
# ToolkitDiscovery connectors are optional
    cat /tmp/excludes-nodejs-all.txt >> /tmp/all-excludes.txt && \
# Toolkit is optional
    cat /tmp/excludes-tools.txt >> /tmp/all-excludes.txt && \
# WSRR nodes are optional
    cat /tmp/excludes-wsrr.txt >> /tmp/all-excludes.txt

# Install ACE and accept the license
RUN apt-get update && apt-get install -y aria2 zip binutils && \
    mkdir /opt/ibm && \
    cd /tmp && \
    echo Downloading package ${DOWNLOAD_URL} && \
    aria2c -s ${DOWNLOAD_CONNECTION_COUNT} -j ${DOWNLOAD_CONNECTION_COUNT} -x ${DOWNLOAD_CONNECTION_COUNT} "${DOWNLOAD_URL}" && \
    tar -zxf *12.0*tar.gz --exclude-from=/tmp/all-excludes.txt --directory /opt/ibm && \
    rm -f *12.0*tar.gz && \
    mv /opt/ibm/ace-12.* /opt/ibm/ace-12 && \
    find /opt/ibm -name "*.so*" -exec strip {} ";" && \
    find /opt/ibm -name "*.wrk" -exec strip {} ";" && \
    find /opt/ibm -name "*.lil" -exec strip {} ";" && \
    ( strip /opt/ibm/ace-12/server/bin/* 2>/dev/null || /bin/true ) && \
    zip -d /opt/ibm/ace-12/common/classes/IntegrationAPI.jar BIPmsgs_de.properties BIPmsgs_es.properties BIPmsgs_fr.properties BIPmsgs_it.properties BIPmsgs_ja.properties BIPmsgs_ko.properties BIPmsgs_pl.properties BIPmsgs_pt_BR.properties BIPmsgs_ru.properties BIPmsgs_tr.properties BIPmsgs_zh.properties BIPmsgs_zh_HK.properties BIPmsgs_zh_TW.properties && \
    if [ $(uname -m) = x86_64 ]; then apt-get remove -y zip binutils binutils-common libbinutils binutils-x86-64-linux-gnu; else apt-get remove -y zip binutils binutils-common libbinutils; fi && \
    rm -rf /var/lib/apt/lists/*  && \
    /opt/ibm/ace-12/ace make registry global accept license deferred

# Source profile automatically
RUN usermod -a -G mqbrkrs vscode \
  && echo ". /opt/ibm/ace-12/server/bin/mqsiprofile" >> /home/vscode/.bashrc \ 
  && echo ". /opt/ibm/ace-12/server/bin/mqsiprofile" >> /home/vscode/.profile

# Install Maven for development use.
#
# The standard Maven packages on most distros bring a lot of extra packages with
# them so we install Maven directly. 
RUN cd /opt && \
    curl -k https://archive.apache.org/dist/maven/maven-3/3.8.4/binaries/apache-maven-3.8.4-bin.tar.gz | tar -xzf - && \
    ln -s /opt/apache-maven-3.8.4/bin/mvn /usr/local/bin/mvn
