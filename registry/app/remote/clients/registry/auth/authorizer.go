// Source: https://github.com/goharbor/harbor

// Copyright 2016 Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	commonhttp "github.com/harness/gitness/registry/app/common/http"
	"github.com/harness/gitness/registry/app/common/http/modifier"
	"github.com/harness/gitness/registry/app/common/lib"
	"github.com/harness/gitness/registry/app/dist_temp/challenge"
	"github.com/harness/gitness/registry/app/remote/clients/registry/auth/basic"
	"github.com/harness/gitness/registry/app/remote/clients/registry/auth/bearer"
	"github.com/harness/gitness/registry/app/remote/clients/registry/auth/null"
)

// NewAuthorizer creates an authorizer that can handle different auth schemes.
func NewAuthorizer(username, password string, insecure bool) lib.Authorizer {
	return &authorizer{
		username: username,
		password: password,
		client: &http.Client{
			Transport: commonhttp.GetHTTPTransport(commonhttp.WithInsecure(insecure)),
		},
	}
}

// authorizer authorizes the request with the provided credential.
// It determines the auth scheme of registry automatically and calls
// different underlying authorizers to do the auth work.
type authorizer struct {
	sync.Mutex
	username   string
	password   string
	client     *http.Client
	url        *url.URL          // registry URL
	authorizer modifier.Modifier // the underlying authorizer
}

func (a *authorizer) Modify(req *http.Request) error {
	// Nil underlying authorizer means this is the first time the authorizer is called
	// Try to connect to the registry and determine the auth scheme
	if a.authorizer == nil {
		// to avoid concurrent issue
		a.Lock()
		defer a.Unlock()
		if err := a.initialize(req.URL); err != nil {
			return err
		}
	}

	// check whether the request targets the registry
	// If it doesn't, no modification is needed, so we return nil.
	if !a.isTarget(req) {
		return nil
	}

	// If the request targets the registry, delegate the modification to the underlying authorizer.
	return a.authorizer.Modify(req)
}

func (a *authorizer) initialize(u *url.URL) error {
	if a.authorizer != nil {
		return nil
	}
	url, err := url.Parse(u.Scheme + "://" + u.Host + "/v2/")
	if err != nil {
		return fmt.Errorf("failed to parse URL for scheme %s and host %s: %w", u.Scheme, u.Host, err)
	}
	a.url = url

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, a.url.String(), nil)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	challenges := challenge.ResponseChallenges(resp)
	// no challenge, mean no auth
	if len(challenges) == 0 {
		a.authorizer = null.NewAuthorizer()
		return nil
	}
	cm := map[string]challenge.Challenge{}
	for _, challenge := range challenges {
		cm[challenge.Scheme] = challenge
	}
	if challenge, exist := cm["bearer"]; exist {
		a.authorizer = bearer.NewAuthorizer(
			challenge.Parameters["realm"],
			challenge.Parameters["service"], basic.NewAuthorizer(a.username, a.password),
			a.client.Transport,
		)
		return nil
	}
	if _, exist := cm["basic"]; exist {
		a.authorizer = basic.NewAuthorizer(a.username, a.password)
		return nil
	}
	return fmt.Errorf("unsupported auth scheme: %v", challenges)
}

// isTarget checks whether the request targets the registry.
// If not, the request shouldn't be handled by the authorizer, e.g., requests sent to backend storage (S3, etc.).
func (a *authorizer) isTarget(req *http.Request) bool {
	// Check if the path contains the versioned API endpoint (e.g., "/v2/")
	const versionedPath = "/v2/"
	if !strings.Contains(req.URL.Path, versionedPath) {
		return false
	}

	// Ensure that the request's host, scheme, and versioned path match the authorizer's URL.
	if req.URL.Host != a.url.Host || req.URL.Scheme != a.url.Scheme ||
		!strings.HasPrefix(req.URL.Path, a.url.Path) {
		return false
	}

	return true
}
