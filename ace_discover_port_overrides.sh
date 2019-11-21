#!/bin/bash -ex
if ! [[ -z "${ACE_HTTP_ROUTE_NAME}" ]] && [[ "${ACE_ENDPOINT_TYPE}" = 'http' ]]; then
  echo "export MQSI_OVERRIDE_HTTP_PORT=80" >> /home/aceuser/portOverrides
  echo "export MQSI_OVERRIDE_HOSTNAME=$(kubectl get route ${ACE_HTTP_ROUTE_NAME} -o jsonpath=\"{.status.ingress[0].host}\")" >> /home/aceuser/portOverrides
elif ! [[ -z "${ACE_HTTPS_ROUTE_NAME}" ]] && [[ "${ACE_ENDPOINT_TYPE}" = 'https' ]]; then
  echo "export MQSI_OVERRIDE_HTTPS_PORT=443" >> /home/aceuser/portOverrides
  echo "export MQSI_OVERRIDE_HOSTNAME=$(kubectl get route ${ACE_HTTPS_ROUTE_NAME} -o jsonpath=\"{.status.ingress[0].host}\")" >> /home/aceuser/portOverrides
elif ! [[ -z "${KUBERNETES_PORT}" ]] && ! [[ -z "${SERVICE_NAME}" ]] ; then
  echo "export MQSI_OVERRIDE_HTTP_PORT=$(kubectl get svc ${SERVICE_NAME} -o jsonpath=\"{.spec.ports[1].nodePort}\")" >> /home/aceuser/portOverrides
  echo "export MQSI_OVERRIDE_HTTPS_PORT=$(kubectl get svc ${SERVICE_NAME} -o jsonpath=\"{.spec.ports[2].nodePort}\")" >> /home/aceuser/portOverrides
fi