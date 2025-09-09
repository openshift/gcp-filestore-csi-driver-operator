#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Usage:
#   ./hack/update-metadata.sh [OCP_VERSION]
#
#   OCP_VERSION is an optional argument. If no argument is provided, it defaults
#   to the version found in .channels[0].currentCSV in PACKAGE_MANIFEST.
#   This means you can run `./hack/update-metadata.sh` to update the manifests
#   using the current package version, or you can for example run
#   `./hack/update-metadata.sh 4.20` to set the package version to 4.20.
#   Both PACKAGE_MANIFEST and CSV_MANIFEST will be updated by this script.


PACKAGE_MANIFEST=config/manifests/gcp-filestore-csi-driver-operator.package.yaml
CHANNEL=$(yq '.channels[0].name' ${PACKAGE_MANIFEST})
CURRENT_CSV=$(yq '.channels[0].currentCSV' ${PACKAGE_MANIFEST})
PACKAGE_NAME=$(echo ${CURRENT_CSV} | sed 's/\.v.*$//')
PACKAGE_VERSION=$(echo ${CURRENT_CSV} | sed 's/^.*\.v//')

if [ -z "${CHANNEL}" ] ||
   [ -z "${PACKAGE_NAME}" ] ||
   [ -z "${PACKAGE_VERSION}" ]; then
	echo "Failed to parse ${PACKAGE_MANIFEST}"
	exit 1
fi

CSV_MANIFEST=config/manifests/${CHANNEL}/${PACKAGE_NAME}.clusterserviceversion.yaml
METADATA_NAME=$(yq '.metadata.name' ${CSV_MANIFEST})
SKIP_RANGE=$(yq '.metadata.annotations["olm.skipRange"]' ${CSV_MANIFEST})
MAX_OCP_VERSION=$(yq '.metadata.annotations["olm.properties"]' ${CSV_MANIFEST})
SPEC_VERSION=$(yq '.spec.version' ${CSV_MANIFEST})
ALM_STATUS_DESC=$(yq '.spec.labels.alm-status-descriptors' ${CSV_MANIFEST})

if [ -z "${METADATA_NAME}" ] ||
   [ -z "${SKIP_RANGE}" ] ||
   [ -z "${MAX_OCP_VERSION}" ] ||
   [ -z "${SPEC_VERSION}" ] ||
   [ -z "${ALM_STATUS_DESC}" ]; then
	echo "Failed to parse ${CSV_MANIFEST}"
	exit 1
fi

OCP_VERSION=${1:-${PACKAGE_VERSION}}
IFS='.' read -r MAJOR_VERSION MINOR_VERSION PATCH_VERSION <<< "${OCP_VERSION}"
PATCH_VERSION=${PATCH_VERSION:-0}
if [ "${OCP_VERSION}" != "${PACKAGE_VERSION}" ]; then
	PACKAGE_VERSION="${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}"
fi

export NEW_CURRENT_CSV="${PACKAGE_NAME}.v${PACKAGE_VERSION}"
export NEW_METADATA_NAME="${PACKAGE_NAME}.v${PACKAGE_VERSION}"
export NEW_SKIP_RANGE=$(echo ${SKIP_RANGE} | sed "s/ <.*$/ <${PACKAGE_VERSION}/")
export NEW_MAX_OCP_VERSION=$(echo ${MAX_OCP_VERSION} | jq -c ". | .[].value = \"${MAJOR_VERSION}.$((MINOR_VERSION + 1))\"")
export NEW_SPEC_VERSION="${PACKAGE_VERSION}"
export NEW_ALM_STATUS_DESC="${PACKAGE_NAME}.v${PACKAGE_VERSION}"

if [ -z "${NEW_METADATA_NAME}" ] ||
   [ -z "${NEW_SKIP_RANGE}" ] ||
   [ -z "${NEW_MAX_OCP_VERSION}" ] ||
   [ -z "${NEW_SPEC_VERSION}" ] ||
   [ -z "${NEW_ALM_STATUS_DESC}" ]; then
	echo "Failed to generate new values for ${CSV_MANIFEST}"
	exit 1
fi

echo "Updating package manifest to ${PACKAGE_VERSION}"
yq -i '.channels[0].currentCSV = strenv(NEW_CURRENT_CSV)' ${PACKAGE_MANIFEST}

echo "Updating OLM metadata to ${PACKAGE_VERSION}"
yq -i '
  .metadata.name = strenv(NEW_METADATA_NAME) |
  .metadata.annotations["olm.skipRange"] = strenv(NEW_SKIP_RANGE) |
  .metadata.annotations["olm.properties"] = strenv(NEW_MAX_OCP_VERSION) |
  .spec.version = strenv(NEW_SPEC_VERSION) |
  .spec.labels.alm-status-descriptors = strenv(NEW_ALM_STATUS_DESC)
' ${CSV_MANIFEST}

MAKEFILE=Makefile
echo "Updating Makefile build-image version to ${MAJOR_VERSION}.${MINOR_VERSION}"
if grep -q "call build-image,gcp-filestore-csi-driver-operator" ${MAKEFILE}; then
    sed -i '' -E "s|ocp/[0-9]+\.[0-9]+:|ocp/${MAJOR_VERSION}.${MINOR_VERSION}:|g" ${MAKEFILE}
else
    echo "build-image call for gcp-filestore-csi-driver-operator not found in ${MAKEFILE}"
fi

README=README.md
echo "Updating README.md version references to ${MAJOR_VERSION}.${MINOR_VERSION}"
if grep -q "registry.ci.openshift.org/ocp/" ${README}; then
    sed -i '' -E "s|registry\.ci\.openshift\.org/ocp/[0-9]+\.[0-9]+:|registry.ci.openshift.org/ocp/${MAJOR_VERSION}.${MINOR_VERSION}:|g" ${README}
else
    echo "registry.ci.openshift.org/ocp/ references not found in ${README}"
fi
