import { test, expect } from '@playwright/test';

let kubernetesId;

test('check operators are installed', async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  kubernetesId = (await kubernetesList.json())[0].id;

  const enginesList = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines`);

  expect(enginesList.ok()).toBeTruthy();

  const engines = (await enginesList.json()).items;
  engines.forEach((engine) => {
    if (engine.spec.type === 'pxc') {
      expect(engine.status.status).toBe('installed');
    }
    if (engine.spec.type === 'psmdb') {
      expect(engine.status.status).toBe('installed');
    }
  });
});

test('get/edit database engine versions', async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  kubernetesId = (await kubernetesList.json())[0].id;


  let engineResponse = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`);
  expect(engineResponse.ok()).toBeTruthy();


  let engineData = await engineResponse.json();
  const availableVersions = engineData.status.availableVersions;

  expect(availableVersions.engine['6.0.5-4'].imageHash).toBe('b6f875974c59d8ea0174675c85f41668460233784cbf2cbe7ce5eca212ac5f6a');
  expect(availableVersions.backup['2.0.5'].status).toBe('recommended');

  const allowedVersions = ['6.0.5-4', '6.0.4-3', '5.0.7-6'];
  delete engineData.status;
  engineData.spec.allowedVersions = allowedVersions;

  const updateResponse = await request.put(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`, {data: engineData});
  expect(updateResponse.ok()).toBeTruthy();

  engineResponse = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`);
  expect(engineResponse.ok()).toBeTruthy();

  expect((await engineResponse.json()).spec.allowedVersions).toEqual(allowedVersions);
});
