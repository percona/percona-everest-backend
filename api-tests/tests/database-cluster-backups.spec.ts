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
import { test, expect } from '@playwright/test'

let kubernetesId

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes')

  kubernetesId = (await kubernetesList.json())[0].id
})

test('create/delete database cluster backups', async ({ request }) => {
  const payload = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseClusterBackup',
    metadata: {
      name: 'backup',
    },
    spec: {
      dbClusterName: 'someCluster',
      backupStorageName: 'someStorageName',
    },
  }

  let response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-backups`, {
    data: payload,
  })

  expect(response.ok()).toBeTruthy()

  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`)
  const result = await response.json()

  expect(result.spec).toMatchObject(payload.spec)

  await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`)
  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`)
  expect(response.status()).toBe(404)
})

test('list backups', async ({ request, page }) => {
  const payloads = [
    {
      apiVersion: 'everest.percona.com/v1alpha1',
      kind: 'DatabaseClusterBackup',
      metadata: {
        name: 'backup4',
      },
      spec: {
        dbClusterName: 'cluster1',
        backupStorageName: 'someStorageName',
      },
    },
    {
      apiVersion: 'everest.percona.com/v1alpha1',
      kind: 'DatabaseClusterBackup',
      metadata: {
        name: 'backup41',
      },
      spec: {
        dbClusterName: 'cluster1',
        backupStorageName: 'someStorageName',
      },
    },
    {
      apiVersion: 'everest.percona.com/v1alpha1',
      kind: 'DatabaseClusterBackup',
      metadata: {
        name: 'backup42',
      },
      spec: {
        dbClusterName: 'cluster2',
        backupStorageName: 'someStorageName',
      },
    },
    {
      apiVersion: 'everest.percona.com/v1alpha1',
      kind: 'DatabaseClusterBackup',
      metadata: {
        name: 'backup43',
      },
      spec: {
        dbClusterName: 'cluster2',
        backupStorageName: 'someStorageName',
      },
    },
  ]

  for (const payload of payloads) {
    const response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-backups`, {
      data: payload,
    })

    expect(response.ok()).toBeTruthy()
  }

  await page.waitForTimeout(1000)
  let response = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/cluster1/backups`)
  let result = await response.json()

  expect(result.items).toHaveLength(2)

  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/cluster2/backups`)
  result = await response.json()

  expect(result.items).toHaveLength(2)

  for (const payload of payloads) {
    await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/${payload.metadata.name}`)
    response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup`)
    expect(response.status()).toBe(404)
  }
})
