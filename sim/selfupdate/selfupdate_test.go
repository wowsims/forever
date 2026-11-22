package selfupdate

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	goselfupdate "github.com/creativeprojects/go-selfupdate"
)

type fakeAsset struct {
	id   int64
	name string
	data []byte
}

func (a *fakeAsset) GetID() int64                  { return a.id }
func (a *fakeAsset) GetName() string               { return a.name }
func (a *fakeAsset) GetSize() int                  { return len(a.data) }
func (a *fakeAsset) GetBrowserDownloadURL() string { return "https://example.invalid/" + a.name }

type fakeRelease struct {
	tag    string
	assets []*fakeAsset
}

func (r *fakeRelease) GetID() int64              { return 1 }
func (r *fakeRelease) GetTagName() string        { return r.tag }
func (r *fakeRelease) GetDraft() bool            { return false }
func (r *fakeRelease) GetPrerelease() bool       { return false }
func (r *fakeRelease) GetPublishedAt() time.Time { return time.Time{} }
func (r *fakeRelease) GetReleaseNotes() string   { return "" }
func (r *fakeRelease) GetName() string           { return r.tag }
func (r *fakeRelease) GetURL() string            { return "https://example.invalid/" + r.tag }
func (r *fakeRelease) GetAssets() []goselfupdate.SourceAsset {
	assets := make([]goselfupdate.SourceAsset, len(r.assets))
	for i, a := range r.assets {
		assets[i] = a
	}
	return assets
}

type fakeSource struct {
	release *fakeRelease
}

func (s *fakeSource) ListReleases(context.Context, goselfupdate.Repository) ([]goselfupdate.SourceRelease, error) {
	return []goselfupdate.SourceRelease{s.release}, nil
}

func (s *fakeSource) DownloadReleaseAsset(_ context.Context, _ *goselfupdate.Release, id int64) (io.ReadCloser, error) {
	for _, a := range s.release.assets {
		if a.id == id {
			return io.NopCloser(bytes.NewReader(a.data)), nil
		}
	}
	return nil, fmt.Errorf("no asset %d", id)
}

func exeName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func zipped(t *testing.T, inner string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	f, err := w.Create(inner)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// A release of both binaries for every platform, each archive holding a binary whose content names its asset, and the
// checksums of all of them. The wowsimcli and other-platform assets come first, so only the filter picks the right one.
func release(t *testing.T, tag string) *fakeRelease {
	t.Helper()
	rel := &fakeRelease{tag: tag}
	var sums bytes.Buffer
	add := func(binary, goos, goarch string) {
		name := fmt.Sprintf("%s_%s_%s.zip", binary, goos, goarch)
		inner := binary
		if goos == "windows" {
			inner += ".exe"
		}
		data := zipped(t, inner, []byte(binary+" "+tag+" "+goos+"/"+goarch))
		rel.assets = append(rel.assets, &fakeAsset{id: int64(len(rel.assets) + 1), name: name, data: data})
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256(data), name)
	}
	for _, binary := range []string{"wowsimcli", "wowsimforever"} {
		for _, goos := range []string{"linux", "windows", "darwin"} {
			for _, goarch := range []string{"amd64", "arm64"} {
				if binary == "wowsimforever" && goos == runtime.GOOS && goarch == runtime.GOARCH {
					continue
				}
				add(binary, goos, goarch)
			}
		}
	}
	add("wowsimforever", runtime.GOOS, runtime.GOARCH)
	rel.assets = append(rel.assets, &fakeAsset{id: int64(len(rel.assets) + 1), name: "checksums.txt", data: sums.Bytes()})
	return rel
}

func newUpdater(t *testing.T, version, exe string, rel *fakeRelease) *Updater {
	t.Helper()
	if exe == "" {
		exe = exeName("wowsimforever")
	}
	path := filepath.Join(t.TempDir(), exe)
	if err := os.WriteFile(path, []byte("old binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	u, err := New(Config{Name: "wowsimforever", Version: version, Executable: path, Source: &fakeSource{release: rel}})
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func readExecutable(t *testing.T, u *Updater) string {
	t.Helper()
	data, err := os.ReadFile(u.executable)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCheckFindsOnlyAStrictlyNewerRelease(t *testing.T) {
	for _, tc := range []struct {
		version string
		newer   bool
	}{
		{"v0.0.99", true},
		{"v0.0.100", false},
		{"v0.0.101", false},
		{"v0.1.0", false},
	} {
		u := newUpdater(t, tc.version, "", release(t, "v0.0.100"))
		rel, err := u.Check(context.Background())
		if err != nil {
			t.Fatalf("%s: %v", tc.version, err)
		}
		if (rel != nil) != tc.newer {
			t.Errorf("%s against v0.0.100: found an update = %v, want %v", tc.version, rel != nil, tc.newer)
		}
		if rel != nil && rel.Tag != "v0.0.100" {
			t.Errorf("%s: release tag %q, want v0.0.100", tc.version, rel.Tag)
		}
	}
}

func TestDevelopmentBuildNeverUpdates(t *testing.T) {
	for _, version := range []string{"", "development", "local-build"} {
		u := newUpdater(t, version, "", release(t, "v0.0.100"))
		if _, err := u.Check(context.Background()); !errors.Is(err, ErrDevelopmentBuild) {
			t.Errorf("version %q: Check gave %v, want ErrDevelopmentBuild", version, err)
		}
	}
}

func TestApplyInstallsThisPlatformsWowsimforever(t *testing.T) {
	u := newUpdater(t, "v0.0.99", "", release(t, "v0.0.100"))
	rel, err := u.Check(context.Background())
	if err != nil || rel == nil {
		t.Fatalf("Check gave %v, %v; want the v0.0.100 release", rel, err)
	}
	if err := u.Apply(context.Background(), rel); err != nil {
		t.Fatal(err)
	}
	want := "wowsimforever v0.0.100 " + runtime.GOOS + "/" + runtime.GOARCH
	if got := readExecutable(t, u); got != want {
		t.Errorf("executable holds %q, want %q", got, want)
	}
}

func TestApplyReplacesARenamedExecutable(t *testing.T) {
	u := newUpdater(t, "v0.0.99", exeName("wowsimforever-windows"), release(t, "v0.0.100"))
	rel, err := u.Check(context.Background())
	if err != nil || rel == nil {
		t.Fatalf("Check gave %v, %v; want the v0.0.100 release", rel, err)
	}
	if err := u.Apply(context.Background(), rel); err != nil {
		t.Fatal(err)
	}
	if got := readExecutable(t, u); got != "wowsimforever v0.0.100 "+runtime.GOOS+"/"+runtime.GOARCH {
		t.Errorf("renamed executable holds %q", got)
	}
}

func TestApplyRejectsAChecksumMismatch(t *testing.T) {
	rel := release(t, "v0.0.100")
	sums := rel.assets[len(rel.assets)-1]
	sums.data = bytes.ReplaceAll(sums.data, []byte(fmt.Sprintf("%x", sha256.Sum256(rel.assets[len(rel.assets)-2].data))),
		[]byte(fmt.Sprintf("%x", sha256.Sum256([]byte("something else")))))
	u := newUpdater(t, "v0.0.99", "", rel)
	found, err := u.Check(context.Background())
	if err != nil || found == nil {
		t.Fatalf("Check gave %v, %v; want the v0.0.100 release", found, err)
	}
	if err := u.Apply(context.Background(), found); !errors.Is(err, goselfupdate.ErrChecksumValidationFailed) {
		t.Errorf("Apply gave %v, want ErrChecksumValidationFailed", err)
	}
	if got := readExecutable(t, u); got != "old binary" {
		t.Errorf("executable holds %q after a rejected update", got)
	}
}

func TestApplyReportsAFolderItCannotWrite(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("folder permissions do not stop this user")
	}
	u := newUpdater(t, "v0.0.99", "", release(t, "v0.0.100"))
	rel, err := u.Check(context.Background())
	if err != nil || rel == nil {
		t.Fatalf("Check gave %v, %v; want the v0.0.100 release", rel, err)
	}
	dir := filepath.Dir(u.executable)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
	if u.CanWrite() {
		t.Error("CanWrite is true in a read-only folder")
	}
	if err := u.Apply(context.Background(), rel); !errors.Is(err, ErrNotWritable) {
		t.Errorf("Apply gave %v, want ErrNotWritable", err)
	}
}

func TestRemoveOldDeletesTheLeftoverExecutable(t *testing.T) {
	u := newUpdater(t, "v0.0.99", "", release(t, "v0.0.100"))
	old := filepath.Join(filepath.Dir(u.executable), "."+filepath.Base(u.executable)+".old")
	if err := os.WriteFile(old, []byte("previous binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	u.RemoveOld()
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("%s is still there: %v", old, err)
	}
}
