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
import { test, expect } from '@playwright/test';

let kubernetesId;

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes');
  kubernetesId = (await kubernetesList.json())[0].id;

});


test('create/edit/delete database cluster backups', async ({ request }) => {
  let payload = {
    apiVersion: "everest.percona.com/v1alpha1",
    kind:"DatabaseClusterBackup",
    metadata:{
      name: "backup"
    },
    spec: {
      dbClusterName: "someCluster",
      objectStorageName: "someStorageName",
    }
  }

  let response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-backups`, {
    data: payload
  });
  expect(response.ok()).toBeTruthy();

  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`)
  let result = await response.json();
  expect(result.spec).toMatchObject(payload.spec);

  payload.spec.dbClusterName = "otherCluster";
  payload.metadata = result.metadata;
  response = await request.put(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`, {
    data: payload
  });
  expect(response.ok()).toBeTruthy();

  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`)
  result = await response.json();
  expect(result.spec).toMatchObject(payload.spec);

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`);
  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`)
  expect(response.status()).toBe(404);
});

