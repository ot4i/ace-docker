# Build and run:
#
# docker build -t ace:12.0.4.0 -f Dockerfile .
# docker run -e LICENSE=accept -p 7600:7600 -p 7800:7800 --rm -ti ace:12.0.2.0
#
# Can also mount a volume for the work directory:
#
# docker run -e LICENSE=accept -v /what/ever/dir:/home/aceuser/ace-server -p 7600:7600 -p 7800:7800 --rm -ti ace:12.0.2.0
#
# This might require a local directory with the right permissions, or changing the userid further down . . .

FROM registry.access.redhat.com/ubi9/ubi-minimal as builder

RUN microdnf update -y && microdnf install -y util-linux tar && microdnf clean all

ARG USERNAME
ARG PASSWORD
# Download and reference the ACE-LINUX64-DEVELOPER.tar.gz from here https://www.ibm.com/resources/mrs/assets?source=swg-wmbfd eg.
ARG DOWNLOAD_URL=<Your downloaded location>/12.0.4.0-ACE-LINUX64-DEVELOPER.tar.gz

RUN mkdir -p /opt/ibm/ace-12 \
    && if [ -z $USERNAME ]; then curl ${DOWNLOAD_URL}; else curl -u "${USERNAME}:${PASSWORD}" ${DOWNLOAD_URL}; fi | \
    tar zx --absolute-names \
    --exclude ace-12.0.*.*/tools \
    --exclude ace-12.0.*.*/server/tools/ibm-dfdl-java.zip \
    --exclude ace-12.0.*.*/server/tools/proxyservlet.war \
    --exclude ace-12.0.*.*/server/bin/TADataCollector.sh \
    --exclude ace-12.0.*.*/server/transformationAdvisor/ta-plugin-ace.jar \
    --strip-components 1 \
    --directory /opt/ibm/ace-12

FROM registry.access.redhat.com/ubi9/ubi-minimal

# Force reinstall tzdata package to get zoneinfo files
RUN microdnf update -y && microdnf install -y findutils util-linux which tar && microdnf reinstall -y tzdata && microdnf clean all

# Install ACE v12.0.4.0 and accept the license
COPY --from=builder /opt/ibm/ace-12 /opt/ibm/ace-12
RUN /opt/ibm/ace-12/ace make registry global accept license deferred \
    && useradd --uid 1001 --create-home --home-dir /home/aceuser --shell /bin/bash -G mqbrkrs aceuser \
    && su - aceuser -c "export LICENSE=accept && . /opt/ibm/ace-12/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server" \
    && echo ". /opt/ibm/ace-12/server/bin/mqsiprofile" >> /home/aceuser/.bashrc

# Add required license as text file in Liceses directory (GPL, MIT, APACHE, Partner End User Agreement, etc)
COPY /licenses/ /licenses/

# aceuser
USER 1001

# Expose ports.  7600, 7800, 7843 for ACE;
EXPOSE 7600 7800 7843

# Set default Integration Server name
ENV ACE_SERVER_NAME ace-server

# Set entrypoint to run the server
ENTRYPOINT ["bash", "-c", ". /opt/ibm/ace-12/server/bin/mqsiprofile && IntegrationServer --name ${ACE_SERVER_NAME} -w /home/aceuser/ace-server"]
