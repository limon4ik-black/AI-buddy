package contracts

import (
	"time"
)

type Profile struct {
	UserID    string    `json:"userId" db:"user_id"`
	FullName  string    `json:"fullName" db:"full_name"`
	TeamID    string    `json:"teamId" db:"team_id"`
	Position  string    `json:"position"`
	StartDate time.Time `json:"startDate" db:"start_date"`
}

type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Buddy       string `json:"buddy"`
	BuddyID     string `json:"buddyId" db:"buddy_id"`
}

type Tool struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	TeamID      *string `json:"teamId" db:"team_id"`
	DocumentID  string  `json:"documentId" db:"document_id"`
}

type Document struct {
	ID         string     `json:"id"`
	TeamID     *string    `json:"teamId" db:"team_id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	Category   string     `json:"category"`
	Visibility string     `json:"visibility"`
	SourceURL  string     `json:"sourceUrl" db:"source_url"`
	Synthetic  bool       `json:"synthetic"`
	UpdatedAt  time.Time  `json:"updatedAt" db:"updated_at"`
	DeletedAt  *time.Time `json:"deletedAt" db:"deleted_at"`
}

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Period      string     `json:"period"`
	DocumentID  *string    `json:"documentId" db:"document_id"`
	SourceURL   string     `json:"sourceUrl" db:"source_url"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completedAt" db:"completed_at"`
}

type Checklist struct {
	Tasks     []Task `json:"tasks"`
	Total     int    `json:"total"`
	Completed int    `json:"completed"`
	Percent   int    `json:"percent"`
	NextTask  *Task  `json:"nextTask"`
}

type UserContext struct {
	Profile            Profile    `json:"profile"`
	Team               TeamDetail `json:"team"`
	Role               string     `json:"role"`
	Onboarding         Checklist  `json:"onboarding"`
	AllowedDocumentIDs []string   `json:"allowedDocumentIds"`
	Tools              []Tool     `json:"tools"`
}

// The directory contains only company-visible synthetic professional data.
// It never exposes private contacts, credentials, or another trainee's progress.
type Member struct {
	ID       string `json:"id"`
	FullName string `json:"fullName" db:"full_name"`
	Position string `json:"position"`
	Role     string `json:"role"`
}

type TechnicalService struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	OwnerTeamID      string    `json:"ownerTeamId" db:"owner_team_id"`
	TechnicalOwnerID string    `json:"technicalOwnerId" db:"technical_owner_id"`
	Lifecycle        string    `json:"lifecycle"`
	RepositoryURL    string    `json:"repositoryUrl" db:"repository_url"`
	DocumentationURL string    `json:"documentationUrl" db:"documentation_url"`
	UpdatedAt        time.Time `json:"updatedAt" db:"updated_at"`
}

type TeamDetail struct {
	Team
	Members  []Member           `json:"members"`
	Services []TechnicalService `json:"services"`
}
