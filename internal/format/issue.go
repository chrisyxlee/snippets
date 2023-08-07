package format

import (
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

func Issue(issue *github.Issue) string {
	// TODO: format the repo?
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

	return fmt.Sprintf(
		"%s %s: %s%s by @%s%s",
		status,
		fmtNumber(issue),
		issue.GetTitle(),
		fmtDuration(issue),
		issue.GetUser().GetLogin(),
		fmtReactions(issue.GetReactions()))
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
	oneDay := time.Hour * 24
	oneWeek := oneDay * 7
	oneMonth := oneDay * 30
	oneYear := oneDay * 365

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

	return fmt.Sprintf(" after %s", roughDuration)

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
