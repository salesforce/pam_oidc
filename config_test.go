// Copyright (c) 2021, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE.txt file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseConfigFromArgs(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    *config
		wantErr string
	}{
		{
			name: "basic with defaults",
			args: []string{"issuer=https://example.com", "aud=example-aud"},
			want: &config{
				Issuer: "https://example.com",
				Aud:    "example-aud",
			},
		},
		{
			name: "basic overriding defaults",
			args: []string{"issuer=https://example.com", "aud=example-aud", "user_template={{.Email}}", "groups_claim_key=roles", "authorized_groups=foo,bar,baz", "require_acr=foo", "http_proxy=http://example.com:8080"},
			want: &config{
				Issuer:           "https://example.com",
				Aud:              "example-aud",
				UserTemplate:     `{{.Email}}`,
				GroupsClaimKey:   "roles",
				AuthorizedGroups: []string{"foo", "bar", "baz"},
				RequireACR:       "foo",
				HTTPProxy:        "http://example.com:8080",
			},
		},
		{
			name:    "invalid option",
			args:    []string{"issuer=https://example.com", "invalid=foo"},
			wantErr: "unknown option: invalid",
		},
	}

	for _, tc := range cases {
		tc := tc

		config, err := configFromArgs(tc.args)
		if err != nil && tc.wantErr == "" {
			t.Fatalf("wanted no error, but got %v", err)
		} else if err != nil && !strings.Contains(err.Error(), tc.wantErr) {
			t.Fatalf("wanted error %v, but got %v", tc.wantErr, err)
		} else if err == nil && tc.wantErr != "" {
			t.Fatalf("wanted error %v, but got none", tc.wantErr)
		}

		if diff := cmp.Diff(config, tc.want); diff != "" {
			t.Errorf("diff: %v", diff)
		}
	}
}
