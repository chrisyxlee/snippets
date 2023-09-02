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
		Bold(true)
)

type CompletedIssue struct {
	Type      string
	ID        string
	Status    string
	Title     string
	Duration  string
	Reactions string
}

type completedIssueWidths struct {
	Type      int
	ID        int
	Status    int
	Title     int
	Duration  int
	Reactions int
}

func GetCompletedIssueParams(issues []*CompletedIssue) completedIssueWidths {
	return completedIssueWidths{
		Type: lo.Max(lo.Map(issues, func(issue *CompletedIssue, _ int) int {
			return len(issue.Type)
		})),
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

func (ci *CompletedIssue) Format(params completedIssueWidths) string {
	typeStr := lipgloss.NewStyle().Align(lipgloss.Right).BorderRight(true).Width(params.Type).Render(ci.Type)
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
	buf.WriteString(typeStr)
	buf.WriteRune(' ')
	buf.WriteString(idStr)
	buf.WriteString(" ")
	buf.WriteString(styleStatus.Render(ci.Status))
	if len(ci.Duration) > 0 {
		buf.WriteRune(' ')
		buf.WriteString(lipgloss.NewStyle().Width(params.Duration).Render(ci.Duration))
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
		Type:      fmtType(issue),
		ID:        styleNumber.Render(fmt.Sprintf("#%d", issue.GetNumber())),
		Status:    fmtStatus(issue),
		Title:     issue.GetTitle(),
		Duration:  fmtDuration(issue),
		Reactions: fmtReactions(issue.GetReactions()),
	}
}

func ParseAllCompleted(issues []*github.Issue) []*CompletedIssue {
	return lo.Map(issues, func(issue *github.Issue, _ int) *CompletedIssue {
		return ParseCompleted(issue)
	})
}

func FormatSection(title string, issues []*github.Issue) string {
	if len(issues) == 0 {
		return ""
	}

	var section bytes.Buffer
	updatedIssues := ParseAllCompleted(issues)
	section.WriteString("## ")
	section.WriteString(title)
	section.WriteString("\n\n")
	updatedIssuesParams := GetCompletedIssueParams(updatedIssues)
	for _, issue := range updatedIssues {
		section.WriteString(issue.Format(updatedIssuesParams))
		section.WriteRune('\n')
	}
	section.WriteRune('\n')

	return section.String()
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
	// rough estimates, doesn't need to be exact
	var dur time.Duration
	if issue.GetState() == "closed" {
		dur = issue.GetClosedAt().Sub(issue.GetCreatedAt().Time)
	} else {
		dur = time.Since(issue.CreatedAt.Time)
	}

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

	if len(roughDuration) > 0 {
		return fmt.Sprintf("(%s)", roughDuration)
	}

	return ""
}

func fmtType(issue *github.Issue) string {
	var label string
	if issue.IsPullRequest() {
		label = "PR"
	} else {
		label = "IS"
	}

	return styleNumber.Render(label)
}
