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
import { expect, test } from '@fixtures';

let req;

test.afterEach(async ({ page }, testInfo) => {
  const request = page.context().request;
  const result = await request.get('/v1/backup-storages');
  const list = await result.json();

  for (const storage of list) {
    await request.delete(`/v1/backup-storages/${storage.id}`);
  }
});

test('add/list/get/delete backup storage success', async ({ request }) => {
  req = request;
  const payload = {
    type: 's3',
    name: 'backup-storage-name',
    url: 'http://custom-url',
    bucketName: 'percona-test-backup-storage',
    region: 'us-east-2',
    accessKey: 'sdfs',
    secretKey: 'sdfsdfsd',
  };

  const response = await request.post('/v1/backup-storages', {
    data: payload,
  });

  // create
  expect(response.ok()).toBeTruthy();
  const created = await response.json();

  const id = created.id;

  expect(created.id.match(expect.any(String)));
  expect(created.name).toBe(payload.name);
  expect(created.url).toBe(payload.url);
  expect(created.bucketName).toBe(payload.bucketName);
  expect(created.region).toBe(payload.region);
  expect(created.type).toBe(payload.type);

  // list
  const listResponse = await request.get('/v1/backup-storages');

  expect(listResponse.ok()).toBeTruthy();
  const list = await listResponse.json();

  expect(list.length).toBe(1);
  expect(list[0].name).toBe(payload.name);

  // get
  const one = await request.get(`/v1/backup-storages/${id}`);

  expect(one.ok()).toBeTruthy();
  expect((await one.json()).name).toBe(payload.name);

  // update
  const updatePayload = {
    name: 'backup-storage-name1',
    bucketName: 'percona-test-backup-storage1',
    accessKey: 'otherAccessKey',
    secretKey: 'otherSecret',
  };
  const updated = await request.patch(`/v1/backup-storages/${id}`, {
    data: updatePayload,
  });

  expect(updated.ok()).toBeTruthy();
  const result = await updated.json();

  expect(result.name).toBe(updatePayload.name);
  expect(result.bucketName).toBe(updatePayload.bucketName);
  expect(result.region).toBe(created.region);
  expect(result.type).toBe(created.type);

  // delete
  const deleted = await request.delete(`/v1/backup-storages/${id}`);

  expect(deleted.ok()).toBeTruthy();
});

test('create backup storage failures', async ({ request }) => {
  req = request;

  const testCases = [
    {
      payload: {},
      errorText: 'property \"name\" is missing',
    },
    {
      payload: {
        type: 's3',
        name: 'backup-storage-name',
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
      },
      errorText: '\'name\' is not RFC 1123 compatible',
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
      },
      errorText: '\'url\' is an invalid URL',
    },
  ];

  for (const testCase of testCases) {
    const response = await request.post('/v1/backup-storages', {
      data: testCase.payload,
    });

    expect(response.status()).toBe(400);
    expect((await response.json()).message).toMatch(testCase.errorText);
  }
});

test('update backup storage failures', async ({ request }) => {
  req = request;
  const createPayload = {
    type: 's3',
    name: 'backup-storage-name',
    bucketName: 'percona-test-backup-storage',
    region: 'us-east-2',
    accessKey: 'sdfsdfs',
    secretKey: 'lkdfslsldfka',
  };
  const response = await request.post('/v1/backup-storages', {
    data: createPayload,
  });

  expect(response.ok()).toBeTruthy();
  const created = await response.json();

  const id = created.id;

  const testCases = [
    {
      payload: {
        name: '-123dfdfs',
      },
      errorText: '\'name\' is not RFC 1123 compatible',
    },
    {
      payload: {
        type: 's3',
        name: 'Bauckup storage',
      },
      errorText: '\'name\' is not RFC 1123 compatible',
    },
    {
      payload: {
        url: '-asldf;asdfk;sadf',
      },
      errorText: '\'url\' is an invalid URL',
    },
  ];

  for (const testCase of testCases) {
    const response = await request.patch(`/v1/backup-storages/${id}`, {
      data: testCase.payload,
    });

    expect(response.status()).toBe(400);
    expect((await response.json()).message).toMatch(testCase.errorText);
  }
});

test('update: backup storage not found', async ({ request }) => {
  const id = '788fd6ee-ec54-4d7f-ae37-beab62064fcc';

  const response = await request.patch(`/v1/backup-storages/${id}`, {
    data: {
      type: 's3',
    },
  });

  expect(response.status()).toBe(404);
});

test('delete: backup storage not found', async ({ request }) => {
  const id = '788fd6ee-ec54-4d7f-ae37-beab62064fcc';

  const response = await request.delete(`/v1/backup-storages/${id}`);

  expect(response.status()).toBe(404);
});

test('get: backup storage not found', async ({ request }) => {
  const id = '788fd6ee-ec54-4d7f-ae37-beab62064fcc';
  const response = await request.get(`/v1/backup-storages/${id}`);

  expect(response.status()).toBe(404);
});
