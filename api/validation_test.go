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
package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRFC1123(t *testing.T) {
	t.Parallel()
	type testCase struct {
		value string
		valid bool
	}

	cases := []testCase{
		{
			value: "abc-sdf12",
			valid: true,
		},
		{
			value: "-abc-sdf12",
			valid: false,
		},
		{
			value: "abc-sdf12-",
			valid: false,
		},
		{
			value: "abc-sAAf12",
			valid: false,
		},
		{
			value: "abc-sAAf12",
			valid: false,
		},
		{
			value: "1abc-sf12",
			valid: false,
		},
		{
			value: "aaa123",
			valid: true,
		},
		{
			value: "asldkafaslkdjfalskdfjaslkdjflsakfjdalskfdjaslkfdjaslkfdjsaklfdassksjdfhskdjfskjdfsdfsdflasdkfasdfk",
			valid: false,
		},
		{
			value: "$%",
			valid: false,
		},
		{
			value: "asdf32$%",
			valid: false,
		},
		{
			value: "",
			valid: false,
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.value, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.valid, validateRFC1123(c.value, "") == nil)
		})
	}
}

func TestValidateCreateDatabaseClusterRequest(t *testing.T) {
	t.Parallel()
	type testCase struct {
		name  string
		value DatabaseCluster
		err   error
	}

	cases := []testCase{
		{
			name:  "empty metadata",
			value: DatabaseCluster{},
			err:   errDBCEmptyMetadata,
		},
		{
			name:  "no dbCluster name",
			value: DatabaseCluster{Metadata: &map[string]interface{}{}},
			err:   errDBCNameEmpty,
		},
		{
			name: "empty dbCluster name",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "",
			}},
			err: ErrNameNotRFC1123Compatible("metadata.name"),
		},
		{
			name: "starts with -",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "-sdfasa",
			}},
			err: ErrNameNotRFC1123Compatible("metadata.name"),
		},
		{
			name: "ends with -",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "sdfasa-",
			}},
			err: ErrNameNotRFC1123Compatible("metadata.name"),
		},
		{
			name: "contains uppercase",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "AAsdf",
			}},
			err: ErrNameNotRFC1123Compatible("metadata.name"),
		},
		{
			name: "valid",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "amsdf-sllla",
			}},
			err: nil,
		},
		{
			name: "dbCluster name wrong format",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": make(map[string]string),
			}},
			err: errDBCNameWrongFormat,
		},
		{
			name: "dbCluster name too long",
			value: DatabaseCluster{Metadata: &map[string]interface{}{
				"name": "a123456789a123456789a12",
			}},
			err: ErrNameTooLong("metadata.name"),
		},
	}

	for _, tc := range cases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			err := validateCreateDatabaseClusterRequest(c.value)
			if c.err == nil {
				require.Nil(t, err)
				return
			}
			require.Equal(t, c.err.Error(), err.Error())
		})
	}
}
