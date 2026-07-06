#!/usr/bin/env bash
#------------------------------------------------------------------------------
#
#     IBM Confidential
#     PIDs 5900-AN0, 5724-J05, 5737-E34, 5737-B81, 5900-AV7
#     Copyright IBM Corp. 2026
#
#------------------------------------------------------------------------------

# Enable strict error handling
set -euo pipefail

# =============================================================================
# === Default Values                                                        ===
# =============================================================================
readonly ACE_BUILD_DEFAULT="ace-13.0.8.0.tar.gz"
readonly ACE_FOLDER_DEFAULT="13.0.8.0"
readonly TAG_VERSION_DEFAULT="13.0.8.0-r1"

# =============================================================================
# === Image Naming and Tagging                                              ===
# =============================================================================
# Container image naming and tag construction. These variables compose the
# final image names and tags used during build and push operations.
export IMAGE_REPOSITORY="ace"
export IMAGE_NAME="${IMAGE_REGISTRY}/${IMAGE_NAMESPACE}/${IMAGE_REPOSITORY}"

# Always use our own TAG_VERSION_DEFAULT, ignoring any TAG_VERSION that may have
# been set by firefly-software-build-scripts/image-tag/cd/tag before this file
# was sourced. ot4i-ace-docker has independent versioning.
export TAG_VERSION="${TAG_VERSION_DEFAULT}"

# =============================================================================
# === Application Metadata                                                  ===
# =============================================================================
# Core application information and build metadata.
export GIT_COMMIT="$(git rev-parse HEAD)"

# =============================================================================
# === Build Parameters                                                      ===
# =============================================================================
ace_build_value="$(get_env ACE_BUILD)"
export ACE_BUILD="${ace_build_value:-$ACE_BUILD_DEFAULT}"

ace_folder_value="$(get_env ACE_FOLDER)"
export ACE_FOLDER="${ace_folder_value:-$ACE_FOLDER_DEFAULT}"

# =============================================================================
# === Image build                                                           ===
# =============================================================================

build_image() {
    log_notice "Preparing to build ${IMAGE_REPOSITORY}"

    log_notice "Logging in to ${IMAGE_REGISTRY}"
    printf '%s' "${ICR_PASS}" | podman login --username iamapikey --password-stdin "${IMAGE_REGISTRY}"

    log_notice "Changing to app directory"
    cd "${WORKSPACE}/$(load_repo app-repo path)" || {
        log_error "Failed to change to app directory"
        return 1
    }
    log_notice "Branch name is ${BRANCH_NAME}"

    # Set image reference and download URL
    local image_reference="${IMAGE_NAME}:${TAG_VERSION}-${ARCHITECTURE}"
    local download_url="https://na.artifactory.swg-devops.com:443/artifactory/appconnect-iiboc/prereqs/builds/${ACE_FOLDER}/${PLATFORM}/${ACE_BUILD}"

    log_notice "Building image with tag: ${image_reference}"

    podman build \
        --no-cache \
        --file ./Dockerfile \
        --tag "${image_reference}" \
        --build-arg DOWNLOAD_URL="${download_url}" \
        --build-arg PASSWORD="${ARTIFACTORY_ACCESS_TOKEN}" \
        --build-arg USERNAME="${ARTIFACTORY_USER}" \
        . | tee docker-prod-build-results.log

    # Check if the build succeeded by examining the exit code of podman build,
    # not tee, which is why we use PIPESTATUS[0])
    if [[ "${PIPESTATUS[0]}" -ne 0 ]]; then
        log_error "Image failed to build!"
        return 1
    fi

    log_notice "Image built successfully"

    # Push image to registry with retry logic
    log_notice "Pushing image ${image_reference}"
    local push_ok=false

    for i in {1..5}; do
        skopeo copy \
            --preserve-digests \
            --dest-creds "iamapikey:${ICR_PASS}" \
            "containers-storage:${image_reference}" \
            "docker://${image_reference}" && {
            push_ok=true
            break
        } || {
            log_warning "Push attempt ${i} failed; retrying in 15s..."
            sleep 15
        }
    done

    if [[ "${push_ok}" != true ]]; then
        log_error "Push failed after 5 attempts"
        return 1
    else
        log_notice "Image pushed successfully"
    fi

    # Tag as "latest" if building main branch
    if [[ "${BRANCH_NAME}" == "main" ]] && [[ "${BUILD_AMD64}" == "true" ]] && [[ "${BUILD_PPC64LE}" == "true" ]] && [[ "${BUILD_S390X}" == "true" ]]; then
        log_notice "Building main branch with all 3 platforms (amd64, ppc64le, s390x) so tagging the image as \"latest\""
        local latest_image_reference="${IMAGE_NAME}:latest-${ARCHITECTURE}"
        local latest_ok=false

        for i in {1..5}; do
            skopeo copy \
                --preserve-digests \
                --dest-creds "iamapikey:${ICR_PASS}" \
                "containers-storage:${image_reference}" \
                "docker://${latest_image_reference}" && {
                latest_ok=true
                break
            } || {
                log_warning "\"latest\" tag push attempt ${i} failed; retrying in 15s..."
                sleep 15
            }
        done

        if [[ "${latest_ok}" != true ]]; then
            log_error "Failed to tag ${latest_image_reference} \"latest\" after 5 attempts"
            return 1
        else
            log_notice "Image successfully tagged \"latest\": ${latest_image_reference}"
        fi
    else
        log_notice "Not tagging image as \"latest\" because branch is not main or all 3 platforms (amd64, ppc64le, s390x) have not been selected"
    fi
}
