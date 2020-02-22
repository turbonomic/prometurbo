FROM quay.io/operator-framework/helm-operator:v0.15.2

# Required OpenShift Labels
LABEL name="Prometurbo Operator" \
      vendor="Turbonomic" \
      version="v7.21.0" \
      release="1" \
      summary="This is the prometurbo operator." \
      description="This operator will deploy an instance of prometurbo." \
### Required labels above - recommended below
      url="https://www.turbonomic.com" \
      run='docker run -tdi --name ${NAME} turbonomic/prometurbo-operator:7.21' \
      io.k8s.description="Turbonomic Workload Automation Platform simultaneously optimizes performance, compliance, and cost in real-time. Workloads are precisely resourced, automatically, to perform while satisfying business constraints.  " \
      io.k8s.display-name="Prometurbo Operator" \
      io.openshift.expose-services="" \
      io.openshift.tags="turbonomic, Multicloud Container"

# Required Licenses
COPY licenses /licenses

COPY watches.yaml ${HOME}/watches.yaml
COPY helm-charts/ ${HOME}/helm-charts/
