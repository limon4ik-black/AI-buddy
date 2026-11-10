package migrations

import (
	_ "embed"
)

// Auth is the initial schema for the authentication database.
//
//go:embed auth/001_init.sql
var Auth []byte

// Chat is the initial schema for chat history.
//
//go:embed chat/001_init.sql
var Chat []byte

// Info is the initial schema and synthetic data for onboarding information.
//
//go:embed info/001_init.sql
var Info []byte

// InfoMVP expands the synthetic company and onboarding data.
//
//go:embed info/002_mvp_data.sql
var InfoMVP []byte
