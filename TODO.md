# TODO (auth-слайс)

## 1. Cookie oauth_state: Secure + домен

Где: `internal/feature/auth/login/handler.go`, метод `Login`:

```go
c.SetCookie("oauth_state", state, 300, "/", "", false, true)
```

Последние два аргумента: `secure` (сейчас `false`), `domain` (сейчас `""`).
На проде, где сервер за HTTPS, поставить `secure = true` и домен сайта:

```go
c.SetCookie("oauth_state", state, 300, "/", "мой-домен.ru", true, true)
```

- `Secure:true` — cookie уходит только по HTTPS, state (половина защиты от CSRF)
  не светится в открытом трафике. Локально с `false`, т.к. localhost работает
  по голому http и браузер не примет Secure-cookie.
- домен — ограничить, каким хостам браузер отдаёт cookie.

## 2. Как подключить ручки в приложении

Зависимости auth-слайса собираются один раз, каждая фича регистрирует свои
роуты сама через `.Register(r gin.IRouter)`:

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/limon4ik-black/AI-buddy/internal/feature/auth"
    "github.com/limon4ik-black/AI-buddy/internal/feature/auth/login"
    "github.com/limon4ik-black/AI-buddy/internal/feature/auth/logout"
    "github.com/limon4ik-black/AI-buddy/internal/feature/auth/refresh"
)

// зависимости
priv := ...                                        // ed25519.PrivateKey (кто подписывает access-JWT)
issuer := auth.NewTokenIssuer(priv)
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"}) // из REDIS_ADDR
refreshStore := auth.NewRedisRefreshStore(rdb)     // refresh-токены, TTL 7 дней
users := auth.NewRedisUserStore(rdb)               // provider+sub -> user_id

// регистрации (каждый пакет вешает свои ручки сам)
conf := login.NewGoogleConf(clientID, clientSecret, redirectURL) // из .env
login.NewHandler(conf, login.NewService(conf, users, issuer, refreshStore)).Register(r)
refresh.NewHandler(refresh.NewService(issuer, refreshStore)).Register(r)
logout.NewHandler(logout.NewService(refreshStore)).Register(r)

// защита любых ручек
r.GET("/что-то", middleware.AuthMiddleware(pub), handler)
```

Роуты, которые появятся:
    GET  /auth/google/login     редирект на Google + state в cookie
    GET  /auth/google/callback  обмен code, выдача пары токенов
    POST /auth/refresh          ротация пары по refresh-токену
    POST /auth/logout           удаление refresh-токена

Нужные переменные окружения (см. .env.example): GOOGLE_CLIENT_ID,
GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL (посимвольно как в Google Cloud
Console), REDIS_ADDR, HTTPS_PROXY (локальный socks5 для исходящих к Google).
