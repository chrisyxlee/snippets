package format

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v53/github"
	"github.com/samber/lo"
)

var (
	styleNumber = lipgloss.NewStyle().
		AlignHorizontal(lipgloss.Right).
		Bold(true).
		Width(9)
)

type CompletedIssue struct {
	ID        string
	Status    string
	Title     string
	Duration  string
	Reactions string
}

type CompletedIssueParams struct {
	ID        int
	Status    int
	Title     int
	Duration  int
	Reactions int
}

func GetCompletedIssueParams(issues []*CompletedIssue) CompletedIssueParams {
	return CompletedIssueParams{
		ID: lo.Max(lo.Map(issues, func(issue *CompletedIssue, _ int) int {
			return len(issue.ID)
		})),
		Status: lo.Max(lo.Map(issues, func(issue *CompletedIssue, _ int) int {
			return len(issue.Status)
		})),
		Title: lo.Max(lo.Map(issues, func(issue *CompletedIssue, _ int) int {
			return len(issue.Title)
		})),
		Duration: lo.Max(lo.Map(issues, func(issue *CompletedIssue, _ int) int {
			return len(issue.Duration)
		})),
		Reactions: lo.Max(lo.Map(issues, func(issue *CompletedIssue, _ int) int {
			return len(issue.Reactions)
		})),
	}
}

func (ci *CompletedIssue) Format(params CompletedIssueParams) string {
	idStr := lipgloss.NewStyle().Align(lipgloss.Left).BorderRight(true).Width(params.ID).Render(ci.ID)
	var styleStatus lipgloss.Style

	statusStyle := lipgloss.NewStyle().Width(params.Status)
	switch ci.Status {
	case "merged":
		styleStatus = statusStyle.Foreground(lipgloss.AdaptiveColor{
			Light: "#a742f5",
			Dark:  "#d194ff",
		})
	case "active":
		styleStatus = statusStyle.Foreground(lipgloss.AdaptiveColor{
			Light: "#ffaa54",
			Dark:  "#ffc994",
		})
	case "done":
		styleStatus = statusStyle.Foreground(lipgloss.AdaptiveColor{
			Light: "#87ff54",
			Dark:  "#caf7b7",
		})
	case "dropped":
		styleStatus = statusStyle.Foreground(lipgloss.AdaptiveColor{
			Light: "#333333",
			Dark:  "#878787",
		})
	}


	var buf bytes.Buffer
	buf.WriteString(idStr)
	buf.WriteString(" ")
	buf.WriteString(styleStatus.Render(ci.Status))
	if len(ci.Duration) > 0 {
		buf.WriteString(" (")
		buf.WriteString(lipgloss.NewStyle().Width(params.Duration).Render(ci.Duration))
		buf.WriteRune(')')
	}
	buf.WriteString(" - ")
	buf.WriteString(ci.Title)
	if len(ci.Reactions) > 0 {
		buf.WriteRune(' ')
		buf.WriteString(ci.Reactions)
	}

	return buf.String()
}

func ParseCompleted(issue *github.Issue) *CompletedIssue {
	// TODO: if only limited to 1 repo, then don't print
	return &CompletedIssue{
		ID:        fmtNumber(issue),
		Status:    fmtStatus(issue),
		Title:     issue.GetTitle(),
		Duration:  fmtDuration(issue),
		Reactions: fmtReactions(issue.GetReactions()),
	}
}

func Issue(issue *github.Issue) string {
	// TODO: format the repo?

	/*
		  PR #1234 | in prog | some title goes here (reactions) | 30m
		             merged
						 dropped

						 orange = in prog
						 green or purple = merged
						 grey = dropped

	*/
	return fmt.Sprintf(
		`%s %s: %s%s by @%s%s`,
		fmtStatus(issue),
		fmtNumber(issue),
		issue.GetTitle(),
		fmtDuration(issue),
		issue.GetUser().GetLogin(),
		fmtReactions(issue.GetReactions()))
}

func fmtStatus(issue *github.Issue) string {
	var status string

	if issue.GetState() == "closed" {
		if issue.IsPullRequest() {
			// TODO: how to tell if the pull request was merged or closed?
			status = "merged"
		} else {
			switch issue.GetStateReason() {
			case "not_planned":
				status = "dropped"
			case "completed":
				status = "done"
			}
		}
	} else if issue.GetStateReason() == "reopened" {
		status = "active"
	} else {
		status = "active"
	}

	return status
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

func fmtDuration(issue *github.Issue) string {
	if issue.GetState() != "closed" {
		return ""
	}

	// rough estimates, doesn't need to be exact
	dur := issue.GetClosedAt().Sub(issue.GetCreatedAt().Time)
	var roughDuration string
	switch {
	case dur > oneYear:
		roughDuration = fmt.Sprintf("%0.1fyr", dur.Seconds()/oneYear.Seconds())
	case dur > oneMonth:
		roughDuration = fmt.Sprintf("%0.1fmo", dur.Seconds()/oneMonth.Seconds())
	case dur > oneWeek:
		roughDuration = fmt.Sprintf("%0.1fw", dur.Seconds()/oneWeek.Seconds())
	case dur > oneDay:
		roughDuration = fmt.Sprintf("%0.1fd", dur.Seconds()/oneDay.Seconds())
	case dur > time.Hour:
		roughDuration = fmt.Sprintf("%0.1fh", dur.Seconds()/time.Hour.Seconds())
	case dur > time.Minute:
		roughDuration = fmt.Sprintf("%0.1fm", dur.Seconds()/time.Minute.Seconds())
	case dur > time.Second:
		roughDuration = fmt.Sprintf("%0.1fs", dur.Seconds()/time.Minute.Seconds())
	}

	return roughDuration

}

func fmtNumber(issue *github.Issue) string {
	var label string
	if issue.IsPullRequest() {
		label = "PR"
	} else {
		label = "IS"
	}

	return styleNumber.Render(fmt.Sprintf("%s #%d", label, issue.GetNumber()))
}
