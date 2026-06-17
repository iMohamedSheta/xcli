package xcli

type InstallSpec struct {
	Command     string
	Args        []string
	FailMessage string
}

var suggestInstallCmds = map[string]InstallSpec{
	"golangci-lint": {
		Command: "go",
		Args: []string{
			"install",
			"github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
		},
	},
	"air": {
		Command: "go",
		Args:    []string{"install", "github.com/cosmtrek/air@latest"},
	},
	"goose": {
		Command: "go",
		Args:    []string{"install", "github.com/pressly/goose/v3/cmd/goose@latest"},
	},
	"npm": {
		FailMessage: "Please install Node.js and npm manually:\n",
	},
	"garble": {
		Command: "go",
		Args:    []string{"install", "mvdan.cc/garble@latest"},
	},
}
