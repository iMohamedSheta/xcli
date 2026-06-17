# XCli - Developer Tooling Made Simple

`XCli` is a powerful developer tooling and code generation command-line interface (CLI) built for modern Go web applications. It automates database migrations, builds, and bootstraps full backend architectures (Models, Repositories, Handlers, Requests, Tasks) and frontend components (Vue 3, Language translation files).

---

## Features

- **Code Generation (`make:*`)**: Instantly bootstrap handlers, models, actions, repositories, requests, tasks, notifications, and Vue components.
- **Full CRUD Generator (`make:crud`)**: Generate all backend layers (Request, Action, Repository, Handler) and a frontend Vue Index page (complete with pagination, filters, and tables) for any entity in a single command.
- **Database Migrations (`migrate:*`)**: Run, rollback, reset, check status, or direct passthrough to the underlying Goose driver. Supports multi-database connections and modular domain-driven layouts.
- **Customizable Code Generation**: Publish embedded templates (stubs) to your local project, edit them, and have the code generator respect your local overrides.

---

## Installation

Since `XCli` is built for Go web projects, it is recommended to install it directly using the Go toolchain. This ensures it is compiled correctly for your system architecture:

```bash
# Install globally
go install github.com/imohamedsheta/xcli/cmd/xcli@latest
```

Ensure your Go bin directory (usually `$HOME/go/bin` or `%USERPROFILE%\go\bin` on Windows) is added to your system `PATH`.

---

## Database Configuration

`XCli` manages database connections dynamically using environment variables.

### 1. Environment Files
When executed, `XCli` automatically searches for environment variables in the following order:
1. `.env.xcli` (Local CLI configuration overrides)
2. `.env` (Standard application configuration)

### 2. Default Connection Settings
By default, database operations (like migrations) look for environment variables prefixed with `DB_`. The supported parameters (with their PostgreSQL defaults) are:

```ini
DB_DIALECT=postgres
DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=xframe
DB_USERNAME=postgres
DB_PASSWORD=123456
DB_SSLMODE=disable
DB_TIMEZONE=UTC
```

### 3. Multiple Database Connections
If your application uses multiple database connections (e.g. `tenant` or `billing`), you can define separate credentials in your `.env` file by prefixing them with the uppercase name of the connection (e.g., `DB_[CONNECTION-NAME]_*`):

```ini
# Default connection
DB_DATABASE=xframe_default
DB_PASSWORD=123456

# Tenant connection credentials
DB_TENANT_DATABASE=xframe_tenant
DB_TENANT_PASSWORD=secretpassword
DB_TENANT_HOST=tenant-db.host
```

To run migrations or status checks on a specific connection, use the `-c` or `--connection` flag:

```bash
# Run migrations on the default database
xcli migrate

# Run migrations specifically on the tenant database connection
xcli migrate --connection tenant
```

### 4. CLI-Only Commands (Optimizing App Boot)
For faster execution and to prevent startup failures when environment configuration files (like `.env`) are missing, you can skip loading database and cache services for commands that don't need them.

`XCli` provides a helper function `xcli.IsCliOnly(cmd)` that returns `true` for all generator and helper commands (like `make:*`, `stub:publish`, `help`, `version`, `build:*`, `dev`, etc.).

In your local `main.go`, use it to conditionally boot your application:

```go
package main

import (
	"fmt"
	"os"
	"github.com/imohamedsheta/xapp/app/registers"
	"github.com/imohamedsheta/xapp/bootstrap"
	"github.com/imohamedsheta/xcli"
)

func main() {
	// 1. Determine environment file (load .env.xcli if it exists, fallback to .env)
	envFile := ".env"
	if _, err := os.Stat(".env.xcli"); err == nil {
		envFile = ".env.xcli"
	}

	// 2. Peek at the requested command argument
	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	// check if command is cli only
	if xcli.IsCliOnly(cmd) {
		// Minimal boot: Config + Logger only (no database, no redis connections needed)
		bootstrap.NewAppBuilder(envFile).
			LoadEnvFile().
			LoadConfig().
			LoadLogger().
			Boot(registers.ServiceProviders)
	} else {
		// Full boot: Load database, redis, websocket, notifications, etc.
		bootstrap.NewAppBuilder(envFile).
			MustLoadEnvFile().
			LoadConfig().
			LoadLogger().
			LoadDatabase().
			LoadStorage().
			LoadValidator(registers.ValidationRules()).
			LoadRedisCache().
			LoadRedisQueue().
			LoadWebsocketServer().
			LoadNotify(registers.NotifyChannels()).
			LoadInertia().
			LoadSocialite().
			LoadXErr().
			Boot(registers.ServiceProviders)
	}

	// 3. Instantiate xcli
	x := xcli.New()

	// 4. Register custom application commands
	// You can load them from registers, or declare them directly:
	x.Register(registers.Commands()...)

	// 5. Run the CLI
	if err := x.Execute(); err != nil {
		fmt.Printf("\033[31m❌ Command failed: %v\033[0m\n", err)
		os.Exit(1)
	}
}

```

---

## Domain-Driven Migration Layout

`XCli` supports both monolithic and modular domain-driven migrations out-of-the-box. When you run `migrate`, `migrate:status`, or `migrate:rollback`, the CLI automatically:
1. Gathers global migrations located in `app/database/migrations/`.
2. Scans for any active subdirectories in `app/domains/*/database/migrations/`.
3. Merges these migrations sequentially in a temporary sandbox to be processed in order.

To generate a migration file inside a specific domain, use the `-d` or `--domain` flag:

```bash
# Create a global database migration
xcli make:migration create_users_table

# Create a migration within a specific domain (e.g. auth domain)
xcli make:migration create_tokens_table --domain auth
```

---

## Customizing Code Generation (Stubs)

All generated files (`models`, `handlers`, etc.) are built from template files called **stubs**. By default, these stubs are embedded inside the `xcli` binary. However, you can export and edit them to match your project's coding standards.

### 1. Publish the Stubs
Export the embedded stubs to your local project directory:

```bash
xcli stub:publish
```

This creates a `stubs/` directory in your project root containing:
- `model.stub` (database models)
- `handler.stub` & `crud_handler.stub` (HTTP handlers)
- `repository.stub` & `crud_repository.stub` (database querying layer)
- `request.stub` & `crud_request.stub` (input validation)
- `vue_component.stub` & `crud_vue_index.stub` (Frontend UI templates)
- `task.stub`, `notification.stub`, `action.stub`, etc.

*To reset or re-publish all stubs and overwrite your local edits, run:*
```bash
xcli stub:publish --force
```

### 2. Resolution Strategy
When generating a file, `XCli` resolves templates in this order:
1. **Local Override**: Looks for `./stubs/[stub_name].stub` in your project root.
2. **Embedded Fallback**: If the local file is missing, it falls back to the default template embedded in the `xcli` binary.

This allows you to keep only the customized stubs you care about (e.g., `model.stub`) and safely delete the rest.

---

## Command Reference

### Stubs Customization
| Command | Description | Flags |
| :--- | :--- | :--- |
| `stub:publish` | Publish all stubs to the local directory | `-f, --force` (overwrite existing stubs) |

### Code Generators (`make`)
| Command | Description | Flags |
| :--- | :--- | :--- |
| `make:model [name]` | Create a database model | |
| `make:handler [name]` | Create an HTTP handler | `-a` (also action), `-r` (also repo), `-R` (also request) |
| `make:action [name]` | Create a business logic action stub | |
| `make:repo [name]` | Create a database repository stub | |
| `make:req [name]` | Create an input request validator | |
| `make:notification [name]` | Create a notification template | |
| `make:task [name]` | Create a background Asynq task handler | |
| `make:vue [name]` | Create a Vue 3 SFC component | `-v, --view` (save in pages folder instead of components) |
| `make:lang [name]` | Create TypeScript lang files in all locales | |
| `make:crud [name]` | Generate full CRUD operation pages/modules | `-p` (subpackage, default: `admin`), `--skip-vue`, `--skip-repo` |

### Database Migrations (`migrate`)
| Command | Description | Flags |
| :--- | :--- | :--- |
| `migrate` | Run all pending migrations | `-d` (domain), `-c` (connection name) |
| `migrate:rollback` | Rollback the last migration batch | `-d` (domain), `-c` (connection name) |
| `migrate:status` | Show status of all migrations | `-d` (domain), `-c` (connection name) |
| `migrate:reset` | Rollback all database migrations | `-d` (domain), `-c` (connection name) |
| `migrate:refresh` | Rollback all migrations and run them again | `-d` (domain), `-c` (connection name) |
| `migrate:next [n]` | Run the next `N` migrations sequentially | `-d` (domain), `-c` (connection name) |
| `migrate:back [n]` | Rollback the last `N` migrations sequentially | `-d` (domain), `-c` (connection name) |
| `migrate:make [name]` | Create a new SQL migration file | `-d` (domain) |
| `goose [args]` | Direct passthrough to the underlying Goose driver | |
