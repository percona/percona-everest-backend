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
import {testsNs} from './helpers'

test('add/list/get/delete s3 backup storage success', async ({request}) => {
    const payload = {
        type: 's3',
        name: 'backup-storage-1',
        url: 'http://custom-url',
        description: 'Dev storage',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: 'sdfs',
        secretKey: 'sdfsdfsd',
        targetNamespaces: [testsNs]
    }

    const response = await request.post(`/v1/backup-storages`, {
        data: payload,
    })

    // create
    expect(response.ok()).toBeTruthy()
    const created = await response.json()

    const name = created.name

    expect(created.name).toBe(payload.name)
    expect(created.url).toBe(payload.url)
    expect(created.bucketName).toBe(payload.bucketName)
    expect(created.region).toBe(payload.region)
    expect(created.type).toBe(payload.type)
    expect(created.description).toBe(payload.description)

    // list
    const listResponse = await request.get(`/v1/backup-storages`)

    expect(listResponse.ok()).toBeTruthy()
    const list = await listResponse.json()

    expect(list.length).toBeGreaterThan(0)

    // get
    const one = await request.get(`/v1/backup-storages${name}`)

    expect(one.ok()).toBeTruthy()
    expect((await one.json()).name).toBe(payload.name)

    // update
    const updatePayload = {
        description: 'some description',
        bucketName: 'percona-test-backup-storage1',
        accessKey: 'otherAccessKey',
        secretKey: 'otherSecret',
        targetNamespaces: [testsNs]
    }
    const updated = await request.patch(`/v1/backup-storages/${name}`, {
        data: updatePayload,
    })

    expect(updated.ok()).toBeTruthy()
    const result = await updated.json()

    expect(result.bucketName).toBe(updatePayload.bucketName)
    expect(result.region).toBe(created.region)
    expect(result.type).toBe(created.type)
    expect(result.description).toBe(updatePayload.description)

    // backup storage already exists
    const createAgain = await request.post(`/v1/backup-storages`, {
        data: payload,
    })

    expect(createAgain.status()).toBe(409)

    // delete
    const deleted = await request.delete(`/v1/backup-storages/${name}`)

    expect(deleted.ok()).toBeTruthy()
})

test('add/list/get/delete azure backup storage success', async ({request}) => {
    const payload = {
        type: 'azure',
        name: 'backup-storage-azure',
        description: 'Dev storage',
        bucketName: 'percona-test-backup-storage',
        accessKey: 'sdfs',
        secretKey: 'sdfsdfsd',
        targetNamespaces: [testsNs]
    }

    const response = await request.post(`/v1/backup-storages`, {
        data: payload,
    })

    // create
    expect(response.ok()).toBeTruthy()
    const created = await response.json()

    const name = created.name

    expect(created.name).toBe(payload.name)
    expect(created.bucketName).toBe(payload.bucketName)
    expect(created.type).toBe(payload.type)
    expect(created.description).toBe(payload.description)

    // list
    const listResponse = await request.get(`/v1/backup-storages`)

    expect(listResponse.ok()).toBeTruthy()
    const list = await listResponse.json()

    expect(list.length).toBeGreaterThan(0)

    // get
    const one = await request.get(`/v1/backup-storages/${name}`)

    expect(one.ok()).toBeTruthy()
    expect((await one.json()).name).toBe(payload.name)

    // update
    const updatePayload = {
        description: 'some description',
        bucketName: 'percona-test-backup-storage1',
    }
    const updated = await request.patch(`/v1/backup-storages/${name}`, {
        data: updatePayload,
    })

    expect(updated.ok()).toBeTruthy()
    const result = await updated.json()

    expect(result.bucketName).toBe(updatePayload.bucketName)
    expect(result.region).toBe(created.region)
    expect(result.type).toBe(created.type)
    expect(result.description).toBe(updatePayload.description)

    // backup storage already exists
    const createAgain = await request.post(`/v1/backup-storages`, {
        data: payload,
    })

    expect(createAgain.status()).toBe(409)

    // delete
    const deleted = await request.delete(`/v1/backup-storages/${name}`)

    expect(deleted.ok()).toBeTruthy()
})

test('create backup storage failures', async ({request}) => {
    const testCases = [
        {
            payload: {},
            errorText: 'property \"name\" is missing',
        },
        {
            payload: {
                type: 's3',
                name: 'backup-storage',
                bucketName: 'percona-test-backup-storage',
                region: 'us-east-2',
                accessKey: 'ssdssd',
            },
            errorText: 'property \"secretKey\" is missing',
        },
        {
            payload: {
                type: 's3',
                name: 'Backup Name',
                bucketName: 'percona-test-backup-storage',
                region: 'us-east-2',
                accessKey: 'ssdssd',
                secretKey: 'ssdssdssdssd',
                targetNamespaces: [testsNs]
            },
            errorText: '\'name\' is not RFC 1035 compatible',
        },
        {
            payload: {
                type: 's3',
                name: 'backup',
                bucketName: 'percona-test-backup-storage',
                url: 'not-valid-url',
                region: 'us-east-2',
                accessKey: 'ssdssd',
                secretKey: 'ssdssdssdssd',
                targetNamespaces: [testsNs]
            },
            errorText: '\'url\' is an invalid URL',
        },
        {
            payload: {
                type: 's3',
                name: 'missing-region',
                bucketName: 'invalid',
                accessKey: 'ssdssd',
                secretKey: 'ssdssdssdssd',
                targetNamespaces: [testsNs]
            },
            errorText: 'Region is required',
        },
        {
            payload: {
                type: 'gcs',
                name: 'invalid',
                region: 'us-east-2',
                bucketName: 'invalid',
                accessKey: 'ssdssd',
                secretKey: 'ssdssdssdssd',
                targetNamespaces: [testsNs]
            },
            errorText: '"/type": value is not one of the allowed values',
        },
    ]

    for (const testCase of testCases) {
        const response = await request.post(`/v1/backup-storages`, {
            data: testCase.payload,
        })

        expect(response.status()).toBe(400)
        expect((await response.json()).message).toMatch(testCase.errorText)
    }
})

test('update backup storage failures', async ({request}) => {
    const createPayload = {
        type: 's3',
        name: 'backup-storage-2',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: 'sdfsdfs',
        secretKey: 'lkdfslsldfka',
        targetNamespaces: [testsNs]
    }
    const response = await request.post(`/v1/backup-storages`, {
        data: createPayload,
    })

    expect(response.ok()).toBeTruthy()
    const created = await response.json()

    const name = created.name

    const testCases = [
        {
            payload: {
                url: '-asldf;asdfk;sadf',
            },
            errorText: '\'url\' is an invalid URL',
        },
        {
            payload: {
                bucket: '-asldf;asdfk;sadf',
            },
            errorText: 'request body has an error: doesn\'t match schema #/components/schemas/UpdateBackupStorageParams: property \"bucket\" is unsupported',
        },
    ]

    for (const testCase of testCases) {
        const response = await request.patch(`/v1/backup-storages/${name}`, {
            data: testCase.payload,
        })

        expect((await response.json()).message).toMatch(testCase.errorText)
        expect(response.status()).toBe(400)
    }

    const deleted = await request.delete(`/v1/backup-storages/${name}`)

    expect(deleted.ok()).toBeTruthy()
})

test('update: backup storage not found', async ({request}) => {
    const name = 'some-storage'

    const response = await request.patch(`/v1/backup-storages/${name}`, {
        data: {
            bucketName: 's3',
        },
    })

    expect(response.status()).toBe(404)
})

test('delete: backup storage not found', async ({request}) => {
    const name = 'backup-storage'

    const response = await request.delete(`/v1/backup-storages/${name}`)

    expect(response.status()).toBe(404)
})

test('get: backup storage not found', async ({request}) => {
    const name = 'backup-storage'
    const response = await request.get(`/v1/backup-storages/${name}`)

    expect(response.status()).toBe(404)
})
