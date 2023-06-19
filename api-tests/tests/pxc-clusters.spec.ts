import { test, expect } from '@playwright/test';

let kubernetesId;

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get(`/kubernetes`);
  kubernetesId = (await kubernetesList.json())[0].ID;

});
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

  expect(expected.metadata.name).toBe(clusterName);
  expect(expected.spec).toMatchObject(pxcPayload.spec);
  expect(expected.status.size).toBe(2);

  // pxcPayload should be overriden because kubernetes adds data into metadata field
  // and uses metadata.generation during updation. It returns 422 HTTP status code if this field is not present
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

test('expose cluster after creation', async({ request, page }) => {
  const clusterName = 'exposed-pxc-cluster';
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
      clusterSize: 3,
      loadBalancer: {
        type: "haproxy", // HAProxy is the default option. However using proxySQL is available
        exposeType: "ClusterIP", // database cluster is not exposed by default
        size: 3, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
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
  await page.waitForTimeout(6000);

  let pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (pxcCluster.ok()).toBeTruthy();

  let expected = (await pxcCluster.json());

  expect(expected.metadata.name).toBe(clusterName);
  expect(expected.spec).toMatchObject(pxcPayload.spec);
  expect(expected.status.size).toBe(6);

  pxcPayload = expected
  delete pxcPayload["status"]

  pxcPayload.spec.loadBalancer.type = "LoadBalancer"
  delete pxcPayload.metadata['finalizers']

  // Update PXC cluster

  let updatedPXCCluster = await request.put(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, {data: pxcPayload});
  expect(updatedPXCCluster.ok()).toBeTruthy();

  pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (pxcCluster.ok()).toBeTruthy();

  expected = (await pxcCluster.json());

  expect((await updatedPXCCluster.json()).spec.loadBalancer.type).toBe("LoadBalancer");

  await request.delete(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.status()).toBe(404);


});
test('expose cluster on EKS to the public internet and scale up', async({ request, page }) => {
  const clusterName = 'eks-pxc-cluster';
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
      clusterSize: 3,
      loadBalancer: {
        type: "haproxy", // HAProxy is the default option. However using proxySQL is available
        exposeType: "LoadBalancer", // database cluster is exposed
        size: 3, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
        image: "percona/percona-xtradb-cluster-operator:1.12.0-haproxy",
        annotations: {
          // Options below needs to be allied for exposed cluster on AWS infra
          "service.beta.kubernetes.io/aws-load-balancer-nlb-target-type": "ip",
          "service.beta.kubernetes.io/aws-load-balancer-scheme": "internet-facing",
          "service.beta.kubernetes.io/aws-load-balancer-target-group-attributes": "preserve_client_ip.enabled=true",
          // This setting is required if the cluster needs to be exposed to the public internet (e.g internet facing)
          "service.beta.kubernetes.io/aws-load-balancer-type": "external"
        }
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
  await page.waitForTimeout(7000);

  let pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (pxcCluster.ok()).toBeTruthy();

  let expected = (await pxcCluster.json());

  expect(expected.metadata.name).toBe(clusterName);
  expect(expected.spec).toMatchObject(pxcPayload.spec);
  expect(expected.status.size).toBe(6);

  pxcPayload = expected
  delete pxcPayload["status"]

  pxcPayload.spec.clusterSize = 5
  delete pxcPayload.metadata['finalizers']

  // Update PXC cluster

  let updatedPXCCluster = await request.put(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, {data: pxcPayload});
  expect(updatedPXCCluster.ok()).toBeTruthy();

  pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (pxcCluster.ok()).toBeTruthy();

  expected = (await pxcCluster.json());

  expect((await updatedPXCCluster.json()).status.size).toBe(8);

  await request.delete(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  pxcCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.status()).toBe(404);


});
