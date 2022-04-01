ARG  FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.4.0-r1
FROM ${FROMIMAGE}

USER root

# Required for the setdbparms script to run
RUN microdnf update && microdnf install python3 && microdnf clean all \
   && ln -s /usr/bin/python3 /usr/local/bin/python

COPY server.conf.yaml /home/aceuser/ace-server/

RUN mkdir -p /home/aceuser/initial-config/setdbparms/
COPY ace_config_*  /home/aceuser/initial-config/
RUN chmod -R ugo+rwx /home/aceuser/

USER 1001