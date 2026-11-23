# Daily profile feed

The README shows the three newest public releases and three newest published research items, with matching light/dark SVGs. Each image links to a full-text Markdown edition with individual publication links. The text edition also provides a screen-reader-friendly alternative to the artwork.

## Sources and publication rules

- Releases come from GitHub's public owner-repository list, excluding forks and private repositories. All release pages are checked, then sorted by `published_at`; drafts and future entries are excluded. Prereleases are included and explicitly labeled. Repository names and tags are displayed, not unverified release-body claims.
- Research comes from [`security-research/publications.json`](https://github.com/CR1MS0N-Operator/security-research/blob/main/publications.json). Entries must have a real title, a `YYYY-MM-DD` publication date, and a raw-source `url` under that repository's `blob/main/` path. A source edit or commit is not a new publication. Future-dated items are withheld; duplicate URLs are deduplicated.
- An optional `summary_url` may point to a **published** executive report under `https://azraelsecurity.vercel.app/research/`. Raw GitHub research remains canonical. Choosing and writing the site's top 3–10 executive reports is a separate editorial workflow, not something this refresh invents or automatically writes.

The research catalog must be merged before the first successful refresh. A valid empty `publications` array renders an honest empty state; a missing or malformed catalog fails the run instead of erasing the last good snapshot.

## How it updates

[`Refresh profile feed`](../.github/workflows/refresh-profile.yml) runs daily at **13:17 UTC**, on changes to its generator on `main`, or via **Run workflow** on its [Actions page](https://github.com/CR1MS0N-Operator/CR1MS0N-Operator/actions/workflows/refresh-profile.yml). Manual publication is restricted to `main` as well.

1. A read-only job fetches public metadata and generates `LATEST.md`, `feed.json`, and two themed SVGs. Only GitHub API calls receive the built-in `GITHUB_TOKEN`; raw content receives no credential. HTTP redirects are rejected.
2. A separate write-permission job publishes only those four data files on `profile-feed`. It never checks out or executes output-branch content. New commits are append-only, and identical snapshots are skipped.
3. The README embeds those public SVGs and links to the text edition. `main` is not edited or pushed by the refresh, so its PR and review requirements remain intact. No personal token, external stats service, model API, or paid service is required.

The checked date is the last successful snapshot's UTC date, not a claim that a release or paper was published that day. A source/API/validation failure stops publication and leaves the previous feed unchanged. Check failed workflow runs before assuming the feed is current. GitHub schedules can be delayed, and inactive public repositories can have scheduled workflows disabled after 60 days; the [GitHub schedule documentation](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#schedule) covers those limits. GitHub's image proxy may also cache an image briefly after publication.

## Local verification and preview

Requires the Go version in `go.mod`, with no third-party modules:

```bash
go test ./...
go vet ./...
go run ./cmd/profile-feed -output .profile-feed
```

Local reads work anonymously. Optionally supply a GitHub token through `GITHUB_TOKEN` to avoid the anonymous API rate limit; public visibility is still checked before releases are fetched. Never paste credentials into arguments, files, or commits.

To preview a proposed catalog, use `-catalog https://raw.githubusercontent.com/CR1MS0N-Operator/security-research/BRANCH/publications.json`. The scheduled workflow always uses the catalog on `main`. No network request is needed for the fixture-based tests, which cover visibility, drafts, future dates, pagination, sorting/deduplication, credential isolation, failed refresh preservation, and Markdown/SVG escaping.

## Recovery

Fix the source catalog or generator through a normal PR, then rerun the workflow on `main`. If the generated branch was deleted, a successful run recreates it automatically. If output-branch protection or Actions permissions prevent publication, the job fails visibly; do not loosen `main` protections or add a privileged personal token to bypass them.
