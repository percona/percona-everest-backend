## Welcome to Percona Everest Backend

Percona Everest is an open-source Database as a Service solution that automates day-1/day-2 operations for Postgres/MySQL/MongoDB databases running in Kubernetes clusters.

## Getting started 

Percona Everest has two main components that help you set up the environment:

1. [CLI](https://github.com/percona/percona-everest-cli) implements the installation of required components for Everest to work
2. Backend that implements DBaaS features

You can start playing with Everest using the following way

```sh
wget https://raw.githubusercontent.com/percona/percona-everest-backend/main/quickstart.yml
docker-compose -f quickstart.yml up -d
```
It will spin up the backend/frontend that will be accessible on the http://127.0.0.1:8080 address

## Creating Kubernetes cluster

You can try creating an EKS cluster using [DBaaS Infrastructure creator](https://percona.community/labs/dbaas-creator/). Also, it works on minikube/kind/k3d

## Installing everything needed into the Kubernetes cluster

To install all required operators in the headless mode you can run these commands

```
git clone git@github.com:percona/percona-everest-cli
cd percona-everest-cli
go run cmd/everest/main.go install operators --backup.enable=false --everest.endpoint=http://127.0.0.1:8080 --monitoring.enable=false --name=minikube --operator.mongodb=true --operator.postgresql=true --operator.xtradb-cluster=true --skip-wizard
```
You can also run it using the wizard:

```
✗ go run cmd/everest/main.go install operators
? Everest URL http://127.0.0.1:8080
? Choose your Kubernetes Cluster name k3d-everest-dev
? Do you want to enable monitoring? No
? Do you want to enable backups? No
? What operators do you want to install? MySQL, MongoDB, PostgreSQL
```
Once provisioning will be finished you can go to http://127.0.0.1:8080 and create your first database cluster!

## Known limitations

1. It supports only the basic creation of database clusters without monitoring integration and backup/restore support. We will add this support soon
2. It supports only one Kubernetes cluster on the user interface, however registering multiple Kubernetes clusters is possible.
3. No authentication/Access control features. You can configure it with the existing authentication solution you're using 
    * [Ambassador](https://github.com/datawire/ambassador) via
  [auth service](https://www.getambassador.io/reference/services/auth-service)
    * [Envoy](https://www.envoyproxy.io) via the
  [External Authorization HTTP Filter](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter.html)
    * AWS API Gateway via
  [Custom Authorizers](https://aws.amazon.com/de/blogs/compute/introducing-custom-authorizers-in-amazon-api-gateway/)
    * [Nginx](https://www.nginx.com) via
  [Authentication Based on Subrequest Result](https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-subrequest-authentication/)
