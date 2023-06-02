## percona-everest-backend

This repo contains the Everest API server source code. It contains two type of methods: 
 - the "own" methods, e.g. register k8s cluster in everest, list the clusters
 - proxy methods for k8s API, which includes all resource-related methods (database-cluster, database-cluster-restore, database-engine)

The API server basic code id generated using [oapi-codegen](https://github.com/deepmap/oapi-codegen) from the docs/spec/openapi.yml file.

The proxy methods are aligned with the corresponding Everest operator methods, however they don't support all the original parameters since there is no need for them.
The definition of the custom resources can be found in the [Everest operator repo](https://github.com/percona/dbaas-operator/tree/main/config/crd/bases)

### Add a new proxy method
1. Copy the corresponding k8s spec to the [openapi.yml](./docs/spec/openapi.yml). Here is an [article](https://jonnylangefeld.com/blog/kubernetes-how-to-view-swagger-ui) about how to observe your cluster API, which will include the operator defined methods (if the operator is installed).
2. Make the spec modifications if needed. Things to keep in mind when designing new methods:
   - the [guidelines](https://opensource.zalando.com/restful-api-guidelines/) describes good practices
   - unlike the operator API the everest API uses kebab-case
   - consider what parameters should be exposed via the proxy method
2. Copy the custom resources schema (if needed) from the [Everest operator](https://github.com/percona/dbaas-operator/tree/main/config/crd/bases) config to the Components section of the [openapi.yml](./docs/spec/openapi.yml).
3. Run the code generation
```
 $ make init
 $ make gen
```
4. Implement the missing `ServerInterface` methods.
5. Run `make format` to format the code and group the imports.
6. Run `make check` to verify your code works and have no style violations.



