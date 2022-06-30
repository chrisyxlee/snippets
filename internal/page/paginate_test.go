package page_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/chrisyxlee/snippets/internal/page"
	"github.com/google/go-github/v45/github"
	"github.com/stretchr/testify/assert"
)

func TestStopEarly(t *testing.T) {
	testCount := 0
	assert.NoError(t, page.Paginate("test", 10, func(listOptions github.ListOptions) (page.Details, *github.Response, error) {
		testCount++

		if listOptions.Page == 2 {
			return page.Details{
				StopEarly: true,
			}, &github.Response{}, nil
		}

		return page.Details{}, &github.Response{
			NextPage: listOptions.Page + 1,
			LastPage: 3,
		}, nil
	}))
	assert.Equal(t, 2, testCount)
}

func TestUntilLastPageWithRateLimitOnEveryCall(t *testing.T) {
	testCount := 0
	totalPages := 3
	visited := make(map[int]bool)
	assert.NoError(t, page.Paginate("test", 10, func(listOptions github.ListOptions) (page.Details, *github.Response, error) {
		testCount++

		if visited[listOptions.Page] {
			return page.Details{}, &github.Response{
				NextPage: listOptions.Page + 1,
				LastPage: totalPages,
			}, nil
		}

		visited[listOptions.Page] = true
		return page.Details{}, &github.Response{
				NextPage: listOptions.Page + 1,
				LastPage: totalPages,
			}, &github.RateLimitError{
				Rate: github.Rate{
					Reset: github.Timestamp{
						Time: time.Now().Add(-5 * time.Minute),
					},
				},
			}
	}))
	// Each page should be visited twice because of the rate limit.
	assert.Equal(t, totalPages*2, testCount)
}

func TestPassAlongError(t *testing.T) {
	assert.Error(t, page.Paginate("test", 10, func(listOptions github.ListOptions) (page.Details, *github.Response, error) {
		return page.Details{}, &github.Response{}, fmt.Errorf("some error")
	}))
}
