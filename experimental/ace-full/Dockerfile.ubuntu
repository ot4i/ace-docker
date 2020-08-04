FROM ubuntu:20.04
MAINTAINER Trevor Dolby <tdolby@uk.ibm.com> (@tdolby)

# docker build -t ace-full:11.0.0.9-ubuntu -f Dockerfile.ubuntu .

ARG DOWNLOAD_URL=http://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/integration/11.0.0.9-ACE-LINUX64-DEVELOPER.tar.gz
ARG PRODUCT_LABEL=ace-11.0.0.9

# Prevent errors about having no terminal when using apt-get
ENV DEBIAN_FRONTEND noninteractive

# Install ACE v11.0.0.9 and accept the license
RUN apt-get update && apt-get install -y --no-install-recommends curl && \
    mkdir /opt/ibm && echo Downloading package ${DOWNLOAD_URL} && \
    curl ${DOWNLOAD_URL} | tar zx --directory /opt/ibm && \
    mv /opt/ibm/${PRODUCT_LABEL} /opt/ibm/ace-11 && \
    /opt/ibm/ace-11/ace make registry global accept license deferred

# Configure the system
RUN echo "ACE_11:" > /etc/debian_chroot \
  && echo ". /opt/ibm/ace-11/server/bin/mqsiprofile" >> /root/.bashrc

# mqsicreatebar prereqs; need to run "Xvfb -ac :99 &" and "export DISPLAY=:99"
RUN apt-get -y install libgtk2.0-0 libxtst6 xvfb

# Set BASH_ENV to source mqsiprofile when using docker exec bash -c
ENV BASH_ENV=/opt/ibm/ace-11/server/bin/mqsiprofile

# Create a user to run as, create the ace workdir, and chmod script files
RUN useradd --create-home --home-dir /home/aceuser --shell /bin/bash -G mqbrkrs,sudo aceuser \
  && su - aceuser -c "export LICENSE=accept && . /opt/ibm/ace-11/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server" \
  && echo ". /opt/ibm/ace-11/server/bin/mqsiprofile" >> /home/aceuser/.bashrc

USER aceuser
ENTRYPOINT ["bash"]
