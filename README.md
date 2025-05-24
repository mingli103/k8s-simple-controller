# k8s-controller

This repository implements a simple controller for annotating a plugin to httproute resources
defined with a CustomResourceDefinition (CRD).

It uses kubebuilder to scaffold a project.

## Usage

To run the operator locally on your Kubernetes cluster:

```bash
make run
```

To build and release a Docker image for the operator:

```bash
make IMG=docker.io/mli1987/ratelimiedconsumer-operator docker-build docker-push
```

To deploy the operator to your Kubernetes cluster:

```bash
make IMG=docker.io/mli1987/ratelimiedconsumer-operator deploy
```

## Testing

To run unit tests:

```bash
make test
```
