package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
	"time"
)

func TestJWT(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	seed := base64.StdEncoding.EncodeToString(priv.Seed())
	public := base64.StdEncoding.EncodeToString(pub)
	keys, e := ParseKeys(seed, public)
	if e != nil {
		t.Fatal(e)
	}
	verifier, e := ParseKeys("", public)
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	keys.Now = func() time.Time { return now }
	verifier.Now = keys.Now
	token := keys.Sign("user", "user@example.test")
	claims, e := verifier.Verify(token)
	if e != nil || claims.Subject != "user" {
		t.Fatal(claims, e)
	}
	verifier.Now = func() time.Time { return now.Add(time.Hour) }
	if _, e = verifier.Verify(token); e == nil {
		t.Fatal("Expired token accepted")
	}
	verifier.Now = keys.Now
	for _, pair := range [][2]string{{"", ""}, {"bad", ""}, {"", "bad"}, {base64.StdEncoding.EncodeToString([]byte("short")), ""}, {"", base64.StdEncoding.EncodeToString([]byte("short"))}} {
		if _, e = ParseKeys(pair[0], pair[1]); e == nil {
			t.Fatal("Invalid key accepted")
		}
	}
	if _, e = ParseKeys(seed, ""); e != nil {
		t.Fatal(e)
	}
	for _, v := range []string{"", token + "extra", "a.b.c", "@@@.a.b", "e30.a.b", "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.a.b", strings.Join([]string{strings.Split(token, ".")[0], strings.Split(token, ".")[1], "@@@"}, ".")} {
		if _, e = keys.Verify(v); e == nil {
			t.Fatal("Invalid JWT accepted", v)
		}
	}
	signRaw := func(body string) string {
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT"}`)) + "." + base64.RawURLEncoding.EncodeToString([]byte(body))
		return payload + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv, []byte(payload)))
	}
	for _, body := range []string{"not json", `{}`, `{"sub":"user","role":"admin"}`} {
		if _, e = keys.Verify(signRaw(body)); e == nil {
			t.Fatal("Invalid claims accepted")
		}
	}
}
