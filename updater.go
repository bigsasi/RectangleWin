// Copyright 2026 bigsasi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/gonutz/w32/v2"
)

// version is set at build time via -ldflags "-X main.version=1.2.3".
var version = "dev"

const (
	githubOwner = "bigsasi"
	githubRepo  = "RectangleWin"
)

var (
	githubAPIBase    = "https://api.github.com"
	updateHTTPClient = &http.Client{Timeout: 30 * time.Second}

	updateMu     sync.Mutex
	cachedUpdate *UpdateInfo
)

// UpdateInfo describes the result of checking GitHub Releases for a newer version.
type UpdateInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	Available      bool   `json:"available"`
	IsDevBuild     bool   `json:"is_dev_build"`
	ReleaseURL     string `json:"release_url"`
	ReleaseNotes   string `json:"release_notes"`
	AssetName      string `json:"asset_name"`
	DownloadURL    string `json:"download_url"`
	ChecksumURL    string `json:"checksum_url"`
}

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	HTMLURL    string        `json:"html_url"`
	Body       string        `json:"body"`
	Draft      bool          `json:"draft"`
	Prerelease bool          `json:"prerelease"`
	Assets     []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetVersion returns the running version string.
func GetVersion() string {
	return version
}

// CleanupOldBinary removes the ".old" backup executable left behind after a self-update.
func CleanupOldBinary() {
	exePath, err := currentExePath()
	if err != nil {
		return
	}
	oldPath := exePath + ".old"
	if _, err := os.Stat(oldPath); err == nil {
		_ = os.Remove(oldPath)
	}
}

func currentExePath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return p, nil
	}
	return resolved, nil
}

// CheckForUpdates queries the GitHub Releases API for the latest release.
func CheckForUpdates() (*UpdateInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", strings.TrimRight(githubAPIBase, "/"), githubOwner, githubRepo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "RectangleWin/"+version)

	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &UpdateInfo{
			CurrentVersion: version,
			IsDevBuild:     isDevVersion(version),
			Available:      false,
		}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to parse release response: %w", err)
	}

	info := &UpdateInfo{
		CurrentVersion: version,
		LatestVersion:  rel.TagName,
		IsDevBuild:     isDevVersion(version),
		ReleaseURL:     rel.HTMLURL,
		ReleaseNotes:   rel.Body,
	}

	if !rel.Draft && !rel.Prerelease && isNewerVersion(version, rel.TagName) {
		info.Available = true
	}

	asset, checksum := selectReleaseAsset(rel.Assets, runtime.GOARCH)
	if asset != nil {
		info.AssetName = asset.Name
		info.DownloadURL = asset.BrowserDownloadURL
	}
	if checksum != nil {
		info.ChecksumURL = checksum.BrowserDownloadURL
	}

	updateMu.Lock()
	cachedUpdate = info
	updateMu.Unlock()

	return info, nil
}

// GetCachedUpdate returns the most recent update check result, if any.
func GetCachedUpdate() *UpdateInfo {
	updateMu.Lock()
	defer updateMu.Unlock()
	if cachedUpdate == nil {
		return nil
	}
	cp := *cachedUpdate
	return &cp
}

// CheckForUpdatesOnStartup checks for updates in the background and prompts the user if one is found.
func CheckForUpdatesOnStartup() {
	if isDevVersion(version) {
		return
	}

	go func() {
		// Brief delay so startup feels instant and network is settled after login.
		time.Sleep(3 * time.Second)

		activeConfigLock.Lock()
		enabled := activeConfig != nil && activeConfig.Settings.AutoCheckUpdates
		activeConfigLock.Unlock()
		if !enabled {
			return
		}

		info, err := CheckForUpdates()
		if err != nil {
			fmt.Printf("warn: startup update check failed: %v\n", err)
			return
		}
		if info.Available {
			promptAndApplyUpdate(info)
		}
	}()
}

// CheckForUpdatesInteractive performs a user-initiated update check and displays the result.
func CheckForUpdatesInteractive() {
	go func() {
		info, err := CheckForUpdates()
		if err != nil {
			showMessageBox("Failed to check for updates:\n\n" + err.Error())
			return
		}

		if info.IsDevBuild {
			msg := "You are running a development build of RectangleWin."
			if info.LatestVersion != "" {
				msg += fmt.Sprintf("\n\nLatest GitHub release: %s", info.LatestVersion)
			}
			showMessageBox(msg)
			return
		}

		if !info.Available {
			showMessageBox(fmt.Sprintf("You're up to date!\n\nRectangleWin %s is the latest version.", formatDisplayVersion(version)))
			return
		}

		promptAndApplyUpdate(info)
	}()
}

func promptAndApplyUpdate(info *UpdateInfo) {
	if info.DownloadURL == "" {
		msg := fmt.Sprintf(
			"A new version of RectangleWin (%s) is available (current: %s).\n\nOpen the GitHub release page to download it?",
			formatDisplayVersion(info.LatestVersion),
			formatDisplayVersion(info.CurrentVersion),
		)
		if w32.MessageBox(0, msg, "RectangleWin Update Available", w32.MB_YESNO|w32.MB_ICONINFORMATION) == w32.IDYES && info.ReleaseURL != "" {
			_ = w32.ShellExecute(0, "open", info.ReleaseURL, "", "", w32.SW_SHOWNORMAL)
		}
		return
	}

	msg := fmt.Sprintf(
		"A new version of RectangleWin (%s) is available (current: %s).\n\nWould you like to download and install it now? RectangleWin will restart automatically.",
		formatDisplayVersion(info.LatestVersion),
		formatDisplayVersion(info.CurrentVersion),
	)
	if w32.MessageBox(0, msg, "RectangleWin Update Available", w32.MB_YESNO|w32.MB_ICONINFORMATION) != w32.IDYES {
		return
	}

	if err := ApplyUpdate(info); err != nil {
		showMessageBox("Failed to install update:\n\n" + err.Error())
	}
}

// ApplyUpdate downloads, verifies, installs the new binary, and restarts RectangleWin.
func ApplyUpdate(info *UpdateInfo) error {
	if info == nil || info.DownloadURL == "" {
		return errors.New("no downloadable release asset found for this architecture")
	}

	if !updateMu.TryLock() {
		return errors.New("an update is already in progress")
	}
	defer updateMu.Unlock()

	exePath, err := currentExePath()
	if err != nil {
		return fmt.Errorf("failed to locate current executable: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Download asset into the same directory so rename is atomic on the same volume.
	dlFile, err := os.CreateTemp(exeDir, "RectangleWin-update-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file (check folder permissions): %w", err)
	}
	dlPath := dlFile.Name()
	dlFile.Close()
	defer os.Remove(dlPath)

	if err := downloadFile(info.DownloadURL, dlPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Verify SHA-256 checksum if the release includes checksums.txt.
	if info.ChecksumURL != "" && info.AssetName != "" {
		checksumBody, err := fetchText(info.ChecksumURL)
		if err != nil {
			return fmt.Errorf("failed to download checksums: %w", err)
		}
		expectedHex, ok := parseChecksums(checksumBody, info.AssetName)
		if !ok {
			return fmt.Errorf("checksum entry not found for %s", info.AssetName)
		}
		if err := verifySHA256(dlPath, expectedHex); err != nil {
			return err
		}
	}

	// Extract or stage the new executable.
	newExePath := exePath + ".new"
	_ = os.Remove(newExePath)
	defer os.Remove(newExePath)

	lowerAsset := strings.ToLower(info.AssetName)
	switch {
	case strings.HasSuffix(lowerAsset, ".zip"):
		if err := extractExeFromZip(dlPath, newExePath); err != nil {
			return fmt.Errorf("failed to extract update zip: %w", err)
		}
	case strings.HasSuffix(lowerAsset, ".tar.gz") || strings.HasSuffix(lowerAsset, ".tgz"):
		if err := extractExeFromTarGz(dlPath, newExePath); err != nil {
			return fmt.Errorf("failed to extract update archive: %w", err)
		}
	case strings.HasSuffix(lowerAsset, ".exe"):
		if err := os.Rename(dlPath, newExePath); err != nil {
			return fmt.Errorf("failed to stage executable: %w", err)
		}
	default:
		return fmt.Errorf("unsupported release asset format: %s", info.AssetName)
	}

	// Swap binaries: Windows allows renaming a running .exe even though it cannot be overwritten.
	oldExePath := exePath + ".old"
	_ = os.Remove(oldExePath)
	if err := os.Rename(exePath, oldExePath); err != nil {
		return fmt.Errorf("failed to back up current executable: %w", err)
	}
	if err := os.Rename(newExePath, exePath); err != nil {
		// Roll back so the user is never left without an executable.
		_ = os.Rename(oldExePath, exePath)
		return fmt.Errorf("failed to install new executable: %w", err)
	}

	// Release single-instance mutex before spawning the new process so it starts cleanly.
	releaseSingleInstanceMutex()

	cmd := exec.Command(exePath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("update installed, but failed to restart automatically: %w", err)
	}

	systray.Quit()
	os.Exit(0)
	return nil
}

func downloadFile(url, destPath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "RectangleWin/"+version)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func fetchText(url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "RectangleWin/"+version)

	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func extractExeFromZip(zipPath, destPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(f.Name), ".exe") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, rc)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		}
	}
	return errors.New("no .exe file found inside zip archive")
}

func extractExeFromTarGz(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeReg && strings.HasSuffix(strings.ToLower(hdr.Name), ".exe") {
			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			return closeErr
		}
	}
	return errors.New("no .exe file found inside tar.gz archive")
}

func parseChecksums(content, assetName string) (string, bool) {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		name := strings.TrimPrefix(fields[1], "*")
		if name == assetName {
			return strings.ToLower(fields[0]), true
		}
	}
	return "", false
}

func verifySHA256(filePath, expectedHex string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, expectedHex) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHex, actual)
	}
	return nil
}

func archTokens(goarch string) []string {
	switch goarch {
	case "amd64":
		return []string{"x64", "amd64", "x86_64"}
	case "386":
		return []string{"x86", "386", "i386"}
	case "arm64":
		return []string{"arm64"}
	default:
		return []string{goarch}
	}
}

func selectReleaseAsset(assets []githubAsset, goarch string) (asset *githubAsset, checksum *githubAsset) {
	tokens := archTokens(goarch)

	var best *githubAsset
	bestScore := -1

	for i := range assets {
		a := &assets[i]
		lower := strings.ToLower(a.Name)

		if lower == "checksums.txt" || strings.HasSuffix(lower, "checksums.txt") {
			checksum = a
			continue
		}

		isZip := strings.HasSuffix(lower, ".zip")
		isTarGz := strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz")
		isExe := strings.HasSuffix(lower, ".exe")
		if !isZip && !isTarGz && !isExe {
			continue
		}

		matchedArch := false
		for _, tok := range tokens {
			if strings.Contains(lower, tok) {
				matchedArch = true
				break
			}
		}
		if !matchedArch {
			continue
		}

		score := 0
		if strings.Contains(lower, "windows") {
			score += 10
		}
		if isZip {
			score += 3
		} else if isExe {
			score += 2
		} else if isTarGz {
			score += 1
		}

		if score > bestScore {
			bestScore = score
			best = a
		}
	}

	return best, checksum
}

func isDevVersion(v string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(v))
	return trimmed == "" || trimmed == "dev" || strings.HasPrefix(trimmed, "dev-")
}

func formatDisplayVersion(v string) string {
	trimmed := strings.TrimSpace(v)
	if trimmed == "" || isDevVersion(trimmed) {
		return "dev"
	}
	if !strings.HasPrefix(trimmed, "v") && !strings.HasPrefix(trimmed, "V") {
		return "v" + trimmed
	}
	return trimmed
}

type semver struct {
	major, minor, patch int
	prerelease          string
}

func parseSemver(v string) (semver, bool) {
	s := strings.TrimSpace(v)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if idx := strings.IndexByte(s, '+'); idx >= 0 {
		s = s[:idx]
	}

	var pre string
	if idx := strings.IndexByte(s, '-'); idx >= 0 {
		pre = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.Split(s, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return semver{}, false
	}

	maj, err := strconv.Atoi(parts[0])
	if err != nil || maj < 0 {
		return semver{}, false
	}
	min, err := strconv.Atoi(parts[1])
	if err != nil || min < 0 {
		return semver{}, false
	}
	patch := 0
	if len(parts) == 3 {
		p, err := strconv.Atoi(parts[2])
		if err != nil || p < 0 {
			return semver{}, false
		}
		patch = p
	}

	return semver{major: maj, minor: min, patch: patch, prerelease: pre}, true
}

// isNewerVersion reports whether latest is strictly newer than current using semver rules.
// Development builds ("dev") always return false.
func isNewerVersion(current, latest string) bool {
	if isDevVersion(current) {
		return false
	}
	cur, ok1 := parseSemver(current)
	lat, ok2 := parseSemver(latest)
	if !ok1 || !ok2 {
		return false
	}

	if lat.major != cur.major {
		return lat.major > cur.major
	}
	if lat.minor != cur.minor {
		return lat.minor > cur.minor
	}
	if lat.patch != cur.patch {
		return lat.patch > cur.patch
	}
	// A stable release (no prerelease) is newer than a prerelease of the same version.
	if cur.prerelease != "" && lat.prerelease == "" {
		return true
	}
	return false
}
