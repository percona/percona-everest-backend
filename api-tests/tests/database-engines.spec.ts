import { test, expect } from '@playwright/test';

const kubernetesId = "a0761de5-3ea8-4269-8d18-f2456c0167de";

test('install and check pxc', async({ request }) => {
  const enginesList = await request.get(`/kubernetes/${kubernetesId}/database-engines`);
  expect(enginesList.ok()).toBeTruthy();

  expect(await enginesList.json()).toContainEqual(expect.objectContaining({
    apiVersion: "dbaas.percona.com/v1"
  }));
});
