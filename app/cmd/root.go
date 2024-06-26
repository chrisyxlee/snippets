package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/chrisyxlee/snippets/internal"
	"github.com/chrisyxlee/snippets/internal/format"
	"github.com/google/go-github/v53/github"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	reUsername        = regexp.MustCompile(`Logged in to .* (account|as) (.*) \(.*\)`)
	reOwnerRepository = regexp.MustCompile(`https://github.com/(.*)/(.*)/(pull|issue)/\d+`)
)

// TODO: create a pie chart for how long you spent on each issue? (length of comment / most number of comments in this cycle) -- gantt??
// don't want this to be a slippery slope into MTTR lol

// TODO: view report (give directory, use glamour?)

func getGitHubToken() (string, error) {
	githubToken, ok := os.LookupEnv("GITHUB_TOKEN")
	if ok && len(githubToken) > 0 {
		internal.Log().Debug().Msg("fetching token from GITHUB_TOKEN")
		return githubToken, nil
	}

	githubOauthToken, ok := os.LookupEnv("GITHUB_OAUTH_TOKEN")
	if ok && len(githubOauthToken) > 0 {
		internal.Log().Debug().Msg("fetching token from GITHUB_OAUTH_TOKEN")
		return githubOauthToken, nil
	}

	if _, err := exec.LookPath("gh"); err == nil {
		internal.Log().Debug().Msg("user has gh installed")
		var b bytes.Buffer
		ghAuthCmd := exec.Command("gh", "auth", "token")
		ghAuthCmd.Stdout = &b
		if err = ghAuthCmd.Run(); err == nil {
			token := strings.Trim(b.String(), "\t\n ")
			if len(token) > 0 {
				return token, nil
			}
			// TODO: ask user for permission to use the token
			internal.Log().Debug().Msg("user has gh installed and is logged in")
		}
	}

	return "", errors.New("github token must be provided through GITHUB_TOKEN or GITHUB_OAUTH_TOKEN environment variables or through the gh CLI")
}

func fmtDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func getUsername() (string, error) {
	if _, err := exec.LookPath("gh"); err == nil {
		internal.Log().Debug().Msg("user has gh installed")
		var b bytes.Buffer
		ghUserCmd := exec.Command("gh", "auth", "status")
		ghUserCmd.Stdout = &b
		if err = ghUserCmd.Run(); err == nil {
			lines := strings.Split(b.String(), "\n")
			for _, line := range lines {
				if group := reUsername.FindStringSubmatch(line); len(group) > 1 {
					// TODO: ask user for permission to use the token
					internal.Log().Debug().Msg("user has gh installed and is logged in")
					return group[2], nil
				}
			}
		} else {
			return "", err
		}
	}

	return "", errors.New("")
}

func getOwnerAndRepository(htmlURL string) (string, string, error) {
	if group := reOwnerRepository.FindStringSubmatch(htmlURL); len(group) > 2 {
		return group[1], group[2], nil
	}

	return "", "", errors.New("no match for owner and repository")
}

var rootCmd = &cobra.Command{
	Use:   "snippet",
	Short: "TODO",
	Long:  `TODO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		githubToken, err := getGitHubToken()
		if err != nil {
			return err
		}

		username, err := getUsername()
		if err != nil {
			internal.Log().Err(err).Msg("no username")
		} else {
			internal.Log().Info().Str("username", username).Msg("got username")
		}

		ctx := cmd.Context()
		client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: githubToken,
		})))

		/* Select repos to search in? */

		/*
		 Issues that were recently created.
		*/
		endTime := time.Now()
		startTime := endTime.Add(-2 * 7 * 24 * time.Hour)

		internal.Log().Debug().
			Str("start time", fmtDate(startTime)).
			Str("end time", fmtDate(endTime)).
			Msg("using time range")

		issCreateQuery := fmt.Sprintf("author:%s created:%s..%s",
			username,
			fmtDate(startTime),
			fmtDate(endTime),
		)
		internal.Log().Debug().
			Str("query", issCreateQuery).
			Msg("query issues by created time")
		issueRes, _, err := client.Search.Issues(
			ctx,
			issCreateQuery,
			&github.SearchOptions{
				Sort:  "updated",
				Order: "asc",
			})
		if err != nil {
			return fmt.Errorf("search issues with query `%s`: %w", issCreateQuery, err)
		}

		addIsMerged := func(issue *github.Issue, _ string) *format.GitHubIssue {
			if issue.IsPullRequest() {
				var isMerged bool
				owner, repo, err := getOwnerAndRepository(issue.GetHTMLURL())
				if err != nil {
					internal.Log().Err(err).
						Str("html_url", issue.GetHTMLURL()).
						Msg("get owner and repository")
				} else {
					isMerged, _, err = client.PullRequests.IsMerged(
						ctx,
						owner,
						repo,
						issue.GetNumber())
					if err != nil {
						internal.Log().Err(err).Msg("check pull request is merged")
						isMerged = false
					}
				}

				return &format.GitHubIssue{
					Merged: isMerged,
					Issue:  issue,
				}
			}

			return &format.GitHubIssue{
				Issue: issue,
			}
		}

		allIssues := lo.SliceToMap(issueRes.Issues, func(issue *github.Issue) (string, *github.Issue) {
			return issue.GetURL(), issue
		})
		ghIssues := lo.MapValues(allIssues, addIsMerged)

		issModQuery := fmt.Sprintf("author:%s updated:%s..%s",
			username,
			fmtDate(startTime),
			fmtDate(endTime),
		)
		internal.Log().Debug().Str("query", issModQuery).Msg("query issues by modified time")
		issueRes, _, err = client.Search.Issues(
			ctx,
			issModQuery,
			&github.SearchOptions{
				Sort:  "updated",
				Order: "asc",
			})
		if err != nil {
			return fmt.Errorf("search issues with query `%s`: %w", issModQuery, err)
		}

		for _, issue := range issueRes.Issues {
			ghIssues[issue.GetURL()] = addIsMerged(issue, issue.GetURL())
		}

		within := func(target time.Time) bool {
			return !startTime.After(target) && !endTime.Before(target)
		}

		// TODO: ask for user to input summary that can be placed in here?

		var report bytes.Buffer
		// weekly report for username: YYYY-mm-dd
		report.WriteString(fmt.Sprintf(`# %s report for %s: %s

`,
			format.DurationAsAdj(endTime.Sub(startTime)),
			username,
			fmtDate(startTime)))

		report.WriteString(format.FormatSection("Completed this cycle",
			moveBy(
				ghIssues,
				func(ghi *format.GitHubIssue) bool {
					return ghi.Issue.GetState() == "closed" && (within(ghi.Issue.GetCreatedAt().Time) || ghi.Issue.GetCreatedAt().Before(startTime))
				})))

		report.WriteString(format.FormatSection("Updated this cycle",
			moveBy(
				ghIssues,
				func(ghi *format.GitHubIssue) bool {
					// TODO: and has a recent comment from this user
					return ghi.Issue.GetClosedAt().Before(startTime)
				})))

		report.WriteString(format.FormatSection("Remaining",
			moveBy(
				ghIssues,
				func(ghi *format.GitHubIssue) bool {
					return true
				})))

		// TODO: allow editing the final report
		// TODO: write the report somewhere (dump into a file?)
		// TODO: optional, allow json so that we can format more

		/* Issues that were commented on
		 */

		/*
		   Find PRs that were updated
		*/

		/* Find PRs that were merged
		 */

		/* Find PRs that were reviewed by username */

		/* Find commits that actually made it through and attach to the PR link?
		 */

		/* Releases that had your changes?
		 */

		/* Slack
		   Find messages in a certain time window
		*/

		//commQuery := fmt.Sprintf("author:%s author-date:>%s merge:true", username, oneWeekAgo)
		//internal.Log().Debug().Str("query", commQuery).Msg("search commits")
		//commRes, _, err := client.Search.Commits(ctx, commQuery, &github.SearchOptions{})
		////internal.Log().Info().Array("issues", zerolog.Arr().Interface(result.Issues)).Msg("")
		//if err != nil {
		//	return fmt.Errorf("search commit: %w", err)
		//}

		fmt.Println(report.String())

		/*
			for category, issues := range categories {
				if len(issues) == 0 {
					continue
				}

				fmt.Println(category)
				for _, issue := range issues {
					fmt.Println(format.Issue(issue))
				}
				fmt.Println("")
			}

			fmt.Println("remaining issues and prs:")
			for _, issue := range ghIssues {
				fmt.Println(format.Issue(issue))
			}
		*/

		// TODO: perhaps commits should just be through the git log
		// fmt.Println("commits:")
		// for _, commit := range commRes.Commits {
		// 	fmt.Println(commit.Commit.GetMessage())
		// }

		return nil
	},
}

// Moves all items passing the filter into the output slice. Items that are
// matching the filter are removed from the original map.
func moveBy[T any](all map[string]T, filterFn func(T) bool) []T {
	out := make([]T, 0, len(all))
	for _, k := range lo.Keys(all) {
		if filterFn(all[k]) {
			out = append(out, all[k])
			delete(all, k)
		}
	}
	return out
}

func Execute() error {
	return rootCmd.Execute()
}

/*
End goal: ??

For the period of <start> to <end> (<duration>)...

You created and merged these PRs:
...

You pulled these PRs across the finish line:
...

You started or continued work on these PRs:
...

You've filed X new issues:
...

You've updated these issues:
...

*/

/* monthly snippets are maybe betteR? or biweekly? */
