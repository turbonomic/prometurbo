FROM registry.access.redhat.com/rhel7
MAINTAINER Enlin Xu <enlin.xu@turbonomic.com>
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

### Add necessary Red Hat repos here
RUN REPOLIST=rhel-7-server-rpms,rhel-7-server-optional-rpms \
    INSTALL_PKGS="golang-github-cpuguy83-go-md2man ca-certificates" && \
    yum -y update-minimal --disablerepo "*" --enablerepo rhel-7-server-rpms --setopt=tsflags=nodocs \
      --security --sec-severity=Important --sec-severity=Critical && \
    yum -y install --disablerepo "*" --enablerepo ${REPOLIST} --setopt=tsflags=nodocs ${INSTALL_PKGS} && \
    yum clean all

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
