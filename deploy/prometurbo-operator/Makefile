# Set shell
SHELL := /usr/bin/env bash
##### Developer Edit #####
OPERATOR_NAME ?= prometurbo
REGISTRY ?= icr.io/cpopen
# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
DEFAULT_VERSION=8.8.0
VERSION=$(or $(PROMETURBO_VERSION), $(DEFAULT_VERSION))
GIT_COMMIT=$(shell git rev-parse HEAD)

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
# CHANNELS define the bundle channels used in the bundle.
# Add a new line here if you would like to change its default config. (E.g CHANNELS = "candidate,fast,stable")
# To re-generate a bundle for other specific channels without changing the standard setup, you can:
# - use the CHANNELS as arg of the bundle target (e.g make bundle CHANNELS=candidate,fast,stable)
# - use environment variables to overwrite this value (e.g export CHANNELS="candidate,fast,stable")
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif

# DEFAULT_CHANNEL defines the default channel used in the bundle.
# Add a new line here if you would like to change its default config. (E.g DEFAULT_CHANNEL = "stable")
# To re-generate a bundle for any other default channel without changing the default setup, you can:
# - use the DEFAULT_CHANNEL as arg of the bundle target (e.g make bundle DEFAULT_CHANNEL=stable)
# - use environment variables to overwrite this value (e.g export DEFAULT_CHANNEL="stable")
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# IMAGE_TAG_BASE defines the docker.io namespace and part of the image name for remote images.
# This variable is used to construct full image tags for bundle and catalog images.
#
# For example, running 'make bundle-build bundle-push catalog-build catalog-push' will build and push both
# helm.k8s.io/prometurbo-operator-bundle:$VERSION and helm.k8s.io/prometurbo-operator-catalog:$VERSION.
IMAGE_TAG_BASE ?= turbonomic/prometurbo-operator

# BUNDLE_IMG defines the image:tag used for the bundle.
# You can use it as an arg. (E.g make bundle-build BUNDLE_IMG=<some-registry>/<project-name-bundle>:<tag>)
BUNDLE_IMG ?= $(IMAGE_TAG_BASE)-bundle:v$(VERSION)

# Image URL to use all building/pushing image targets
IMG ?= $(IMAGE_TAG_BASE):$(VERSION)

.PHONY: all
all: docker-build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: run
run: helm-operator ## Run against the configured Kubernetes cluster in ~/.kube/config
	$(HELM_OPERATOR) run

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build --build-arg VERSION=${VERSION} -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

.PHONY: install
install: kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

.PHONY: deploy
deploy: kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m | sed 's/x86_64/amd64/')

.PHONY: kustomize
KUSTOMIZE = $(LOCALBIN)/kustomize
kustomize: ## Download kustomize locally if necessary.
ifeq (,$(wildcard $(KUSTOMIZE)))
ifeq (,$(shell which kustomize 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(KUSTOMIZE)) ;\
	curl -sSLo - https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v3.8.7/kustomize_v3.8.7_$(OS)_$(ARCH).tar.gz | \
	tar xzf - -C bin/ ;\
	}
else
KUSTOMIZE = $(shell which kustomize)
endif
endif

.PHONY: helm-operator
HELM_OPERATOR = $(LOCALBIN)/helm-operator
helm-operator: ## Download helm-operator locally if necessary, preferring the $(pwd)/bin path over global if both exist.
ifeq (,$(wildcard $(HELM_OPERATOR)))
ifeq (,$(shell which helm-operator 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(HELM_OPERATOR)) ;\
	curl -sSLo $(HELM_OPERATOR) https://github.com/operator-framework/operator-sdk/releases/download/v1.18.1/helm-operator_$(OS)_$(ARCH) ;\
	chmod +x $(HELM_OPERATOR) ;\
	}
else
HELM_OPERATOR = $(shell which helm-operator)
endif
endif

.PHONY: bundle
bundle: kustomize ## Generate bundle manifests and metadata, then validate generated files.
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

.PHONY: bundle-build
bundle-build: ## Build the bundle image.
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

.PHONY: bundle-push
bundle-push: ## Push the bundle image.
	$(MAKE) docker-push IMG=$(BUNDLE_IMG)

YQ ?= $(LOCALBIN)/yq
YQ_TOOLS_VERSION ?= v4.30.4

.PHONY: yq
yq: $(YQ) ## Download yq locally if necessary.
$(YQ): $(LOCALBIN)
	test -s $(LOCALBIN)/yq || GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@$(YQ_TOOLS_VERSION)


.PHONY: python
PYTHON = $(LOCALBIN)/python3
python: $(PYTHON)  ## Install Python locally if necessary. Darwin OS is specific to mac users if running locally
$(PYTHON):
	@if ! command -v python3 >/dev/null 2>&1; then \
		mkdir -p $(LOCALBIN); \
		if [ `uname -s` = "Darwin" ]; then \
			brew install python@3; \
		else \
			sudo apt update && sudo apt install python3; \
		fi \
	fi
	ln -sf `command -v python3` $(PYTHON)


OPERATOR_OLM_INCLUSIVE_RANGE_VERSION ?= 8.7.5
OPERATOR_OLM_INCLUSIVE_BETA_VERSION ?= beta.1
OPERATOR_BETA_RELEASE_FILTER ?= beta
OPERATOR_BETA_RELEASE_VERSION_SNAPSHOT ?= SNAPSHOT
OPERATOR_CERTIFIED ?= prometurbo-certified
OPERATOR_BUNDLE_DIR ?= certified-operator-bundle
OPERATOR_BUNDLE_TEMPLATE_DIR ?= bundle-template
OPERATOR_CRD_FILE_PATH ?= deploy/crds/charts.helm.k8s.io_prometurbos_crd.yaml
CLUSTER_PERMISSION_ROLE_YAML_FILE_PATH ?= deploy/prometurbo-operator-cluster-role.yaml
CERTIFIED_OPERATOR_CLUSTER_SERVICE_VERSION_YAML_FILE_PATH ?= $(OPERATOR_BUNDLE_DIR)/manifest/prometurbo-certified.clusterserviceversion.yaml
GITHUB_REPO_URL := https://api.github.com/repos/turbodeploy/certified-operators/contents/operators/prometurbo-certified
OPERATOR_VERSION_BETA := $(shell \
	OPERATOR_BETA_RELEASE_MINOR_VERSION=$$(curl -s "$(GITHUB_REPO_URL)" | \
	jq -r 'map(select(.type == "dir")) | .[].name | match("$(OPERATOR_RELEASE_VERSION)-beta\\.[0-9]+") | try .string catch "0"' | \
	awk -F'.' 'BEGIN{max=0} {n=substr($$0, index($$0, "beta.")+5)+0; if (n>max) max=n} END{print max}'); \
	if [ -z "$$OPERATOR_BETA_RELEASE_MINOR_VERSION" ]; then \
		OPERATOR_BETA_RELEASE_MINOR_VERSION=1; \
	else \
		OPERATOR_BETA_RELEASE_MINOR_VERSION=$$(($$OPERATOR_BETA_RELEASE_MINOR_VERSION + 1)); \
	fi; \
	OPERATOR_VERSION_BETA=beta.$$OPERATOR_BETA_RELEASE_MINOR_VERSION; \
	echo "$$OPERATOR_VERSION_BETA" \
)
.PHONY: build-certified-operator-bundle
build-certified-operator-bundle:yq python verify-operator-release-channel verify-stable-operator-release-version verify_image_digest_version create_certified_operator_bundle_directory update_image_digest_in_operator_bundle update_operator_version_in_operator_bundle update_olm_skipRange_in_operator_bundle update_cluster_permissions_in_operator_bundle update_release_channel_in_operator_bundle
## verify operator release channel, to allow only valid releases
verify-operator-release-channel:
    ifneq ($(filter $(OPERATOR_RELEASE_CHANNEL),stable beta),$(OPERATOR_RELEASE_CHANNEL))
        $(error Invalid operator release channel parameter - $(OPERATOR_RELEASE_CHANNEL). valid release channels are either "stable" or "beta only".)
    endif
## verify operator release version on stable channel, to avoid multiple releases of same version
verify-stable-operator-release-version:
	if [ "$(OPERATOR_RELEASE_CHANNEL)" = "stable" ]; then \
        echo "Checking if the stable release version $(OPERATOR_RELEASE_VERSION) exists..."; \
         if [ -n "$$(curl -s "$(GITHUB_REPO_URL)" | jq -r 'map(select(.type == "dir" and .name == "$(OPERATOR_RELEASE_VERSION)")) | .[].name')" ]; then \
            echo "Error: The operator release version already exists for stable channel: $(OPERATOR_RELEASE_VERSION)."; \
            exit 1; \
        fi; \
    fi
## verify if the version field value is present in the image to proceed, if empty exit the execution
verify_image_digest_version:
	@echo "Verify Image Digest version field value"
	$(eval OPERATOR_IMAGE_RELEASE_VERSION := $(if $(filter $(OPERATOR_BETA_RELEASE_FILTER),$(OPERATOR_RELEASE_CHANNEL)),$(OPERATOR_RELEASE_VERSION)-$(OPERATOR_BETA_RELEASE_VERSION_SNAPSHOT),$(OPERATOR_RELEASE_VERSION)))
	docker pull $(REGISTRY)/$(OPERATOR_NAME)-operator:$(OPERATOR_IMAGE_RELEASE_VERSION)
	version=$$(docker inspect $(REGISTRY)/$(OPERATOR_NAME)-operator:$(OPERATOR_IMAGE_RELEASE_VERSION) | grep '"version":' | awk '{print $$2}' | tr -d '",'); \
	if [ -z "$$version" ]; then \
		echo "Error: Image digest version field is empty, cannot procced with $(OPERATOR_CERTIFIED)-operator bundle release."; \
		exit 1; \
	elif [ "$$version" != "$(OPERATOR_IMAGE_RELEASE_VERSION)" ]; then \
		echo "Error: Image digest version field: ($$version) does not match operator release version: ($(OPERATOR_IMAGE_RELEASE_VERSION))."; \
		exit 1; \
	else \
		echo "Image digest validation successful, proceeding with next steps."; \
	fi

## create certified operator bundle dir and copy the base files to update the clusterserviceversion and metadata contents as required for releasing
create_certified_operator_bundle_directory:
	@echo "Creating certified operator bundle files for $(OPERATOR_CERTIFIED) clusterserviceversion..."
	mkdir -p $(OPERATOR_BUNDLE_DIR)/manifest/
	cp $(OPERATOR_CRD_FILE_PATH) $(OPERATOR_BUNDLE_DIR)/manifest/prometurbos.charts.helm.k8s.io.crd.yaml
	cp $(OPERATOR_BUNDLE_TEMPLATE_DIR)/template/$(OPERATOR_CERTIFIED).clusterserviceversion.yaml $(OPERATOR_BUNDLE_DIR)/manifest/$(OPERATOR_CERTIFIED).clusterserviceversion.yaml
	mkdir -p $(OPERATOR_BUNDLE_DIR)/metadata/
	cp $(OPERATOR_BUNDLE_TEMPLATE_DIR)/template/annotations.yaml $(OPERATOR_BUNDLE_DIR)/metadata/annotations.yaml
	@echo "$(OPERATOR_CERTIFIED)-operator bundle directory created successfully."

## update image digest key
update_image_digest_in_operator_bundle:
	@echo "Updating image digest in $(OPERATOR_CERTIFIED)-clusterserviceversion..."
	$(eval OPERATOR_IMAGE_RELEASE_VERSION := $(if $(filter $(OPERATOR_BETA_RELEASE_FILTER),$(OPERATOR_RELEASE_CHANNEL)),$(OPERATOR_RELEASE_VERSION)-$(OPERATOR_BETA_RELEASE_VERSION_SNAPSHOT),$(OPERATOR_RELEASE_VERSION)))
	digest=$$(docker inspect --format='{{index .RepoDigests 0}}' $(REGISTRY)/$(OPERATOR_NAME)-operator:$(OPERATOR_IMAGE_RELEASE_VERSION) | awk -F@ '{print $$2}'); \
	if [ -z "$$digest" ]; then \
		echo "Error: Image digest is empty, cannot proceed with $(OPERATOR_CERTIFIED)-operator bundle release."; \
		exit 1; \
	else \
	    $(YQ) eval -i '.spec.install.spec.deployments[0].spec.template.spec.containers[0].image |= sub("sha256:.*", "'$$digest'") | .spec.relatedImages[0].image |= sub("sha256:.*", "'$$digest'")' \
		    $(OPERATOR_BUNDLE_DIR)/manifest/$(OPERATOR_CERTIFIED).clusterserviceversion.yaml; \
	    echo "$(OPERATOR_CERTIFIED)-clusterserviceversion image digest updated."; \
	fi

## update release version
update_operator_version_in_operator_bundle:
	@echo "Updating release versions in $(OPERATOR_CERTIFIED)-clusterserviceversion..."
	$(eval OPERATOR_RELEASE_CHANNEL_VERSION := $(if $(filter $(OPERATOR_BETA_RELEASE_FILTER),$(OPERATOR_RELEASE_CHANNEL)),$(OPERATOR_RELEASE_VERSION)-$(OPERATOR_VERSION_BETA),$(OPERATOR_RELEASE_VERSION)))
	$(YQ) eval -i '.metadata.name |= sub("prometurbo-operator.v.*", "prometurbo-operator.v$(OPERATOR_RELEASE_CHANNEL_VERSION)") | .spec.version = "$(OPERATOR_RELEASE_CHANNEL_VERSION)"' \
		$(OPERATOR_BUNDLE_DIR)/manifest/$(OPERATOR_CERTIFIED).clusterserviceversion.yaml
	@echo "$(OPERATOR_CERTIFIED)-clusterserviceversion release versions updated successfully."

## update skipRange based on inclusive and exclusive release versions set
update_olm_skipRange_in_operator_bundle:
	@echo "Updating olm.skipRange in $(OPERATOR_CERTIFIED)-clusterserviceversion..."
	$(YQ) eval -i '.metadata.annotations."olm.skipRange" |= sub(">=[^<]+", ">=$(if $(filter $(OPERATOR_BETA_RELEASE_FILTER),$(OPERATOR_RELEASE_CHANNEL)),$(OPERATOR_OLM_INCLUSIVE_RANGE_VERSION)-$(OPERATOR_OLM_INCLUSIVE_BETA_VERSION),$(OPERATOR_OLM_INCLUSIVE_RANGE_VERSION))") \
	| .metadata.annotations."olm.skipRange" |= sub("<[^<]+", " <$(if $(filter $(OPERATOR_BETA_RELEASE_FILTER),$(OPERATOR_RELEASE_CHANNEL)),$(OPERATOR_RELEASE_VERSION)-$(OPERATOR_VERSION_BETA),$(OPERATOR_RELEASE_VERSION))")' \
	   $(OPERATOR_BUNDLE_DIR)/manifest/$(OPERATOR_CERTIFIED).clusterserviceversion.yaml
	@echo "$(OPERATOR_CERTIFIED)-clusterserviceversion olm.skipRange updated successfully."

## update cluster permissions roles
update_cluster_permissions_in_operator_bundle:
	@echo "Updating cluster permissions roles in $(OPERATOR_CERTIFIED)-clusterserviceversion..."
	$(PYTHON) $(OPERATOR_BUNDLE_TEMPLATE_DIR)/cluster_permissions_automation.py \
	$(CLUSTER_PERMISSION_ROLE_YAML_FILE_PATH) \
	$(CERTIFIED_OPERATOR_CLUSTER_SERVICE_VERSION_YAML_FILE_PATH)
	@echo "$(OPERATOR_CERTIFIED)-clusterserviceversion cluster permissions roles updated successfully."

## update release channel
update_release_channel_in_operator_bundle:
	@echo "Updating release channel in  $(OPERATOR_CERTIFIED)-annotations..."
	$(YQ) eval -i '.annotations."operators.operatorframework.io.bundle.channels.v1" = "$(OPERATOR_RELEASE_CHANNEL)"' \
		$(OPERATOR_BUNDLE_DIR)/metadata/annotations.yaml
	@echo "$(OPERATOR_CERTIFIED)-annotations release channel updated successfully."


.PHONY: opm
OPM = ./bin/opm
opm: ## Download opm locally if necessary.
ifeq (,$(wildcard $(OPM)))
ifeq (,$(shell which opm 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPM)) ;\
	curl -sSLo $(OPM) https://github.com/operator-framework/operator-registry/releases/download/v1.19.1/$(OS)-$(ARCH)-opm ;\
	chmod +x $(OPM) ;\
	}
else
OPM = $(shell which opm)
endif
endif

# A comma-separated list of bundle images (e.g. make catalog-build BUNDLE_IMGS=example.com/operator-bundle:v0.1.0,example.com/operator-bundle:v0.2.0).
# These images MUST exist in a registry and be pull-able.
BUNDLE_IMGS ?= $(BUNDLE_IMG)

# The image tag given to the resulting catalog image (e.g. make catalog-build CATALOG_IMG=example.com/operator-catalog:v0.2.0).
CATALOG_IMG ?= $(IMAGE_TAG_BASE)-catalog:$(VERSION)

# Set CATALOG_BASE_IMG to an existing catalog image tag to add $BUNDLE_IMGS to that image.
ifneq ($(origin CATALOG_BASE_IMG), undefined)
FROM_INDEX_OPT := --from-index $(CATALOG_BASE_IMG)
endif

# Build a catalog image by adding bundle images to an empty catalog using the operator package manager tool, 'opm'.
# This recipe invokes 'opm' in 'semver' bundle add mode. For more information on add modes, see:
# https://github.com/operator-framework/community-operators/blob/7f1438c/docs/packaging-operator.md#updating-your-existing-operator
.PHONY: catalog-build
catalog-build: opm ## Build a catalog image.
	$(OPM) index add --container-tool docker --mode semver --tag $(CATALOG_IMG) --bundles $(BUNDLE_IMGS) $(FROM_INDEX_OPT)

# Push the catalog image.
.PHONY: catalog-push
catalog-push: ## Push a catalog image.
	$(MAKE) docker-push IMG=$(CATALOG_IMG)

# Build the image for multi-platforms
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
REPO_NAME ?= icr.io/cpopen
.PHONY: docker-buildx
docker-buildx:
	docker buildx create --name prometurbo-operator-builder
	- docker buildx use prometurbo-operator-builder
	- docker buildx build --platform=$(PLATFORMS) --label "git-commit=$(GIT_COMMIT)" --label "git-version=$(VERSION)" --provenance=false --push --tag $(REPO_NAME)/prometurbo-operator:$(VERSION) --build-arg VERSION=$(VERSION) .
	docker buildx rm prometurbo-operator-builder
