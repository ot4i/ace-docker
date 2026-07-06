# Build and run:
#
# docker build -t ace:13.0.8.0 -f Dockerfile .
# docker run -e LICENSE=accept -p 7600:7600 -p 7800:7800 --rm -ti ace:13.0.8.0
#
# Can also mount a volume for the work directory:
#
# docker run -e LICENSE=accept -v /what/ever/dir:/home/aceuser/ace-server -p 7600:7600 -p 7800:7800 --rm -ti ace:13.0.8.0
#
# This might require a local directory with the right permissions, or changing the userid further down . . .

FROM registry.access.redhat.com/ubi9/ubi-minimal AS builder

RUN microdnf update -y && microdnf install -y util-linux tar && microdnf clean all

ARG USERNAME
ARG PASSWORD
# Download and reference the ACE-LINUX64-DEVELOPER.tar.gz from here https://www.ibm.com/resources/mrs/assets?source=swg-wmbfd eg.
ARG DOWNLOAD_URL=<Your downloaded location>/13.0.8.0-ACE-LINUX64-DEVELOPER.tar.gz

# Note: We skip extracting any files from the ACE tar that we don't want, but we have to be explicit on
# the "ace" folder name and not use "--exclude ace-*.*.*.*/foo" as "*" matches across directory boundaries
WORKDIR /opt/ibm
RUN mkdir ace-13 \
    && if [ -z $USERNAME ]; then curl ${DOWNLOAD_URL} -o ace-13.tar.gz; else curl -u "${USERNAME}:${PASSWORD}" ${DOWNLOAD_URL} -o ace-13.tar.gz; fi \
    && TAR_ROOT_DIR=$(tar tf ace-13.tar.gz | head -1 | cut -d'/' -f1) \
    && tar -xzf ace-13.tar.gz --absolute-names \
        --exclude ${TAR_ROOT_DIR}/tools \
        --exclude ${TAR_ROOT_DIR}/server/tools/ibm-dfdl-java.zip \
        --exclude ${TAR_ROOT_DIR}/server/tools/proxyservlet.war \
        --exclude ${TAR_ROOT_DIR}/server/bin/TADataCollector.sh \
        --exclude ${TAR_ROOT_DIR}/server/transformationAdvisor/ta-plugin-ace.jar \
        --exclude ${TAR_ROOT_DIR}/server/nodejs_partial \
        --strip-components 1 \
        --directory /opt/ibm/ace-13 \
    && rm ace-13.tar.gz

FROM registry.access.redhat.com/ubi9/ubi-minimal

# Force reinstall tzdata package to get zoneinfo files
RUN microdnf update -y && microdnf install -y findutils util-linux which tar && microdnf reinstall -y tzdata && microdnf clean all

# Install ACE v13.0.8.0 and accept the license
COPY --from=builder /opt/ibm/ace-13 /opt/ibm/ace-13
RUN export MQSI_USE_CALL_HOME_TEST_SOURCE='true';  /opt/ibm/ace-13/ace accept license --silently --make-registry-global --company-name REPLACE_ME_COMPANY_NAME  \
    && useradd --uid 1001 --create-home --home-dir /home/aceuser --shell /bin/bash -G mqbrkrs aceuser \
    && su - aceuser -c "export LICENSE=accept && . /opt/ibm/ace-13/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server" \
    && chmod -R 777 /home/aceuser /var/mqsi \
    && echo ". /opt/ibm/ace-13/server/bin/mqsiprofile" >> /home/aceuser/.bashrc

# Add required license as text file in Liceses directory (GPL, MIT, APACHE, Partner End User Agreement, etc)
COPY /licenses/ /licenses/

# aceuser
USER 1001

# Expose ports.  7600, 7800, 7843 for ACE;
EXPOSE 7600 7800 7843

# Set default Integration Server name
ENV ACE_SERVER_NAME=ace-server

# Set entrypoint to run the server
ENTRYPOINT ["bash", "-c", ". /opt/ibm/ace-13/server/bin/mqsiprofile && IntegrationServer --name ${ACE_SERVER_NAME} -w /home/aceuser/ace-server"]
