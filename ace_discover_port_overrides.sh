#!/bin/bash -ex

if ! [[ -z "${KUBERNETES_PORT}" ]] && ! [[ -z "${SERVICE_NAME}" ]] ; then
  echo "export MQSI_OVERRIDE_HTTP_PORT=$(kubectl get svc ${SERVICE_NAME} -o jsonpath=\"{.spec.ports[1].nodePort}\")" >> /home/aceuser/portOverrides
  echo "export MQSI_OVERRIDE_HTTPS_PORT=$(kubectl get svc ${SERVICE_NAME} -o jsonpath=\"{.spec.ports[2].nodePort}\")" >> /home/aceuser/portOverrides
fi