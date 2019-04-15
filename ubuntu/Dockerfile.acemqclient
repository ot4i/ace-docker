ARG BASE_IMAGE=ibmcom/ace
FROM $BASE_IMAGE

# The URL to download the MQ installer from in tar.gz format
ARG MQ_URL=https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqadv/mqadv_dev911_ubuntu_x86-64.tar.gz

USER root
COPY ubuntu/installMQClient.sh /tmp/mq/
RUN /tmp/mq/installMQClient.sh 

USER aceuser
