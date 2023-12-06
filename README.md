# gcp-filestore-csi-driver-operator

An operator to deploy the [GCP Filestore CSI Driver](https://github.com/openshift/gcp-filestore-csi-driver) in OKD.

# Quick start

To build and run the operator locally:

```shell
# Create only the resources the operator needs to run via CLI
oc apply -f - <<EOF
apiVersion: operator.openshift.io/v1
kind: ClusterCSIDriver
metadata:
    name: filestore.csi.storage.gke.io
spec:
  logLevel: Normal
  managementState: Managed
  operatorLogLevel: Trace
EOF

# Build the operator
make

# Set the environment variables
export DRIVER_IMAGE=quay.io/openshift/origin-gcp-filestore-csi-driver:latest
export OPERATOR_NAME=gcp-filestore-csi-driver-operator
export PROVISIONER_IMAGE=quay.io/openshift/origin-csi-external-provisioner:latest
export RESIZER_IMAGE=quay.io/openshift/origin-csi-external-resizer:latest
export SNAPSHOTTER_IMAGE=quay.io/openshift/origin-csi-external-snapshotter:latest
export NODE_DRIVER_REGISTRAR_IMAGE=quay.io/openshift/origin-csi-node-driver-registrar:latest
export LIVENESS_PROBE_IMAGE=quay.io/openshift/origin-csi-livenessprobe:latest
export KUBE_RBAC_PROXY_IMAGE=quay.io/openshift/origin-kube-rbac-proxy:latest

# Run the operator via CLI
./gce-filestore-csi-driver-operator start --kubeconfig $KUBECONFIG --namespace openshift-cluster-csi-drivers
```

# OLM

To build an bundle and index images, use the `hack/create-bundle` script:

```shell
cd hack
./create-bundle registry.ci.openshift.org/ocp/4.14:gcp-filestore-csi-driver registry.ci.openshift.org/ocp/4.14:gcp-filestore-csi-driver-operator quay.io/<my_user>/filestore-bundle quay.io/<my_user>/filestore-index
```

At the end it will print a command that creates `Subscription` for the newly created index image.
