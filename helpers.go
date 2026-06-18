package xcli

import (
	"os"
	"strings"
)

func ucFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func toCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func getEnvPath(envKey, fallback string) string {
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	return fallback
}
