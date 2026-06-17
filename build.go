package xcli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func (x *XCli) buildCliCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build:cli",
		Short: "Build CLI for current OS (root xcli or xcli.exe) and all platforms into builds/xcli/",
		Run: func(cmd *cobra.Command, args []string) {
			currentOS := runtime.GOOS
			currentArch := runtime.GOARCH

			targets := []struct {
				GOOS   string
				GOARCH string
				Output string
			}{
				{"windows", "amd64", "xcli-windows.exe"},
				{"linux", "amd64", "xcli-linux"},
				{"darwin", "amd64", "xcli-mac"},
			}

			// Ensure builds/ folder exists
			buildDir := "builds/xcli"
			if err := os.MkdirAll(buildDir, 0755); err != nil {
				fmt.Printf("Failed to create builds folder: %v\n", err)
				return
			}

			// Build native binary at root ./xcli or ./xcli.exe
			nativeName := "xcli"
			if currentOS == "windows" {
				nativeName += ".exe"
			}
			fmt.Printf("Building native binary: %s\n", nativeName)
			if err := runBuild(nil, nativeName); err != nil {
				fmt.Printf("Build failed: %v\n", err)
				return
			}

			// Build all targets into builds/xcli/ and copy linux+windows to project root
			rootCopy := map[string]string{
				"linux":   "xcli",
				"windows": "xcli.exe",
			}

			for _, target := range targets {
				outputPath := filepath.Join(buildDir, target.Output)

				if target.GOOS == currentOS && target.GOARCH == currentArch {
					fmt.Printf("Copying native binary to: %s\n", outputPath)
					if err := copyFile(nativeName, outputPath); err != nil {
						fmt.Printf("Copy failed: %v\n", err)
					}
				} else {
					// Cross-compile into builds/xcli/
					fmt.Printf("Cross-compiling for %s/%s -> %s\n", target.GOOS, target.GOARCH, outputPath)
					env := []string{
						"GOOS=" + target.GOOS,
						"GOARCH=" + target.GOARCH,
					}
					if err := runBuild(env, outputPath); err != nil {
						fmt.Printf("Cross-compile failed: %v\n", err)
						continue
					}
				}

				// Copy linux and windows binaries to project root
				if rootName, ok := rootCopy[target.GOOS]; ok {
					fmt.Printf("Copying %s -> ./%s\n", outputPath, rootName)
					if err := copyFile(outputPath, rootName); err != nil {
						fmt.Printf("Root copy failed: %v\n", err)
					}
				}
			}
		},
	}
}

// buildAppCommand builds the full application following the same steps as the Dockerfile:
//  1. npm install + npm run build  (frontend)
//  2. CGO_ENABLED=0 go build       (Go server, for linux → app and windows → app.exe)
//  3. Copy xfig/freeradius, public, storage
//  4. Remove public/hot
//  5. Copy .env.prod -> .env
//
// By default (no --os flag) it builds for BOTH linux and windows into builds/app/.
// Pass --os linux|windows|darwin to build only that target.
func (x *XCli) buildAppCommand() *cobra.Command {
	var targetOS string
	var obfuscate bool
	var skipFrontend bool

	cmd := &cobra.Command{
		Use:   "build:app",
		Short: "Build full app for Linux (app) and Windows (app.exe) into builds/app/",
		Long: `Builds the full application following the same steps as the Dockerfile build stage:
  1. npm install && npm run build   (Vite frontend — run once)
  2. CGO_ENABLED=0 go build         (Go server binary, cross-compiled)
  3. Copy xfig/freeradius, public, storage
  4. Remove public/hot
  5. Copy .env.prod -> .env

By default builds for both Linux and Windows into builds/app/:
  builds/app/app       (Linux binary)
  builds/app/app.exe   (Windows binary)

Use --os linux|windows|darwin to build a single target only.
Use --obfuscate to obfuscate the binary.
Use --skip-frontend to skip frontend build.`,
		Run: func(cmd *cobra.Command, args []string) {
			validOS := map[string]bool{"linux": true, "windows": true, "darwin": true}

			// Determine which targets to build
			var targets []string
			if targetOS == "" {
				// Default: build both linux and windows
				targets = []string{"linux", "windows"}
			} else {
				if !validOS[targetOS] {
					fmt.Printf("Invalid --os value %q. Must be one of: linux, windows, darwin\n", targetOS)
					return
				}
				targets = []string{targetOS}
			}

			outDir := filepath.Join("builds", "app")
			if err := os.MkdirAll(outDir, 0755); err != nil {
				fmt.Println("Failed to create output folder:", err)
				return
			}

			if !skipFrontend {
				// ── Step 1: npm install ────────────────────────────────────────────
				fmt.Println("\n[1/5] Installing npm dependencies (npm install)...")
				npmInstall := exec.Command("npm", "install")
				npmInstall.Stdout = os.Stdout
				npmInstall.Stderr = os.Stderr
				if err := npmInstall.Run(); err != nil {
					fmt.Println("npm install failed:", err)
					return
				}

				// ── Step 2: npm run build (Vite) ──────────────────────────────────
				fmt.Println("\n[2/5] Building frontend assets (npm run build)...")
				npmBuild := exec.Command("npm", "run", "build")
				npmBuild.Stdout = os.Stdout
				npmBuild.Stderr = os.Stderr
				if err := npmBuild.Run(); err != nil {
					fmt.Println("Frontend build failed:", err)
					return
				}

			}

			// ── Step 3: go build for each target ──────────────────────────────
			fmt.Printf("\n[3/5] Building Go server binaries...\n")
			for _, goos := range targets {
				exeName := "app"
				if goos == "windows" {
					exeName = "app.exe"
				}
				exePath := filepath.Join(outDir, exeName)
				fmt.Printf("  Cross-compiling %s/amd64 -> %s\n", goos, exePath)

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
				goBuild.Stdout = os.Stdout
				goBuild.Stderr = os.Stderr
				if err := goBuild.Run(); err != nil {
					fmt.Printf("  Go build failed for %s: %v\n", goos, err)
					return
				}
				fmt.Printf("  ✓ %s\n", exePath)
			}

			// ── Step 4: Copy static assets ────────────────────────────────────
			fmt.Println("\n[4/5] Copying static assets...")
			assetsToCopy := []string{
				"xfig/freeradius",
				"public",
			}
			for _, asset := range assetsToCopy {
				dst := filepath.Join(outDir, asset)
				fmt.Printf("  Copying %s -> %s\n", asset, dst)
				if err := copyDir(asset, dst); err != nil {
					fmt.Printf("  Failed to copy %s: %v\n", asset, err)
					return
				}
			}

			// Remove public/hot (Vite hot-reload file — not needed in production)
			hotFile := filepath.Join(outDir, "public", "hot")
			_ = os.Remove(hotFile)
			fmt.Println("  Removed public/hot (if present)")

			// ── Step 5: Copy .env.prod -> .env ────────────────────────────────
			fmt.Println("\n[5/5] Setting up .env file...")
			envDst := filepath.Join(outDir, ".env")
			if err := copyFile(".env.prod", envDst); err != nil {
				fmt.Println("  .env.prod not found, writing minimal .env...")
				minimal := "APP_ENV=production\nAPP_PORT=8080\nAPP_URL=localhost\nAPP_BIND_ADDRESS=0.0.0.0\nAPP_BIND_PORT=8080\n\nREDIS_HOST=\"127.0.0.1\"\nREDIS_PORT=6379\nREDIS_PASSWORD=\nREDIS_DB=10\n"
				_ = os.WriteFile(envDst, []byte(minimal), 0644)
			} else {
				fmt.Printf("  Copied .env.prod -> %s\n", envDst)
			}

			// ── Done ──────────────────────────────────────────────────────────
			fmt.Printf("\nBuild complete! Output folder: %s\n", outDir)
			fmt.Println("Contents:")
			for _, goos := range targets {
				exeName := "app"
				if goos == "windows" {
					exeName = "app.exe"
				}
				fmt.Printf("  %s/%s\n", outDir, exeName)
			}
		},
	}

	cmd.Flags().StringVar(&targetOS, "os", "", "Target OS: linux, windows, darwin (default: builds both linux and windows)")
	cmd.Flags().BoolVar(&obfuscate, "obfuscate", false, "Obfuscate the binary")
	cmd.Flags().BoolVar(&skipFrontend, "skip-frontend", false, "Skip frontend build (npm install + npm run build)")
	return cmd
}

// Unified build method for xcli binaries
func runBuild(env []string, output string) error {
	cmd := exec.Command("go", "build", "-o", output, "cmd/xcli/main.go")
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	} else {
		cmd.Env = os.Environ()
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0755)
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		sPath := filepath.Join(src, e.Name())
		dPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(sPath, dPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(sPath, dPath); err != nil {
				return err
			}
		}
	}
	return nil
}
