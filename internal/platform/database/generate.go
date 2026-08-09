// Package database contains PostgreSQL connection lifecycle and generated SQL
// access. Business queries remain in their owning feature packages.
package database

//go:generate go tool sqlc generate -f ../../../sqlc.yaml
