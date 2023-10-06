export MQSI_SIGNAL_EXCLUSIONS=11
export MQSI_NO_CACHE_SUPPORT=1

# Needed to avoid errors in ibmint
export _JAVA_OPTIONS="-Dmqsipackagebar.noExtendClasspath=1"
# Needed for javax.xml packages
export MQSI_EXTRA_JAR_DIRECTORY=/opt/ibm/ace-12/common/jakarta/lib

# Make sure Maven and others can find javac
export PATH=/opt/ibm/ace-12/common/java11/bin:$PATH

# Not really java-related, but still needed
export LD_LIBRARY_PATH=/usr/glibc-compat/zlib-only:$LD_LIBRARY_PATH
