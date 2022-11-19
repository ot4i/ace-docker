FROM mcr.microsoft.com/devcontainers/base:ubuntu-22.04

MAINTAINER Trevor Dolby <trevor.dolby@ibm.com> (@trevor-dolby-at-ibm-com)

# Build and run:
#
# docker build -t ace-devcontainer:12.0.4.0-ubuntu -f Dockerfile .

ARG USERNAME
ARG PASSWORD
ARG DOWNLOAD_URL=http://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/integration/12.0.4.0-ACE-LINUX64-DEVELOPER.tar.gz
ARG PRODUCT_LABEL=ace-12.0.4.0

# Prevent errors about having no terminal when using apt-get
ENV DEBIAN_FRONTEND noninteractive

# Exclude tools, languages, WSRR, transformation advisor, etc
COPY excludes.txt /tmp/

# Install ACE v12.0.4.0 and accept the license
RUN apt-get update && apt-get install -y --no-install-recommends curl && \
    mkdir /opt/ibm && echo Downloading package ${DOWNLOAD_URL} && \
    if [ -z $USERNAME ]; then curl ${DOWNLOAD_URL}; else curl -u ${USERNAME}:{PASSWORD} ${DOWNLOAD_URL}; fi | \
    tar zx --exclude-from=/tmp/excludes.txt --directory /opt/ibm && \
    mv /opt/ibm/${PRODUCT_LABEL} /opt/ibm/ace-12 && \
    rm -rf /var/lib/apt/lists/*  && \
    /opt/ibm/ace-12/ace make registry global accept license deferred

# Create a user to run as, create the ace workdir, and chmod script files
RUN usermod -a -G mqbrkrs vscode \
  && echo ". /opt/ibm/ace-12/server/bin/mqsiprofile" >> /home/vscode/.bashrc

# The standard Maven packages on most distros bring a lot of extra packages with
# them so we install Maven directly. 
RUN cd /opt && \
    curl -k https://archive.apache.org/dist/maven/maven-3/3.8.4/binaries/apache-maven-3.8.4-bin.tar.gz | tar -xzf - && \
    ln -s /opt/apache-maven-3.8.4/bin/mvn /usr/local/bin/mvn

