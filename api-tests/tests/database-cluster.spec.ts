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
import { expect, test } from '@fixtures'

// testPrefix is used to differentiate between several workers
// running this test to avoid conflicts in instance names
const testPrefix = `${(Math.random() + 1).toString(36).substring(10)}`

let kubernetesId
let recommendedVersion
const monitoringConfigName1 = `a${testPrefix}-1`
const monitoringConfigName2 = `b${testPrefix}-2`

test.setTimeout(360 * 1000)

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes')

  kubernetesId = (await kubernetesList.json())[0].id

  const engineResponse = await request.get(`/v1/kubernetes/${kubernetesId}/database-engines/percona-server-mongodb-operator`)
  const availableVersions = (await engineResponse.json()).status.availableVersions.engine

  for (const k in availableVersions) {
    if (availableVersions[k].status === 'recommended' && k.startsWith('6')) {
      recommendedVersion = k
    }
  }

  expect(recommendedVersion).not.toBe('')

  const miData = {
    type: 'pmm',
    name: monitoringConfigName1,
    url: 'http://monitoring',
    pmm: {
      apiKey: '123',
    },
  }

  // Monitoring configs
  let res = await request.post('/v1/monitoring-instances', { data: miData })

  expect(res.ok()).toBeTruthy()

  miData.name = monitoringConfigName2
  res = await request.post('/v1/monitoring-instances', { data: miData })
  expect(res.ok()).toBeTruthy()
})

test.afterAll(async ({ request }) => {
  let res = await request.delete(`/v1/monitoring-instances/${monitoringConfigName1}`)
  console.log(await res.text())

  expect(res.ok()).toBeTruthy()

  res = await request.delete(`/v1/monitoring-instances/${monitoringConfigName2}`)
  console.log(await res.text())
  expect(res.ok()).toBeTruthy()
})

test('create db cluster with monitoring config', async ({ request }) => {
  const clusterName = 'db-monitoring-create'
  const data = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      monitoring: {
        monitoringConfigName: monitoringConfigName1,
      },
      engine: {
        type: 'psmdb',
        replicas: 1,
        version: recommendedVersion,
        storage: {
          size: '4G',
        },
        resources: {
          cpu: '1',
          memory: '1G',
        },
      },
      proxy: {
        type: 'mongos',
        replicas: 1,
        expose: {
          type: 'internal',
        },
      },
    },
  }

  const postReq = await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, { data })

  expect(postReq.ok()).toBeTruthy()

  try {
    await expect(async () => {
      const pgCluster = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)

      expect(pgCluster.ok()).toBeTruthy()
      const res = (await pgCluster.json())

      expect(res?.status?.size).toBeGreaterThanOrEqual(1)
    }).toPass({
      intervals: [1000],
      timeout: 60 * 1000,
    })
  } finally {
    await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)
  }
})

test('update db cluster with a new monitoring config', async ({ request }) => {
  const clusterName = 'dbc-monitoring-put'
  const data = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      monitoring: {
        monitoringConfigName: monitoringConfigName1,
      },
      engine: {
        type: 'psmdb',
        replicas: 1,
        version: recommendedVersion,
        storage: {
          size: '4G',
        },
        resources: {
          cpu: '1',
          memory: '1G',
        },
      },
      proxy: {
        type: 'mongos',
        replicas: 1,
        expose: {
          type: 'internal',
        },
      },
    },
  }

  const postReq = await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, { data })
  expect(postReq.ok()).toBeTruthy()

  try {
    let res

    await expect(async () => {
      const req = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)

      expect(req.ok()).toBeTruthy()
      res = (await req.json())
      expect(res?.status?.size).toBeGreaterThanOrEqual(1)
    }).toPass({
      intervals: [1000],
      timeout: 60 * 1000,
    })

    expect(res?.spec?.monitoring?.monitoringConfigName).toBe(monitoringConfigName1)

    const putData = data

    putData.metadata = res.metadata
    putData.spec.monitoring.monitoringConfigName = monitoringConfigName2

    const putReq = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: putData })

    expect(putReq.ok()).toBeTruthy()
    res = (await putReq.json())
    expect(res?.spec?.monitoring?.monitoringConfigName).toBe(monitoringConfigName2)
  } finally {
    await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)
  }
})

test('update db cluster without monitoring config with a new monitoring config', async ({ request }) => {
  const clusterName = 'monitoring-put-empty'
  const data = {
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
          size: '4G',
        },
        resources: {
          cpu: '1',
          memory: '1G',
        },
      },
      proxy: {
        type: 'mongos',
        replicas: 1,
        expose: {
          type: 'internal',
        },
      },
    },
  }

  const postReq = await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, { data })

  expect(postReq.ok()).toBeTruthy()

  try {
    let res

    await expect(async () => {
      const req = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)

      expect(req.ok()).toBeTruthy()
      res = (await req.json())
      expect(res?.status?.size).toBeGreaterThanOrEqual(1)
    }).toPass({
      intervals: [1000],
      timeout: 60 * 1000,
    })

    expect(res?.spec?.monitoring?.monitoringConfigName).toBeFalsy()

    const putData = data

    putData.metadata = res.metadata;
    (putData.spec as any).monitoring = { monitoringConfigName: monitoringConfigName2 }

    const putReq = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: putData })

    expect(putReq.ok()).toBeTruthy()
    res = (await putReq.json())
    expect(res?.spec?.monitoring?.monitoringConfigName).toBe(monitoringConfigName2)
  } finally {
    await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)
  }
})

test('update db cluster monitoring config with an empty monitoring config', async ({ request }) => {
  const clusterName = 'monit-put-to-empty'
  const data = {
    apiVersion: 'everest.percona.com/v1alpha1',
    kind: 'DatabaseCluster',
    metadata: {
      name: clusterName,
    },
    spec: {
      monitoring: {
        monitoringConfigName: monitoringConfigName1,
      },
      engine: {
        type: 'psmdb',
        replicas: 1,
        version: recommendedVersion,
        storage: {
          size: '4G',
        },
        resources: {
          cpu: '1',
          memory: '1G',
        },
      },
      proxy: {
        type: 'mongos',
        replicas: 1,
        expose: {
          type: 'internal',
        },
      },
    },
  }

  const postReq = await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, { data })
  expect(postReq.ok()).toBeTruthy()

  try {
    let res

    await expect(async () => {
      const req = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)

      expect(req.ok()).toBeTruthy()
      res = (await req.json())
      expect(res?.status?.size).toBeGreaterThanOrEqual(1)
    }).toPass({
      intervals: [1000],
      timeout: 60 * 1000,
    })

    const putData = data

    putData.metadata = res.metadata;
    (putData.spec.monitoring as any) = {}

    const putReq = await request.put(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`, { data: putData })

    expect(putReq.ok()).toBeTruthy()
    res = (await putReq.json())
    expect(res?.spec?.monitoring?.monitoringConfigName).toBeFalsy()
  } finally {
    await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${clusterName}`)
  }
})
