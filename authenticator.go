// Copyright (c) 2021, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE.txt file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/pardot/oidc"
)

type authenticator struct {
	// UserTemplate is a template that, when rendered with the JWT claims, should
	// match the user being authenticated.
	//
	// `{{.Subject}}` is used by default if not set.
	UserTemplate string

	verifier *oidc.Verifier
	aud      string
}

func discoverAuthenticator(ctx context.Context, issuer string, aud string) (*authenticator, error) {
	verifier, err := oidc.DiscoverVerifier(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discovering verifier: %v", err)
	}

	return &authenticator{
		verifier: verifier,
		aud:      aud,
	}, nil
}

// Authenticate authenticates a user with the provided token.
func (a *authenticator) Authenticate(ctx context.Context, user string, token string) error {
	claims, err := a.verifier.VerifyRaw(ctx, a.aud, token)
	if err != nil {
		return fmt.Errorf("verifying token: %v", err)
	}

	userTemplate := "{{.Subject}}"
	if a.UserTemplate != "" {
		userTemplate = a.UserTemplate
	}

	userTmpl, err := template.New("").Funcs(template.FuncMap{
		"trimPrefix": func(prefix, s string) string { return strings.TrimPrefix(s, prefix) },
		"trimSuffix": func(suffix, s string) string { return strings.TrimSuffix(s, suffix) },
	}).Parse(userTemplate)
	if err != nil {
		return fmt.Errorf("parsing user template: %v", err)
	}

	buf := new(bytes.Buffer)
	if err := userTmpl.Execute(buf, claims); err != nil {
		return fmt.Errorf("executing user template: %v", err)
	}

	wantUser := buf.String()
	if wantUser != user {
		return fmt.Errorf("expected user %q but is authenticating as %q", wantUser, user)
	}

	return nil
}
