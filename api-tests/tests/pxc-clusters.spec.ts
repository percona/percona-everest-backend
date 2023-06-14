import { test, expect } from '@playwright/test';

const kubernetesId = "a0761de5-3ea8-4269-8d18-f2456c0167de";

test('create/edit/delete single node cluster', async({ request, page }) => {
  const clusterName = 'test-pxc-cluster';
  let pxcPayload =  {
    apiVersion: "dbaas.percona.com/v1",
    kind: "DatabaseCluster",
    metadata: {
      "name": clusterName,
      "finalizers": [ "delete-pxc-pvc" ] // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: "pxc",
      databaseImage: "percona/percona-xtradb-cluster:8.0.23-14.1",
      databaseConfig: "",
      secretsName: "test-pxc-cluster-secrets",
      clusterSize: 1,
      loadBalancer: {
        type: "haproxy", // HAProxy is the default option. However using proxySQL is available
        exposeType: "ClusterIP", // database cluster is not exposed by default
        size: 1, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
        image: "percona/percona-xtradb-cluster-operator:1.12.0-haproxy",
      },
      dbInstance: {
        cpu: "1",
        memory: "1G",
        diskSize: "15G"
      }
    }
  }
  await request.post(`/kubernetes/${kubernetesId}/database-clusters`, {
    data: pxcPayload
  });
  await page.waitForTimeout(5000);

  let pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (pxcCluster.ok()).toBeTruthy();

  let expected = (await pxcCluster.json());

  expect(expected.metadata.name).toBe("test-pxc-cluster");
  expect(expected.spec).toMatchObject(pxcPayload.spec);
  expect(expected.status.size).toBe(2);

  // pxcPayload should be overrided because kubernetes adds data into metadata field
  // and uses metadata.generation for the actual updation
  //
  // kubectl under the hood merges everything hence the UX is seemless
  pxcPayload = expected
  delete pxcPayload["status"]

  pxcPayload.spec.databaseConfig ="[mysqld]\nwsrep_provider_options=\"debug=1;gcache.size=1G\"\n"
  delete pxcPayload.metadata['finalizers']

  // Update PXC cluster

  let updatedPXCCluster = await request.put(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, {data: pxcPayload});
  expect(updatedPXCCluster.ok()).toBeTruthy();

  pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (pxcCluster.ok()).toBeTruthy();

  expected = (await pxcCluster.json());

  expect((await updatedPXCCluster.json()).spec.databaseConfig).toBe(pxcPayload.spec.databaseConfig);

  await request.delete(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.status()).toBe(404);


});
