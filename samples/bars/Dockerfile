ARG  FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.4.0-r1
FROM ${FROMIMAGE}

USER root

COPY *.bar /tmp
RUN export LICENSE=accept \
    && . /opt/ibm/ace-12/server/bin/mqsiprofile \
    && set -x && for FILE in /tmp/*.bar; do ibmint deploy --input-bar-file "$FILE" --output-work-directory /home/aceuser/ace-server/ 2>&1 >  /tmp/deploys; done \
    && chmod -R ugo+rwx /home/aceuser/

USER 1001