// CDN-backed TACT reader: the remote counterpart of the local-CASC path,
// after TACTSharp's CDN + IndexInstance (https://github.com/wowdev/TACTSharp,
// v0.0.13-alpha, commit d0ab516eb98b5db35682467b6e4977d88955046d).
// Copyright (c) 2024 Martin Benjamins. MIT License — see tools/db2tool/NOTICES.md.
//
// Blizzard's version service (https://<region>.version.battle.net/<product>/
// versions and /cdns) names the current build and the CDN hosts. Configs,
// archive indexes and loose files come from <host>/<path>/{config,data}/
// <xx>/<yy>/<hash>; a file that lives inside an archive is an HTTP range read
// out of data/<xx>/<yy>/<archive>. The CDN serves no archive-group index for
// this product (404), so the per-archive .index files are merged in memory.
//
// Every hash-addressed download is verified. Immutable downloads are kept
// under cacheDir so a re-run of the same build is offline.

package tact

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const versionServiceURL = "https://%s.version.battle.net/%s/%s"

type archiveEntry struct {
	archive uint16
	offset  uint32
	size    uint32
}

// cdnStore serves eKeys from the CDN: archived files by range read, the rest
// as loose files.
type cdnStore struct {
	bases    []string // "<scheme>://<host>/<path>", tried in order
	cacheDir string
	client   *http.Client
	archives []string
	entries  map[[16]byte]archiveEntry
}

// OpenCDN loads the current build of product from the CDN, using the
// version service of region (e.g. "us") to find both.
func OpenCDN(product, region, cacheDir string) (*Build, error) {
	s := &cdnStore{
		cacheDir: cacheDir,
		client:   &http.Client{Timeout: 10 * time.Minute},
	}
	versions, err := s.versionService(product, region, "versions")
	if err != nil {
		return nil, err
	}
	entry := AvailableBuild{
		Product:     product,
		Version:     versions["VersionsName"],
		BuildConfig: versions["BuildConfig"],
	}
	buildNumber, err := BuildNumber(entry.Version)
	if err != nil {
		return nil, err
	}

	cdns, err := s.versionService(product, region, "cdns")
	if err != nil {
		return nil, err
	}
	// The Servers column carries full origins (http://host/?maxhosts=8); the
	// Hosts column bare hostnames.
	for server := range strings.FieldsSeq(cdns["Servers"]) {
		if u, _, ok := strings.Cut(server, "/?"); ok {
			s.bases = append(s.bases, u+"/"+cdns["Path"])
		}
	}
	for host := range strings.FieldsSeq(cdns["Hosts"]) {
		s.bases = append(s.bases, "https://"+host+"/"+cdns["Path"])
	}
	if len(s.bases) == 0 {
		return nil, fmt.Errorf("cdns for %s lists no servers", product)
	}

	buildConfig, err := s.config(entry.BuildConfig)
	if err != nil {
		return nil, fmt.Errorf("loading build config: %w", err)
	}
	cdnConfig, err := s.config(versions["CDNConfig"])
	if err != nil {
		return nil, fmt.Errorf("loading CDN config: %w", err)
	}
	s.archives = cdnConfig["archives"]
	s.entries = make(map[[16]byte]archiveEntry)
	for i, hash := range s.archives {
		raw, err := s.cached("data/"+hashPath(hash)+".index", "")
		if err != nil {
			return nil, fmt.Errorf("archive index %s: %w", hash, err)
		}
		if err := s.mergeArchiveIndex(uint16(i), raw); err != nil {
			return nil, fmt.Errorf("archive index %s: %w", hash, err)
		}
	}

	return load(entry, buildNumber, buildConfig, s)
}

// versionService fetches one of the Ribbit-over-HTTP files (versions, cdns)
// and returns the row for region, or the first row when region is absent.
func (s *cdnStore) versionService(product, region, file string) (map[string]string, error) {
	url := fmt.Sprintf(versionServiceURL, region, product, file)
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: HTTP %s", url, resp.Status)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rows := parsePSV(string(raw))
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s: no entries (unknown product?)", url)
	}
	for _, row := range rows {
		if row["Region"] == region || row["Name"] == region {
			return row, nil
		}
	}
	return rows[0], nil
}

// parsePSV reads the version service's pipe-separated tables: a header line
// of "Name!TYPE:len" columns, "## seqn" comment lines, then one row per line.
func parsePSV(raw string) []map[string]string {
	var header []string
	var rows []map[string]string
	for line := range strings.SplitSeq(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "|")
		if header == nil {
			for _, f := range fields {
				header = append(header, strings.SplitN(f, "!", 2)[0])
			}
			continue
		}
		row := make(map[string]string, len(header))
		for i, f := range fields {
			if i < len(header) {
				row[header[i]] = f
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func hashPath(hash string) string {
	return hash[0:2] + "/" + hash[2:4] + "/" + hash
}

func (s *cdnStore) config(hash string) (map[string][]string, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("invalid config hash %q", hash)
	}
	raw, err := s.cached("config/"+hashPath(hash), hash)
	if err != nil {
		return nil, err
	}
	return parseConfig(raw)
}

// get fetches rel from the first CDN base that serves it. A Range header
// asks for a partial response.
func (s *cdnStore) get(rel, byteRange string) ([]byte, error) {
	var lastErr error
	for _, base := range s.bases {
		url := base + "/" + rel
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		want := http.StatusOK
		if byteRange != "" {
			req.Header.Set("Range", byteRange)
			want = http.StatusPartialContent
		}
		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		switch {
		case err != nil:
			lastErr = fmt.Errorf("%s: %w", url, err)
		case resp.StatusCode != want:
			lastErr = fmt.Errorf("%s: HTTP %s", url, resp.Status)
		default:
			return body, nil
		}
	}
	return nil, lastErr
}

// cached fetches an immutable hash-addressed file, keeping a copy under
// cacheDir. A non-empty md5 (hex) is checked against both the cached copy
// and a fresh download.
func (s *cdnStore) cached(rel, md5hex string) ([]byte, error) {
	path := filepath.Join(s.cacheDir, filepath.FromSlash(rel))
	if data, err := os.ReadFile(path); err == nil && (md5hex == "" || md5Hex(data) == md5hex) {
		return data, nil
	}
	data, err := s.get(rel, "")
	if err != nil {
		return nil, err
	}
	if md5hex != "" && md5Hex(data) != md5hex {
		return nil, fmt.Errorf("%s: md5 mismatch", rel)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	// A temp file of its own, so two reads of one file at the same time never write the same one.
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return nil, err
	}
	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmp.Name())
		return nil, err
	}
	return data, os.Rename(tmp.Name(), path)
}

// mergeArchiveIndex adds every entry of one archive's .index to the store.
// Layout (IndexInstance): fixed-size blocks of packed entries, a TOC of each
// block's last key and hash, then a 28-byte footer: toc hash(8),
// formatRevision, flags0, flags1, blockSizeKB, offsetBytes, sizeBytes,
// keyBytes, hashBytes, numElements(u32 LE), footer hash(8). Entries are
// key, big-endian size, big-endian offset.
func (s *cdnStore) mergeArchiveIndex(archive uint16, raw []byte) error {
	if len(raw) < 28 {
		return fmt.Errorf("too small for an index footer")
	}
	f := raw[len(raw)-20:]
	if f[0] != 1 {
		return fmt.Errorf("unsupported index format revision %d", f[0])
	}
	blockSize := int(f[3]) << 10
	offsetBytes, sizeBytes, keyBytes, hashBytes := int(f[4]), int(f[5]), int(f[6]), int(f[7])
	numElements := int(binary.LittleEndian.Uint32(f[8:12]))
	if keyBytes != 16 || sizeBytes != 4 || offsetBytes != 4 {
		return fmt.Errorf("unexpected entry layout key=%d size=%d offset=%d", keyBytes, sizeBytes, offsetBytes)
	}
	entrySize := keyBytes + sizeBytes + offsetBytes
	entriesPerBlock := blockSize / entrySize
	numBlocks := (numElements + entriesPerBlock - 1) / entriesPerBlock
	if numBlocks*blockSize+numBlocks*(keyBytes+hashBytes)+28 != len(raw) {
		return fmt.Errorf("size %d does not match %d elements in %d-byte blocks", len(raw), numElements, blockSize)
	}
	for block := range numBlocks {
		count := entriesPerBlock
		if block == numBlocks-1 {
			count = numElements - block*entriesPerBlock
		}
		p := raw[block*blockSize:]
		for range count {
			var key [16]byte
			copy(key[:], p[:16])
			s.entries[key] = archiveEntry{
				archive: archive,
				size:    binary.BigEndian.Uint32(p[16:20]),
				offset:  binary.BigEndian.Uint32(p[20:24]),
			}
			p = p[entrySize:]
		}
	}
	return nil
}

// readEKey returns the raw BLTE bytes for an eKey: a range read from its
// archive, or the loose data/<xx>/<yy>/<ekey> file when no archive holds it.
func (s *cdnStore) readEKey(eKey []byte) ([]byte, error) {
	if len(eKey) != 16 {
		return nil, fmt.Errorf("unsupported ekey length %d", len(eKey))
	}
	var key [16]byte
	copy(key[:], eKey)
	hexKey := hex.EncodeToString(eKey)
	var raw []byte
	if e, ok := s.entries[key]; ok {
		var err error
		raw, err = s.get("data/"+hashPath(s.archives[e.archive]), fmt.Sprintf("bytes=%d-%d", e.offset, int64(e.offset)+int64(e.size)-1))
		if err != nil {
			return nil, fmt.Errorf("ekey %s: %w", hexKey, err)
		}
	} else {
		var err error
		raw, err = s.cached("data/"+hashPath(hexKey), "")
		if err != nil {
			return nil, fmt.Errorf("ekey %s not in any archive and not a loose file: %w", hexKey, err)
		}
	}
	if got := encodedKey(raw); got != hexKey {
		return nil, fmt.Errorf("ekey %s: downloaded content hashes to %s", hexKey, got)
	}
	return raw, nil
}

// encodedKey is the MD5 the encoding key of a BLTE blob covers: the whole
// blob for a headerless single chunk, otherwise just the header (whose
// chunk table already carries each chunk's checksum).
func encodedKey(raw []byte) string {
	if len(raw) >= 8 && string(raw[:4]) == "BLTE" {
		if headerSize := int(be32(raw[4:])); headerSize > 0 && headerSize <= len(raw) {
			return md5Hex(raw[:headerSize])
		}
	}
	return md5Hex(raw)
}

func md5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}
