import { test, expect } from '@playwright/test';

let kubernetesId;

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  kubernetesId = (await kubernetesList.json())[0].ID;
});

test('create/edit/delete single node cluster', async ({ request, page }) => {
  const clusterName = 'test-psmdb-cluster';
  const psmdbPayload = {
    apiVersion: 'dbaas.percona.com/v1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
      finalizers: ['delete-psmdb-pvc'], // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: 'psmdb',
      databaseImage: 'percona/percona-server-mongodb:4.4.10-11',
      databaseConfig: '',
      secretsName: 'test-psmdb-cluster-secrets',
      clusterSize: 1,
      loadBalancer: {
        type: 'mongos',
        exposeType: 'ClusterIP', // database cluster is not exposed by default
        size: 1, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
      },
      dbInstance: {
        cpu: '1',
        memory: '1G',
        diskSize: '15G',
      },
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: psmdbPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(1000);

    const psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(psmdbCluster.ok()).toBeTruthy();

    const result = (await psmdbCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(psmdbPayload.spec);
    expect(result.status.size).toBe(1);

    // psmdbPayload should be overriden because kubernetes adds data into metadata field
    // and uses metadata.generation during updation. It returns 422 HTTP status code if this field is not present
    //
    // kubectl under the hood merges everything hence the UX is seemless
    psmdbPayload.spec = result.spec;
    psmdbPayload.metadata = result.metadata;
    break;
  }

  psmdbPayload.spec.databaseConfig = 'operationProfiling:\nmode: slowOp';
  delete psmdbPayload.metadata.finalizers;

  // Update PSMDB cluster

  const updatedPSMDBCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: psmdbPayload });
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  let psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.ok()).toBeTruthy();

  expect((await updatedPSMDBCluster.json()).spec.databaseConfig).toBe(psmdbPayload.spec.databaseConfig);

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);
});

test('expose cluster after creation', async ({ request, page }) => {
  const clusterName = 'exposed-psmdb-cluster';
  const psmdbPayload = {
    apiVersion: 'dbaas.percona.com/v1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
      finalizers: ['delete-psmdb-pvc'], // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: 'psmdb',
      databaseImage: 'percona/percona-server-mongodb:4.4.10-11',
      databaseConfig: '',
      secretsName: 'test-psmdb-cluster-secrets',
      clusterSize: 3,
      loadBalancer: {
        type: 'mongos',
        exposeType: 'ClusterIP', // database cluster is not exposed by default
        size: 3, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
      },
      dbInstance: {
        cpu: '1',
        memory: '1G',
        diskSize: '15G',
      },
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: psmdbPayload,
  });

  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(1000);

    const psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(psmdbCluster.ok()).toBeTruthy();

    const result = (await psmdbCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(psmdbPayload.spec);
    expect(result.status.size).toBe(6);

    psmdbPayload.spec = result.spec;
    psmdbPayload.metadata = result.metadata;
    break;
  }

  psmdbPayload.spec.loadBalancer.type = 'LoadBalancer';
  delete psmdbPayload.metadata.finalizers;

  // Update PSMDB cluster

  const updatedPSMDBCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: psmdbPayload });
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  let psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.ok()).toBeTruthy();

  expect((await updatedPSMDBCluster.json()).spec.loadBalancer.type).toBe('LoadBalancer');

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);
});
test('expose cluster on EKS to the public internet and scale up', async ({ request, page }) => {
  const clusterName = 'eks-psmdb-cluster';
  const psmdbPayload = {
    apiVersion: 'dbaas.percona.com/v1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
      finalizers: ['delete-psmdb-pvc'], // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: 'psmdb',
      databaseImage: 'percona/percona-server-mongodb:4.4.10-11',
      databaseConfig: '',
      secretsName: 'test-psmdb-cluster-secrets',
      clusterSize: 3,
      loadBalancer: {
        type: 'mongos',
        exposeType: 'LoadBalancer', // database cluster is exposed
        size: 3, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
        annotations: {
          // Options below needs to be allied for exposed cluster on AWS infra
          'service.beta.kubernetes.io/aws-load-balancer-nlb-target-type': 'ip',
          'service.beta.kubernetes.io/aws-load-balancer-scheme': 'internet-facing',
          'service.beta.kubernetes.io/aws-load-balancer-target-group-attributes': 'preserve_client_ip.enabled=true',
          // This setting is required if the cluster needs to be exposed to the public internet (e.g internet facing)
          'service.beta.kubernetes.io/aws-load-balancer-type': 'external',
        },
      },
      dbInstance: {
        cpu: '1',
        memory: '1G',
        diskSize: '15G',
      },
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: psmdbPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(1000);

    const psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(psmdbCluster.ok()).toBeTruthy();

    const result = (await psmdbCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(psmdbPayload.spec);
    expect(result.status.size).toBe(6);

    psmdbPayload.spec = result.spec;
    psmdbPayload.metadata = result.metadata;
    break;
  }

  psmdbPayload.spec.clusterSize = 5;
  delete psmdbPayload.metadata.finalizers;

  // Update PSMDB cluster

  const updatedPSMDBCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: psmdbPayload });
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  let psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.ok()).toBeTruthy();

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  await page.waitForTimeout(1000);

  psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);
});
