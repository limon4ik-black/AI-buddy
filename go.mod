module github.com/limon4ik-black/AI-buddy

go 1.27.1

require (
	github.com/jackc/pgx/v5 v5.7.6
	github.com/limon4ik-black/AI-buddy/chat-service v0.0.0
	github.com/limon4ik-black/AI-buddy/note-service v0.0.0
	golang.org/x/crypto v0.48.0
	golang.org/x/net v0.51.0
	golang.org/x/oauth2 v0.35.0
)

replace github.com/limon4ik-black/AI-buddy/chat-service => ./chat-service

replace github.com/limon4ik-black/AI-buddy/note-service => ./note-service

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)
