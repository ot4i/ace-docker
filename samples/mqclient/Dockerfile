ARG  FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.10.0-r1
FROM ${FROMIMAGE}

USER root

# The MQ packages to install - see install-mq.sh for default value
ARG MQ_URL
ARG MQ_URL_USER
ARG MQ_URL_PASS
ARG MQ_PACKAGES="MQSeriesRuntime*.rpm MQSeriesJava*.rpm MQSeriesJRE*.rpm MQSeriesGSKit*.rpm MQSeriesClient*.rpm"
ARG INSTALL_JRE=0

ARG MQM_UID=888

COPY install-mq.sh /usr/local/bin/
COPY install-mq-client-prereqs.sh /usr/local/bin/
# Install MQ.  To avoid a "text file busy" error here, we sleep before installing.
RUN chmod u+x /usr/local/bin/install-*.sh \
  && sleep 1 \
  && install-mq-client-prereqs.sh $MQM_UID \
  && install-mq.sh $MQM_UID \
  && chown -R 1001:root /opt/mqm/*  \
  && chown 1001:root /usr/local/bin/*mq* \
  && mkdir -p /var/mqm/data \
  && chown -R 1001:root /var/mqm \
  && chmod -R 777 /var/mqm

ENV MQCERTLABL=aceclient

USER 1001