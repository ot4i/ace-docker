#!/usr/bin/env bash
#------------------------------------------------------------------------------
#
#     IBM Confidential
#     PIDs 5900-AN0, 5724-J05, 5737-E34, 5737-B81, 5900-AV7
#     Copyright IBM Corp. 2026
#
#------------------------------------------------------------------------------

# Validate required environment variables
if [ -z "${ONE_PIPELINE_PATH}" ] || [ -z "${WORKSPACE}" ]; then
  echo "ERROR: Required environment variables (ONE_PIPELINE_PATH, WORKSPACE) not set" >&2
  exit 1
fi

source "${ONE_PIPELINE_PATH}/internal/tools/logging"

GIT_TOKEN=$(get_env git-token)
if [ -z "${GIT_TOKEN}" ]; then
  error "Failed to retrieve git-token"
  exit 1
fi

notice "Bootstrapping firefly-software-build-scripts"
if ! git clone "https://${GIT_TOKEN}@github.ibm.com/Cloud-Integration/firefly-software-build-scripts.git" "${WORKSPACE}/firefly-software-build-scripts"; then
  error "Failed to clone firefly-software-build-scripts repository"
  exit 1
fi

repo_branch="${1:-main}"
if [ "${repo_branch}" != "main" ]; then
  notice "Checking out branch ${repo_branch} of firefly-software-build-scripts"
  pushd "${WORKSPACE}/firefly-software-build-scripts"
  if ! git checkout "${repo_branch}"; then
    error "Failed to checkout branch ${repo_branch}"
    exit 1
  fi
  popd
fi

source "${WORKSPACE}/firefly-software-build-scripts/scripts/acecc-pipeline-scripts/bootstrap.sh"
