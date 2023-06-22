import { test, expect } from '@playwright/test';

test('install and check pxc', async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  const kubernetesId = (await kubernetesList.json())[0].id;

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
