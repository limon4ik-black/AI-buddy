// Команда для локальной проверки руками: печатает подписанный токен для Postman.
// Требует JWT_PRIVATE_SEED (лежит в .env репы у auth); сервис приватный ключ не знает.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/limon4ik-black/AI-buddy/note-service/internal/core/security"
)

func main() {
	userID := flag.String("user_id", "", "id пользователя, пойдёт в sub токена")
	email := flag.String("email", "dev@local", "email для claim")
	ttl := flag.Duration("ttl", 0, "срок жизни; по умолчанию как у auth — 15m")
	flag.Parse()

	if *userID == "" {
		fmt.Fprintln(os.Stderr, "укажи -user_id <id>")
		os.Exit(1)
	}
	keys, err := security.ParseKeys(os.Getenv("JWT_PRIVATE_SEED"), "")
	if err != nil {
		fmt.Fprintln(os.Stderr, "нужен JWT_PRIVATE_SEED в окружении:", err)
		os.Exit(1)
	}
	if *ttl != 0 {
		keys.TTL = *ttl
	}
	fmt.Fprintf(os.Stderr, "токен для %s, действует %s\n", *userID, keys.TTL)
	fmt.Println(keys.Sign(*userID, *email))
}
