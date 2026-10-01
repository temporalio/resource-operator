This repository contains the Temporal Resource Operator: a Kubernetes
[Operator][k8s-operator] that manages the lifecycle of one or more [Temporal
Cloud][tcloud] resources.

[k8s-operator]: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[tcloud]: https://saas-api.tmprl.cloud/docs/httpapi.html#description/introduction

# How It Works

After installing the Temporal Resource Operator (TCO), a user can create, modify
or delete Temporal Cloud resources using the `kubectl` CLI.

For example, to create a new Temporal Cloud Namespace called
`my-org-namespace`, you might do:

```bash
$ kubectl apply -f - <<EOF
apiVersion: cloud.temporal.io/v1alpha1
kind: Namespace
metadata:
  name: my-org-namespace
  namespace: temporal-cloud  # NOTE: this is the Kubernetes Namespace, not the Temporal Cloud Namespace.
EOF
```

TRO watches the Kubernetes API for [Custom Resources][cr] (CRs) that represent
Temporal Cloud resources and performs all the necessary actions to managed the
lifecycle of those Temporal Cloud resources associated with the Custom
Resource.

If you are familiar with projects like [Crossplane][crossplane] or [AWS
Controllers for Kubernetes][ack] (ACK), this is the same concept applied to
Temporal Cloud resources.

[cr]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/
[crossplane]: https://www.crossplane.io/
[ack]: https://aws-controllers-k8s.github.io/docs/intro/

# Installing the Operator

## Kustomize

The recommended way to install Temporal Resource Operator is to use Kustomize.

The following will install the latest Temporal Resource Operator into your
Kubernetes Cluster in the `temporal-resource-operator` Kubernetes Namespace:

```bash
$ export TRO_TAG=main
$ kubectl apply -n temporal-resource-operator -k "https://github.com/temporalio/resource-operator/deploy/kustomize?ref=$TRO_TAG"
```

## Helm

# Upgrading the Operator

# Using the Operator

## Create a new Temporal Cloud Namespace

## Modify a Temporal Cloud Namespace

## Delete a Temporal Cloud Namespace

# Contributing

## Developing

### Generating code

There are several Makefile targets to assist you in generating code.

```
make gen-all
```

Will generate all Go code and Kubernetes deployment manifests.

## Testing

This repository does *not* use Ginkgo or the upstream Kubernetes
controller-runtime `envtest` package for testing the binaries and configuration
artifacts for Temporal Resource Operator.

### Test environment

Several Makefile targets automate common tasks for setting up, tearing down,
and using a [KinD][kind] cluster for local testing of Temporal Resource
Operator.

```
make test-cluster-create
```

Will create the KinD cluster for local testing.

```
make test-cluster-delete
```

Will tear down the KinD cluster for local testing.

```
make test-cluster-load-image
```

Will build the Temporal Resource Operator container image with your local
changes and load the container image into the KinD cluster's Docker registry.

```
make test-cluster-reset
```

Can be used to delete any existing test cluster environment, create a new one,
build the Temporal Resource Operator container image and load it into the KinD
cluster's Docker registry.

```
make test-deploy
```

Will install the Temporal Resource Operator running your local code along with
all of the CRDs, RBAC resources and related manifests in the `deploy/kustomize`
directory.

```
make test-undeploy
```

Will remove all CRDs, RBAC roles and Temporal Resource Operator deployment
resources from the test cluster environment.

[kind]: https://kind.sigs.k8s.io/

### Unit testing

For unit testing, we use simple, [table-driven][table-driven-testing] tests and
the following test libraries:

* `github.com/stretchr/testify/`
* `github.com/stretchr/testify/mock`

Unit tests should validate the smallest testable unit of work possible. If your
unit tests cross lots of function/method boundaries, you might not be writing a
unit test but instead building a larger functional or integration test.

To run unit tests:

```
make test-unit
```

[table-driven-testing]: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests 

### Functional and end-to-end testing

For functional and integration testing, we use a declarative testing framework
called [`gdt`][gdt] which allows test authors to focus on clearly defining the
*assertions* that their tests should verify instead of cluttering tests with
lots of setup and implementation details.

`gdt` test cases are YAML files in the `test/` directory.

To run all functional end-to-end tests:

```
make test-e2e
```

To run just the functional tests that verify configuration artifacts are
installed properly:

```
make test-e2e-install
```

To run functional tests that verify that the Temporal Resource Operator properly
manages the lifecycle of Temporal Cloud resources:

```
make test-e2e-cluster
```

[gdt]: https://github.com/gdt-dev/gdt
