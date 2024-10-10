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

package cleanup

import (
	"context"
	"fmt"
	"time"

	"github.com/harness/gitness/app/store"
	"github.com/harness/gitness/job"
	"github.com/harness/gitness/types/enum"

	"github.com/rs/zerolog/log"
)

const (
	jobTypeTokens = "gitness:cleanup:tokens"
	//nolint:gosec
	jobCronTokens        = "42 */4 * * *" // At minute 42 past every 4th hour.
	jobMaxDurationTokens = 1 * time.Minute

	// tokenRetentionTime specifies the time for which session tokens are kept even after they expired.
	// This ensures that users can still trace them after expiry for some time.
	// NOTE: I don't expect this to change much, so make it a constant instead of exposing it via config.
	tokenRetentionTime = 72 * time.Hour // 3d
)

type tokensCleanupJob struct {
	tokenStore store.TokenStore
}

func newTokensCleanupJob(
	tokenStore store.TokenStore,
) *tokensCleanupJob {
	return &tokensCleanupJob{
		tokenStore: tokenStore,
	}
}

// Handle purges old token that are expired.
func (j *tokensCleanupJob) Handle(ctx context.Context, _ string, _ job.ProgressReporter) (string, error) {
	// Don't remove PAT / SAT as they were explicitly created and are managed by user.
	expiredBefore := time.Now().Add(-tokenRetentionTime)
	log.Ctx(ctx).Info().Msgf(
		"start purging expired tokens (expired before: %s)",
		expiredBefore.Format(time.RFC3339Nano),
	)

	n, err := j.tokenStore.DeleteExpiredBefore(ctx, expiredBefore, []enum.TokenType{enum.TokenTypeSession})
	if err != nil {
		return "", fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	result := "no expired tokens found"
	if n > 0 {
		result = fmt.Sprintf("deleted %d tokens", n)
	}

	log.Ctx(ctx).Info().Msg(result)

	return result, nil
}
