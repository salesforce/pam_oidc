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

	// GroupsClaimKey is the name of the key within the token claims that
	// specifies which groups a user is a member of.
	//
	// `groups` is used by default if not set.
	GroupsClaimKey string

	// AuthorizedGroups is a list of groups required for authentication to pass.
	// A user must be a member of at least one of the groups in the list, if
	// specified.
	//
	// If the list is empty, group membership is not required for authentication
	// to pass.
	AuthorizedGroups []string

	// RequireACR is the required value of the acr claim in the token for
	// authentication to pass.
	//
	// If empty, the ACR value is not checked.
	RequireACR string

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

	// Validate AuthorizedGroups / GroupClaimsKey
	if len(a.AuthorizedGroups) > 0 {
		groupsClaimKey := "groups"
		if len(a.GroupsClaimKey) > 0 {
			groupsClaimKey = a.GroupsClaimKey
		}

		groupsClaim, ok := claims.Extra[groupsClaimKey].([]interface{})
		if !ok {
			return fmt.Errorf("user is not member of any groups, but one of %v is required", a.AuthorizedGroups)
		}

		groups := make([]string, 0, len(groupsClaim))
		for _, groupVal := range groupsClaim {
			if group, ok := groupVal.(string); ok {
				groups = append(groups, group)
			}
		}
		if !isMemberOfAtLeastOneGroup(a.AuthorizedGroups, groups) {
			return fmt.Errorf("user is member of %v, but one of %v is required", groups, a.AuthorizedGroups)
		}
	}

	// Validate RequireACR
	if len(a.RequireACR) > 0 && a.RequireACR != claims.ACR {
		return fmt.Errorf("acr is %q, but %q is required", claims.ACR, a.RequireACR)
	}

	return nil
}

func isMemberOfAtLeastOneGroup(authorizedGroups []string, groups []string) bool {
	for _, wantGroup := range authorizedGroups {
		for _, group := range groups {
			if wantGroup == group {
				return true
			}
		}
	}

	return false
}
