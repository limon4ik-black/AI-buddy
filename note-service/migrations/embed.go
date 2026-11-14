package migrations

import _ "embed"

//go:embed 000003_onboarding.up.sql
var Onboarding []byte
