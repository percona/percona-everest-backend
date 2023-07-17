import { test, expect } from '@playwright/test';

let kubernetesId;
let recommendedVersion;

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  kubernetesId = (await kubernetesList.json())[0].id;

  const engineResponse = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`);
  const availableVersions =  (await engineResponse.json()).status.availableVersions.engine;

  for (const k in availableVersions) {
    if (availableVersions[k].status === 'recommended' && k.startsWith('6')) {
      recommendedVersion = k
    }
  }
  expect(recommendedVersion).not.toBe('');
});

test('create/edit/delete single node psmdb cluster', async ({ request, page }) => {
  const clusterName = 'test-psmdb-cluster';
  const psmdbPayload = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      engine: {
        type: 'psmdb',
        replicas: 1,
        version: recommendedVersion,
        storage: {
          size: '25G'
        },
        resources: {
          cpu: '1',
          memory: '1G'
        }
      },
      proxy: {
        type: 'mongos', // HAProxy is the default option. However using proxySQL is available
        replicas: 1,
        expose: {
          type: 'internal',
        }
      }
    },
  }
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

  psmdbPayload.spec.engine.config = 'operationProfiling:\nmode: slowOp';

  // Update PSMDB cluster

  const updatedPSMDBCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: psmdbPayload });
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  let psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.ok()).toBeTruthy();

  expect((await updatedPSMDBCluster.json()).spec.engine.config).toBe(psmdbPayload.spec.engine.config);

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);
});

test('expose psmdb cluster after creation', async ({ request, page }) => {
  const clusterName = 'exposed-psmdb-cluster';
  const psmdbPayload = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      engine: {
        type: 'psmdb',
        replicas: 3,
        version: recommendedVersion,
        storage: {
          size: '25G'
        },
        resources: {
          cpu: '1',
          memory: '1G'
        }
      },
      proxy: {
        type: 'mongos', // HAProxy is the default option. However using proxySQL is available
        replicas: 3,
        expose: {
          type: 'internal',
        }
      }
    },
  }
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

  psmdbPayload.spec.proxy.expose.type = 'external';

  // Update PSMDB cluster

  const updatedPSMDBCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: psmdbPayload });
  expect(updatedPSMDBCluster.ok()).toBeTruthy();

  let psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.ok()).toBeTruthy();

  expect((await updatedPSMDBCluster.json()).spec.proxy.expose.type).toBe('external');

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  psmdbCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(psmdbCluster.status()).toBe(404);
});

test('expose psmdb cluster on EKS to the public internet and scale up', async ({ request, page }) => {
  const clusterName = 'eks-psmdb-cluster';
  const psmdbPayload = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      engine: {
        type: 'psmdb',
        replicas: 3,
        version: recommendedVersion,
        storage: {
          size: '25G'
        },
        resources: {
          cpu: '1',
          memory: '1G'
        }
      },
      proxy: {
        type: 'mongos', // HAProxy is the default option. However using proxySQL is available
        replicas: 3,
        expose: {
          type: 'external', // FIXME: Add internetfacing once it'll be implemented
        }
      }
    },
  }
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: psmdbPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(2000);

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

  psmdbPayload.spec.engine.replicas = 5;

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
