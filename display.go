package main

import (
	"fmt"
	"strings"
	"time"
)

const maxPreviewLen = 72

func DisplayList(notes []Note) {
	if len(notes) == 0 {
		fmt.Println("no notes yet")
		return
	}

	idWidth := len(fmt.Sprintf("%d", notes[0].ID))

	for _, n := range notes {
		body := n.Body
		if len(body) > maxPreviewLen {
			body = body[:maxPreviewLen-1] + "…"
		}
		age := formatAge(n.CreatedAt)
		fmt.Printf("  %*d  %-*s  %s\n", idWidth, n.ID, maxPreviewLen, body, dim(age))
	}
}

func DisplayNote(n *Note) {
	age := formatAge(n.CreatedAt)
	fmt.Printf("%s  %s\n\n%s\n", dim(fmt.Sprintf("#%d", n.ID)), dim(age), n.Body)
}

func dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}

func formatAge(createdAt string) string {
	t, err := time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		return createdAt
	}

	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		return plural(m, "min")
	case d < 24*time.Hour:
		h := int(d.Hours())
		return plural(h, "hr")
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		if days < 30 {
			return plural(days, "day")
		}
		return t.Format("Jan 2")
	}
}

func plural(n int, unit string) string {
	s := fmt.Sprintf("%d %s", n, unit)
	if n != 1 {
		s += "s"
	}
	s += " ago"
	return strings.TrimSpace(s)
}
