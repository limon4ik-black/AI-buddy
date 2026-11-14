package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	fmt.Println("JWT_PRIVATE_SEED=" + base64.StdEncoding.EncodeToString(private.Seed()))
	fmt.Println("JWT_PUBLIC_KEY=" + base64.StdEncoding.EncodeToString(public))
	for _, name := range []string{"GATEWAY_SERVICE_TOKEN", "CHAT_SERVICE_TOKEN", "INDEXER_SERVICE_TOKEN", "RAG_SERVICE_TOKEN", "CORP_SERVICE_TOKEN", "POSTGRES_PASSWORD", "AUTH_DB_PASSWORD", "INFO_DB_PASSWORD", "CHAT_DB_PASSWORD"} {
		b := make([]byte, 32)
		if _, err = rand.Read(b); err != nil {
			panic(err)
		}
		fmt.Printf("%s=%x\n", name, b)
	}
	fmt.Println("APP_ORIGIN=http://localhost:8080\nCOOKIE_SECURE=false\nSEED_DEMO=true")
}
