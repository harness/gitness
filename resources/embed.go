// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"embed"
	"fmt"
	"strings"
)

var (
	//go:embed gitignore
	gitignore embed.FS

	//go:embed license
	licence embed.FS
)

// Licenses returns map of licences in license folder.
func Licenses() ([]byte, error) {
	return licence.ReadFile("license/index.json")
}

// ReadLicense reads licence from license folder.
func ReadLicense(name string) ([]byte, error) {
	content, err := licence.ReadFile(fmt.Sprintf("license/%s.txt", name))
	if err != nil {
		return nil, err
	}
	return content, err
}

// GitIgnores lists all files in gitignore folder and return file names.
func GitIgnores() ([]string, error) {
	entries, err := gitignore.ReadDir("gitignore")
	if err != nil {
		return []string{}, err
	}
	var files []string
	for _, filename := range entries {
		if strings.HasSuffix(filename.Name(), ".gitignore") {
			files = append(files, strings.ReplaceAll(filename.Name(), ".gitignore", ""))
		}
	}
	return files, nil
}

// ReadGitIgnore reads gitignore file from license folder.
func ReadGitIgnore(name string) ([]byte, error) {
	return gitignore.ReadFile(fmt.Sprintf("gitignore/%s.gitignore", name))
}
