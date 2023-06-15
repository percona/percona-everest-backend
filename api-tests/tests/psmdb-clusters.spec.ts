import { test, expect } from '@playwright/test';

let kubernetesId;

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get(`/kubernetes`);
  kubernetesId = (await kubernetesList.json())[0].ID;

});

test('create/edit/delete single node cluster', async({ request, page }) => {
  const clusterName = 'test-psmdb-cluster';
  let psmdbPayload =  {
    apiVersion: "dbaas.percona.com/v1",
    kind: "DatabaseCluster",
    metadata: {
      "name": clusterName,
      "finalizers": [ "delete-psmdb-pvc" ] // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: "psmdb",
      databaseImage: "percona/percona-server-mongodb:4.4.10-11",
      databaseConfig: "",
      secretsName: "test-psmdb-cluster-secrets",
      clusterSize: 1,
      loadBalancer: {
        type: "mongos",
        exposeType: "ClusterIP", // database cluster is not exposed by default
        size: 1, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
      },
      dbInstance: {
        cpu: "1",
        memory: "1G",
        diskSize: "15G"
      }
    }
  }
  await request.post(`/kubernetes/${kubernetesId}/database-clusters`, {
    data: psmdbPayload
  });
  await page.waitForTimeout(5000);

  let psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (psmdbCluster.ok()).toBeTruthy();

  let expected = (await psmdbCluster.json());

  expect(expected.metadata.name).toBe(clusterName);
  expect(expected.spec).toMatchObject(psmdbPayload.spec);
  expect(expected.status.size).toBe(1);

  // psmdbPayload should be overriden because kubernetes adds data into metadata field
  // and uses metadata.generation during updation. It returns 422 HTTP status code if this field is not present
  //
  // kubectl under the hood merges everything hence the UX is seemless
  psmdbPayload = expected
  delete psmdbPayload["status"]

  psmdbPayload.spec.databaseConfig ="[mysqld]\nwsrep_provider_options=\"debug=1;gcache.size=1G\"\n"
  delete psmdbPayload.metadata['finalizers']

  // Update PSMDB cluster

  let updatedPSMDBCluster = await request.put(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, {data: psmdbPayload});
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (psmdbCluster.ok()).toBeTruthy();

  expected = (await psmdbCluster.json());

  expect((await updatedPSMDBCluster.json()).spec.databaseConfig).toBe(psmdbPayload.spec.databaseConfig);

  await request.delete(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);


});

test('expose cluster after creation', async({ request, page }) => {
  const clusterName = 'exposed-psmdb-cluster';
  let psmdbPayload =  {
    apiVersion: "dbaas.percona.com/v1",
    kind: "DatabaseCluster",
    metadata: {
      "name": clusterName,
      "finalizers": [ "delete-psmdb-pvc" ] // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: "psmdb",
      databaseImage: "percona/percona-server-mongodb:4.4.10-11",
      databaseConfig: "",
      secretsName: "test-psmdb-cluster-secrets",
      clusterSize: 3,
      loadBalancer: {
        type: "mongos",
        exposeType: "ClusterIP", // database cluster is not exposed by default
        size: 3, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
      },
      dbInstance: {
        cpu: "1",
        memory: "1G",
        diskSize: "15G"
      }
    }
  }
  await request.post(`/kubernetes/${kubernetesId}/database-clusters`, {
    data: psmdbPayload
  });
  await page.waitForTimeout(6000);

  let psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (psmdbCluster.ok()).toBeTruthy();

  let expected = (await psmdbCluster.json());

  expect(expected.metadata.name).toBe(clusterName);
  expect(expected.spec).toMatchObject(psmdbPayload.spec);
  expect(expected.status.size).toBe(6);

  psmdbPayload = expected
  delete psmdbPayload["status"]

  psmdbPayload.spec.loadBalancer.type = "LoadBalancer"
  delete psmdbPayload.metadata['finalizers']

  // Update PSMDB cluster

  let updatedPSMDBCluster = await request.put(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, {data: psmdbPayload});
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (psmdbCluster.ok()).toBeTruthy();

  expected = (await psmdbCluster.json());

  expect((await updatedPSMDBCluster.json()).spec.loadBalancer.type).toBe("LoadBalancer");

  await request.delete(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);


});
test('expose cluster on EKS to the public internet and scale up', async({ request, page }) => {
  const clusterName = 'eks-psmdb-cluster';
  let psmdbPayload =  {
    apiVersion: "dbaas.percona.com/v1",
    kind: "DatabaseCluster",
    metadata: {
      "name": clusterName,
      "finalizers": [ "delete-psmdb-pvc" ] // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: "psmdb",
      databaseImage: "percona/percona-server-mongodb:4.4.10-11",
      databaseConfig: "",
      secretsName: "test-psmdb-cluster-secrets",
      clusterSize: 3,
      loadBalancer: {
        type: "mongos",
        exposeType: "LoadBalancer", // database cluster is exposed
        size: 3, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
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
    data: psmdbPayload
  });
  await page.waitForTimeout(7000);

  let psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (psmdbCluster.ok()).toBeTruthy();

  let expected = (await psmdbCluster.json());

  expect(expected.metadata.name).toBe(clusterName);
  expect(expected.spec).toMatchObject(psmdbPayload.spec);
  expect(expected.status.size).toBe(6);

  psmdbPayload = expected
  delete psmdbPayload["status"]

  psmdbPayload.spec.clusterSize = 5
  delete psmdbPayload.metadata['finalizers']

  // Update PSMDB cluster

  let updatedPSMDBCluster = await request.put(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, {data: psmdbPayload});
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect (psmdbCluster.ok()).toBeTruthy();

  expected = (await psmdbCluster.json());

  expect((await updatedPSMDBCluster.json()).status.size).toBe(8);

  await request.delete(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  psmdbCluster = await request.get(`/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);


});
