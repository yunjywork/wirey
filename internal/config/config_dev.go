//go:build dev || (!dev && !production)

package config

// IsProduction is false when running with `wails dev` or no tags (default)
const IsProduction = false
