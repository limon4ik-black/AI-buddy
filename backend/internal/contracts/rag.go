package contracts

type Source struct {
	DocumentID string `json:"documentId"`
	Title      string `json:"title"`
	Section    string `json:"section"`
}

type Action struct {
	Type   string `json:"type"`
	Label  string `json:"label"`
	System string `json:"system"`
}

type Turn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateInput struct {
	RequestID   string      `json:"requestId"`
	Message     string      `json:"message"`
	History     []Turn      `json:"history"`
	UserContext UserContext `json:"userContext"`
}

type GenerateOutput struct {
	Answer          string   `json:"answer"`
	Sources         []Source `json:"sources"`
	SuggestedAction *Action  `json:"suggestedAction"`
	Mode            string   `json:"mode"`
}
