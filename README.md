## Welcome to Percona Everest Backend

Percona Everest is an open source Database-as-a-Service solution that automates day-one and day-two operations for Postgres, MySQL, and MongoDB databases within Kubernetes clusters.

## Prerequisites

A Kubernetes cluster is available for public use, but we do not offer support for creating one.

## Creating Kubernetes cluster

You must have a publicly accessible Kubernetes cluster to use Everest. EKS or GKE is recommended, as it may be difficult to make it work with local installations of Kubernetes such as minikube, kind, k3d, or similar products. Everest does not help with spinning up a Kubernetes cluster but assists with installing all the necessary components for Everest to run.


## Getting started

The Percona Everest has two primary components that assist you in creating the environment:

1. [CLI](https://github.com/percona/percona-everest-cli), which installs Everest's required components.
2. Backend, which installs DBaaS features.

To start using Everest, use the following commands:

```sh
wget https://raw.githubusercontent.com/percona/percona-everest-backend/main/quickstart.yml
docker-compose -f quickstart.yml up -d
```
This will spin up the backend/frontend, accessible at http://127.0.0.1:8080.



### Everest provisioning

1. Download the latest release of [everestctl](https://github.com/percona/percona-everest-cli/releases) command for your operating system 

2. Modify the permissions of the file:

  ```sh
  chmod +x everestctl
  ```

3. Run the following command to install all the required operators in headless mode:

  ```sh
   ./everestctl-darwin-amd64 install operators --backup.enable=false --everest.endpoint=http://127.0.0.1:8080 --monitoring.enable=false --operator.mongodb=true --operator.postgresql=true --operator.xtradb-cluster=true --skip-wizard
  ```

Alternatively, use the wizard to run it:

✗ ./everestctl install operators
? Everest URL http://127.0.0.1:8080
? Choose your Kubernetes Cluster name k3d-everest-dev
? Do you want to enable monitoring? No
? Do you want to enable backups? No
? What operators do you want to install? MySQL, MongoDB, PostgreSQL
```
Once provisioning is complete, you can visit http://127.0.0.1:8080 to create your first database cluster!

## Known limitations

- Currently, Everest only allows for the basic creation of database clusters without monitoring integration or backup/restore support. However, we will be adding this functionality in the near future.
- It is possible to register multiple Kubernetes clusters, but the user interface only supports one.
- There are no authentication or access control features, but you can integrate Everest with your existing solution.
    * [Ambassador](https://github.com/datawire/ambassador) via
  [auth service](https://www.getambassador.io/reference/services/auth-service)
    * [Envoy](https://www.envoyproxy.io) via the
  [External Authorization HTTP Filter](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter.html)
    * AWS API Gateway via
  [Custom Authorizers](https://aws.amazon.com/de/blogs/compute/introducing-custom-authorizers-in-amazon-api-gateway/)
    * [Nginx](https://www.nginx.com) via
  [Authentication Based on Subrequest Result](https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-subrequest-authentication/)
