package xcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
)

const (
	version = "1.0.1"
)

type XCli struct {
	preRegisteredCmds map[string][]*cobra.Command
	externalCmds      []*cobra.Command
	root              *cobra.Command
}

func loadEnv() {
	if _, err := os.Stat(".env.xcli"); err == nil {
		_ = godotenv.Overload(".env.xcli")
	} else if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load(".env")
	}
}

func New() *XCli {
	loadEnv()

	x := &XCli{
		preRegisteredCmds: make(map[string][]*cobra.Command),
	}

	// Cobra root command settings
	root := &cobra.Command{
		Use:   "",
		Short: "Command runner",
		Long:  "A simple CLI tool for running application commands",
	}

	root.CompletionOptions.DisableDefaultCmd = true
	x.root = root

	// tool pre registered commands
	x.preRegisteredCmds = map[string][]*cobra.Command{
		"App": {
			x.devCommand(),
			x.helpCommand(),
			x.lintCommand(),
			x.scanCommand(),
			x.versionCommand(),
		},
		"build": {
			x.buildCliCommand(),
			x.buildAppCommand(),
			x.buildReleaseCommand(),
		},
		"migrate": {
			x.migrateCommand(),
			x.migrateRollbackCommand(),
			x.makeMigrationCommand(),
			x.migrateStatusCommand(),
			x.migrateResetCommand(),
			x.migrateRefreshCommand(),
			x.migrateUpByCommand(),
			x.migrateDownByCommand(),
			x.goosePassthroughCommand(),
		},
		"make": {
			x.makeVueComponentCommand(),
			x.makeModelCommand(),
			x.makeLangFilesCommand(),
			x.makeHandlerCommand(),
			x.makeActionCommand(),
			x.makeRepoCommand(),
			x.makeRequestCommand(),
			x.makeNotificationCommand(),
			x.makeTaskCommand(),
			x.makeCrudCommand(),
		},
		"stubs": {
			x.stubPublishCommand(),
		},
	}

	for _, cmds := range x.preRegisteredCmds {
		for _, cmd := range cmds {
			x.root.AddCommand(cmd)
		}
	}

	return x
}

func (x *XCli) Register(cmd ...*cobra.Command) {
	if len(cmd) > 0 {
		x.externalCmds = append(x.externalCmds, cmd...)
		for _, c := range cmd {
			x.root.AddCommand(c)
		}
	}
}

func (x *XCli) runInCli(ctx context.Context, cmd string, args []string) error {
	if _, err := exec.LookPath(cmd); err != nil {
		fmt.Printf("❌ Required CLI '%s' not found.\n", cmd)

		if !suggestInstall(cmd) {
			return err
		}

		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("tool '%s' still not found in PATH", cmd)
		}
	}

	cli := exec.CommandContext(ctx, cmd, args...)
	cli.Stdout = os.Stdout
	cli.Stderr = os.Stderr
	cli.Stdin = os.Stdin

	return cli.Run()
}

func suggestInstall(cmd string) bool {
	fmt.Printf(Yellow+"❓ Install %s now? [y/N]: "+Reset, cmd)

	var input string
	if _, err := fmt.Scanln(&input); err != nil || strings.ToLower(input) != "y" {
		fmt.Println(Red + "❌ Aborted." + Reset)
		return false
	}

	spec, ok := suggestInstallCmds[cmd]
	if !ok {
		fmt.Println(Red + "⚠️ Unknown tool: " + cmd + Reset)
		return false
	}

	if spec.Command == "" {
		fmt.Println(Red + "❌ Cannot install automatically: " + cmd)
		fmt.Println(spec.FailMessage)
		return false
	}

	fmt.Println(Green + "📦 Installing " + cmd + "..." + Reset)
	c := exec.Command(spec.Command, spec.Args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fmt.Println(Red+"⚠️ Failed to install "+cmd+":", err.Error(), Reset)
		return false
	}

	fmt.Println(Green + "✅ Installed " + cmd + "." + Reset)
	return true
}

// IsCliOnly returns true if the command is a built-in framework utility
// that does not require database, redis, or service connections.
func IsCliOnly(cmd string) bool {
	switch cmd {
	case "", "help", "version", "stub:publish", "dev", "lint", "scan":
		return true
	case "build:app", "build:release", "build:cli":
		return true
	case "make:model", "make:vue", "make:lang", "make:handler", "make:action", "make:repo", "make:req", "make:notification", "make:task", "make:crud", "make:migration":
		return true
	}
	return false
}
