// Package selfupdate replaces a running wowsimforever or wowsimcli with the newest release on GitHub and
// restarts it.
package selfupdate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"syscall"
	"time"

	"github.com/Masterminds/semver/v3"
	goselfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/creativeprojects/go-selfupdate/update"
)

var (
	ErrDevelopmentBuild = errors.New("a development build does not update itself")
	ErrNotWritable      = errors.New("the folder holding the executable is not writable")
)

// Set to 1 in the environment of a process Restart started.
const restartedFromEnv = "WOWSIMS_FOREVER_UPDATED_FROM"

var repository = goselfupdate.NewRepositorySlug("wowsims", "forever")

type Config struct {
	Name       string              // The binary's name in the release assets and inside their archives, e.g. "wowsimforever".
	Version    string              // The running build's version tag; "" or "development" is a local build.
	Executable string              // The file an update replaces; ForExecutable passes os.Executable().
	Source     goselfupdate.Source // Where releases come from; ForExecutable passes the releases on GitHub.
}

// An updater for the running executable, reading the releases of wowsims/forever on GitHub with $GITHUB_TOKEN when
// it is set.
func ForExecutable(name, version string) (*Updater, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	source, err := goselfupdate.NewGitHubSource(goselfupdate.GitHubConfig{})
	if err != nil {
		return nil, err
	}
	return New(Config{Name: name, Version: version, Executable: executable, Source: source})
}

type Updater struct {
	name       string
	version    *semver.Version // nil for a build that is not a release.
	executable string
	source     goselfupdate.Source
	validator  goselfupdate.Validator
	updater    *goselfupdate.Updater
}

// A newer release than the running build.
type Release struct {
	Tag string // The release's tag, e.g. "v0.0.100".
	URL string // The release's page.
	rel *goselfupdate.Release
}

func New(cfg Config) (*Updater, error) {
	// The update writes next to the file itself, so a symlink is followed to it.
	executable, err := filepath.EvalSymlinks(cfg.Executable)
	if err != nil {
		return nil, err
	}

	// With a filter set, the library matches assets by the filter alone and skips its own OS and arch check, so the
	// filter names both; asset names are lowercased before matching.
	filter := fmt.Sprintf(`^%s_%s_%s\.zip$`, regexp.QuoteMeta(cfg.Name), runtime.GOOS, runtime.GOARCH)
	validator := &goselfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"}
	updater, err := goselfupdate.NewUpdater(goselfupdate.Config{
		Source:    cfg.Source,
		Validator: validator,
		Filters:   []string{filter},
	})
	if err != nil {
		return nil, err
	}

	u := &Updater{name: cfg.Name, executable: executable, source: cfg.Source, validator: validator, updater: updater}
	if cfg.Version != "" && cfg.Version != "development" {
		// A version that is not semver cannot be compared with a release, so it is treated as a local build.
		u.version, _ = semver.NewVersion(cfg.Version)
	}
	return u, nil
}

// The newest release when it is strictly newer than the running build, or nil when there is none.
func (u *Updater) Check(ctx context.Context) (*Release, error) {
	if u.version == nil {
		return nil, ErrDevelopmentBuild
	}
	rel, found, err := u.updater.DetectLatest(ctx, repository)
	if err != nil || !found {
		return nil, err
	}
	if !semver.MustParse(rel.Version()).GreaterThan(u.version) {
		return nil, nil
	}
	return &Release{Tag: "v" + rel.Version(), URL: rel.URL, rel: rel}, nil
}

// The library's own UpdateTo looks inside the archive for a file named like the executable, which fails when the
// user renamed it, so the archive is read here by the release name instead.
func (u *Updater) Apply(ctx context.Context, release *Release) error {
	if !u.CanWrite() {
		return ErrNotWritable
	}
	rel := release.rel
	data, err := u.download(ctx, rel, rel.AssetID)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", rel.AssetName, err)
	}
	sums, err := u.download(ctx, rel, rel.ValidationAssetID)
	if err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}
	if err := u.validator.Validate(rel.AssetName, data, sums); err != nil {
		return fmt.Errorf("checking %s: %w", rel.AssetName, err)
	}
	binary, err := goselfupdate.DecompressCommand(bytes.NewReader(data), rel.AssetName, u.name, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	if err := update.Apply(binary, update.Options{TargetPath: u.executable}); err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("%w: %w", ErrNotWritable, err)
		}
		return err
	}
	return nil
}

func (u *Updater) download(ctx context.Context, rel *goselfupdate.Release, assetID int64) ([]byte, error) {
	reader, err := u.source.DownloadReleaseAsset(ctx, rel, assetID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// Whether a file can be created next to the executable, which replacing it needs.
func (u *Updater) CanWrite() bool {
	f, err := os.CreateTemp(filepath.Dir(u.executable), ".wowsims-write-check-*")
	if err != nil {
		return false
	}
	f.Close()
	os.Remove(f.Name())
	return true
}

// Starts the executable again with args. On Windows it starts as a child with the same console and this process
// exits at once, so a shell that started the simulator gets its prompt back; the process must not wait, because a
// second update renames the executable onto the .old file it would still run from. Returns only on failure.
func (u *Updater) Restart(args []string) error {
	env := append(os.Environ(), restartedFromEnv+"=1")
	if runtime.GOOS != "windows" {
		return syscall.Exec(u.executable, append([]string{u.executable}, args...), env)
	}
	cmd := exec.Command(u.executable, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}

// Whether this process was started by Restart after an update. A caller skips updating again on startup, so a
// release stamped with the wrong version cannot restart in a loop.
func Restarted() bool {
	return os.Getenv(restartedFromEnv) != ""
}

// Removes the previous executable an update left behind. Windows cannot delete a running executable, so the update
// renames it to a hidden .<name>.old, and the process that replaced it may still be exiting: the removal is retried
// for a few seconds.
func (u *Updater) RemoveOld() {
	old := filepath.Join(filepath.Dir(u.executable), "."+filepath.Base(u.executable)+".old")
	for range 10 {
		if err := os.Remove(old); err == nil || errors.Is(err, fs.ErrNotExist) {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}
