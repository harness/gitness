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

package request

import (
	"net/http"
	"strconv"

	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/enum"
)

const (
	PathParamRemainder = "*"

	QueryParamCreatedBy = "created_by"
	QueryParamSort      = "sort"
	QueryParamOrder     = "order"
	QueryParamQuery     = "query"
	QueryParamRecursive = "recursive"

	QueryParamState = "state"
	QueryParamKind  = "kind"
	QueryParamType  = "type"

	QueryParamAfter  = "after"
	QueryParamBefore = "before"

	QueryParamDeletedBeforeOrAt = "deleted_before_or_at"
	QueryParamDeletedAt         = "deleted_at"

	QueryParamPage  = "page"
	QueryParamLimit = "limit"
	PerPageDefault  = 30
	PerPageMax      = 100

	// TODO: have shared constants across all services?
	HeaderRequestID       = "X-Request-Id"
	HeaderUserAgent       = "User-Agent"
	HeaderAuthorization   = "Authorization"
	HeaderContentEncoding = "Content-Encoding"
)

// GetOptionalRemainderFromPath returns the remainder ("*") from the path or an empty string if it doesn't exist.
func GetOptionalRemainderFromPath(r *http.Request) string {
	return PathParamOrEmpty(r, PathParamRemainder)
}

// GetRemainderFromPath returns the remainder ("*") from the path or an error if it doesn't exist.
func GetRemainderFromPath(r *http.Request) (string, error) {
	return PathParamOrError(r, PathParamRemainder)
}

// ParseQuery extracts the query parameter from the url.
func ParseQuery(r *http.Request) string {
	return r.URL.Query().Get(QueryParamQuery)
}

// ParsePage extracts the page parameter from the url.
func ParsePage(r *http.Request) int {
	s := r.URL.Query().Get(QueryParamPage)
	i, _ := strconv.Atoi(s)
	if i <= 0 {
		i = 1
	}
	return i
}

// ParseLimit extracts the limit parameter from the url.
func ParseLimit(r *http.Request) int {
	s := r.URL.Query().Get(QueryParamLimit)
	i, _ := strconv.Atoi(s)
	if i <= 0 {
		i = PerPageDefault
	} else if i > PerPageMax {
		i = PerPageMax
	}
	return i
}

// ParseOrder extracts the order parameter from the url.
func ParseOrder(r *http.Request) enum.Order {
	return enum.ParseOrder(
		r.URL.Query().Get(QueryParamOrder),
	)
}

// ParseSort extracts the sort parameter from the url.
func ParseSort(r *http.Request) string {
	return r.URL.Query().Get(QueryParamSort)
}

// ParsePaginationFromRequest parses pagination related info from the url.
func ParsePaginationFromRequest(r *http.Request) types.Pagination {
	return types.Pagination{
		Page: ParsePage(r),
		Size: ParseLimit(r),
	}
}

// ParseListQueryFilterFromRequest parses pagination and query related info from the url.
func ParseListQueryFilterFromRequest(r *http.Request) types.ListQueryFilter {
	return types.ListQueryFilter{
		Query:      ParseQuery(r),
		Pagination: ParsePaginationFromRequest(r),
	}
}

// GetContentEncodingFromHeadersOrDefault returns the content encoding from the request headers.
func GetContentEncodingFromHeadersOrDefault(r *http.Request, dflt string) string {
	return GetHeaderOrDefault(r, HeaderContentEncoding, dflt)
}

// ParseRecursiveFromQuery extracts the recursive option from the URL query.
func ParseRecursiveFromQuery(r *http.Request) (bool, error) {
	return QueryParamAsBoolOrDefault(r, QueryParamRecursive, false)
}

// GetDeletedAtFromQueryOrError gets the exact resource deletion timestamp from the query.
func GetDeletedAtFromQueryOrError(r *http.Request) (int64, error) {
	return QueryParamAsPositiveInt64(r, QueryParamDeletedAt)
}

// GetDeletedBeforeOrAtFromQuery gets the resource deletion timestamp from the query as an optional parameter.
func GetDeletedBeforeOrAtFromQuery(r *http.Request) (int64, bool, error) {
	value, err := QueryParamAsPositiveInt64OrDefault(r, QueryParamDeletedBeforeOrAt, 0)
	if err != nil {
		return 0, false, err
	}
	if value == 0 {
		return value, false, nil
	}

	return value, true, nil
}
