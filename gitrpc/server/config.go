// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package server

import (
	"errors"
)

// Config represents the configuration for the gitrpc server.
type Config struct {
	// Bind specifies the addr used to bind the grpc server.
	Bind string `envconfig:"GITRPC_SERVER_BIND" default:":3001"`
	// GitRoot specifies the directory containing git related data (e.g. repos, ...)
	GitRoot string `envconfig:"GITRPC_SERVER_GIT_ROOT"`
	// TmpDir (optional) specifies the directory for temporary data (e.g. repo clones, ...)
	TmpDir string `envconfig:"GITRPC_SERVER_TMP_DIR"`
	// GitHookPath points to the binary used as git server hook.
	GitHookPath string `envconfig:"GITRPC_SERVER_GIT_HOOK_PATH"`
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required")
	}
	if c.Bind == "" {
		return errors.New("config.Bind is required")
	}
	if c.GitRoot == "" {
		return errors.New("config.GitRoot is required")
	}
	if c.GitHookPath == "" {
		return errors.New("config.GitHookPath is required")
	}

	return nil
}