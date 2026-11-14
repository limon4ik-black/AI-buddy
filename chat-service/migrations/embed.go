package migrations

import _ "embed"

//go:embed 001_init.sql
var Initial string

//go:embed 002_context.sql
var Context string
