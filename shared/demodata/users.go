// Package demodata contains synthetic identities shared by the demo seeders.
package demodata

import (
	_ "embed"
	"encoding/json"
	"slices"
)

// Password is public test data, used only when SEED_DEMO=true.
const Password = "BuddyDemo2026!"

// Intern links an Auth identity to its synthetic Info profile.
type Intern struct {
	ID         string   `json:"id"`
	Email      string   `json:"email"`
	FullName   string   `json:"fullName"`
	TeamID     string   `json:"teamId"`
	Position   string   `json:"position"`
	StartDate  string   `json:"startDate"`
	Completed  []string `json:"completed"`
	InProgress []string `json:"inProgress"`
}

//go:embed interns.json
var data []byte

// Interns decodes a fresh copy so callers cannot mutate shared seed state.
func Interns() []Intern {
	var result []Intern
	if err := json.Unmarshal(data, &result); err != nil {
		panic(err)
	}
	return result
}

// SupportsTool describes the systems accepted by the corporate demo adapter.
func SupportsTool(id string) bool {
	return slices.Contains([]string{"stash", "confluence", "vpn", "figma", "storybook", "metabase", "jupyter", "testrail", "test-environment"}, id)
}
