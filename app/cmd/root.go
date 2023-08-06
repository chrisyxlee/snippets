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
	"github.com/google/go-github/v53/github"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	reUsername = regexp.MustCompile(`Logged in to .* as (.*) \(.*\)`)
)

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
					return group[1], nil
				}
			}
		} else {
			return "", err
		}
	}

	return "", errors.New("")
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

		categories := make(map[string][]*github.Issue)

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

		allIssues := lo.SliceToMap(issueRes.Issues, func(issue *github.Issue) (string, *github.Issue) {
			return issue.GetURL(), issue
		})

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
			allIssues[issue.GetURL()] = issue
		}

		within := func(target time.Time) bool {
			return !startTime.After(target) && !endTime.Before(target)
		}

		categories[categoryCreatedAndCompletedWithin] = moveBy(allIssues, func(issue *github.Issue) bool {
			return within(issue.GetCreatedAt().Time) && issue.GetState() == "closed"
		})
		fmt.Println(categories[categoryCreatedAndCompletedWithin][0])
		categories[categoryGeneralUpdate] = moveBy(allIssues, func(issue *github.Issue) bool {
			// TODO: and has a comment from this user
			return issue.GetClosedAt().Before(startTime)
		})
		categories[categoryLongTermFinished] = moveBy(allIssues, func(issue *github.Issue) bool {
			return issue.GetCreatedAt().Before(startTime) && issue.GetState() == "closed"
		})
		categories[categoryLongTermContinue] = moveBy(allIssues, func(issue *github.Issue) bool {
			return issue.GetState() == "open"
		})

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

		for category, issues := range categories {
			if len(issues) == 0 {
				continue
			}

			fmt.Println(category)
			for _, issue := range issues {
				fmt.Println(fmtIssue(issue))
			}
			fmt.Println("")
		}

		fmt.Println("remaining issues and prs:")
		for _, issue := range allIssues {
			fmt.Println(fmtIssue(issue))
		}

		// TODO: perhaps commits should just be through the git log
		// fmt.Println("commits:")
		// for _, commit := range commRes.Commits {
		// 	fmt.Println(commit.Commit.GetMessage())
		// }

		return nil
	},
}

func moveBy(all map[string]*github.Issue, filterFn func(*github.Issue) bool) []*github.Issue {
	out := make([]*github.Issue, 0, len(all))
	for _, k := range lo.Keys(all) {
		if filterFn(all[k]) {
			out = append(out, all[k])
			delete(all, k)
		}
	}
	return out
}

func fmtReaction(emoji string, count int) string {
	if count == 0 {
		return ""
	}

	return fmt.Sprintf("%d %s", count, emoji)
}

func fmtReactions(reactions *github.Reactions) string {
	content := strings.Join(lo.Filter([]string{
		fmtReaction("❤️", reactions.GetHeart()),
		fmtReaction("👀", reactions.GetEyes()),
		fmtReaction("👍", reactions.GetPlusOne()),
		fmtReaction("👎", reactions.GetMinusOne()),
		fmtReaction("🚀", reactions.GetRocket()),
		fmtReaction("🎉", reactions.GetHooray()),
		fmtReaction("😃", reactions.GetLaugh()),
		fmtReaction("😕", reactions.GetConfused()),
	}, func(s string, _ int) bool {
		return len(s) > 0
	}), " ")

	if len(content) > 0 {
		return fmt.Sprintf(" (%s)", content)
	}

	return ""
}

func fmtIssue(issue *github.Issue) string {
	var label string
	var status string

	if issue.GetState() == "closed" {
		if issue.IsPullRequest() {
			// TODO: how to tell if the pull request was merged or closed?
			status = "✅"
		} else {
			switch issue.GetStateReason() {
			case "not_planned":
				status = "🗑"
			case "completed":
				status = "✅"
			}
		}
	} else if issue.GetStateReason() == "reopened" {
		status = "🔁"
	} else {
		if issue.IsPullRequest() {
			status = "🚧"
		} else {
			status = "📂"
		}
	}

	if issue.IsPullRequest() {
		label = "PR"
	} else {
		label = "Issue"
	}

	return fmt.Sprintf(
		"%s %s #%d: %s by @%s%s",
		status,
		label,
		issue.GetNumber(),
		issue.GetTitle(),
		issue.GetUser().GetLogin(),
		fmtReactions(issue.GetReactions()))
}

func Execute() error {
	return rootCmd.Execute()
}

const (
	categoryCreatedAndCompletedWithin = "create_complete_within"
	categoryLongTermFinished          = "pulled_across_finish_line"
	categoryLongTermContinue          = "long_term_continue"
	categoryGeneralUpdate             = "general_update"
)

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
