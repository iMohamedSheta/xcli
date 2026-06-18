# XCli - Developer Tooling Made Simple

`XCli` is a powerful developer tooling and code generation command-line interface (CLI) built for xapp framework for golang fullstack web applications. It automates database migrations, builds, and bootstraps modular backend architectures (Handlers, Repositories, Requests, Actions, Tasks) and frontend components (Vue 3, Language translation files).

---

## Features

- **Modular Architecture Support**: Generates code structured for feature-oriented modules (`app/modules/<module>`) natively.
- **Code Scaffolding (`make:module`)**: Scaffold a new application module containing handler, repository, requests, and router configurations in a single run.
- **Full CRUD Generator (`make:crud`)**: Generate all backend layers (Request, Action, Repository, Handler) and a frontend Vue Index page (complete with pagination, filters, and tables) for any entity within a module.
- **Database Migrations (`migrate:*`)**: Run, rollback, reset, check status, or direct passthrough to the underlying Goose driver. Supports multi-database connections and a unified global migrations path.
- **Customizable Code Generation**: Publish embedded templates (stubs) to your local project, edit them, and have the code generator respect your local overrides.
- **Flexible Path Configuration**: Customize output directories and Go import packages dynamically via environment variables.

---

## Installation

Since `XCli` is built for Go web projects, install it directly using the Go toolchain:

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
	x.Register(registers.Commands()...)

	// 5. Run the CLI
	if err := x.Execute(); err != nil {
		fmt.Printf("\033[31m❌ Command failed: %v\033[0m\n", err)
		os.Exit(1)
	}
}
```

---

## Overriding Paths & Go Imports

You can customize directory paths and Go package import targets for code generation using environment variables. This is useful when adapting `XCli` to custom project layouts.

### 1. Directory Path Overrides
| Environment Variable | Default Value | Description |
| :--- | :--- | :--- |
| `XCLI_PATH_MODULES` | `app/modules` | Output path for modules and their Go files |
| `XCLI_PATH_MODELS` | `app/models` | Output path for DB models |
| `XCLI_PATH_MIGRATIONS` | `app/database/migrations` | Output path for Goose SQL migrations |
| `XCLI_PATH_TASKS` | `app/domain/tasks` | Output path for background tasks |
| `XCLI_PATH_NOTIFICATIONS`| `app/domain/notifications` | Output path for notification modules |
| `XCLI_PATH_VUE_PAGES` | `resources/js/pages` | Output path for CRUD Vue pages |

### 2. Import Path Replacement Overrides
| Environment Variable | Default Value | Replaces in Stubs |
| :--- | :--- | :--- |
| `XCLI_IMPORT_HANDLER` | `[module]/app/http/handler` | `{{HANDLER_PATH}}` |
| `XCLI_IMPORT_MODELS` | `[module]/app/models` | `{{MODELS_PATH}}` |
| `XCLI_IMPORT_ENUMS` | `[module]/app/domain/enums` | `{{ENUMS_PATH}}` |
| `XCLI_IMPORT_UTILS` | `[module]/app/domain/utils` | `{{UTILS_PATH}}` |
| `XCLI_IMPORT_X_APP` | `[module]/app/x` | `{{X_APP_PATH}}` |
| `XCLI_IMPORT_PKG` | `[module]/pkg` | `{{PKG_PATH}}` |
| `XCLI_IMPORT_INERTIA` | `[module]/pkg/inertia` | `{{INERTIA_PATH}}` |
| `XCLI_IMPORT_REQUESTS` | `[module]/app/http/requests` | `{{REQUESTS_PATH}}` |

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
- `task.stub`, `notification.stub`, `action.stub`, `routes.stub`, etc.

*To reset or re-publish all stubs and overwrite your local edits, run:*
```bash
xcli stub:publish --force
```

### 2. Resolution Strategy
When generating a file, `XCli` resolves templates in this order:
1. **Local Override**: Looks for `./stubs/[stub_name].stub` in your project root.
2. **Embedded Fallback**: If the local file is missing, it falls back to the default template embedded in the `xcli` binary.

---

## Command Reference

### Stubs Customization
| Command | Description | Flags |
| :--- | :--- | :--- |
| `stub:publish` | Publish all stubs to the local directory | `-f, --force` (overwrite existing stubs) |

### Code Generators (`make`)
| Command | Description | Flags |
| :--- | :--- | :--- |
| `make:module [name]` | Create a new application module structure | |
| `make:crud [name]` | Generate full CRUD operation pages/modules | `-m` (module name, default: name pluralized), `--skip-vue`, `--skip-repo` |
| `make:model [name]` | Create a database model | |
| `make:handler [name]` | Create an HTTP handler within a module | `-m` (module name), `-a` (also action), `-r` (also repo), `-R` (also request) |
| `make:action [name]` | Create a business logic action stub in a module | `-m` (module name) |
| `make:repo [name]` | Create a database repository stub in a module | `-m` (module name) |
| `make:req [name]` | Create an input request validator in a module | `-m` (module name) |
| `make:notification [name]`| Create a notification template | |
| `make:task [name]` | Create a background Asynq task handler | |
| `make:vue [name]` | Create a Vue 3 SFC component | `-v, --view` (save in pages folder instead of components) |
| `make:lang [name]` | Create TypeScript lang files in all locales | |

### Database Migrations (`migrate`)
| Command | Description | Flags |
| :--- | :--- | :--- |
| `migrate` | Run all pending migrations | `-c` (connection name) |
| `migrate:rollback` | Rollback the last migration batch | `-c` (connection name) |
| `migrate:status` | Show status of all migrations | `-c` (connection name) |
| `migrate:reset` | Rollback all database migrations | `-c` (connection name) |
| `migrate:refresh` | Rollback all migrations and run them again | `-c` (connection name) |
| `migrate:next [n]` | Run the next `N` migrations sequentially | `-c` (connection name) |
| `migrate:back [n]` | Rollback the last `N` migrations sequentially | `-c` (connection name) |
| `migrate:make [name]` | Create a new SQL migration file | |
| `goose [args]` | Direct passthrough to the underlying Goose driver | |
