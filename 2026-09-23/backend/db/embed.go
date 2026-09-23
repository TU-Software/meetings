package db

import "embed"

// Migrations holds the contents of the migrations directory, embedded at
// compile time so the migrate binary is fully self-contained.
//
//go:embed migrations/*.sql
var Migrations embed.FS
