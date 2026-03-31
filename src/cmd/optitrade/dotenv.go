package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// loadDashboardEnv looks for `.env` in the current working directory, then each parent,
// and loads the first file found. It does not override variables already set in the process
// environment (same semantics as `export` before launch).
func loadDashboardEnv(log *slog.Logger) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	for d := wd; d != ""; {
		p := filepath.Join(d, ".env")
		st, err := os.Stat(p)
		if err == nil && !st.IsDir() {
			if err := godotenv.Load(p); err != nil {
				if log != nil {
					log.Warn("dotenv: could not load .env file", "path", p, "err", err)
				}
				return
			}
			return
		}
		parent := filepath.Dir(d)
		if parent == d {
			break
		}
		d = parent
	}
}
