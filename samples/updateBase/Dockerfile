ARG  FROMIMAGE=cp.icr.io/cp/appc/ace:13.0.1.0-r1
FROM ${FROMIMAGE}

USER root

RUN microdnf update && microdnf clean all

USER 1001
