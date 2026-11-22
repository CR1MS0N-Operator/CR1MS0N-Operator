package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 11, 22, 19, 0, 0, 0, time.UTC)

func TestGeneratePublicOnlySnapshot(t *testing.T) {
	var requested []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = append(requested, r.URL.Path)
		switch r.URL.Path {
		case "/users/" + owner + "/repos":
			fmt.Fprintf(w, `[{"full_name":%q,"private":false},{"full_name":"%s/nightforge","private":false},{"full_name":"%s/secret","private":true},{"full_name":"%s/fork","private":false,"fork":true}]`, researchRepo, owner, owner, owner)
		case "/repos/" + researchRepo + "/releases":
			io.WriteString(w, "[]")
		case "/repos/" + owner + "/nightforge/releases":
			fmt.Fprintf(w, `[
			{"tag_name":"v2","html_url":"https://github.com/%s/nightforge/releases/tag/v2","published_at":"2026-11-20T00:00:00Z","prerelease":true},
			{"tag_name":"v1","html_url":"https://github.com/%s/nightforge/releases/tag/v1","published_at":"2026-10-01T00:00:00Z"},
			{"tag_name":"draft","draft":true,"published_at":"2026-11-22T00:00:00Z"},
			{"tag_name":"v3","html_url":"https://github.com/%s/nightforge/releases/tag/v3","published_at":"2027-01-01T00:00:00Z"}]`, owner, owner, owner)
		case "/catalog":
			fmt.Fprintf(w, `{"publications":[
			{"title":"Old paper","published_at":"2026-02-08","url":"https://github.com/%s/blob/main/writeups/old.md"},
			{"title":"New paper","published_at":"2026-04-12","url":"https://github.com/%s/blob/main/research/paper.md","summary_url":"https://azraelsecurity.vercel.app/research/paper"},
			{"title":"Future paper","published_at":"2027-01-01","url":"https://github.com/%s/blob/main/research/future.md"}]}`, researchRepo, researchRepo, researchRepo)
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	out := filepath.Join(t.TempDir(), "feed")
	s := source{client: server.Client(), api: server.URL}
	if err := generate(s, server.URL+"/catalog", out, testNow); err != nil {
		t.Fatal(err)
	}
	encoded, err := os.ReadFile(filepath.Join(out, "feed.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got snapshot
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Releases) != 2 || got.Releases[0].Title != "nightforge v2" || !got.Releases[0].Prerelease {
		t.Fatalf("wrong release filtering/sorting: %+v", got.Releases)
	}
	if len(got.Research) != 2 || got.Research[0].Title != "New paper" || got.Research[0].SummaryURL == "" {
		t.Fatalf("wrong research sorting: %+v", got.Research)
	}
	for _, path := range requested {
		if strings.Contains(path, "/secret/") || strings.Contains(path, "/fork/") {
			t.Fatalf("non-public or fork release endpoint fetched: %s", path)
		}
	}
	files, err := os.ReadDir(out)
	if err != nil || len(files) != 4 {
		t.Fatalf("expected exactly four artifacts: %v, %v", files, err)
	}
	if err := generate(s, server.URL+"/catalog", out, testNow.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	again, _ := os.ReadFile(filepath.Join(out, "feed.json"))
	if string(again) != string(encoded) {
		t.Fatal("unchanged same-day refresh should be deterministic")
	}
}

func TestFailurePreservesExistingFeed(t *testing.T) {
	for _, failure := range []string{"repos", "release", "catalog", "private-research"} {
		t.Run(failure, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasPrefix(r.URL.Path, "/users/"):
					if failure == "repos" {
						w.WriteHeader(http.StatusTooManyRequests)
					} else {
						fmt.Fprintf(w, `[{"full_name":%q,"private":%t}]`, researchRepo, failure == "private-research")
					}
				case strings.HasSuffix(r.URL.Path, "/releases"):
					if failure == "release" {
						w.WriteHeader(http.StatusServiceUnavailable)
					} else {
						io.WriteString(w, "[]")
					}
				default:
					io.WriteString(w, `{"publications":[{"published_at":"not-a-date"}]}`)
				}
			}))
			defer server.Close()
			out := t.TempDir()
			for _, name := range []string{"LATEST.md", "latest-dark.svg", "latest-light.svg", "feed.json"} {
				if err := os.WriteFile(filepath.Join(out, name), []byte("last good snapshot"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if err := generate(source{client: server.Client(), api: server.URL}, server.URL+"/catalog", out, testNow); err == nil {
				t.Fatal("expected refresh to fail")
			}
			files, _ := os.ReadDir(out)
			for _, file := range files {
				body, _ := os.ReadFile(filepath.Join(out, file.Name()))
				if string(body) != "last good snapshot" {
					t.Fatalf("previous feed changed on failure: %s", file.Name())
				}
			}
		})
	}
}

func TestCatalogValidation(t *testing.T) {
	for _, body := range []string{
		`{}`, `{"publications":null}`, `[]`,
		`{"publications":[{"title":"Paper","published_at":"2026-02-30","url":"https://github.com/CR1MS0N-Operator/security-research/blob/main/paper.md"}]}`,
		`{"publications":[{"title":"Paper","published_at":"2026-04-12","url":"https://github.com/other/research/blob/main/paper.md"}]}`,
		`{"publications":[{"title":"Paper","published_at":"2026-04-12","url":"javascript:alert(1)"}]}`,
		`{"publications":[{"title":"Paper","published_at":"2026-04-12","url":"https://github.com/CR1MS0N-Operator/security-research/blob/main/paper.md","summary_url":"https://evil.example/research/paper"}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, body) }))
			defer server.Close()
			if _, err := (source{client: server.Client(), api: server.URL}).research(server.URL, testNow); err == nil {
				t.Fatal("invalid catalog accepted")
			}
		})
	}
}

func TestPaginationAndLatestLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/users/") {
			count := 100
			if r.URL.Query().Get("page") == "2" {
				count = 1
			}
			rows := make([]repository, count)
			public := false
			for i := range rows {
				rows[i] = repository{FullName: owner + "/test", Private: &public}
			}
			json.NewEncoder(w).Encode(rows)
			return
		}
		rows := make([]release, 100)
		for i := range rows {
			rows[i] = release{Tag: "v1", Published: "2026-01-01T00:00:00Z", URL: "https://github.com/" + owner + "/test/releases/tag/v1"}
		}
		if r.URL.Query().Get("page") == "2" {
			rows = []release{{Tag: "v2", Published: "2026-11-21T00:00:00Z", URL: "https://github.com/" + owner + "/test/releases/tag/v2"}}
		}
		json.NewEncoder(w).Encode(rows)
	}))
	defer server.Close()
	s := source{client: server.Client(), api: server.URL}
	repos, err := s.publicRepositories()
	if err != nil || len(repos) != 101 {
		t.Fatalf("repository pagination: %d, %v", len(repos), err)
	}
	items, err := s.releases([]string{owner + "/test"}, testNow)
	if err != nil || len(items) != 2 || items[0].Title != "test v2" {
		t.Fatalf("release pagination/deduplication: %+v, %v", items, err)
	}
	if got := latest([]item{{URL: "a"}, {URL: "b"}, {URL: "c"}, {URL: "d"}}); len(got) != 3 {
		t.Fatalf("limit not enforced: %+v", got)
	}
}

func TestTokenDoesNotReachRawContent(t *testing.T) {
	raw := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("token leaked to content host")
		}
		io.WriteString(w, `{"publications":[]}`)
	}))
	defer raw.Close()
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("API token missing")
		}
		io.WriteString(w, `[]`)
	}))
	defer api.Close()
	s := source{client: api.Client(), api: api.URL, token: "test-token"}
	if _, err := s.publicRepositories(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.research(raw.URL, testNow); err != nil {
		t.Fatal(err)
	}
}

func TestRenderingEscapesMetadata(t *testing.T) {
	data := snapshot{Checked: "2026-11-22", Research: []item{{Title: `[click](https://evil.example) <script>& "'`, Published: "2026-04-12", URL: "https://github.com/" + researchRepo + "/blob/main/a(b).md"}}}
	md := markdown(data)
	if strings.Contains(md, "[click](https://evil.example)") || strings.Contains(md, "<script>") || !strings.Contains(md, "a%28b%29.md") {
		t.Fatalf("unsafe Markdown: %s", md)
	}
	for _, dark := range []bool{false, true} {
		body := svg(data, dark)
		if strings.Contains(body, "<script>") || !strings.Contains(body, "&lt;") {
			t.Fatalf("unsafe SVG: %s", body)
		}
		decoder := xml.NewDecoder(strings.NewReader(body))
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("invalid SVG XML: %v", err)
			}
		}
	}
	if lines := wrap(strings.Repeat("longtitle", 100)); len(lines) != 2 || !strings.HasSuffix(lines[1], "…") {
		t.Fatalf("unbounded title: %v", lines)
	}
}
