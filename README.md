# k8s-simple-controller

This repository implements a simple controller for annotating Kong rate-limit plugins to HTTPRoute resources defined with a CustomResourceDefinition (CRD).

The controller watches `RateLimitedConsumer` custom resources and automatically annotates the target HTTPRoute with Kong rate-limit plugin configurations.

## API Group

The controller uses the API group: `ratelimit.test.annotation.com/v1alpha1`

## Prerequisites

- Kubernetes cluster with Gateway API CRDs installed
- cert-manager (for webhook TLS certificates in production)

## Installation

### 1. Install cert-manager (required for webhooks)

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.0/cert-manager.yaml
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=cert-manager -n cert-manager --timeout=120s
```

### 2. Install the CRD

```bash
make install
```

### 3. Deploy the controller

```bash
make deploy
```

## Usage

### Running Locally

To run the controller locally with webhooks disabled (for development):

```bash
ENABLE_WEBHOOKS=false make run
```

### Building and Deploying

To build and release a Docker image for the controller:

```bash
make IMG=docker.io/mli1987/ratelimiedconsumer-controller docker-build docker-push
```

To deploy the controller to your Kubernetes cluster:

```bash
make IMG=docker.io/mli1987/ratelimiedconsumer-controller deploy
```

### Example RateLimitedConsumer

```yaml
apiVersion: ratelimit.test.annotation.com/v1alpha1
kind: RateLimitedConsumer
metadata:
  name: example-ratelimitedconsumer
  namespace: default
spec:
  rateLimit:
    names: 
      - user-rate-limit
  targetRoute:
    name: my-httproute
```

## Architecture

### Cert-manager Integration

The controller uses cert-manager to automatically generate TLS certificates for webhooks:

1. **Issuer**: `selfsigned-issuer` - Generates self-signed certificates
2. **Certificate**: `serving-cert` - Requests certificates for webhook service
3. **Secret**: `webhook-server-cert` - Contains the TLS certificate and key

Cert-manager automatically:

- Watches for Certificate resources
- Generates certificates using the specified Issuer
- Creates the secret with proper TLS format
- Renews certificates before expiration

### Webhook Configuration

Webhooks are enabled by default in production deployments. For local development, disable them using:

```bash
ENABLE_WEBHOOKS=false make run
```

## Testing

To run unit tests:

```bash
make test
```

## Development

### Project Structure

- `api/v1alpha1/` - CRD definitions and types
- `internal/controller/` - Controller logic
- `internal/webhook/` - Webhook validation logic
- `config/` - Kubernetes manifests and kustomize configs

### Key Commands

```bash
# Generate manifests
make manifests

# Build the controller
make build

# Run locally without webhooks
ENABLE_WEBHOOKS=false make run

# Deploy to cluster
make deploy

# Uninstall from cluster
make uninstall
```
