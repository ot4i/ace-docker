# Combined file for easier scripting
export MQSI_SIGNAL_EXCLUSIONS=11
export MQSI_NO_CACHE_SUPPORT=1

. /opt/ibm/ace-11/server/bin/mqsiprofile

export LD_LIBRARY_PATH=/lib:/opt/ibm/java/jre/lib/amd64/compressedrefs:/opt/ibm/java/jre/lib/amd64:$LD_LIBRARY_PATH

# Not really ibmjava-related, but still needed
export LD_LIBRARY_PATH=/usr/glibc-compat/zlib-only:$LD_LIBRARY_PATH
