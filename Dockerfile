FROM ubuntu:18.04 AS base

RUN mkdir -p /usr/local/datakit \
    && mkdir -p /usr/local/datakit/externals \
    && mkdir -p /opt/oracle

COPY dist/datakit-linux-amd64/datakit /usr/local/datakit/datakit
COPY dist/datakit-linux-amd64/externals /usr/local/datakit/externals

RUN sed -i 's/\(archive\|security\).ubuntu.com/mirrors.aliyun.com/' /etc/apt/sources.list \
    && apt-get update \
    && apt-get install -y libaio-dev libaio1 unzip wget curl

# download 3rd party libraries
RUN wget -q https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/otn_software/instantclient/instantclient-basiclite-linux.x64-19.8.0.0.0dbru.zip -O /usr/local/datakit/externals/instantclient-basiclite-linux.zip \
    && unzip /usr/local/datakit/externals/instantclient-basiclite-linux.zip -d /opt/oracle

# download data files required by datakit
RUN wget -q -O data.tar.gz https://zhuyun-static-files-production.oss-cn-hangzhou.aliyuncs.com/datakit/data.tar.gz \
	&& tar -xzf data.tar.gz -C /usr/local/datakit && rm -rf data.tar.gz

ARG dataway=""
ARG uuid=""
ARG loglevel=""
ARG global_tags=""
ARG hostname=""

ENV ENV_UUID=$uuid \
    ENV_DATAWAY=$dataway \
    ENV_LOG_LEVEL=$loglevel \
    ENV_GLOBAL_TAGS=$global_tags \
    ENV_HOSTNAME=$hostname

CMD ["/usr/local/datakit/datakit", "--docker"]
