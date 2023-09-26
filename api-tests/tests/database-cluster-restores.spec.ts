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
import { expect, test } from '@playwright/test'
import * as th from './helpers'

let kubernetesId

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes')

  kubernetesId = (await kubernetesList.json())[0].id
})

test('create/update/delete database cluster restore', async ({ request }) => {
  const bsName = th.suffixedName('storage')
  const clName = th.suffixedName('cluster')
  const clName2 = th.suffixedName('cluster2')
  const backupName = th.suffixedName('backup')

  await th.createBackupStorage(request, bsName)
  await th.createDBCluster(request, kubernetesId, clName)
  await th.createDBCluster(request, kubernetesId, clName2)
  await th.createBackup(request, kubernetesId, clName, backupName, bsName)

  const restoreName = th.suffixedName('restore')

  const payloadRestore = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseClusterRestore',
    metadata: {
      name: restoreName,
    },
    spec: {
      dataSource: {
        dbClusterBackupName: backupName,
      },
      dbClusterName: clName,
    },
  }

  let response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-restores`, {
    data: payloadRestore,
  })

  expect(response.ok()).toBeTruthy()
  const restore = await response.json()

  expect(restore.spec).toMatchObject(payloadRestore.spec)

  // update restore
  restore.spec.dbClusterName = clName2
  response = await request.put(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${restoreName}`, {
    data: restore,
  })
  expect(response.ok()).toBeTruthy()
  const result = await response.json()

  expect(result.spec).toMatchObject(restore.spec)

  // update restore with not existing dbClusterName
  restore.spec.dbClusterName = 'not-existing-cluster'
  response = await request.put(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${restoreName}`, {
    data: restore,
  })
  expect(response.status()).toBe(400)
  expect(await response.text()).toContain('{"message":"DatabaseCluster \'not-existing-cluster\' is not found"}')

  // delete restore
  await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${restoreName}`)
  // check it couldn't be found anymore
  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${restoreName}`)
  expect(response.status()).toBe(404)

  await th.deleteRestore(request, kubernetesId, restoreName)
  await th.deleteDBCluster(request, kubernetesId, clName)
  await th.deleteDBCluster(request, kubernetesId, clName2)
  await th.deleteBackup(request, kubernetesId, backupName)
  await th.deleteBackupStorage(request, bsName)
})

test('list restores', async ({ request, page }) => {
  const bsName = th.suffixedName('storage')
  const clName1 = th.suffixedName('cluster1')
  const clName2 = th.suffixedName('cluster2')
  const backupName = th.suffixedName('backup')

  await th.createBackupStorage(request, bsName)
  await th.createDBCluster(request, kubernetesId, clName1)
  await th.createDBCluster(request, kubernetesId, clName2)
  await th.createBackup(request, kubernetesId, clName1, backupName, bsName)

  const restoreName1 = th.suffixedName('restore1')
  const restoreName2 = th.suffixedName('restore2')
  const restoreName3 = th.suffixedName('restore3')

  const payloads = [
    {
      apiVersion: 'everest.percona.com/v1alpha1',
      kind: 'DatabaseClusterRestore',
      metadata: {
        name: restoreName1,
      },
      spec: {
        dataSource: {
          dbClusterBackupName: backupName,
        },
        dbClusterName: clName1,
      },
    },
    {
      apiVersion: 'everest.percona.com/v1alpha1',
      kind: 'DatabaseClusterRestore',
      metadata: {
        name: restoreName2,
      },
      spec: {
        dataSource: {
          dbClusterBackupName: backupName,
        },
        dbClusterName: clName1,
      },
    },
    {
      apiVersion: 'everest.percona.com/v1alpha1',
      kind: 'DatabaseClusterRestore',
      metadata: {
        name: restoreName3,
      },
      spec: {
        dataSource: {
          dbClusterBackupName: backupName,
        },
        dbClusterName: clName2,
      },
    },
  ]

  for (const payload of payloads) {
    const response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-restores`, {
      data: payload,
    })

    expect(response.ok()).toBeTruthy()
  }

  await page.waitForTimeout(6000)

  // check if the restores are available when being requested via database-clusters/{cluster-name}/restores path
  let response = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clName1}/restores`)
  let result = await response.json()

  expect(result.items).toHaveLength(2)

  response = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clName2}/restores`)
  result = await response.json()

  expect(result.items).toHaveLength(1)

  // delete the created restores
  for (const payload of payloads) {
    await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${payload.metadata.name}`)
    response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${payload.metadata.name}`)
    expect(response.status()).toBe(404)
  }

  await th.deleteBackup(request, kubernetesId, backupName)
  await th.deleteRestore(request, kubernetesId, restoreName1)
  await th.deleteRestore(request, kubernetesId, restoreName2)
  await th.deleteRestore(request, kubernetesId, restoreName3)
  await th.deleteDBCluster(request, kubernetesId, clName1)
  await th.deleteDBCluster(request, kubernetesId, clName2)
  await th.deleteBackupStorage(request, bsName)
})

test('create restore: validation errors', async ({ request, page }) => {
  const bsName = th.suffixedName('storage')
  const backupName = th.suffixedName('backup')
  const clName = th.suffixedName('cl')

  await th.createBackupStorage(request, bsName)
  await th.createDBCluster(request, kubernetesId, clName)
  await th.createBackup(request, kubernetesId, clName, backupName, bsName)

  // dbcluster not found
  const restoreName = th.suffixedName('restore')
  const payloadRestore = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseClusterRestore',
    metadata: {
      name: restoreName,
    },
    spec: {
      dataSource: {
        dbClusterBackupName: backupName,
      },
      dbClusterName: 'not-existing-cluster',
    },
  }

  let response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-restores`, {
    data: payloadRestore,
  })

  expect(response.status()).toBe(400)
  expect(await response.text()).toContain('{"message":"DatabaseCluster \'not-existing-cluster\' is not found"}')

  // empty spec
  const payloadEmptySpec = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseClusterRestore',
    metadata: {
      name: restoreName,
    },
  }

  response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-restores`, {
    data: payloadEmptySpec,
  })
  expect(response.status()).toBe(400)
  expect(await response.text()).toContain('{"message":"\'Spec\' field should not be empty"}')

  await th.deleteBackup(request, kubernetesId, backupName)
  await th.deleteRestore(request, kubernetesId, restoreName)
  await th.deleteBackupStorage(request, bsName)
  await th.deleteDBCluster(request, kubernetesId, clName)
  await th.deleteBackup(request, kubernetesId, backupName)
})
