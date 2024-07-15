# alert-operator

The Alert-Operator is the missing component of the [kube-prometheus stack](https://github.com/prometheus-operator/kube-prometheus/) for observability on Kubernetes:
it allows you to view active **Alerts** from the comfort of your `kubectl` CLI and lets you manage associated **Silences** declaratively, for example with GitOps.

```sh
$ kubectl get alerts
NAME                 STATE   VALUE  SINCE  LABELS
ContainerOOM-asxi    firing  1      3h17m  pod=prometheus-k8s-db-prometheus-k8s-0,severity=warning
KubeJobFailed-kp2md  firing  3      42m    alertname=KubeJobFailed,job_name=image-pruner-28679172,namespace=openshift-image-registry

$ kubectl get alert ContainerOOM-asxi -o yaml
apiVersion: alertmanager.prometheus.io/v1alpha1
kind: Alert
metadata:
  name: ContainerOOM-asxi
spec: {}
status:
  since: 2018-07-04 20:27:12.60602144 +0200 CEST
  state: firing
  value: "1e+00"
  labels:
    alertname: ContainerOOM
    endpoint: https-metrics
    instance: 1.2.3.4:10250
    job: kubelet
    metrics_path: /metrics
    namespace: openshift-monitoring
    node: infra-avz-a-5phgg
    pod: prometheus-k8s-db-prometheus-k8s-0
    service: kubelet
    severity: warning
  annotations:
    summary: The container 'prometheus' of pod 'prometheus-k8s-db-prometheus-k8s-0' has been restarted multiple times due to running out of memory.
```

```sh
$ kubectl get silences
NAME                   STATE    CREATOR  COMMENT
out-of-memory-issues   active   foobar   Currently scaling up the cluster and waiting for new nodes
```

## Development

### Prerequisites
- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/alert-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install

# to uninstall (delete CRDs from the cluster):
make uninstall
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/alert-operator:tag

# to uninstall (remove controller from the cluster):
make undeploy
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/

# to uninstall:
kubectl delete -k config/samples/
```

> **NOTE**: Ensure that the samples has default values to test it out.

> **NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/alert-operator:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/alert-operator/<tag or branch>/dist/install.yaml
```

## Contributing

Contributions are welcome! Feel free to submit feedback and ideas by opening [GitHub issues](https://github.com/jacksgt/alert-operator/issues).
It is recommended to discuss your feature in an issue before opening a pull request.

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

