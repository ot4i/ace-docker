
# May also need to change JDBC policy settings to avoid hitting
# "Unsupported protocolSSL_TLSv2" errors from DB2 as described in
# https://www.ibm.com/mysupport/s/defect/aCI3p000000XlXvGAK/dt179015
#
# <environmentParms>sslConnection=true;sslVersion=TLSv1.3</environmentParms>

. /opt/ibm/ace-13/server/bin/mqsiprofile

# Make sure Maven and others can find javac
export PATH=/opt/ibm/ace-13/common/java17/bin:$PATH

