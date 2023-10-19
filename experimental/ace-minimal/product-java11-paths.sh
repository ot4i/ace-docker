export MQSI_SIGNAL_EXCLUSIONS=11
export MQSI_NO_CACHE_SUPPORT=1

# Don't use JSSE2
export MQSI_JAVA_AVOID_SSL_SETTINGS=1
# May also need to change JDBC policy settings to avoid hitting
# "Unsupported protocolSSL_TLSv2" errors from DB2 as described in
# https://www.ibm.com/mysupport/s/defect/aCI3p000000XlXvGAK/dt179015
#
# <environmentParms>sslConnection=true;sslVersion=TLSv1.3</environmentParms>

# Needed to avoid errors in ibmint
export _JAVA_OPTIONS="-Dmqsipackagebar.noExtendClasspath=1"
# Needed for javax.xml packages
export MQSI_EXTRA_JAR_DIRECTORY=/opt/ibm/ace-12/common/jakarta/lib

# Make sure Maven and others can find javac
export PATH=/opt/ibm/ace-12/common/java11/bin:$PATH

# Not really java-related, but still needed
export LD_LIBRARY_PATH=/usr/glibc-compat/zlib-only:$LD_LIBRARY_PATH
