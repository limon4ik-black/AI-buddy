package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func main() {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	values := map[string]string{
		"JWT_PRIVATE_SEED": base64.StdEncoding.EncodeToString(private.Seed()),
		"JWT_PUBLIC_KEY": base64.StdEncoding.EncodeToString(public),
	}
	for _, name := range []string{"GATEWAY_SERVICE_TOKEN", "CHAT_SERVICE_TOKEN", "INDEXER_SERVICE_TOKEN", "RAG_SERVICE_TOKEN", "CORP_SERVICE_TOKEN", "POSTGRES_PASSWORD", "AUTH_DB_PASSWORD", "INFO_DB_PASSWORD", "CHAT_DB_PASSWORD"} {
		b := make([]byte, 32)
		if _, err = rand.Read(b); err != nil {
			panic(err)
		}
		values[name] = fmt.Sprintf("%x", b)
	}
	template, err := os.ReadFile(".env.example")
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(template), "\n"), "\n") {
		key, _, ok := strings.Cut(line, "=")
		if value, generated := values[key]; ok && generated {
			line = key + "=" + value
		}
		fmt.Println(line)
	}
}
