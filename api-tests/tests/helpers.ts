import {APIRequestContext, expect} from '@playwright/test'

// testPrefix is used to differentiate between several workers
// running this test to avoid conflicts in instance names
const testSuffix = () => `${(Math.random() + 1).toString(36).substring(8)}`

export const randomName = (prefix = "rnd") => {
    return `${prefix}-${testSuffix()}`
}

export const createDBCluster = async (request, kubernetesId, name) => {
    const data = {
        apiVersion: 'everest.percona.com/v1alpha1',
        kind: 'DatabaseCluster',
        metadata: {
            name,
        },
        spec: {
            engine: {
                type: 'pxc',
                replicas: 1,
                storage: {
                    size: '4G',
                },
                resources: {
                    cpu: '1',
                    memory: '1G',
                },
            },
            proxy: {
                type: 'haproxy',
                replicas: 1,
                expose: {
                    type: 'internal',
                },
            },
        },
    }

    const postReq = await request.post(`/v1/kubernetes/${kubernetesId}/database-clusters`, {data})

    expect(postReq.ok()).toBeTruthy()
}

export const deleteDBCluster = async (request, kubernetesId, name) => {
    const res = await request.delete(`/v1/kubernetes/${kubernetesId}/database-clusters/${name}`)

    expect(res.ok()).toBeTruthy()
}

export const createBackupStorage = async (request, name) => {
    const storagePayload = {
        type: 's3',
        name,
        url: 'http://custom-url',
        description: 'Dev storage',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: 'sdfs',
        secretKey: 'sdfsdfsd',
    }

    const response = await request.post('/v1/backup-storages', {data: storagePayload})

    expect(response.ok()).toBeTruthy()
}

export const deleteBackupStorage = async (request, name) => {
    const res = await request.delete(`/v1/backup-storages/${name}`)

    expect(res.ok()).toBeTruthy()
}

export const createBackup = async (request, kubernetesId, clusterName, backupName, storageName) => {
    const payloadBackup = {
        apiVersion: 'everest.percona.com/v1alpha1',
        kind: 'DatabaseClusterBackup',
        metadata: {
            name: backupName,
        },
        spec: {
            dbClusterName: clusterName,
            backupStorageName: storageName,
        },
    }

    const responseBackup = await request.post(`/v1/kubernetes/${kubernetesId}/database-cluster-backups`, {
        data: payloadBackup,
    })

    expect(responseBackup.ok()).toBeTruthy()
}

export const deleteBackup = async (request, kubernetesId, backupName) => {
    const res = await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-backups/${backupName}`)

    expect(res.ok()).toBeTruthy()
}

export const deleteRestore = async (request, kubernetesId, restoreName) => {
    const res = await request.delete(`/v1/kubernetes/${kubernetesId}/database-cluster-restores/${restoreName}`)

    expect(res.ok()).toBeTruthy()
}

export const deleteMonitoringInstances = async (request, names) => {
    for (let i = 0; i < names.length; i++) {
        await deleteMonitoringInstance(request, names[i])
    }
}

export const deleteMonitoringInstance = async (request, name) => {
    const res = await request.delete(`/v1/monitoring-instances/${name}`)
    expect(res.ok()).toBeTruthy()
}


export const createMonitoringInstance = async (request: APIRequestContext): Promise<string> => {
    const name  = randomName()
    const data = {
        type: 'pmm',
        name: name,
        url: 'http://monitoring-instance',
        pmm: {
            apiKey: '123',
        },
    }

    const response = await request.post('/v1/monitoring-instances', {data})
    expect(response.ok()).toBeTruthy()

    return name
}

export const createMonitoringInstances = async (request: APIRequestContext, prefix: string, count = 3): Promise<string[]> => {
    const data = {
        type: 'pmm',
        name: '',
        url: 'http://monitoring-instance',
        pmm: {
            apiKey: '123',
        },
    }

    const res = []

    for (let i = 1; i <= count; i++) {
        data.name = randomName(prefix)
        res.push(data.name)
        const response = await request.post('/v1/monitoring-instances', {data})

        expect(response.ok()).toBeTruthy()
    }

    return res
}
