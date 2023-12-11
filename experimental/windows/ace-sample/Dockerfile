# Override for other variants
ARG BASE_IMAGE=ace-basic:12.0.10.0-windows
FROM $BASE_IMAGE

# docker build -t ace-sample:12.0.10.0-windows .
# docker run -p 7800:7800 --rm -ti ace-sample:12.0.10.0-windows

COPY aceFunction.bar c:\\tmp\\aceFunction.bar

RUN c:\\ace-12\\ace.cmd mqsicreateworkdir c:\\tmp\\ace-server
RUN c:\\ace-12\\ace.cmd mqsibar -c -a c:\\tmp\\aceFunction.bar -w c:\\tmp\\ace-server
#RUN c:\\ace-12\\ace.cmd ibmint optimize server --work-dir c:\\tmp\\ace-server

# Set entrypoint to run the server
ENTRYPOINT ["c:\\ace-12\\ace.cmd", "IntegrationServer", "-w", "c:\\tmp\\ace-server", "--no-nodejs", "--admin-rest-api", "-1"]
