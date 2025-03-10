#!/usr/bin/env bash

set -e

if [ -z "$1" ]; then
    echo "Error: script argument VERSION is not specified."
    exit 1
fi
VERSION=$1

if [ -z "${PUBLIC_GITHUB_TOKEN}" ]; then
    echo "Error: PUBLIC_GITHUB_TOKEN environment variable is not set"
    exit 1
fi
TC_PUBLIC_REPO=turbonomic-container-platform

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SRC_DIR=${SCRIPT_DIR}/../../deploy
OUTPUT_DIR=${SCRIPT_DIR}/../../_output
HELM=${SCRIPT_DIR}/../../bin/helm

if ! command -v ${HELM} > /dev/null 2>&1; then
    HELM=helm
    if ! command -v helm > /dev/null 2>&1; then
        echo "Error: helm could not be found."
        exit 1
    fi
fi

if ! command -v git > /dev/null 2>&1; then
    echo "Error: git could not be found."
    exit 1
fi

echo "===> Cloning public repo..."; 
mkdir ${OUTPUT_DIR}
cd ${OUTPUT_DIR}
git clone https://${PUBLIC_GITHUB_TOKEN}@github.com/IBM/${TC_PUBLIC_REPO}.git
cd ${TC_PUBLIC_REPO}

echo "===> Cleanup existing files"
rm -rf prometurbo 
mkdir -p prometurbo/operator
mkdir -p prometurbo/yamls
cd prometurbo

# copy helm chart
echo "===> Copy helm chart files"
cp -r ${SRC_DIR}/prometurbo helm_chart 

# copy operator files
echo "===> Copy Operator files"
cp -r ${SRC_DIR}/prometurbo-operator/deploy/prometurbo_operator_full.yaml operator/
cp -r ${SRC_DIR}/prometurbo-operator/deploy/crds/* operator/

# copy yaml files
echo "===> Copy yaml files"
cp -r ${SRC_DIR}/prometurbo_yamls/prometurbo_full.yaml yamls/

# Insert current version
echo "===> Updating Turbo version in yaml files"
sed -i.bak "s|version: 1.0.0|version: ${VERSION}|" helm_chart/Chart.yaml 
find ./ -type f -name '*.y*' -exec sed -i.bak -e "s|<PROMETURBO_IMAGE_TAG>|${VERSION}|g; s|<PROMETURBO_OPERATOR_IMAGE_TAG>|${VERSION}|g; s|<TURBODIF_IMAGE_TAG>|${VERSION}|g; s|<TURBONOMIC_SERVER_VERSION>|${VERSION}|g; s|VERSION|${VERSION}|g" {} +
find ./ -name '*.bak' -type f -delete 

# commit all modified source files to the public repo
echo "===> Commit modified files to public repo"
cd .. 
git add .
if ! git diff --quiet --cached; then
    git commit -m "prometurbo deployment ${VERSION}"
    git push
else
    echo "No changed files"
fi

# package the helm chart and upload to helm repo
echo "===> Package helm chart"
${HELM} package prometurbo/helm_chart -d ${OUTPUT_DIR}

echo "===> Update helm chart index"
git switch gh-pages
cp index.yaml ..
mkdir -p downloads/prometurbo
cp ${OUTPUT_DIR}/prometurbo-${VERSION}.* downloads/prometurbo/ 
${HELM} repo index .. --url https://ibm.github.io/${TC_PUBLIC_REPO}/downloads/prometurbo --merge index.yaml
cp ../index.yaml . 

# commit packaged helm chart
echo "===> Commit packaged helm chart to helm chart repo"
git add .
git commit -m "prometurbo helm chart ${VERSION}"
git push

# cleanup
rm -rf ${OUTPUT_DIR}

echo ""
echo "Update public repo complete."
