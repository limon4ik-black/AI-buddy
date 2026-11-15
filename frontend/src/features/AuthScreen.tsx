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
  onForgot,
  onResetConfirm,
}: {
  busy: boolean;
  error: string;
  notice: string;
  codeSent: boolean;
  onLogin: (email: string, password: string) => Promise<void>;
  onConfirm: (code: string) => Promise<void>;
  onForgot: (email: string) => Promise<boolean>;
  onResetConfirm: (code: string, password: string) => Promise<void>;
}) {
  const [email, setEmail] = useState(demoAccounts[0].email),
    [password, setPassword] = useState(demoPassword),
    [code, setCode] = useState(""),
    [reset, setReset] = useState<"ask" | "code" | null>(null);
  function submit(e: FormEvent) {
    e.preventDefault();
    void onLogin(email, password);
  }
  function submitCode(e: FormEvent) {
    e.preventDefault();
    void onConfirm(code);
  }
  async function askReset(e: FormEvent) {
    e.preventDefault();
    if (await onForgot(email)) setReset("code");
  }
  function submitReset(e: FormEvent) {
    e.preventDefault();
    void onResetConfirm(code, password);
  }
  function backToLogin() {
    setReset(null);
    setPassword("");
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
            Больше, чем
            <br />
            просто <span>AI-buddy.</span>
          </h1>
          <p>
            Наставник, заметки, доски и чек-лист адаптации в одном месте. Всё
            нужное для уверенного старта в новой команде.
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
          <div className="auth-stickers" aria-label="Возможности помощника">
            <span>💬 Ответы на вопросы</span>
            <span>🔑 Доступы</span>
            <span>📋 Чек-лист адаптации</span>
            <span>📝 Заметки и доска</span>
          </div>
        </section>
        <section className="auth-card">
          <div className="auth-card-title">
            <span className="status-chip">
              <i /> Демо-пространство
            </span>
            <h2>С возвращением</h2>
            <p>Войди с рабочей почтой и паролем.</p>
          </div>
          {reset === "code" ? (
            <form onSubmit={submitReset}>
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
              <label>
                Новый пароль
                <input
                  type="password"
                  minLength={10}
                  maxLength={72}
                  autoComplete="new-password"
                  required
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </label>
              <button className="primary-button full-width" disabled={busy}>
                {busy ? "Меняем пароль…" : "Сменить пароль"}
                <Icon name="arrow" size={18} />
              </button>
              <button
                type="button"
                className="text-button auth-switch"
                disabled={busy}
                onClick={backToLogin}
              >
                Вернуться ко входу
              </button>
            </form>
          ) : reset === "ask" ? (
            <form onSubmit={askReset}>
              {error && (
                <div className="inline-error" role="alert">
                  <Icon name="shield" size={18} />
                  {error}
                </div>
              )}
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
              <button className="primary-button full-width" disabled={busy}>
                {busy ? "Отправляем…" : "Получить код"}
                <Icon name="arrow" size={18} />
              </button>
              <button
                type="button"
                className="text-button auth-switch"
                disabled={busy}
                onClick={backToLogin}
              >
                Вернуться ко входу
              </button>
            </form>
          ) : codeSent ? (
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
                    autoComplete="current-password"
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
                  {busy ? "Подключаемся…" : "Войти в пространство"}
                  <Icon name="arrow" size={18} />
                </button>
              </form>
              <button
                className="text-button auth-switch"
                type="button"
                disabled={busy}
                onClick={() => {
                  setReset("ask");
                  setPassword("");
                }}
              >
                Забыли пароль?
              </button>
            </>
          )}
          <p className="auth-footnote">
            Нет аккаунта? Обратись к администратору команды.
            <br />
            Новых пользователей добавляет администратор.
          </p>
        </section>
      </div>
      <footer className="auth-footer">
        AI Buddy · Пространство адаптации<span>Создано для хакатона</span>
      </footer>
    </main>
  );
}
