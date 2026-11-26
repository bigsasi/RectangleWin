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
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"1.0.0", "1.0.1", true},
		{"v1.0.0", "v1.1.0", true},
		{"1.2.0", "2.0.0", true},
		{"1.2.0", "1.2.0", false},
		{"1.2.1", "1.2.0", false},
		{"1.2.0-rc1", "1.2.0", true},
		{"1.2.0", "1.2.0-rc1", false},
		{"dev", "1.2.0", false},
		{"", "1.2.0", false},
		{"1.2.0", "invalid", false},
	}

	for _, tt := range tests {
		if got := isNewerVersion(tt.current, tt.latest); got != tt.want {
			t.Errorf("isNewerVersion(%q, %q) = %v, want %v", tt.current, tt.latest, got, tt.want)
		}
	}
}

func TestSelectReleaseAsset(t *testing.T) {
	assets := []githubAsset{
		{Name: "RectangleWin_1.3.0_windows_amd64.zip", BrowserDownloadURL: "https://example.com/amd64.zip"},
		{Name: "RectangleWin_1.3.0_windows_386.zip", BrowserDownloadURL: "https://example.com/386.zip"},
		{Name: "RectangleWin_1.3.0_windows_arm64.zip", BrowserDownloadURL: "https://example.com/arm64.zip"},
		{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
	}

	asset, checksum := selectReleaseAsset(assets, "amd64")
	if asset == nil || asset.Name != "RectangleWin_1.3.0_windows_amd64.zip" {
		t.Fatalf("amd64 asset = %+v, want amd64 zip", asset)
	}
	if checksum == nil || checksum.Name != "checksums.txt" {
		t.Fatalf("checksum asset = %+v, want checksums.txt", checksum)
	}

	asset386, _ := selectReleaseAsset(assets, "386")
	if asset386 == nil || asset386.Name != "RectangleWin_1.3.0_windows_386.zip" {
		t.Fatalf("386 asset = %+v, want 386 zip", asset386)
	}

	assetArm, _ := selectReleaseAsset(assets, "arm64")
	if assetArm == nil || assetArm.Name != "RectangleWin_1.3.0_windows_arm64.zip" {
		t.Fatalf("arm64 asset = %+v, want arm64 zip", assetArm)
	}
}

func TestParseChecksumsAndVerifySHA256(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "RectangleWin.zip")
	payload := []byte("rectanglewin-release-payload")
	if err := os.WriteFile(filePath, payload, 0644); err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(payload)
	hexSum := hex.EncodeToString(sum[:])

	checksums := hexSum + "  RectangleWin.zip\n0000000000000000  other.zip\n"
	got, ok := parseChecksums(checksums, "RectangleWin.zip")
	if !ok || got != hexSum {
		t.Fatalf("parseChecksums got (%q, %v), want (%q, true)", got, ok, hexSum)
	}

	if err := verifySHA256(filePath, got); err != nil {
		t.Fatalf("verifySHA256 failed on valid file: %v", err)
	}

	if err := verifySHA256(filePath, "deadbeef"); err == nil {
		t.Fatalf("verifySHA256 expected mismatch error, got nil")
	}
}

func TestExtractExeFromZip(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "release.zip")
	outExe := filepath.Join(dir, "extracted.exe")

	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	w, err := zw.Create("RectangleWin-x64-v1.3.0.exe")
	if err != nil {
		t.Fatal(err)
	}
	wantBytes := []byte("MZ-fake-pe-binary")
	if _, err := w.Write(wantBytes); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zf.Close(); err != nil {
		t.Fatal(err)
	}

	if err := extractExeFromZip(zipPath, outExe); err != nil {
		t.Fatalf("extractExeFromZip failed: %v", err)
	}
	got, err := os.ReadFile(outExe)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(wantBytes) {
		t.Fatalf("extracted bytes = %q, want %q", got, wantBytes)
	}
}

func TestCheckForUpdatesMockAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := githubRelease{
			TagName: "v2.0.0",
			HTMLURL: "https://github.com/bigsasi/RectangleWin/releases/tag/v2.0.0",
			Body:    "New features",
			Assets: []githubAsset{
				{Name: "RectangleWin_2.0.0_windows_amd64.zip", BrowserDownloadURL: "https://example.com/amd64.zip"},
				{Name: "RectangleWin_2.0.0_windows_386.zip", BrowserDownloadURL: "https://example.com/386.zip"},
				{Name: "RectangleWin_2.0.0_windows_arm64.zip", BrowserDownloadURL: "https://example.com/arm64.zip"},
				{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
			},
		}
		_ = json.NewEncoder(w).Encode(rel)
	}))
	defer srv.Close()

	origBase := githubAPIBase
	origVersion := version
	defer func() {
		githubAPIBase = origBase
		version = origVersion
	}()

	githubAPIBase = srv.URL
	version = "1.2.0"

	info, err := CheckForUpdates()
	if err != nil {
		t.Fatalf("CheckForUpdates failed: %v", err)
	}
	if !info.Available {
		t.Fatalf("expected Available=true, got false: %+v", info)
	}
	if info.LatestVersion != "v2.0.0" {
		t.Errorf("LatestVersion = %q, want v2.0.0", info.LatestVersion)
	}
	if info.DownloadURL == "" || info.ChecksumURL == "" {
		t.Errorf("expected DownloadURL and ChecksumURL to be populated: %+v", info)
	}
}
