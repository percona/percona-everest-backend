import { test, expect } from '@playwright/test';

let kubernetesId;

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  kubernetesId = (await kubernetesList.json())[0].id;
});

test('create/edit/delete single node pg cluster', async ({ request, page }) => {
  const clusterName = 'test-pg-cluster';
  const pgPayload = {
    apiVersion: 'dbaas.percona.com/v1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
      finalizers: ['percona.com/delete-pvc'], // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: 'postgresql',
      databaseImage: 'percona/percona-postgresql-operator:2.0.0-ppg14-postgres',
      databaseConfig: '',
      secretsName: 'test-pg-cluster-secrets',
      clusterSize: 1,
      loadBalancer: {
        type: 'pgbouncer',
        exposeType: 'ClusterIP', // database cluster is not exposed by default
        size: 1, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
        image: 'percona/percona-postgresql-operator:2.0.0-ppg14-pgbouncer',
      },
      dbInstance: {
        cpu: '1',
        memory: '1G',
        diskSize: '15G',
      },
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: pgPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(1000);

    const pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(pgCluster.ok()).toBeTruthy();

    const result = (await pgCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(pgPayload.spec);
    expect(result.status.size).toBe(2);

    // pgPayload should be overriden because kubernetes adds data into metadata field
    // and uses metadata.generation during updation. It returns 422 HTTP status code if this field is not present
    //
    // kubectl under the hood merges everything hence the UX is seemless
    pgPayload.spec = result.spec;
    pgPayload.metadata = result.metadata;
    break;
  }

  pgPayload.spec.clusterSize = 3;
  delete pgPayload.metadata.finalizers;

  // Update PG cluster

  const updatedPGCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: pgPayload });
  expect(updatedPGCluster.ok()).toBeTruthy();

  let pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pgCluster.ok()).toBeTruthy();

  expect((await updatedPGCluster.json()).spec.clusterSize).toBe(pgPayload.spec.clusterSize);

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pgCluster.status()).toBe(404);
});

test('expose pg cluster after creation', async ({ request, page }) => {
  const clusterName = 'exposed-pg-cluster';
  const pgPayload = {
    apiVersion: 'dbaas.percona.com/v1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
      finalizers: ['percona.com/delete-pvc'], // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: 'postgresql',
      databaseImage: 'percona/percona-postgresql-operator:2.0.0-ppg14-postgres',
      databaseConfig: '',
      secretsName: 'test-pg-cluster-secrets',
      clusterSize: 1,
      loadBalancer: {
        type: 'pgbouncer',
        exposeType: 'ClusterIP', // database cluster is not exposed by default
        size: 1, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
        image: 'percona/percona-postgresql-operator:2.0.0-ppg14-pgbouncer',
      },
      dbInstance: {
        cpu: '1',
        memory: '1G',
        diskSize: '15G',
      },
    },
  };
  await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {
    data: pgPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(1000);

    const pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(pgCluster.ok()).toBeTruthy();

    const result = (await pgCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(pgPayload.spec);
    expect(result.status.size).toBe(2);

    pgPayload.spec = result.spec;
    pgPayload.metadata = result.metadata;
    break;
  }

  pgPayload.spec.loadBalancer.type = 'LoadBalancer';
  delete pgPayload.metadata.finalizers;

  // Update PG cluster

  const updatedPGCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: pgPayload });
  expect(updatedPGCluster.ok()).toBeTruthy();

  let pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pgCluster.ok()).toBeTruthy();

  expect((await updatedPGCluster.json()).spec.loadBalancer.type).toBe('LoadBalancer');

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);

  pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pgCluster.status()).toBe(404);
});

test('expose pg cluster on EKS to the public internet and scale up', async ({ request, page }) => {
  const clusterName = 'eks-pg-cluster';
  const pgPayload = {
    apiVersion: 'dbaas.percona.com/v1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
      finalizers: ['percona.com/delete-pvc'], // Required for the CI/CD workflows. For the end user we should keep unset, unless she set it explicitly
    },
    spec: {
      databaseType: 'postgresql',
      databaseImage: 'percona/percona-postgresql-operator:2.0.0-ppg14-postgres',
      databaseConfig: '',
      secretsName: 'test-pg-cluster-secrets',
      clusterSize: 3,
      loadBalancer: {
        type: 'pgbouncer',
        exposeType: 'LoadBalancer', // database cluster is not exposed by default
        size: 3, // Usually, a cluster size is equals to a load balancer set of nodes and any database nodes
        image: 'percona/percona-xtradb-cluster-operator:1.12.0-haproxy',
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
    data: pgPayload,
  });
  for (let i = 0; i < 15; i++) {
    await page.waitForTimeout(15000);

    const pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
    expect(pgCluster.ok()).toBeTruthy();

    const result = (await pgCluster.json());
    if (typeof result.status === 'undefined' || typeof result.status.size === 'undefined') {
      continue;
    }

    expect(result.metadata.name).toBe(clusterName);
    expect(result.spec).toMatchObject(pgPayload.spec);
    expect(result.status.size).toBe(6);

    pgPayload.spec = result.spec;
    pgPayload.metadata = result.metadata;
    break;
  }

  pgPayload.spec.clusterSize = 5;
  delete pgPayload.metadata.finalizers;

  // Update PG cluster

  const updatedPGCluster = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: pgPayload });
  expect(updatedPGCluster.ok()).toBeTruthy();

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  await page.waitForTimeout(1000);

  const pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`);
  expect(pgCluster.status()).toBe(404);
});
