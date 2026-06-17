package xcli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// releaseManifest is the JSON written into every release zip as version.json.
type releaseManifest struct {
	Version     string    `json:"version"`
	Build       string    `json:"build"`
	OS          string    `json:"os"`
	ReleasedAt  time.Time `json:"released_at"`
	Description string    `json:"description"`
}

// buildReleaseCommand packages a versioned, deployable release zip for Linux and/or Windows.
//
// The zip contains:
//
//	app / app.exe      — compiled Go binary
//	public/            — built frontend assets
//	xfig/              — FreeRADIUS config templates
//	application.json   — updated with the new version
//	version.json       — release manifest
//
// Usage:
//
//	./xcli build:release --version 1.2.0
//	./xcli build:release --version 1.2.0 --os linux
func (x *XCli) buildReleaseCommand() *cobra.Command {
	var (
		targetOS     string
		versionFlag  string
		skipFrontend bool
		obfuscate    bool
	)

	cmd := &cobra.Command{
		Use:   "build:release",
		Short: "Build versioned release zip(s) for Linux and/or Windows into builds/releases/",
		Long: `Builds a production-ready release package following the same steps as build:app,
then bundles everything into a zip suitable for the auto-update system at local.connect4ar.com.

Steps:
  1. (optional) npm install && npm run build — Vite frontend
  2. CGO_ENABLED=0 go build              — Go binary per target OS
  3. Copy public/ and xfig/
  4. Bump version in application.json
  5. Write version.json manifest
  6. Zip everything into builds/releases/connect_local-<version>-<os>.zip

Examples:
  ./xcli build:release --version 1.2.0
  ./xcli build:release --version 1.2.0 --os linux --skip-frontend`,
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag == "" {
				fmt.Println("❌ --version is required. Example: --version 1.2.0")
				return
			}

			validOS := map[string]bool{"linux": true, "windows": true, "darwin": true}
			var targets []string
			if targetOS == "" {
				targets = []string{"linux", "windows"}
			} else {
				if !validOS[targetOS] {
					fmt.Printf("❌ Invalid --os %q. Must be one of: linux, windows, darwin\n", targetOS)
					return
				}
				targets = []string{targetOS}
			}

			releaseDir := filepath.Join("builds", "releases")
			if err := os.MkdirAll(releaseDir, 0755); err != nil {
				fmt.Println("❌ Failed to create builds/releases:", err)
				return
			}

			// ── Step 1 & 2: Frontend ─────────────────────────────────────────
			if !skipFrontend {
				fmt.Println("\n[1/5] Installing npm dependencies…")
				npmInstall := exec.Command("npm", "install")
				npmInstall.Stdout, npmInstall.Stderr = os.Stdout, os.Stderr
				if err := npmInstall.Run(); err != nil {
					fmt.Println("❌ npm install failed:", err)
					return
				}

				fmt.Println("\n[2/5] Building frontend assets…")
				npmBuild := exec.Command("npm", "run", "build")
				npmBuild.Stdout, npmBuild.Stderr = os.Stdout, os.Stderr
				if err := npmBuild.Run(); err != nil {
					fmt.Println("❌ Frontend build failed:", err)
					return
				}
			} else {
				fmt.Println("\n[1-2/5] Skipping frontend build (--skip-frontend)")
			}

			// ── Step 3: Bump application.json ────────────────────────────────
			fmt.Printf("\n[3/5] Updating application.json to version %s…\n", versionFlag)
			if err := bumpAppVersion(versionFlag); err != nil {
				fmt.Println("❌ Failed to update application.json:", err)
				return
			}

			// ── Step 4: Build binaries + create zips ─────────────────────────
			fmt.Println("\n[4/5] Building Go binaries and creating release zips…")
			for _, goos := range targets {
				if err := buildReleaseTarget(goos, versionFlag, releaseDir, obfuscate); err != nil {
					fmt.Printf("❌ Release build failed for %s: %v\n", goos, err)
					return
				}
			}

			// ── Done ─────────────────────────────────────────────────────────
			fmt.Printf("\n✅ Release v%s complete!\n", versionFlag)
			fmt.Println("Output:")
			for _, goos := range targets {
				fmt.Printf("  builds/releases/connect_local-%s-%s.zip\n", versionFlag, goos)
			}
			fmt.Println("\nUpload the zip(s) to local.connect4ar.com/releases/ and update the version manifest.")
		},
	}

	cmd.Flags().StringVar(&versionFlag, "version", "", "Release version string, e.g. 1.2.0 (required)")
	cmd.Flags().StringVar(&targetOS, "os", "", "Target OS: linux, windows, darwin (default: linux + windows)")
	cmd.Flags().BoolVar(&skipFrontend, "skip-frontend", false, "Skip npm install and npm run build")
	cmd.Flags().BoolVar(&obfuscate, "obfuscate", false, "Obfuscate the binary")

	return cmd
}

// buildReleaseTarget compiles the binary for goos, assembles a staging directory, then zips it.
func buildReleaseTarget(goos, version, releaseDir string, obfuscate bool) error {
	stagingDir := filepath.Join("builds", "releases", fmt.Sprintf(".stage-%s", goos))
	defer os.RemoveAll(stagingDir) // always clean up

	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return err
	}

	// Binary name
	exeName := "app"
	if goos == "windows" {
		exeName = "app.exe"
	}
	exePath := filepath.Join(stagingDir, exeName)

	fmt.Printf("  Compiling %s/amd64 → %s\n", goos, exePath)

	var goBuild *exec.Cmd
	// Check if obfuscate flag is set
	if obfuscate {
		goBuild = exec.Command(
			"garble",
			"-literals",
			"-tiny",        // removes extra debug info
			"-seed=random", // different seed each build
			"build",
			"-trimpath",
			"-ldflags", "-s -w",
			"-o", exePath,
			"cmd/server/main.go",
		)
	} else {
		goBuild = exec.Command(
			"go",
			"build",
			"-trimpath",
			"-ldflags", "-s -w",
			"-o", exePath,
			"cmd/server/main.go",
		)
	}

	goBuild.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+goos,
		"GOARCH=amd64",
	)
	goBuild.Stdout, goBuild.Stderr = os.Stdout, os.Stderr
	if err := goBuild.Run(); err != nil {
		return fmt.Errorf("go build: %w", err)
	}

	// Copy public/ and xfig into staging
	for _, pair := range []struct{ src, dst string }{
		{"public", filepath.Join(stagingDir, "public")},
		{filepath.Join("xfig"), filepath.Join(stagingDir, "xfig")},
	} {
		if _, err := os.Stat(pair.src); err == nil {
			if err := copyDir(pair.src, pair.dst); err != nil {
				return fmt.Errorf("copy %s: %w", pair.src, err)
			}
		}
	}

	// Remove Vite's hot file if present
	_ = os.Remove(filepath.Join(stagingDir, "public", "hot"))

	// Copy application.json (already bumped)
	if err := copyFile("application.json", filepath.Join(stagingDir, "application.json")); err != nil {
		return fmt.Errorf("copy application.json: %w", err)
	}

	// Write version.json manifest
	manifest := releaseManifest{
		Version:     version,
		Build:       time.Now().Format("2006.01.02"),
		OS:          goos,
		ReleasedAt:  time.Now().UTC(),
		Description: fmt.Sprintf("Connect Local v%s release for %s", version, goos),
	}
	manifestBytes, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(stagingDir, "version.json"), manifestBytes, 0644); err != nil {
		return fmt.Errorf("write version.json: %w", err)
	}

	// Zip the staging directory
	zipName := fmt.Sprintf("connect_local-%s-%s.zip", version, goos)
	zipPath := filepath.Join(releaseDir, zipName)
	if err := zipDirectory(stagingDir, zipPath); err != nil {
		return fmt.Errorf("zip: %w", err)
	}

	fmt.Printf("  ✓ %s\n", zipPath)
	return nil
}

// bumpAppVersion reads application.json, updates "version", and writes it back.
func bumpAppVersion(version string) error {
	data, err := os.ReadFile("application.json")
	if err != nil {
		return err
	}
	var info map[string]any
	if err := json.Unmarshal(data, &info); err != nil {
		return err
	}
	info["version"] = version
	info["build"] = time.Now().Format("2006.01.02")

	out, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile("application.json", out, 0644)
}

// zipDirectory creates a zip at dest containing all files under srcDir (relative paths inside zip).
func zipDirectory(srcDir, dest string) error {
	zf, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zf.Close()

	w := zip.NewWriter(zf)
	defer w.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		// Normalize separators to forward-slash for cross-platform zips
		zipEntry := filepath.ToSlash(rel)

		if info.IsDir() {
			if zipEntry == "." {
				return nil
			}
			_, err = w.Create(zipEntry + "/")
			return err
		}

		f, err := w.Create(zipEntry)
		if err != nil {
			return err
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(f, src)
		return err
	})
}
