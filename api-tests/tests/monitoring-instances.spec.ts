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
import http from 'http'
import {expect, test} from '@fixtures'
import * as th from './helpers'
import {createMonitoringInstance, createMonitoringInstances} from './helpers'

test('create monitoring instance with api key', async ({request}) => {
    let name = th.randomName()
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

    expect(created.name).toBe(data.name)
    expect(created.url).toBe(data.url)
    expect(created.type).toBe(data.type)

    await th.deleteMonitoringInstance(request, name)
})

test('create monitoring instance with user/password', async ({request}) => {
    const server = http.createServer((_, res) => {
        res.statusCode = 200
        res.setHeader('Content-Type', 'application/json')
        res.end(JSON.stringify({key: 'test-api-key'}))
    })

    try {
        let s

        await new Promise<void>((resolve) => {
            s = server.listen(0, '127.0.0.1', () => resolve())
        })

        const port = s.address()?.port
        let name = th.randomName()
        const data = {
            type: 'pmm',
            name: name,
            url: `http://127.0.0.1:${port}`,
            pmm: {
                user: 'admin',
                password: 'admin',
            },
        }

        const response = await request.post('/v1/monitoring-instances', {data})

        expect(response.ok()).toBeTruthy()
        const created = await response.json()

        expect(created.name).toBe(data.name)
        expect(created.url).toBe(data.url)
        expect(created.type).toBe(data.type)
        await th.deleteMonitoringInstance(request, name)
    } finally {
        server.closeAllConnections()
        await new Promise<void>((resolve) => server.close(() => resolve()))
    }
})

test('create monitoring instance with user/password cannot connect to PMM', async ({request}) => {
    const server = http.createServer((_, res) => {
        res.statusCode = 404
        res.setHeader('Content-Type', 'application/json')
        res.end('{}')
    })

    try {
        let s

        await new Promise<void>((resolve) => {
            s = server.listen(0, '127.0.0.1', () => resolve())
        })

        const port = s.address()?.port
        const data = {
            type: 'pmm',
            name: 'monitoring-fail',
            url: `http://127.0.0.1:${port}`,
            pmm: {
                user: 'admin',
                password: 'admin',
            },
        }

        const response = await request.post('/v1/monitoring-instances', {data})

        expect(response.status()).toBe(400)
    } finally {
        server.closeAllConnections()
        await new Promise<void>((resolve) => server.close(() => resolve()))
    }
})

test('create monitoring instance missing pmm', async ({request}) => {
    const data = {
        type: 'pmm',
        name: 'monitoring-fail',
        url: 'http://monitoring-instance',
    }

    const response = await request.post('/v1/monitoring-instances', {data})

    expect(response.status()).toBe(400)
})

test('create monitoring instance missing pmm credentials', async ({request}) => {
    const data = {
        type: 'pmm',
        name: 'monitoring-fail',
        url: 'http://monitoring-instance',
        pmm: {},
    }

    const response = await request.post('/v1/monitoring-instances', {data})

    expect(response.status()).toBe(400)
})

test('list monitoring instances', async ({request}) => {
    const testPrefix = th.randomName()

    const names = await th.createMonitoringInstances(request, testPrefix)

    const response = await request.get('/v1/monitoring-instances')

    expect(response.ok()).toBeTruthy()
    const list = await response.json()

    expect(list.filter((i) => i.name.startsWith(`${testPrefix}`)).length).toBe(3)


    await th.deleteMonitoringInstances(request, names)
})

test('get monitoring instance', async ({request}) => {
    const name = await th.createMonitoringInstance(request)

    const response = await request.get(`/v1/monitoring-instances/${name}`)

    expect(response.ok()).toBeTruthy()
    const i = await response.json()

    expect(i.name).toBe(name)
    await th.deleteMonitoringInstance(request, name)
})

test('delete monitoring instance', async ({request}) => {
    const testPrefix = th.randomName()
    const names = await createMonitoringInstances(request, testPrefix)
    // we had a bug in the implementation where delete would delete the first instance, not the one selected by name
    const name = names[1]

    let response = await request.get('/v1/monitoring-instances')
    let list = await response.json()

    expect(list.filter((i) => i.name.startsWith(`${testPrefix}`)).length).toBe(3)

    response = await request.delete(`/v1/monitoring-instances/${name}`)
    expect(response.ok()).toBeTruthy()

    response = await request.get(`/v1/monitoring-instances/${name}`)
    expect(response.status()).toBe(404)

    response = await request.get('/v1/monitoring-instances')
    list = await response.json()

    expect(list.filter((i) => i.name.startsWith(`${testPrefix}`)).length).toBe(2)

    await th.deleteMonitoringInstances(request, names.filter((n) => n != name))
})

test('patch monitoring instance', async ({request}) => {
    const name = await createMonitoringInstance(request)

    const response = await request.get(`/v1/monitoring-instances/${name}`)

    expect(response.ok()).toBeTruthy()
    const created = await response.json()

    const patchData = {url: 'http://monitoring'}
    const updated = await request.patch(`/v1/monitoring-instances/${name}`, {data: patchData})

    expect(updated.ok()).toBeTruthy()
    const getJson = await updated.json()

    expect(getJson.url).toBe(patchData.url)
    expect(getJson.apiKeySecretId).toBe(created.apiKeySecretId)
    await th.deleteMonitoringInstance(request, name)
})

test('patch monitoring instance secret key changes', async ({request}) => {
    const name = await createMonitoringInstance(request)

    const response = await request.get(`/v1/monitoring-instances/${name}`)

    expect(response.ok()).toBeTruthy()

    const patchData = {
        url: 'http://monitoring2',
        pmm: {
            apiKey: 'asd',
        },
    }
    const updated = await request.patch(`/v1/monitoring-instances/${name}`, {data: patchData})

    expect(updated.ok()).toBeTruthy()
    const getJson = await updated.json()

    expect(getJson.url).toBe(patchData.url)
    await th.deleteMonitoringInstance(request, name)
})

test('patch monitoring instance type updates properly', async ({request}) => {
    const name = await createMonitoringInstance(request)

    const response = await request.get(`/v1/monitoring-instances/${name}`)

    expect(response.ok()).toBeTruthy()

    const patchData = {
        type: 'pmm',
        pmm: {
            apiKey: 'asd',
        },
    }
    const updated = await request.patch(`/v1/monitoring-instances/${name}`, {data: patchData})

    expect(updated.ok()).toBeTruthy()
    await th.deleteMonitoringInstance(request, name)
})

test('patch monitoring instance type fails on missing key', async ({request}) => {
    const name = await createMonitoringInstance(request)

    const response = await request.get(`/v1/monitoring-instances/${name}`)

    expect(response.ok()).toBeTruthy()

    const patchData = {
        type: 'pmm',
    }
    const updated = await request.patch(`/v1/monitoring-instances/${name}`, {data: patchData})

    expect(updated.status()).toBe(400)

    const getJson = await updated.json()
    expect(getJson.message).toMatch('Pmm key is required')

    await th.deleteMonitoringInstance(request, name)
})

test('create monitoring instance failures', async ({request}) => {
    const testCases = [
        {
            payload: {},
            errorText: 'doesn\'t match schema',
        },
    ]

    for (const testCase of testCases) {
        const response = await request.post('/v1/monitoring-instances', {data: testCase.payload})

        expect(response.status()).toBe(400)
        expect((await response.json()).message).toMatch(testCase.errorText)
    }
})

test('update monitoring instances failures', async ({request}) => {
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

    const testCases = [
        {
            payload: {url: 'not-url'},
            errorText: '\'url\' is an invalid URL',
        },
        {
            payload: {pmm: {apiKey: ''}},
            errorText: 'Error at "/pmm/apiKey"',
        },
    ]

    for (const testCase of testCases) {
        const response = await request.patch(`/v1/monitoring-instances/${name}`, {data: testCase.payload})

        expect(response.status()).toBe(400)
        expect((await response.json()).message).toMatch(testCase.errorText)
    }

    await th.deleteMonitoringInstance(request, name)
})

test('update: monitoring instance not found', async ({request}) => {
    const name = 'non-existent'
    const response = await request.patch(`/v1/monitoring-instances/${name}`, {data: {url: 'http://monitoring'}})

    expect(response.status()).toBe(404)
})

test('delete: monitoring instance not found', async ({request}) => {
    const name = 'non-existent'
    const response = await request.delete(`/v1/monitoring-instances/${name}`)

    expect(response.status()).toBe(404)
})

test('get: monitoring instance not found', async ({request}) => {
    const name = 'non-existent'
    const response = await request.get(`/v1/monitoring-instances/${name}`)

    expect(response.status()).toBe(404)
})

