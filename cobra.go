package xcli

import (
	"os"
	"os/exec"
)

// Execute runs the root command
func (x *XCli) Execute() error {
	printBanner()
	// If we are not already running inside the local command proxy, and a local main.go exists, proxy to it.
	if os.Getenv("XCLI_RUNNING_LOCAL") != "true" {
		localPaths := []string{"cmd/cli/main.go", "cmd/xcli/main.go"}
		for _, path := range localPaths {
			if _, err := os.Stat(path); err == nil {
				args := os.Args[1:]
				cmd := exec.Command("go", append([]string{"run", path}, args...)...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Stdin = os.Stdin
				cmd.Env = append(os.Environ(), "XCLI_RUNNING_LOCAL=true")
				if err := cmd.Run(); err != nil {
					if exitError, ok := err.(*exec.ExitError); ok {
						os.Exit(exitError.ExitCode())
					}
					os.Exit(1)
				}
				return nil
			}
		}
	}
	return x.root.Execute()
}
