package page

import (
	"github.com/chrisyxlee/snippets/internal"
	"github.com/chrisyxlee/snippets/internal/ratelimit"
	"github.com/google/go-github/v45/github"
)

// Details are options that the caller can return to describe how to proceed
// with pagination.
type Details struct {
	// Whether pagination should stop now at the caller's request.
	StopEarly bool
}

// Paginate handles pagination according to the google/go-github API. It handles the
// rate limit error by waiting if it receives one. This allows the caller to handle
// just the business logic without having to worry about pagination. The details the
// caller can return control whether the function returns early.
func Paginate(detail string, perPage int, fn func(listOptions github.ListOptions) (Details, *github.Response, error)) error {
	pageNum := 1
	for {
		internal.Log().Debug().Int("items", perPage).Int("page", pageNum).Str("detail", detail).Msgf("paginating")

		options := github.ListOptions{
			PerPage: perPage,
			Page:    pageNum,
		}

		details, resp, err := fn(options)
		if ratelimit.WaitIfRateLimited(err) {
			// Don't increase the page number since we were rate limited and need to try the request again.
			continue
		}

		if err != nil {
			return err
		}

		if details.StopEarly || resp.LastPage == pageNum {
			return nil
		}

		pageNum = resp.NextPage
	}
}
