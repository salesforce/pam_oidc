// Copyright (c) 2021, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE.txt file in the repo root or https://opensource.org/licenses/BSD-3-Clause

package main

import (
	"fmt"
	"strings"
)

type config struct {
	// Issuer is the OpenID Connect issuer
	Issuer string
	// Aud is the expected aud(ience) value for valid OIDC tokens
	Aud string
	// UserTemplate is a template that, when rendered with the JWT claims, should
	// match the user being authenticated.
	UserTemplate string
	// GroupsClaimKey is the name of the key within the token claims that
	// specifies which groups a user is a member of.
	GroupsClaimKey string
	// AuthorizedGroups is a list of groups required for authentication to pass.
	// A user must be a member of at least one of the groups in the list, if
	// specified.
	AuthorizedGroups []string
	// RequireACRs is a list of required ACRs required for authentication to pass.
	// one of the acr values must be present in the claims.
	RequireACRs []string
	// HTTPProxy is the HTTP proxy server used to connect to HTTP services.
	HTTPProxy string
}

func configFromArgs(args []string) (*config, error) {
	c := &config{}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed arg: %v", arg)
		}

		switch parts[0] {
		case "issuer":
			c.Issuer = parts[1]
		case "aud":
			c.Aud = parts[1]
		case "user_template":
			c.UserTemplate = parts[1]
		case "groups_claim_key":
			c.GroupsClaimKey = parts[1]
		case "authorized_groups":
			c.AuthorizedGroups = strings.Split(parts[1], ",")
		case "require_acr":
			c.RequireACRs = []string{parts[1]}
		case "require_acrs":
			c.RequireACRs = strings.Split(parts[1], ",")
		case "http_proxy":
			c.HTTPProxy = parts[1]
		default:
			return nil, fmt.Errorf("unknown option: %v", parts[0])
		}
	}

	return c, nil
}
