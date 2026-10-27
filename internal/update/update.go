// Package update checks for and installs new claude-dx releases from GitHub Releases (ADR-005).
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/mod/semver"
)

// LatestURL is the GitHub API URL of the latest release.
const LatestURL = "https://api.github.com/repos/EqualSoftware/claude-dx/releases/latest"

const (
	timeout   = 10 * time.Second
	interval  = 24 * time.Hour
	stateFile = "update.json"
	maxSize   = 200 << 20 // limit for one download
)

// client limits the connection and the wait for the response headers, but not the body,
// because a release archive can take more than 10 seconds on a slow network.
var client = &http.Client{Transport: &http.Transport{
	Proxy:                 http.ProxyFromEnvironment,
	DialContext:           (&net.Dialer{Timeout: timeout}).DialContext,
	TLSHandshakeTimeout:   timeout,
	ResponseHeaderTimeout: timeout,
}}

// Updater checks and installs releases for one version of claude-dx.
type Updater struct {
	URL     string // latest release API URL
	Version string // current version; a version that is not semver (example: "dev") never updates
	Cache   string // cache directory that holds update.json
	Now     func() time.Time
}

// Release is the part of the GitHub release that update uses.
type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []Asset `json:"assets"`
}

// Asset is one file of a release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// State is the content of <cache>/update.json.
type State struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

// Newer reports whether latest is a newer semver version than current.
func Newer(latest, current string) bool {
	return semver.IsValid(latest) && semver.IsValid(current) && semver.Compare(latest, current) > 0
}

// Latest gets the latest release.
func (u Updater) Latest(ctx context.Context) (Release, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	body, err := u.get(ctx, u.URL)
	if err != nil {
		return Release{}, err
	}
	var r Release
	if err := json.Unmarshal(body, &r); err != nil {
		return Release{}, fmt.Errorf("the release data is not valid: %w", err)
	}
	if !semver.IsValid(r.Tag) {
		return Release{}, fmt.Errorf("the latest release tag %q is not a semver version", r.Tag)
	}
	return r, nil
}

// Update installs the latest release in place of binary, if it is newer. It returns a message for the user.
func (u Updater) Update(ctx context.Context, binary string) (string, error) {
	if !semver.IsValid(u.Version) {
		return "", fmt.Errorf("%s builds do not update. Install a release", u.Version)
	}
	r, err := u.Latest(ctx)
	if err != nil {
		return "", err
	}
	if !Newer(r.Tag, u.Version) {
		return fmt.Sprintf("claude-dx is up to date (%s).", u.Version), u.writeState(r.Tag)
	}
	if err := u.Apply(ctx, r, binary); err != nil {
		return "", err
	}
	return fmt.Sprintf("claude-dx is updated to %s.", r.Tag), nil
}

// Check gets the latest release and writes update.json, if the last check is older than 24 hours.
// It ignores all errors, because it runs in the background and the user did not ask for it.
func (u Updater) Check(ctx context.Context) {
	if !semver.IsValid(u.Version) {
		return
	}
	if s, err := ReadState(u.Cache); err == nil && u.now().Sub(s.CheckedAt) < interval {
		return
	}
	r, err := u.Latest(ctx)
	if err != nil {
		return
	}
	_ = u.writeState(r.Tag)
}

// Notice returns the line that tells the user about a new release, or "".
func Notice(cache, version string) string {
	s, err := ReadState(cache)
	if err != nil || !Newer(s.Latest, version) {
		return ""
	}
	return fmt.Sprintf("claude-dx %s is available. Run: claude-dx update", s.Latest)
}

// ReadState reads <cache>/update.json.
func ReadState(cache string) (State, error) {
	data, err := os.ReadFile(filepath.Join(cache, stateFile))
	if err != nil {
		return State{}, err
	}
	var s State
	return s, json.Unmarshal(data, &s)
}

// writeState writes update.json to a temporary file and renames it, because sessions share the cache.
func (u Updater) writeState(latest string) error {
	data, err := json.Marshal(State{CheckedAt: u.now().UTC(), Latest: latest})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(u.Cache, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(u.Cache, "update-*.json")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err := errors.Join(err, f.Close()); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	return os.Rename(f.Name(), filepath.Join(u.Cache, stateFile))
}

func (u Updater) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// The GitHub API refuses requests with no User-Agent.
	req.Header.Set("User-Agent", "claude-dx/"+u.Version)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxSize))
}

func (u Updater) now() time.Time {
	if u.Now != nil {
		return u.Now()
	}
	return time.Now()
}
