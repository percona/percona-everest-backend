// percona-everest-backend
// Copyright (C) 2023 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { test, expect } from '@fixtures';

let kubernetesId;

test('check operators are installed', async ({ request, cli }) => {
  const kubernetesList = await request.get('/v1/kubernetes');

  kubernetesId = (await kubernetesList.json())[0].id;

  const enginesList = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines`);

  expect(enginesList.ok()).toBeTruthy();

  const engines = (await enginesList.json()).items;

  engines.forEach((engine) => {
    if (engine.spec.type === 'pxc') {
      expect(engine.status?.status).toBe('installed');
    }

    if (engine.spec.type === 'psmdb') {
      expect(engine.status?.status).toBe('installed');
    }
  });

  const output = await cli.execSilent('kubectl get pods --namespace=percona-everest');

  await output.outContainsNormalizedMany([
    'everest-operator-controller-manager',
  ]);
});

test('get/edit database engine versions', async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');

  kubernetesId = (await kubernetesList.json())[0].id;

  let engineResponse = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`);

  expect(engineResponse.ok()).toBeTruthy();

  const engineData = await engineResponse.json();
  const availableVersions = engineData.status.availableVersions;

  expect(availableVersions.engine['6.0.5-4'].imageHash).toBe('b6f875974c59d8ea0174675c85f41668460233784cbf2cbe7ce5eca212ac5f6a');
  expect(availableVersions.backup['2.0.5'].status).toBe('recommended');

  const allowedVersions = ['6.0.5-4', '6.0.4-3', '5.0.7-6'];

  delete engineData.status;
  engineData.spec.allowedVersions = allowedVersions;

  const updateResponse = await request.put(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`, {
    data: engineData,
  });

  expect(updateResponse.ok()).toBeTruthy();

  engineResponse = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`);
  expect(engineResponse.ok()).toBeTruthy();

  expect((await engineResponse.json()).spec.allowedVersions).toEqual(allowedVersions);
});
