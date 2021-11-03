export MQSI_SIGNAL_EXCLUSIONS=11
export MQSI_NO_CACHE_SUPPORT=1

# Make sure Maven and others can find javac
export PATH=/opt/ibm/ace-12/common/jdk/bin:$PATH

# Not really java-related, but still needed
export LD_LIBRARY_PATH=/usr/glibc-compat/zlib-only:$LD_LIBRARY_PATH
