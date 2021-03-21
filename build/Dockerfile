FROM registry.access.redhat.com/ubi8-minimal
MAINTAINER Enlin Xu <enlin.xu@turbonomic.com>
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

### Atomic/OpenShift Labels - https://github.com/projectatomic/ContainerApplicationGenericLabels
LABEL name="Prometurbo Container" \
      vendor="Turbonomic" \
      version="v8.0.0" \
      release="1" \
      summary="Performance assurance for the applications in Openshift" \
      description="Prometurbo Container leverages Turbonomic control platform, to assure the performance of micro-services running in OpenShift, as well as the efficiency of underlying infrastructure." \
### Required labels above - recommended below
      url="https://www.turbonomic.com" \
      run='docker run -tdi --name ${NAME} vmturbo/prometurbo:latest' \
      io.k8s.description="Prometurbo Container will monitor and control the entire stack, from OpenShift down to your underlying infrastructure. " \
      io.k8s.display-name="Prometurbo Container" \
      io.openshift.expose-services="" \
      io.openshift.tags="turbonomic, Multicloud Container"

### add licenses to this directory
COPY licenses /licenses

COPY Dockerfile /Dockerfile

### Setup user for build execution and application runtime
ENV APP_ROOT=/opt/turbonomic
ENV PATH=$PATH:${APP_ROOT}/bin

RUN mkdir -p ${APP_ROOT}/bin
COPY prometurbo.linux ${APP_ROOT}/bin/prometurbo
RUN chmod -R ug+x ${APP_ROOT}/bin && sync && \
    chmod -R g=u ${APP_ROOT}

####### Add app-specific needs below. #######
USER 10001
WORKDIR ${APP_ROOT}
ENTRYPOINT ["/opt/turbonomic/bin/prometurbo"]
