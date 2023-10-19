# Combined file for easier scripting
export MQSI_SIGNAL_EXCLUSIONS=11
export MQSI_NO_CACHE_SUPPORT=1

# Don't use JSSE2
export MQSI_JAVA_AVOID_SSL_SETTINGS=1

# Needed to avoid errors in ibmint
export _JAVA_OPTIONS="-Dmqsipackagebar.noExtendClasspath=1"
# Needed for javax.xml packages
export MQSI_EXTRA_JAR_DIRECTORY=/opt/ibm/ace-12/common/jakarta/lib

. /opt/ibm/ace-12/server/bin/mqsiprofile

# Make sure Maven and others can find javac
export PATH=/opt/ibm/ace-12/common/java11/bin:$PATH

# Not really java-related, but still needed
export LD_LIBRARY_PATH=/usr/glibc-compat/zlib-only:$LD_LIBRARY_PATH
