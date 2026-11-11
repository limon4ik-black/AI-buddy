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
  onLogin,
}: {
  busy: boolean;
  error: string;
  onLogin: (
    email: string,
    password: string,
    register: boolean,
  ) => Promise<void>;
}) {
  const [email, setEmail] = useState(demoAccounts[0].email),
    [password, setPassword] = useState(demoPassword),
    [register, setRegister] = useState(false);
  function submit(e: FormEvent) {
    e.preventDefault();
    void onLogin(email, password, register);
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
            className="text-button full-width"
            type="button"
            disabled={busy}
            onClick={() => {
              window.location.assign("/api/v1/auth/google/login");
            }}
          >
            Войти через Google
          </button>
          <button
            className="text-button auth-switch"
            disabled={busy}
            onClick={() => setRegister(!register)}
          >
            {register ? "Уже есть аккаунт? Войти" : "Создать свой аккаунт"}
          </button>
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
