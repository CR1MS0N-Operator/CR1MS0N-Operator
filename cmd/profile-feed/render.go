package main

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

func markdown(data snapshot) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Latest releases & research\n\nChecked %s (UTC). Refreshed daily from public sources.\n\n", data.Checked)
	for _, section := range []struct {
		title string
		items []item
	}{{"Releases", data.Releases}, {"Research", data.Research}} {
		fmt.Fprintf(&b, "## %s\n\n", section.title)
		if len(section.items) == 0 {
			b.WriteString("No published entries yet.\n\n")
		}
		for _, p := range section.items {
			fmt.Fprintf(&b, "- **[%s](%s)** · %s", escapeMarkdown(p.Title), markdownURL(p.URL), p.Published[:10])
			if p.Prerelease {
				b.WriteString(" · Prerelease")
			}
			if p.SummaryURL != "" {
				fmt.Fprintf(&b, " · [Executive report](%s)", markdownURL(p.SummaryURL))
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
	b.WriteString("Raw research lives in [security-research](https://github.com/CR1MS0N-Operator/security-research). [Azrael Security](https://azraelsecurity.vercel.app/research) is the research portfolio; executive-report links appear above only after publication.\n\n[Profile](https://github.com/CR1MS0N-Operator) · [Refresh workflow](https://github.com/CR1MS0N-Operator/CR1MS0N-Operator/actions/workflows/refresh-profile.yml)\n")
	return b.String()
}

func escapeMarkdown(s string) string {
	return strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "*", "\\*", "_", "\\_", "`", "\\`", "<", "&lt;", ">", "&gt;", "!", "\\!", "|", "\\|").Replace(s)
}

func markdownURL(s string) string {
	u, _ := url.Parse(s) // Already validated before rendering.
	return strings.NewReplacer("(", "%28", ")", "%29").Replace(u.String())
}

// wrap bounds image height; the linked Markdown edition retains complete titles.
func wrap(s string) []string {
	const width = 34
	var lines []string
	runes := []rune(s)
	for len(runes) > width && len(lines) < 2 {
		cut := width
		for i := width; i > width/2; i-- {
			if runes[i] == ' ' {
				cut = i
				break
			}
		}
		lines = append(lines, string(runes[:cut]))
		runes = []rune(strings.TrimSpace(string(runes[cut:])))
	}
	if len(lines) < 2 && len(runes) > 0 {
		lines = append(lines, string(runes))
	} else if len(runes) > 0 {
		lines[1] = strings.TrimRight(lines[1], " .") + "…"
	}
	return lines
}

func svg(data snapshot, dark bool) string {
	background, foreground, muted, border, accent := "#f6f8fa", "#1f2328", "#59636e", "#d1d9e0", "#a91639"
	if dark {
		background, foreground, muted, border, accent = "#0d1117", "#f0f6fc", "#a0aab5", "#30363d", "#fb7185"
	}
	var content strings.Builder
	y := 40
	for _, section := range []struct {
		title string
		items []item
	}{{"LATEST RELEASES", data.Releases}, {"RECENT RESEARCH", data.Research}} {
		fmt.Fprintf(&content, `<text x="24" y="%d" fill="%s" font-size="18" font-family="monospace" letter-spacing="1">%s</text>`, y, accent, section.title)
		y += 38
		if len(section.items) == 0 {
			fmt.Fprintf(&content, `<text x="24" y="%d" fill="%s" font-size="24">No published entries yet.</text>`, y, muted)
			y += 45
		}
		for _, p := range section.items {
			for _, line := range wrap(p.Title) {
				fmt.Fprintf(&content, `<text x="24" y="%d" fill="%s" font-size="26" font-weight="600">%s</text>`, y, foreground, html.EscapeString(line))
				y += 32
			}
			date := p.Published[:10]
			if p.Prerelease {
				date += " · Prerelease"
			}
			fmt.Fprintf(&content, `<text x="24" y="%d" fill="%s" font-size="19">%s</text>`, y-3, muted, date)
			y += 36
		}
		fmt.Fprintf(&content, `<path d="M24 %dH576" stroke="%s"/>`, y-10, border)
		y += 26
	}
	fmt.Fprintf(&content, `<text x="24" y="%d" fill="%s" font-size="18">Checked %s UTC · Open text edition for links</text>`, y, muted, data.Checked)
	height := y + 24
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="%d" viewBox="0 0 600 %d" role="img" aria-labelledby="title desc"><title id="title">Latest releases and research</title><desc id="desc">Daily public-source digest, checked %s UTC. Full titles and individual links are in the linked text edition.</desc><rect width="600" height="%d" rx="12" fill="%s"/><g font-family="-apple-system,BlinkMacSystemFont,Segoe UI,Helvetica,Arial,sans-serif">%s</g></svg>`+"\n", height, height, data.Checked, height, background, content.String())
}
