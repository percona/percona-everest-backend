import {expect, test} from '@playwright/test';

let id
let req

test.afterAll(async ({}, testInfo) =>{
    if (testInfo.status != "passed") {
        req.delete(`/backup-storages/` + id)
    }
})

test('add/list/get/delete backup storage', async ({request}) => {
    req = request
    const payload = {
        type: 's3',
        name: 'backup-storage-name',
        bucketName: 'percona-test-backup-storage',
        region: 'us-east-2',
        accessKey: "AKIA2QEXCXDVSAGAYX7X",
        secretKey: "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"
    }

    const response = await request.post(`/backup-storages`, {
        data: payload
    });

    // create
    expect(response.ok()).toBeTruthy();
    const created = await response.json()

    id = created.id

    expect(created.id.match(expect.any(String)))
    expect(created.name).toBe(payload.name)
    expect(created.bucketName).toBe(payload.bucketName)
    expect(created.region).toBe(payload.region)
    expect(created.type).toBe(payload.type)

    // list
    const listResponse = await request.get(`/backup-storages`);
    expect(listResponse.ok()).toBeTruthy();
    const list = await listResponse.json()
    expect(list.length).toBe(1)
    expect(list[0].name).toBe(payload.name)

    // get
    const one = await request.get(`/backup-storages/` + id);
    expect(one.ok()).toBeTruthy();
    expect((await one.json()).name).toBe(payload.name)

    // update
    const updatePayload = {
        name: 'backup-storage-name1',
        bucketName: 'percona-test-backup-storage1',
        accessKey: "otherAccessKey",
        secretKey: "otherSecret"
    }
    const updated = await request.patch(`/backup-storages/` + id, {data: updatePayload});
    expect(updated.ok()).toBeTruthy();
    const result = await updated.json()

    console.log(result)
    console.log(updatePayload)

    expect(result.name).toBe(updatePayload.name)
    expect(result.bucketName).toBe(updatePayload.bucketName)
    expect(result.region).toBe(created.region)
    expect(result.type).toBe(created.type)

    // delete
    const deleted = await request.delete(`/backup-storages/` + id);
    expect(deleted.ok()).toBeTruthy();
});
