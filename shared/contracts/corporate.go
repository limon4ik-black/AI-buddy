package contracts

type CorpInput struct {
	UserID         string `json:"userId"`
	System         string `json:"system"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type CorpResult struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
