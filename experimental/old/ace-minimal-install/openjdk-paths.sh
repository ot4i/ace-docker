export MQSI_SIGNAL_EXCLUSIONS=11
export MQSI_NON_IBM_JAVA=1
export MQSI_NO_CACHE_SUPPORT=1

export LD_LIBRARY_PATH=/lib:/usr/lib/jvm/default-jvm/jre/lib/amd64/server:/usr/lib/jvm/default-jvm/jre/lib/amd64:$LD_LIBRARY_PATH

# Not really openjdk-related, but still needed
export LD_LIBRARY_PATH=/usr/glibc-compat/zlib-only:$LD_LIBRARY_PATH
