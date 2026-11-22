// profile-feed publishes a small, public-only snapshot; it never edits the profile source.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const owner = "CR1MS0N-Operator"
const researchRepo = owner + "/security-research"
const catalogURL = "https://raw.githubusercontent.com/" + researchRepo + "/main/publications.json"
const limit = 3

type item struct {
	Title      string `json:"title"`
	Published  string `json:"published_at"`
	URL        string `json:"url"`
	SummaryURL string `json:"summary_url,omitempty"`
	Prerelease bool   `json:"prerelease,omitempty"`
}

type snapshot struct {
	Checked  string `json:"checked_at"`
	Releases []item `json:"releases"`
	Research []item `json:"research"`
}

type repository struct {
	FullName string `json:"full_name"`
	Private  *bool  `json:"private"`
	Fork     bool   `json:"fork"`
}

type release struct {
	Tag        string `json:"tag_name"`
	URL        string `json:"html_url"`
	Published  string `json:"published_at"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

type source struct {
	client *http.Client
	api    string
	token  string
}

func (s source) get(address string, target any) error {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	// Never send the workflow token to raw content or redirected hosts.
	if strings.HasPrefix(address, s.api+"/") && s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	req.Header.Set("User-Agent", "CR1MS0N-profile-feed")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", address, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch %s: HTTP %d", address, resp.StatusCode)
	}
	const maxBytes = 2 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return err
	}
	if len(body) > maxBytes {
		return fmt.Errorf("response too large: %s", address)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("invalid JSON from %s: %w", address, err)
	}
	return nil
}

func (s source) publicRepositories() ([]string, error) {
	var names []string
	for page := 1; page <= 100; page++ {
		var repos []repository
		address := fmt.Sprintf("%s/users/%s/repos?type=owner&per_page=100&page=%d", s.api, owner, page)
		if err := s.get(address, &repos); err != nil {
			return nil, err
		}
		if repos == nil {
			return nil, errors.New("expected a public repository array")
		}
		for _, repo := range repos {
			if repo.Private == nil {
				return nil, errors.New("repository visibility missing")
			}
			parts := strings.Split(repo.FullName, "/")
			if len(parts) != 2 || !strings.EqualFold(parts[0], owner) || parts[1] == "" {
				return nil, errors.New("unexpected repository owner or name")
			}
			if !*repo.Private && !repo.Fork {
				names = append(names, repo.FullName)
			}
		}
		if len(repos) < 100 {
			return names, nil
		}
	}
	return nil, errors.New("too many repository pages")
}

func (s source) releases(repos []string, now time.Time) ([]item, error) {
	var items []item
	for _, repo := range repos {
		for page := 1; ; page++ {
			if page > 100 {
				return nil, errors.New("too many release pages")
			}
			var releases []release
			address := fmt.Sprintf("%s/repos/%s/releases?per_page=100&page=%d", s.api, repo, page)
			if err := s.get(address, &releases); err != nil {
				return nil, err
			}
			if releases == nil {
				return nil, fmt.Errorf("expected releases array for %s", repo)
			}
			for _, r := range releases {
				if r.Draft || r.Published == "" {
					continue
				}
				date, err := time.Parse(time.RFC3339, r.Published)
				if err != nil || !validURL(r.URL, "github.com", "/"+repo+"/releases/tag/") || clean(r.Tag) == "" {
					return nil, fmt.Errorf("invalid release metadata for %s", repo)
				}
				if date.After(now) {
					continue
				}
				items = append(items, item{Title: strings.Split(repo, "/")[1] + " " + clean(r.Tag), Published: date.UTC().Format(time.RFC3339), URL: r.URL, Prerelease: r.Prerelease})
			}
			if len(releases) < 100 {
				break
			}
		}
	}
	return latest(items), nil
}

func (s source) research(address string, now time.Time) ([]item, error) {
	var catalog struct {
		Publications *[]item `json:"publications"`
	}
	if err := s.get(address, &catalog); err != nil {
		return nil, err
	}
	if catalog.Publications == nil {
		return nil, errors.New("research catalog must contain a publications array")
	}
	var items []item
	for _, p := range *catalog.Publications {
		date, err := time.Parse(time.DateOnly, p.Published)
		if err != nil || clean(p.Title) == "" || !validURL(p.URL, "github.com", "/"+researchRepo+"/blob/main/") {
			return nil, errors.New("invalid research title, publication date, or raw source URL")
		}
		if p.SummaryURL != "" && !validURL(p.SummaryURL, "azraelsecurity.vercel.app", "/research/") {
			return nil, errors.New("invalid executive report URL")
		}
		if date.After(now) {
			continue
		}
		p.Title = clean(p.Title)
		p.Prerelease = false
		items = append(items, p)
	}
	return latest(items), nil
}

func validURL(raw, host, prefix string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != host || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	return strings.HasPrefix(u.Path, prefix) && len(u.Path) > len(prefix) && !strings.Contains(u.Path, "/../")
}

func clean(s string) string { return strings.Join(strings.Fields(s), " ") }

func latest(items []item) []item {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Published == items[j].Published {
			return items[i].URL < items[j].URL
		}
		return items[i].Published > items[j].Published
	})
	result := make([]item, 0, limit)
	seen := make(map[string]bool)
	for _, p := range items {
		if !seen[p.URL] {
			result = append(result, p)
			seen[p.URL] = true
		}
		if len(result) == limit {
			break
		}
	}
	return result
}

func generate(s source, researchURL, output string, now time.Time) error {
	repos, err := s.publicRepositories()
	if err != nil {
		return err
	}
	publicResearch := false
	for _, repo := range repos {
		publicResearch = publicResearch || strings.EqualFold(repo, researchRepo)
	}
	if !publicResearch {
		return errors.New("research repository is not public; preserving previous feed")
	}
	releases, err := s.releases(repos, now)
	if err != nil {
		return err
	}
	research, err := s.research(researchURL, now)
	if err != nil {
		return err
	}
	data := snapshot{Checked: now.UTC().Format(time.DateOnly), Releases: releases, Research: research}
	return writeSnapshot(output, data)
}

func main() {
	output := flag.String("output", "feed", "directory for generated feed files")
	catalog := flag.String("catalog", catalogURL, "public publication catalog URL (override to preview a catalog branch)")
	flag.Parse()
	s := source{client: &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, api: "https://api.github.com", token: os.Getenv("GITHUB_TOKEN")}
	if err := generate(s, *catalog, *output, time.Now()); err != nil {
		fmt.Fprintln(os.Stderr, "profile-feed:", err)
		os.Exit(1)
	}
}

func writeSnapshot(output string, data snapshot) error {
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	files := map[string][]byte{
		"feed.json":        append(encoded, '\n'),
		"LATEST.md":        []byte(markdown(data)),
		"latest-light.svg": []byte(svg(data, false)),
		"latest-dark.svg":  []byte(svg(data, true)),
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(output, name), body, 0644); err != nil {
			return err
		}
	}
	return nil
}
