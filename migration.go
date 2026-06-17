package xcli

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func (x *XCli) migrateCommand() *cobra.Command {
	var connection string
	var domain string
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run all database migrations in proper order",
		Run: func(cmd *cobra.Command, args []string) {
			dialect, dsn, tmpDir, err := x.collectMigrations(connection, domain)
			if err != nil {
				fmt.Println(Red + "❌ Error: " + err.Error() + Reset)
				os.Exit(1)
			}

			fmt.Println("🔄 Running combined migrations from temp folder...")
			// run combined migrations
			if err := x.runInCli(cmd.Context(), "goose", []string{dialect, dsn, "up", "--dir", tmpDir}); err != nil {
				fmt.Println(Red+"❌ Goose make migration failed:", err, Reset)
				os.Exit(1)
			}

			fmt.Println(Green + "✅ All migrations executed." + Reset)
		},
	}
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Run migration for specific domain")
	cmd.Flags().StringVarP(&connection, "connection", "c", "default", "Specify a connection to run migrations in it (default: default)")
	return cmd
}

func (x *XCli) migrateRollbackCommand() *cobra.Command {
	var domain string
	var connection string

	cmd := &cobra.Command{
		Use:   "migrate:rollback",
		Short: "Rollback the last database migration across all domains",
		Run: func(cmd *cobra.Command, args []string) {
			dialect, dsn, tmpDir, err := x.collectMigrations(connection, domain)
			if err != nil {
				fmt.Println(Red + "❌ Error: " + err.Error() + Reset)
				os.Exit(1)
			}

			fmt.Println("🔄 Rolling back last migration from temp folder...")
			if err := x.runInCli(cmd.Context(), "goose", []string{dialect, dsn, "reset", "--dir", tmpDir}); err != nil {
				fmt.Println(Red+"❌ Goose rollback failed:", err, Reset)
				os.Exit(1)
			}

			fmt.Println(Green + "✅ Rollback completed." + Reset)
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Rollback migration for specific domain")
	cmd.Flags().StringVarP(&connection, "connection", "c", "default", "Specify connection to use")
	return cmd
}

func (x *XCli) collectMigrations(connection string, domain string) (string, string, string, error) {
	if connection == "" {
		return "", "", "", fmt.Errorf("no connection is specified")
	}

	dialect, dsn := getDBConfig(connection)

	var migrationDirs []string

	if domain != "" {
		dir := filepath.Join("app/domains", domain, "database/migrations")
		migrationDirs = append(migrationDirs, dir)
	} else {
		// Add global migrations
		migrationDirs = append(migrationDirs, "app/database/migrations")

		if x.IsFolderExists("app/domains") {
			// Add domain migrations
			entries, err := os.ReadDir("app/domains")
			if err != nil {
				return "", "", "", errors.New("Failed to read domains:" + err.Error())
			}
			for _, entry := range entries {
				if entry.IsDir() {
					dir := filepath.Join("app/domains", entry.Name(), "database/migrations")
					if _, err := os.Stat(dir); err == nil {
						migrationDirs = append(migrationDirs, dir)
					}
				}
			}
		}
	}

	// Merge all .sql files into temp folder
	tmpDir, err := os.MkdirTemp("", "xmigration_combined_migrations_")
	if err != nil {
		log.Fatalf("failed to create temporary directory: %v", err)
	}

	for _, dir := range migrationDirs {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
				src := filepath.Join(dir, file.Name())
				dst := filepath.Join(tmpDir, file.Name())
				data, err := os.ReadFile(src)
				if err != nil {
					log.Println("error reading:", src, err)
					continue
				}
				if err := os.WriteFile(dst, data, os.ModePerm); err != nil {
					log.Println("error writing:", dst, err)
				}
			}
		}
	}

	return dialect, dsn, tmpDir, nil
}

func (x *XCli) makeMigrationCommand() *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "make:migration [name]",
		Short: "Create a new database migration",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			var dir string
			if domain != "" {
				dir = fmt.Sprintf("app/domains/%s/database/migrations", domain)
			} else {
				dir = "app/database/migrations"
			}

			// Ensure directory exists
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, os.ModePerm); err != nil {
					fmt.Println(Red+"❌ Failed to create migration directory:", err, Reset)
					os.Exit(1)
				}
			}

			if err := x.runInCli(cmd.Context(), "goose", []string{"create", name, "sql", "--dir", dir}); err != nil {
				fmt.Println(Red+"❌ Goose make migration failed:", err, Reset)
				os.Exit(1)
			}
		},
	}

	// Register the --domain flag
	cmd.Flags().StringVarP(&domain, "domain", "d", "", "The domain name (e.g., auth)")
	return cmd
}

func (x *XCli) migrateStatusCommand() *cobra.Command {
	var domain string
	var connection string

	cmd := &cobra.Command{
		Use:   "migrate:status",
		Short: "Show the status of each migration (applied/pending)",
		Run: func(cmd *cobra.Command, args []string) {
			dialect, dsn, tmpDir, err := x.collectMigrations(connection, domain)
			if err != nil {
				fmt.Println(Red + "❌ Error: " + err.Error() + Reset)
				os.Exit(1)
			}

			fmt.Println("📄 Checking migration status from temp folder...")

			if err := x.runInCli(cmd.Context(), "goose", []string{"-dir", tmpDir, dialect, dsn, "status"}); err != nil {
				fmt.Println(Red+"❌ Goose migration status failed:", err, Reset)
				os.Exit(1)
			}
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Check status for specific domain")
	cmd.Flags().StringVarP(&connection, "connection", "c", "default", "Specify connection to use")
	return cmd
}

func (x *XCli) migrateResetCommand() *cobra.Command {
	var domain, connection string

	cmd := &cobra.Command{
		Use:   "migrate:reset",
		Short: "Rollback all migrations (one by one)",
		Run: func(cmd *cobra.Command, args []string) {
			dialect, dsn, tmpDir, err := x.collectMigrations(connection, domain)
			if err != nil {
				fmt.Println(Red + "❌ Error: " + err.Error() + Reset)
				os.Exit(1)
			}
			fmt.Println("🔁 Resetting all migrations...")

			if err := x.runInCli(cmd.Context(), "goose", []string{"-dir", tmpDir, dialect, dsn, "reset"}); err != nil {
				fmt.Println(Red+"❌ Goose migration reset failed:", err, Reset)
				os.Exit(1)
			}

			fmt.Println(Green + "✅ All migrations rolled back." + Reset)
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Target domain")
	cmd.Flags().StringVarP(&connection, "connection", "c", "default", "Target connection")
	return cmd
}

func (x *XCli) migrateRefreshCommand() *cobra.Command {
	var domain, connection string

	cmd := &cobra.Command{
		Use:   "migrate:refresh",
		Short: "Reset and re-run all migrations",
		Run: func(cmd *cobra.Command, args []string) {
			dialect, dsn, tmpDir, err := x.collectMigrations(connection, domain)
			if err != nil {
				fmt.Println(Red + "❌ Error: " + err.Error() + Reset)
				os.Exit(1)
			}
			fmt.Println("♻️  Refreshing migrations...")

			if err := x.runInCli(cmd.Context(), "goose", []string{"-dir", tmpDir, dialect, dsn, "reset"}); err != nil {
				fmt.Println(Red+"❌ Goose migration refresh failed on reset:", err, Reset)
				os.Exit(1)
			}

			if err := x.runInCli(cmd.Context(), "goose", []string{"-dir", tmpDir, dialect, dsn, "up"}); err != nil {
				fmt.Println(Red+"❌ Goose migration refresh faild on up failed:", err, Reset)
				os.Exit(1)
			}

			fmt.Println(Green + "✅ Migrations refreshed." + Reset)
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Target domain")
	cmd.Flags().StringVarP(&connection, "connection", "c", "default", "Target connection")
	return cmd
}

func (x *XCli) migrateUpByCommand() *cobra.Command {
	var domain, connection string

	cmd := &cobra.Command{
		Use:   "migrate:next [n]",
		Short: "Run the next N migrations (calls goose up-by-one N times)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dialect, dsn, tmpDir, err := x.collectMigrations(connection, domain)
			if err != nil {
				fmt.Println(Red + "❌ Error: " + err.Error() + Reset)
				os.Exit(1)
			}
			n, err := strconv.Atoi(args[0])
			if err != nil || n < 1 {
				fmt.Println(Red + "❌ Please provide a valid number greater than 0" + Reset)
				os.Exit(1)
			}

			for i := range n {
				fmt.Printf("🔼 Running migration step %d of %d...\n", i+1, n)

				if err := x.runInCli(cmd.Context(), "goose", []string{"-dir", tmpDir, dialect, dsn, "up-by-one"}); err != nil {
					fmt.Println(Red+"❌ Goose migration up by one failed:", err, Reset)
					os.Exit(1)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Target domain")
	cmd.Flags().StringVarP(&connection, "connection", "c", "default", "Target connection")
	return cmd
}

func (x *XCli) migrateDownByCommand() *cobra.Command {
	var domain, connection string

	cmd := &cobra.Command{
		Use:   "migrate:back [n]",
		Short: "Rollback the last N migrations (calls goose down N times)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			dialect, dsn, tmpDir, err := x.collectMigrations(connection, domain)
			if err != nil {
				fmt.Println(Red + "❌ Error: " + err.Error() + Reset)
				os.Exit(1)
			}
			n, err := strconv.Atoi(args[0])
			if err != nil || n < 1 {
				fmt.Println(Red + "❌ Please provide a valid number greater than 0" + Reset)
				os.Exit(1)
			}

			for i := range n {
				fmt.Printf("🔽 Rolling back step %d of %d...\n", i+1, n)

				if err := x.runInCli(cmd.Context(), "goose", []string{"-dir", tmpDir, dialect, dsn, "down"}); err != nil {
					fmt.Println(Red+"❌ Goose migration down failed:", err, Reset)
					os.Exit(1)
				}
			}
		},
	}

	cmd.Flags().StringVarP(&domain, "domain", "d", "", "Target domain")
	cmd.Flags().StringVarP(&connection, "connection", "c", "default", "Target connection")
	return cmd
}

func (x *XCli) goosePassthroughCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "goose",
		Short:              "Direct access to goose CLI (useful for advanced control)",
		DisableFlagParsing: true, // allows passing flags to goose directly
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println(Red + "❌ Error: No arguments provided for goose" + Reset)
				os.Exit(1)
			}

			fmt.Println("📤 Forwarding to goose:", strings.Join(args, " "))
			x.runInCli(cmd.Context(), "goose", args)
		},
	}
	return cmd
}

func getDBConfig(connection string) (string, string) {
	prefix := "DB_"
	if connection != "" && connection != "default" {
		prefix = "DB_" + strings.ToUpper(connection) + "_"
	}

	dialect := getEnv(prefix+"DIALECT", "postgres")
	user := getEnv(prefix+"USERNAME", getEnv("DB_USERNAME", "postgres"))
	pass := getEnv(prefix+"PASSWORD", getEnv("DB_PASSWORD", "123456"))
	host := getEnv(prefix+"HOST", getEnv("DB_HOST", "localhost"))
	port := getEnv(prefix+"PORT", getEnv("DB_PORT", "5432"))
	database := getEnv(prefix+"DATABASE", getEnv("DB_DATABASE", "xframe"))
	sslmode := getEnv(prefix+"SSLMODE", "disable")
	timezone := getEnv(prefix+"TIMEZONE", "UTC")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&TimeZone=%s",
		url.PathEscape(user),
		url.PathEscape(pass),
		host, port,
		database, sslmode,
		url.QueryEscape(timezone),
	)

	return dialect, dsn
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
