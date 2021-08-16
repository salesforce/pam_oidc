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
}

func configFromArgs(args []string) (*config, error) {
	c := &config{}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		switch parts[0] {
		case "issuer":
			c.Issuer = parts[1]
		case "aud":
			c.Aud = parts[1]
		case "user_template":
			c.UserTemplate = parts[1]
		default:
			return nil, fmt.Errorf("unknown option: %v", parts[0])
		}
	}

	return c, nil
}
