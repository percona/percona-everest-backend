import { test, expect } from '@playwright/test';

let kubernetesId;
let recommendedVersion;

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  kubernetesId = (await kubernetesList.json())[0].id;


  const engineResponse = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines/percona-xtradb-cluster-operator`);
  const availableVersions =  (await engineResponse.json()).status.availableVersions.engine;

  for (const k in availableVersions) {
    if (k.startsWith('5')) {
      continue
    }
    if (availableVersions[k].status === 'recommended') {
      recommendedVersion = k
    }
  }
  expect(recommendedVersion).not.toBe('');
});

test('create/edit/delete pxc single node cluster', async ({ request, page }) => {
  const clusterName = 'test-pxc-cluster';
  const pxcPayload = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      engine: {
        type: 'pxc',
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
        type: 'haproxy', // HAProxy is the default option. However using proxySQL is available
        replicas: 1,
        expose: {
          type: 'internal',
        }
      }
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: pxcPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(1000);

    const pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(pxcCluster.ok()).toBeTruthy();

    const result = (await pxcCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(pxcPayload.spec);
    expect(result.status.size).toBe(2);

    // pxcPayload should be overriden because kubernetes adds data into metadata field
    // and uses metadata.generation during updation. It returns 422 HTTP status code if this field is not present
    //
    // kubectl under the hood merges everything hence the UX is seemless
    pxcPayload.spec = result.spec;
    pxcPayload.metadata = result.metadata;
    break;
  }

  pxcPayload.spec.engine.config = '[mysqld]\nwsrep_provider_options="debug=1;gcache.size=1G"\n';

  // Update PXC cluster

  const updatedPXCCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: pxcPayload });
  expect(updatedPXCCluster.ok()).toBeTruthy();

  let pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.ok()).toBeTruthy();

  expect((await updatedPXCCluster.json()).spec.databaseConfig).toBe(pxcPayload.spec.databaseConfig);

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.status()).toBe(404);
});

test('expose pxc cluster after creation', async ({ request, page }) => {
  const clusterName = 'exposed-pxc-cluster';
  const pxcPayload = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      engine: {
        type: 'pxc',
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
        type: 'haproxy', // HAProxy is the default option. However using proxySQL is available
        replicas: 3,
        expose: {
          type: 'internal',
        }
      }
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: pxcPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(1000);

    const pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(pxcCluster.ok()).toBeTruthy();

    const result = (await pxcCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(pxcPayload.spec);
    expect(result.status.size).toBe(6);

    pxcPayload.spec = result.spec;
    pxcPayload.metadata = result.metadata;
    break;
  }

  pxcPayload.spec.proxy.expose.type = 'external';

  // Update PXC cluster

  const updatedPXCCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: pxcPayload });
  expect(updatedPXCCluster.ok()).toBeTruthy();

  let pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.ok()).toBeTruthy();

  expect((await updatedPXCCluster.json()).spec.proxy.expose.type).toBe('external');

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.status()).toBe(404);
});

test('expose pxc cluster on EKS to the public internet and scale up', async ({ request, page }) => {
  test.setTimeout(60000);
  const clusterName = 'eks-pxc-cluster';
  const pxcPayload = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      engine: {
        type: 'pxc',
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
        type: 'haproxy', // HAProxy is the default option. However using proxySQL is available
        replicas: 3,
        expose: {
          type: 'external', // FIXME: Add Internetfacing once it'll be implemented
        }
      }
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: pxcPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(10000);

    const pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(pxcCluster.ok()).toBeTruthy();

    const result = (await pxcCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(pxcPayload.spec);
    expect(result.status.size).toBe(6);

    pxcPayload.spec = result.spec;
    pxcPayload.metadata = result.metadata;
    break;
  }

  pxcPayload.spec.engine.replicas = 5;

  // Update PXC cluster

  const updatedPXCCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: pxcPayload });
  expect(updatedPXCCluster.ok()).toBeTruthy();

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  await page.waitForTimeout(1000);

  const pxcCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pxcCluster.status()).toBe(404);
});
