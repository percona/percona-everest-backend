# Percona everest API integration tests

## Before running tests

Before running tests one needs to have provisioned kubernetes cluster, Everest backend

Running Percona Everest backend. Run these commands in the root of the project

```
   docker-compose up -d  # or make local-env-up
   go run cmd/main.go
```
Running minikube cluster

```
   make k8s
```
Provisioning kubernetes cluster

```
   git clone git@github.com:percona/percona-everest-cli
   cd percona-everest-cli
   go run cmd/everest/main.go install operators --enable_backup=false --everest.endpoint=http://127.0.0.1:8081 --install_olm=true --monitoring.enabled=false --name=minikube --operator.mongodb=true --operator.postgresql=true --operator.xtradb_cluster=true --skip-wizard
```
Using these commands you'll have installed the following operators

1. Postgres operator
2. PXC operator
3. PSMDB operator
4. DBaaS operator

After these commands you're ready to run integration tests

## Running integration tests
There are several ways running tests
```
  npx playwright test
    Runs the end-to-end tests.

  npx playwright test tests/test.spec.ts
    Runs the specific tests.

  npx playwright test --ui
    Starts the interactive UI mode.

  npx playwright test --debug
    Runs the tests in debug mode.

  npx playwright codegen
    Auto generate tests with Codegen.
```
