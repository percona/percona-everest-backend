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
import { test, expect } from '@fixtures'

// testPrefix is used to differentiate between several workers
// running this test to avoid conflicts in instance names
const testPrefix = `${Date.now()}-${process.env.TEST_WORKER_INDEX}`

let kubernetesId

test.beforeAll(async ({ request }) => {
  const kubernetesList = await request.get('/v1/kubernetes')

  kubernetesId = (await kubernetesList.json())[0].id
})

test('get resource usage', async ({ request }) => {
  const r = await request.get(`/v1/kubernetes/${kubernetesId}/resources`)
  const resources = await r.json()

  expect(r.ok()).toBeTruthy()

  expect(resources).toBeTruthy()

  expect(resources?.capacity).toBeTruthy()
  expect(resources?.available).toBeTruthy()
})

test('enable/disable cluster-monitoring', async ({ request }) => {
  const data = {
    type: 'pmm',
    name: `${testPrefix}-cluster-monitoring`,
    url: 'http://monitoring',
    pmm: {
      apiKey: '123',
    },
  }

  const response = await request.post('/v1/monitoring-instances', { data })

  expect(response.ok()).toBeTruthy()
  const created = await response.json()

  const rEnable = await request.post(`/v1/kubernetes/${kubernetesId}/cluster-monitoring`, {
    data: {
      enable: true,
      monitoringInstance: created.name,
    },
  })

  expect(rEnable.ok()).toBeTruthy()

  const rDisable = await request.post(`/v1/kubernetes/${kubernetesId}/cluster-monitoring`, {
    data: { enable: false },
  })

  expect(rDisable.ok()).toBeTruthy()
})
