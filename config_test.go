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
			args: []string{"issuer=https://example.com", "aud=example-aud", "user_template={{.Email}}"},
			want: &config{
				Issuer:       "https://example.com",
				Aud:          "example-aud",
				UserTemplate: `{{.Email}}`,
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
