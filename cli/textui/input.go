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

package textui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// Registration returns the userID, displayName, email and password from stdin.
func Registration() (string, string, string, string) {
	return UserID(), DisplayName(), Email(), Password()
}

// Credentials returns the login identifier and password from stdin.
func Credentials() (string, string) {
	return LoginIdentifier(), Password()
}

// UserID returns the user ID from stdin.
func UserID() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter User ID: ")
	uid, _ := reader.ReadString('\n')

	return strings.TrimSpace(uid)
}

// LoginIdentifier returns the login identifier from stdin.
func LoginIdentifier() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter User ID or Email: ")
	id, _ := reader.ReadString('\n')

	return strings.TrimSpace(id)
}

// DisplayName returns the display name from stdin.
func DisplayName() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Display Name: ")
	name, _ := reader.ReadString('\n')

	return strings.TrimSpace(name)
}

// Email returns the email from stdin.
func Email() string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Email: ")
	email, _ := reader.ReadString('\n')

	return strings.TrimSpace(email)
}

// Password returns the password from stdin.
func Password() string {
	fmt.Print("Enter Password: ")
	passwordb, _ := term.ReadPassword(int(syscall.Stdin))
	password := string(passwordb)

	return strings.TrimSpace(password)
}
