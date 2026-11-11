import { useState, type FormEvent } from "react";
import { Brand, Icon } from "../components/Icon";
import {
  demoAccounts,
  demoPassword,
  initials,
  teamColors,
  teamNames,
} from "../model";
export function AuthScreen({
  busy,
  error,
  notice,
  codeSent,
  onLogin,
  onConfirm,
}: {
  busy: boolean;
  error: string;
  notice: string;
  codeSent: boolean;
  onLogin: (
    email: string,
    password: string,
    register: boolean,
  ) => Promise<void>;
  onConfirm: (code: string) => Promise<void>;
}) {
  const [email, setEmail] = useState(demoAccounts[0].email),
    [password, setPassword] = useState(demoPassword),
    [register, setRegister] = useState(false),
    [code, setCode] = useState("");
  function submit(e: FormEvent) {
    e.preventDefault();
    void onLogin(email, password, register);
  }
  function submitCode(e: FormEvent) {
    e.preventDefault();
    void onConfirm(code);
  }
  return (
    <main className="auth-page">
      <div className="auth-top">
        <Brand />
        <span className="quiet-label">Hack&Change 2026</span>
      </div>
      <div className="auth-layout">
        <section className="auth-story">
          <span className="eyebrow">ТВОЙ ЧЕЛОВЕК В НОВОЙ КОМАНДЕ</span>
          <h1>
            Осваивайся.
            <br />
            Спрашивай.
            <br />
            <span>У тебя есть buddy.</span>
          </h1>
          <p>
            Команда, инструменты и первые задачи —<br />в одном пространстве, в
            твоём темпе.
          </p>
          <div className="auth-note">
            <span className="mini-orbit" aria-hidden="true">
              <i />
              <i />
              <i />
              <i />
            </span>
            <div>
              <strong>Первый день — уже проще</strong>
              <span>Ответы с источниками и понятный следующий шаг.</span>
            </div>
          </div>
          <div className="auth-team-preview">
            <div className="avatar-stack">
              {demoAccounts.slice(0, 4).map((a) => (
                <span
                  key={a.email}
                  className={"avatar " + teamColors[a.teamId]}
                >
                  {initials(a.name)}
                </span>
              ))}
            </div>
            <span>
              Четыре команды. Разные пути.
              <br />
              <strong>Один помощник рядом.</strong>
            </span>
          </div>
        </section>
        <section className="auth-card">
          <div className="auth-card-title">
            <span className="status-chip">
              <i /> Демо-пространство
            </span>
            <h2>{register ? "Давай познакомимся" : "Рады тебя видеть"}</h2>
            <p>
              {register
                ? "Создай аккаунт для знакомства с демо-командой."
                : "Выбери стажёра или войди со своим аккаунтом."}
            </p>
          </div>
          {codeSent ? (
            <form onSubmit={submitCode}>
              {notice && (
                <div className="inline-notice" role="status">
                  {notice}
                </div>
              )}
              {error && (
                <div className="inline-error" role="alert">
                  <Icon name="shield" size={18} />
                  {error}
                </div>
              )}
              <label>
                Код из письма
                <input
                  className="code-input"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  maxLength={6}
                  required
                  value={code}
                  onChange={(e) => setCode(e.target.value)}
                  placeholder="000000"
                />
              </label>
              <button className="primary-button full-width" disabled={busy}>
                {busy ? "Проверяем код…" : "Подтвердить вход"}
                <Icon name="arrow" size={18} />
              </button>
            </form>
          ) : (
            <>
          {!register && (
            <div className="demo-picker" aria-label="Тестовые стажёры">
              {demoAccounts.map((a) => (
                <button
                  key={a.email}
                  type="button"
                  disabled={busy}
                  className={
                    "demo-person " + (email === a.email ? "chosen" : "")
                  }
                  onClick={() => {
                    setEmail(a.email);
                    setPassword(demoPassword);
                  }}
                >
                  <span
                    aria-hidden="true"
                    className={"avatar " + teamColors[a.teamId]}
                  >
                    {initials(a.name)}
                  </span>
                  <span>
                    <strong>{a.name}</strong>
                    <small>{teamNames[a.teamId]}</small>
                  </span>
                  {email === a.email && <Icon name="check" size={16} />}
                </button>
              ))}
            </div>
          )}
          <form onSubmit={submit}>
            <label>
              Email
              <input
                type="email"
                autoComplete="username"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="name@company.ru"
              />
            </label>
            <label>
              Пароль
              <input
                type="password"
                minLength={10}
                maxLength={72}
                autoComplete={register ? "new-password" : "current-password"}
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </label>
            {error && (
              <div className="inline-error" role="alert">
                <Icon name="shield" size={18} />
                {error}
              </div>
            )}
            <button className="primary-button full-width" disabled={busy}>
              {busy
                ? "Подключаемся…"
                : register
                  ? "Создать аккаунт"
                  : "Войти в пространство"}
              <Icon name="arrow" size={18} />
            </button>
          </form>
          <div className="auth-divider">
            <span>или</span>
          </div>
          <button
            className="google-button"
            type="button"
            disabled={busy}
            onClick={() => {
              window.location.assign("/api/v1/auth/google/login");
            }}
          >
            <svg viewBox="0 0 48 48" aria-hidden="true">
              <path
                fill="#FFC107"
                d="M43.6 20.1H42V20H24v8h11.3C33.7 32.7 29.2 36 24 36c-6.6 0-12-5.4-12-12s5.4-12 12-12c3.1 0 5.9 1.2 8 3l5.7-5.7C34.1 6.1 29.3 4 24 4 13 4 4 13 4 24s9 20 20 20 20-9 20-20c0-1.3-.1-2.6-.4-3.9z"
              />
              <path
                fill="#FF3D00"
                d="M6.3 14.7l6.6 4.8C14.7 15.1 19 12 24 12c3.1 0 5.9 1.2 8 3l5.7-5.7C34.1 6.1 29.3 4 24 4 16.3 4 9.7 8.3 6.3 14.7z"
              />
              <path
                fill="#4CAF50"
                d="M24 44c5.2 0 9.9-2 13.4-5.2l-6.2-5.2C29.2 35.1 26.7 36 24 36c-5.2 0-9.6-3.3-11.3-8l-6.5 5C9.5 39.6 16.2 44 24 44z"
              />
              <path
                fill="#1976D2"
                d="M43.6 20.1H42V20H24v8h11.3c-.8 2.2-2.2 4.1-4.1 5.5l6.2 5.2C40.9 35.6 44 30.3 44 24c0-1.3-.1-2.6-.4-3.9z"
              />
            </svg>
            Войти через Google
          </button>
          <button
            className="text-button auth-switch"
            disabled={busy}
            onClick={() => setRegister(!register)}
          >
            {register ? "Уже есть аккаунт? Войти" : "Создать свой аккаунт"}
          </button>
            </>
          )}
          <p className="auth-footnote">
            Тестовые команды и вымышленные сотрудники.
            <br />
            Ответы помощника пока работают в демо-режиме.
          </p>
        </section>
      </div>
      <footer className="auth-footer">
        AI Buddy · Пространство адаптации<span>Создано для хакатона</span>
      </footer>
    </main>
  );
}
