import {expect, test} from '@playwright/test';

let req

test.afterEach(async ({page}, testInfo) => {
    const request = page.context().request;
    const result = await request.get(`/v1/backup-storages`)
    const list = await result.json()

    for (const storage of list) {
        await request.delete(`/v1/backup-storages/` + storage.name)
    }
})

test('add/list/get/delete backup storage success', async ({request}) => {
    req = request
    const payload = {
        type: 's3',
        name: 'backup-storage-name-1',
        url: 'http://custom-url',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: "sdfs",
        secretKey: "sdfsdfsd"
    }

    const response = await request.post(`/v1/backup-storages`, {
        data: payload
    });

    // create
    expect(response.ok()).toBeTruthy();
    const created = await response.json()

    const name = created.name

    expect(created.name).toBe(payload.name)
    expect(created.url).toBe(payload.url)
    expect(created.bucketName).toBe(payload.bucketName)
    expect(created.region).toBe(payload.region)
    expect(created.type).toBe(payload.type)

    // list
    const listResponse = await request.get(`/v1/backup-storages`);
    expect(listResponse.ok()).toBeTruthy();
    const list = await listResponse.json()
    expect(list.length).toBe(1)
    expect(list[0].name).toBe(payload.name)

    // get
    const one = await request.get(`/v1/backup-storages/` + name);
    expect(one.ok()).toBeTruthy();
    expect((await one.json()).name).toBe(payload.name)

    // update
    const updatePayload = {
        bucketName: 'percona-test-backup-storage1',
        accessKey: "otherAccessKey",
        secretKey: "otherSecret"
    }
    const updated = await request.patch(`/v1/backup-storages/` + name, {data: updatePayload});
    expect(updated.ok()).toBeTruthy();
    const result = await updated.json()

    expect(result.bucketName).toBe(updatePayload.bucketName)
    expect(result.region).toBe(created.region)
    expect(result.type).toBe(created.type)

    // delete
    const deleted = await request.delete(`/v1/backup-storages/` + name);
    expect(deleted.ok()).toBeTruthy();
});


test('create backup storage failures', async ({request}) => {
    req = request

    const testCases = [
        {
            payload: {},
            errorText: `property \"name\" is missing`,
        },
        {
            payload: {
                type: 's3',
                name: 'backup-storage-name',
                bucketName: 'percona-test-backup-storage',
                region: 'us-east-2',
                accessKey: "ssdssd"
            },
            errorText: `property \"secretKey\" is missing`,
        },
        {
            payload: {
                type: 's3',
                name: 'Backup Name',
                bucketName: 'percona-test-backup-storage',
                region: 'us-east-2',
                accessKey: "ssdssd",
                secretKey: "ssdssdssdssd"
            },
            errorText: `'name' is not RFC 1123 compatible`,
        },
        {
            payload: {
                type: 's3',
                name: 'backup',
                bucketName: 'percona-test-backup-storage',
                url: 'not-valid-url',
                region: 'us-east-2',
                accessKey: "ssdssd",
                secretKey: "ssdssdssdssd"
            },
            errorText: `'url' is an invalid URL`,
        },
    ];

    for (const testCase of testCases) {
        const response = await request.post(`/v1/backup-storages`, {
            data: testCase.payload
        });
        expect(response.status()).toBe(400)
        expect((await response.json()).message).toMatch(testCase.errorText)
    }
});

test('update backup storage failures', async ({request}) => {
    req = request
    const createPayload = {
        type: 's3',
        name: 'backup-storage-name-2',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: "sdfsdfs",
        secretKey: "lkdfslsldfka"
    }
    const response = await request.post(`/v1/backup-storages`, {
        data: createPayload
    });
    expect(response.ok()).toBeTruthy();
    const created = await response.json()

    const name = created.name

    const testCases = [
        {
            payload: {
                url: '-asldf;asdfk;sadf'
            },
            errorText: `'url' is an invalid URL`,
        },
    ];

    for (const testCase of testCases) {
        const response = await request.patch(`/v1/backup-storages/` + name, {
            data: testCase.payload
        });
        expect((await response.json()).message).toMatch(testCase.errorText)
        expect(response.status()).toBe(400)
    }
});


test('backup storage already exists', async ({request}) => {
    req = request
    const createPayload = {
        type: 's3',
        name: 'storage-to-be-duplicated',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: "sdfsdfs",
        secretKey: "lkdfslsldfka"
    }
    const response = await request.post(`/v1/backup-storages`, {
        data: createPayload
    });
    expect(response.ok()).toBeTruthy();

    const createAgain = await request.post(`/v1/backup-storages`, {
        data: createPayload
    });
    expect(createAgain.status()).toBe(409)

    // delete
    const deleted = await request.delete(`/v1/backup-storages/` + createPayload.name);
    expect(deleted.ok()).toBeTruthy();
});


test('update: backup storage not found', async ({request}) => {
    const name = "some-storage"

    const response = await request.patch(`/v1/backup-storages/` + name, {
        data: {type: "s3"}
    });
    expect(response.status()).toBe(404)
});


test('delete: backup storage not found', async ({request}) => {
    const name = "backup-storage-name"

    const response = await request.delete(`/v1/backup-storages/` + name);
    expect(response.status()).toBe(404)
});

test('get: backup storage not found', async ({request}) => {
    const name = "backup-storage-name"
    const response = await request.get(`/v1/backup-storages/` + name);
    expect(response.status()).toBe(404)
});
