ARG  FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.4.0-r1
ARG  IFIX
FROM ${FROMIMAGE}

ADD ifix/<iFixName>.tar.gz /home/aceuser/ifix

USER root

RUN  cd /home/aceuser/ifix \
     && ./mqsifixinst.sh /opt/ibm/ace-12 install <iFixName> \
     && cd /home/aceuser \
     && rm -rf /home/aceuser/fix

USER 1001
