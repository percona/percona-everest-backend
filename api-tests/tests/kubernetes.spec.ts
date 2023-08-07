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

test('get resource usage', async ({ request, page }) => {
  const r = await request.get(`/v1/kubernetes/${kubernetesId}/resources`, {})
  const resources = await r.json()

  expect(resources).toBeTruthy()
  
  expect(resources?.capacity).toHaveProperty('cpuMillis')
  expect(resources?.capacity).toHaveProperty('memoryBytes')
  expect(resources?.capacity).toHaveProperty('diskSize')

  expect(resources?.available).toHaveProperty('cpuMillis')
  expect(resources?.available).toHaveProperty('memoryBytes')
  expect(resources?.available).toHaveProperty('diskSize')
})
