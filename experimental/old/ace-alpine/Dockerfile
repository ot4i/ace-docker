# FROM alpine:3.8
FROM openjdk:8-jre-alpine
# FROM ibmcom/ibmjava:8-sfj-alpine

MAINTAINER Trevor Dolby <tdolby@uk.ibm.com> (@tdolby)

RUN apk --update add --no-cache --virtual .build-deps curl binutils \
    && GLIBC_VER="2.29-r0" \
    && ALPINE_GLIBC_REPO="https://github.com/sgerrand/alpine-pkg-glibc/releases/download" \
    && GCC_LIBS_URL="https://archive.archlinux.org/packages/g/gcc-libs/gcc-libs-8.2.1%2B20180831-1-x86_64.pkg.tar.xz" \
    && GCC_LIBS_SHA256=e4b39fb1f5957c5aab5c2ce0c46e03d30426f3b94b9992b009d417ff2d56af4d \
    && ZLIB_URL="https://archive.archlinux.org/packages/z/zlib/zlib-1%3A1.2.11-3-x86_64.pkg.tar.xz" \
    && ZLIB_SHA256=17aede0b9f8baa789c5aa3f358fbf8c68a5f1228c5e6cba1a5dd34102ef4d4e5 \
    && curl -Ls https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub -o /etc/apk/keys/sgerrand.rsa.pub \
    && curl -Ls ${ALPINE_GLIBC_REPO}/${GLIBC_VER}/glibc-${GLIBC_VER}.apk > /tmp/${GLIBC_VER}.apk \
    && apk add /tmp/${GLIBC_VER}.apk \
    && curl -Ls ${GCC_LIBS_URL} -o /tmp/gcc-libs.tar.xz \
    && echo "${GCC_LIBS_SHA256}  /tmp/gcc-libs.tar.xz" | sha256sum -c - \
    && mkdir /tmp/gcc \
    && tar -xf /tmp/gcc-libs.tar.xz -C /tmp/gcc \
    && mv /tmp/gcc/usr/lib/libgcc* /tmp/gcc/usr/lib/libstdc++* /usr/glibc-compat/lib \
    && strip /usr/glibc-compat/lib/libgcc_s.so.* /usr/glibc-compat/lib/libstdc++.so* \
    && curl -Ls ${ZLIB_URL} -o /tmp/libz.tar.xz \
    && echo "${ZLIB_SHA256}  /tmp/libz.tar.xz" | sha256sum -c - \
    && mkdir /tmp/libz \
    && tar -xf /tmp/libz.tar.xz -C /tmp/libz \
    && mv /tmp/libz/usr/lib/libz.so* /usr/glibc-compat/lib \
    && apk del --purge .build-deps \
    && rm -rf /tmp/${GLIBC_VER}.apk /tmp/gcc /tmp/gcc-libs.tar.xz /tmp/libz /tmp/libz.tar.xz /var/cache/apk/* \
    && apk add bash curl \
    && mkdir /usr/glibc-compat/zlib-only \
    && ( cd /usr/glibc-compat/lib && tar -cf - libz* ) | ( cd /usr/glibc-compat/zlib-only && tar -xf - )


ENV TZ=Europe/London

# Set up exclusions
COPY excludes*txt /tmp/

RUN cat /tmp/excludes-base.txt >> /tmp/all-excludes.txt && \
# Adapters are optional
    cat /tmp/excludes-adapters.txt >> /tmp/all-excludes.txt && \
# CICS is optional
    cat /tmp/excludes-cics.txt >> /tmp/all-excludes.txt && \
# FTE is optional
    cat /tmp/excludes-fte.txt >> /tmp/all-excludes.txt && \
# Global Cache is optional; set MQSI_NO_CACHE_SUPPORT=1
    cat /tmp/excludes-globalcache.txt >> /tmp/all-excludes.txt && \
# JDK provided by base layer; use MQSI_SIGNAL_EXCLUSIONS=11 and 
# MQSI_NON_IBM_JAVA=1 for non-IBM Java layers
    cat /tmp/excludes-jdk.txt >> /tmp/all-excludes.txt && \
# NPM install support is optional
    cat /tmp/excludes-npm.txt >> /tmp/all-excludes.txt && \
# ODBC drivers are optional
    cat /tmp/excludes-odbc.txt >> /tmp/all-excludes.txt && \
# REST Web UI support optional; use --admin-rest-api -1 on server start line
    cat /tmp/excludes-webui.txt >> /tmp/all-excludes.txt && \
# XSLT nodes are optional
    cat /tmp/excludes-xmlt.txt >> /tmp/all-excludes.txt && \
# Toolkit is optional
    cat /tmp/excludes-tools.txt >> /tmp/all-excludes.txt

# Set the env vars mentioned above
COPY openjdk-paths.sh /etc/profile.d/openjdk-paths.sh

# Install ACE v11.0.0.5
RUN mkdir -p /opt/ibm && \
    apk add binutils zip && \
    curl http://kenya.hursley.uk.ibm.com:52367/ace-11.0.16836.5.tar.gz \
    | tar zx --exclude-from=/tmp/all-excludes.txt --directory /opt/ibm && \
    mv /opt/ibm/ace-11.0.16836.5 /opt/ibm/ace-11 && \
    find /opt/ibm -name "*.so*" -exec strip {} ";" && \
    find /opt/ibm -name "*.wrk" -exec strip {} ";" && \
    find /opt/ibm -name "*.lil" -exec strip {} ";" && \
    ( strip /opt/ibm/ace-11/server/bin/* 2>/dev/null || /bin/true ) && \
    zip -d /opt/ibm/ace-11/common/classes/IntegrationAPI.jar BIPmsgs_de.properties BIPmsgs_es.properties BIPmsgs_fr.properties BIPmsgs_it.properties BIPmsgs_ja.properties BIPmsgs_ko.properties BIPmsgs_pl.properties BIPmsgs_pt_BR.properties BIPmsgs_ru.properties BIPmsgs_tr.properties BIPmsgs_zh.properties BIPmsgs_zh_HK.properties BIPmsgs_zh_TW.properties && \
    apk del --purge binutils zip && \
    addgroup mqbrkrs && \
    cat /etc/profile.d/openjdk-paths.sh >> /root/.bashrc && \
    ln -s /usr/lib/jvm/default-jvm /opt/ibm/ace-11/common/jdk

# Copy in script files - this is used to prevent the "mqsiprofile 
# repetition disallowed" message from being printed on shell start.
COPY ace_env.sh /usr/local/bin/

RUN bash -c '. /etc/profile.d/openjdk-paths.sh && /opt/ibm/ace-11/ace make registry global accept license silently' && \
    echo ". /opt/ibm/ace-11/server/bin/mqsiprofile" >> /root/.bashrc

# Set BASH_ENV to source mqsiprofile when using docker exec bash -c
ENV BASH_ENV=/usr/local/bin/ace_env.sh

# Create a user to run as, create the ace workdir, and chmod script files
RUN ( echo "Passw0rd" ; echo "Passw0rd" ) | adduser -h /home/aceuser -s /bin/bash aceuser mqbrkrs && \
    adduser aceuser mqbrkrs && \
    chmod 755 /usr/local/bin/* && \
    cat /etc/profile.d/openjdk-paths.sh >> /home/aceuser/.bashrc && \
    chown aceuser /home/aceuser/.bashrc && \
    chmod 664 /home/aceuser/.bashrc && \
    su - aceuser -c '. /opt/ibm/ace-11/server/bin/mqsiprofile && mqsicreateworkdir /home/aceuser/ace-server' && \
    echo ". /usr/local/bin/ace_env.sh" >> /home/aceuser/.bashrc

USER aceuser
ENTRYPOINT ["bash"]
