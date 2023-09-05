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
import {expect, test} from '@playwright/test'
import * as cluster from "cluster";

// testPrefix is used to differentiate between several workers
// running this test to avoid conflicts in instance names
const testPrefix = `${Date.now()}-${process.env.TEST_WORKER_INDEX}`

let kubernetesId
const bsName = `${testPrefix}-bs`

test.beforeAll(async ({request}) => {
    const kubernetesList = await request.get('/v1/kubernetes')
    kubernetesId = (await kubernetesList.json())[0].id
})


test('create/update/delete database cluster restore', async ({request}) => {
    // Backup storage
    const payload = {
        type: 's3',
        name: bsName,
        url: 'http://custom-url',
        description: 'Dev storage',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: 'sdfs',
        secretKey: 'sdfsdfsd',
    }

    let response = await request.post('/v1/backup-storages', {data: payload})
    expect(response.ok()).toBeTruthy()


    const payloadBackup = {
        apiVersion: 'everest.percona.com/v1alpha1',
        kind: 'DatabaseClusterBackup',
        metadata: {
            name: 'backup-for-restore',
        },
        spec: {
            dbClusterName: 'cluster11',
            backupStorageName: bsName,
        },
    }

    let responseBackup = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-backups`, {
        data: payloadBackup,
    })
    expect(responseBackup.ok()).toBeTruthy()

    const payloadRestore = {
        apiVersion: 'everest.percona.com/v1alpha1',
        kind: 'DatabaseClusterRestore',
        metadata: {
            name: 'restore',
        },
        spec: {
            dataSource: {
                dbClusterBackupName: "backup-for-restore",
            },
            dbClusterName: 'cluster11',
        },
    }

    response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-restores`, {
        data: payloadRestore,
    })
    expect(response.ok()).toBeTruthy()
    const restore = await response.json()
    expect(restore.spec).toMatchObject(payloadRestore.spec)

    // update restore
    restore.spec.dbClusterName = "cluster22"
    response = await request.put(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${restore.metadata.name}`,{
        data: restore,
    })
    expect(response.ok()).toBeTruthy()
    const result = await response.json()
    expect(result.spec).toMatchObject(restore.spec)

    // delete restore
    await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${restore.metadata.name}`)
    // check it couldn't be found anymore
    response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/restore`)
    expect(response.status()).toBe(404)

    let res = await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup-for-restore`)
    expect(res.ok()).toBeTruthy()
    res = await request.delete(`/v1/backup-storages/${bsName}`)
    expect(res.ok()).toBeTruthy()
})

test('list restores', async ({request, page}) => {
    // Backup storage
    const payload = {
        type: 's3',
        name: bsName,
        url: 'http://custom-url',
        description: 'Dev storage',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: 'sdfs',
        secretKey: 'sdfsdfsd',
    }

    let response = await request.post('/v1/backup-storages', {data: payload})
    expect(response.ok()).toBeTruthy()

    const payloadBackup = {
        apiVersion: 'everest.percona.com/v1alpha1',
        kind: 'DatabaseClusterBackup',
        metadata: {
            name: 'backup1',
        },
        spec: {
            dbClusterName: 'cluster11',
            backupStorageName: bsName,
        },
    }

    let responseBackup = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-backups`, {
        data: payloadBackup,
    })
    expect(responseBackup.ok()).toBeTruthy()

    const payloads = [
        {
            apiVersion: 'everest.percona.com/v1alpha1',
            kind: 'DatabaseClusterRestore',
            metadata: {
                name: 'restore1',
            },
            spec: {
                dataSource: {
                    dbClusterBackupName: "backup1",
                },
                dbClusterName: 'cluster11',
            },
        },
        {
            apiVersion: 'everest.percona.com/v1alpha1',
            kind: 'DatabaseClusterRestore',
            metadata: {
                name: 'restore11',
            },
            spec: {
                dataSource: {
                    dbClusterBackupName: "backup1",
                },
                dbClusterName: 'cluster11',
            },
        },
        {
            apiVersion: 'everest.percona.com/v1alpha1',
            kind: 'DatabaseClusterRestore',
            metadata: {
                name: 'restore2',
            },
            spec: {
                dataSource: {
                    dbClusterBackupName: "backup1",
                },
                dbClusterName: 'cluster22',
            },
        },
        {
            apiVersion: 'everest.percona.com/v1alpha1',
            kind: 'DatabaseClusterRestore',
            metadata: {
                name: 'restore22',
            },
            spec: {
                dataSource: {
                    dbClusterBackupName: "backup1",
                },
                dbClusterName: 'cluster22',
            },
        },
    ]

    for (const payload of payloads) {
        const response = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-restores`, {
            data: payload,
        })
        expect(response.ok()).toBeTruthy()
    }

    await page.waitForTimeout(5000)

    // check if the restores are available when being requested via database-clusters/{cluster-name}/restores path
    response = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/cluster11/restores`)
    let result = await response.json()

    expect(result.items).toHaveLength(2)

    response = await request.get(`/v1/kubernetes/${kubernetesId}/database-clusters/cluster22/restores`)
    result = await response.json()

    expect(result.items).toHaveLength(2)

    // delete the created restores
    for (const payload of payloads) {
        await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${payload.metadata.name}`)
        response = await request.get(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${payload.metadata.name}`)
        expect(response.status()).toBe(404)
    }

    // delete the created backup
    let res = await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/backup1`)
    expect(res.ok()).toBeTruthy()

    res = await request.delete(`/v1/backup-storages/${bsName}`)
    expect(res.ok()).toBeTruthy()
})
