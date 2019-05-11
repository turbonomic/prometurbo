FROM registry.access.redhat.com/ubi8
MAINTAINER Enlin Xu <enlin.xu@turbonomic.com>
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

### Setup user for build execution and application runtime
ENV APP_ROOT=/opt/turbonomic \
    USER_NAME=default \
    USER_UID=10001
ENV PATH=$PATH:${APP_ROOT}/bin

RUN mkdir -p ${APP_ROOT}/bin
COPY prometurbo.linux ${APP_ROOT}/bin/prometurbo
RUN chmod -R ug+x ${APP_ROOT}/bin && sync && \
    useradd -l -u ${USER_UID} -r -g 0 -d ${APP_ROOT} -s /sbin/nologin -c "${USER_NAME} user" ${USER_NAME} && \
    chown -R ${USER_UID}:0 ${APP_ROOT} && \
    chmod -R g=u ${APP_ROOT}

####### Add app-specific needs below. #######
USER 10001
WORKDIR ${APP_ROOT}
ENTRYPOINT ["/opt/turbonomic/bin/prometurbo"]
