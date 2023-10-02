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
import {expect, test} from '@fixtures'
import * as th from './helpers'


// testPrefix is used to differentiate between several workers
// running this test to avoid conflicts in instance names
const testPrefix = `${(Math.random() + 1).toString(36).substring(10)}`

let kubernetesId

test.beforeAll(async ({request}) => {
    const kubernetesList = await request.get('/v1/kubernetes')

    kubernetesId = (await kubernetesList.json())[0].id
})

test('get resource usage', async ({request}) => {
    const r = await request.get(`/v1/kubernetes/${kubernetesId}/resources`)
    const resources = await r.json()

    expect(r.ok()).toBeTruthy()

    expect(resources).toBeTruthy()

    expect(resources?.capacity).toBeTruthy()
    expect(resources?.available).toBeTruthy()
})

test('enable/disable cluster-monitoring', async ({request}) => {
    const name = th.randomName()
    const data = {
        type: 'pmm',
        name: name,
        url: 'http://monitoring',
        pmm: {
            apiKey: '123',
        },
    }

    const response = await request.post('/v1/monitoring-instances', {data})

    expect(response.ok()).toBeTruthy()
    const created = await response.json()

    const rEnable = await request.post(`/v1/kubernetes/${kubernetesId}/cluster-monitoring`, {
        data: {
            enable: true,
            monitoringInstanceName: created.name,
        },
    })

    expect(rEnable.ok()).toBeTruthy()

    const rDisable = await request.post(`/v1/kubernetes/${kubernetesId}/cluster-monitoring`, {
        data: {enable: false},
    })

    expect(rDisable.ok()).toBeTruthy()

    await th.deleteMonitoringInstance(request, name)
})

test('get cluster info', async ({request}) => {
    const r = await request.get(`/v1/kubernetes/${kubernetesId}/cluster-info`)
    const info = await r.json()

    expect(r.ok()).toBeTruthy()

    expect(info).toBeTruthy()

    expect(info?.clusterType).toBeTruthy()
    expect(info?.storageClassNames).toBeTruthy()
    expect(info?.storageClassNames).toHaveLength(1)
})
