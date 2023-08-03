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

	githubToken, ok = os.LookupEnv("GITHUB_OAUTH_TOKEN")
	if ok && len(githubToken) > 0 {
		internal.Log().Debug().Msg("fetching token from GITHUB_OAUTH_TOKEN")
		return githubToken, nil
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
					return group[1], nil
				}
			}
			// TODO: ask user for permission to use the token
			internal.Log().Debug().Msg("user has gh installed and is logged in")
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

		oneWeekAgo := time.Now().Add(-7 * 24 * time.Hour).Format("2006-01-02")
		issueRes, _, err := client.Search.Issues(ctx, fmt.Sprintf("author:%s created:>%s", username, oneWeekAgo), &github.SearchOptions{
			Sort:  "updated",
			Order: "asc",
		})
		if err != nil {
			return fmt.Errorf("search issues: %w", err)
		}

		commRes, _, err := client.Search.Commits(ctx, fmt.Sprintf("author:%s author-date:>%s merge:true", username, oneWeekAgo), &github.SearchOptions{})
		//internal.Log().Info().Array("issues", zerolog.Arr().Interface(result.Issues)).Msg("")
		if err != nil {
			return fmt.Errorf("search commit: %w", err)
		}

		fmt.Println("issues:")
		for _, issue := range issueRes.Issues {
			fmt.Println(issue.GetTitle())
		}

		fmt.Println("commits:")
		for _, commit := range commRes.Commits {
			fmt.Println(commit.Commit.GetMessage())
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
