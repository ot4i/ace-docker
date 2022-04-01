ARG  FROMIMAGE=cp.icr.io/cp/appc/ace:12.0.4.0-r1
FROM ${FROMIMAGE}

USER root

RUN RUN microdnf update && microdnf clean all

USER 1001