import {expect, test} from '@playwright/test';

let req

test.afterEach(async ({page}, testInfo) => {
    const request = page.context().request;
    const result = await request.get(`/v1/pmm-instances`)
    const list = await result.json()

    for (const pmm of list) {
        await request.delete(`/v1/pmm-instances/` + pmm.id)
    }
})

test('add/list/get/delete pmm instance success', async ({request}) => {
    req = request
    const data = {
        url: 'http://pmm-instance',
        apiKey: '123',
    }

    const response = await request.post(`/v1/pmm-instances`, { data });

    // create
    expect(response.ok()).toBeTruthy();
    const created = await response.json()

    const id = created.id

    expect(created.id.match(expect.any(String)))
    expect(created.url).toBe(data.url)
    expect(created.apiKeySecretId.match(expect.any(String)))

    // list
    const listResponse = await request.get(`/v1/pmm-instances`);
    expect(listResponse.ok()).toBeTruthy();
    const list = await listResponse.json()
    expect(list.length).toBe(1)
    expect(list[0].url).toBe(data.url)

    // get
    const one = await request.get(`/v1/pmm-instances/` + id);
    expect(one.ok()).toBeTruthy();
    expect((await one.json()).url).toBe(data.url)

    // patch 1
    const patch1Data = {
        url: 'http://pmm'
    }
    const updated1 = await request.patch(`/v1/pmm-instances/` + id, {data: patch1Data});
    expect(updated1.ok()).toBeTruthy();

    // get 1
    const get1 = await request.get(`/v1/pmm-instances/` + id);
    expect(get1.ok()).toBeTruthy();
    const get1Json = await get1.json()
    expect(get1Json.url).toBe(patch1Data.url)
    expect(get1Json.apiKeySecretId).toBe(created.apiKeySecretId)

    // patch 2
    const patch2Data = {
        url: 'http://pmm2',
        apiKey: 'asd',
    }
    const updated2 = await request.patch(`/v1/pmm-instances/` + id, {data: patch2Data});
    expect(updated2.ok()).toBeTruthy();

    // get 2
    const get2 = await request.get(`/v1/pmm-instances/` + id);
    expect(get2.ok()).toBeTruthy();
    const get2Json = await get2.json()
    expect(get2Json.url).toBe(patch2Data.url)
    expect(get2Json.apiKeySecretId).not.toBe(get1Json.apiKeySecretId)

    // delete
    const deleted = await request.delete(`/v1/pmm-instances/` + id);
    expect(deleted.ok()).toBeTruthy();
});


test('create pmm instance failures', async ({request}) => {
    req = request

    const testCases = [
        {
            payload: {},
            errorText: `property \"url\" is missing`,
        },
    ];

    for (const testCase of testCases) {
        const response = await request.post(`/v1/pmm-instances`, {
            data: testCase.payload
        });
        expect(response.status()).toBe(400)
        expect((await response.json()).message).toMatch(testCase.errorText)
    }
});

test('update pmm instances failures', async ({request}) => {
    req = request
    const data = {
        url: 'http://pmm',
        apiKey: '123',
    }
    const response = await request.post(`/v1/pmm-instances`, { data });
    expect(response.ok()).toBeTruthy();
    const created = await response.json()

    const id = created.id

    const testCases = [
        {
            payload: {
                url: 'not-url',
            },
            errorText: `'url' is an invalid URL`,
        },
        {
            payload: {
                apiKey: '',
            },
            errorText: `Error at "/apiKey"`,
        },
    ];

    for (const testCase of testCases) {
        const response = await request.patch(`/v1/pmm-instances/` + id, {
            data: testCase.payload
        });
        expect(response.status()).toBe(400)
        expect((await response.json()).message).toMatch(testCase.errorText)
    }
});


test('update: pmm instance not found', async ({request}) => {
    const id = "788fd6ee-ec54-4d7f-ae37-beab62064fcc"

    const response = await request.patch(`/v1/pmm-instances/` + id, {
        data: {url: 'http://pmm'}
    });
    expect(response.status()).toBe(404)
});


test('delete: pmm instance not found', async ({request}) => {
    const id = "788fd6ee-ec54-4d7f-ae37-beab62064fcc"

    const response = await request.delete(`/v1/pmm-instances/` + id);
    expect(response.status()).toBe(404)
});

test('get: backup storage not found', async ({request}) => {
    const id = "788fd6ee-ec54-4d7f-ae37-beab62064fcc"
    const response = await request.get(`/v1/pmm-instances/` + id);
    expect(response.status()).toBe(404)
});
