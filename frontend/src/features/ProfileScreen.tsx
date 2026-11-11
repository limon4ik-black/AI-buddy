import { useState, type FormEvent } from "react";
import type { Buddy } from "../hooks/useBuddy";
import { api, type Document } from "../api";
import { errorText } from "../hooks/useBuddy";
import { Icon } from "../components/Icon";
import { dateLabel, initials, roleLabel, teamColors } from "../model";
export function ProfileScreen({
  buddy,
  onDocument,
}: {
  buddy: Buddy;
  onDocument: (doc: Document) => void;
}) {
  const d = buddy.data!,
    team = d.teamDetails[d.profile.teamId],
    color = teamColors[d.profile.teamId];
  const [filter, setFilter] = useState("all"),
    [editing, setEditing] = useState(false),
    [fullName, setFullName] = useState(d.profile.fullName),
    [position, setPosition] = useState(d.profile.position);
  const filtered = d.checklist.tasks.filter(
    (t) =>
      filter === "all" ||
      (filter === "done" ? t.status === "completed" : t.status !== "completed"),
  );
  const periods = Array.from(new Set(filtered.map((t) => t.period)));
  const buddyPerson = team.members.find((m) => m.id === team.buddyId);
  async function save(e: FormEvent) {
    e.preventDefault();
    if (await buddy.saveProfile(fullName, position)) setEditing(false);
  }
  return (
    <div className="page-content profile-page">
      <div className="page-heading">
        <span className="eyebrow">МОЁ ПРОСТРАНСТВО</span>
        <h1>Профиль и адаптация</h1>
        <p>Всё, что помогает освоиться в новой команде.</p>
      </div>
      <section className="profile-hero">
        <span className={"avatar avatar-xl " + color}>
          {initials(d.profile.fullName)}
        </span>
        <div className="profile-identity">
          <span className={"team-label " + color}>{team.name}</span>
          <h2>{d.profile.fullName}</h2>
          <p>{d.profile.position}</p>
          <div className="profile-meta">
            <span>
              <Icon name="mail" size={14} />
              {d.account.email}
            </span>
            <span>
              <Icon name="calendar" size={14} />
              Начало: {dateLabel(d.profile.startDate)}
            </span>
          </div>
        </div>
        <button
          className="secondary-button"
          onClick={() => {
            setFullName(d.profile.fullName);
            setPosition(d.profile.position);
            setEditing(!editing);
          }}
        >
          <Icon name="edit" size={16} />
          Изменить профиль
        </button>
      </section>
      {editing && (
        <form className="profile-editor" onSubmit={save}>
          <label>
            Имя и фамилия
            <input
              value={fullName}
              required
              maxLength={60}
              onChange={(e) => setFullName(e.target.value)}
            />
          </label>
          <label>
            Должность
            <input
              value={position}
              required
              maxLength={60}
              onChange={(e) => setPosition(e.target.value)}
            />
          </label>
          <button className="primary-button" disabled={buddy.savingProfile}>
            Сохранить
          </button>
          <button
            type="button"
            className="text-button"
            onClick={() => setEditing(false)}
          >
            Отмена
          </button>
        </form>
      )}
      <SecuritySection email={d.account.email} onNotice={buddy.setNotice} />
      <div className="profile-grid">
        <div className="profile-main">
          <section className="progress-summary">
            <div>
              <span className="section-kicker">ТВОЙ ПУТЬ В КОМАНДЕ</span>
              <h2>Каждый шаг считается</h2>
              <p>
                {d.checklist.completed} из {d.checklist.total} шагов уже позади
              </p>
            </div>
            <div
              className="progress-ring"
              role="progressbar"
              aria-label="Прогресс адаптации"
              aria-valuenow={d.checklist.percent}
              aria-valuemin={0}
              aria-valuemax={100}
              style={{
                background: `conic-gradient(var(--blue) ${d.checklist.percent}%, #e7f4fc 0)`,
              }}
            >
              <span>
                {d.checklist.percent}
                <small>%</small>
              </span>
            </div>
          </section>
          <section className="checklist-section">
            <div className="section-heading">
              <h2>Чек-лист адаптации</h2>
              <span className="count-label">{d.checklist.total} шагов</span>
            </div>
            <div className="segmented" aria-label="Фильтр задач">
              {[
                ["all", "Все шаги"],
                ["pending", "Осталось"],
                ["done", "Готово"],
              ].map(([id, title]) => (
                <button
                  key={id}
                  className={filter === id ? "selected" : ""}
                  onClick={() => setFilter(id)}
                >
                  {title}
                </button>
              ))}
            </div>
            {periods.length === 0 ? (
              <div className="empty-panel">
                <Icon name="check" size={28} />
                <h3>
                  {filter === "done"
                    ? "Твои победы ещё впереди"
                    : "Все шаги пройдены"}
                </h3>
                <p>Отмечай выполненные задачи — прогресс сохранится.</p>
              </div>
            ) : (
              periods.map((period) => (
                <div className="task-group" key={period}>
                  <h3>
                    <span /> {period}
                  </h3>
                  {filtered
                    .filter((t) => t.period === period)
                    .map((t) => (
                      <div
                        className={
                          "task-row " +
                          (t.status === "completed" ? "is-complete" : "")
                        }
                        key={t.id}
                      >
                        <label className="task-check">
                          <input
                            type="checkbox"
                            aria-label={t.title}
                            checked={t.status === "completed"}
                            disabled={buddy.updatingTask}
                            onChange={() =>
                              void buddy.updateTask(
                                t.id,
                                t.status === "completed"
                                  ? "pending"
                                  : "completed",
                              )
                            }
                          />
                          <span>
                            <Icon name="check" size={13} />
                          </span>
                        </label>
                        <div className="task-copy">
                          <strong>{t.title}</strong>
                          <p>{t.description}</p>
                          <div className="task-controls">
                            {t.documentId &&
                              d.documents.some(
                                (doc) => doc.id === t.documentId,
                              ) && (
                                <button
                                  onClick={() =>
                                    onDocument(
                                      d.documents.find(
                                        (doc) => doc.id === t.documentId,
                                      )!,
                                    )
                                  }
                                >
                                  <Icon name="book" size={12} />
                                  Инструкция
                                </button>
                              )}
                            {t.status !== "completed" && (
                              <button
                                className={
                                  t.status === "in_progress"
                                    ? "in-progress"
                                    : ""
                                }
                                disabled={buddy.updatingTask}
                                onClick={() =>
                                  void buddy.updateTask(
                                    t.id,
                                    t.status === "in_progress"
                                      ? "pending"
                                      : "in_progress",
                                  )
                                }
                              >
                                <Icon name="clock" size={12} />
                                {t.status === "in_progress"
                                  ? "В работе"
                                  : "Взять в работу"}
                              </button>
                            )}
                            {t.completedAt && (
                              <span className="completed-date">
                                Готово {dateLabel(t.completedAt)}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                </div>
              ))
            )}
          </section>
          <section className="requests-section">
            <div className="section-heading">
              <h2>Мои заявки</h2>
              <span className="count-label">{d.requests.length}</span>
            </div>
            {d.requests.length === 0 ? (
              <div className="requests-empty">
                <Icon name="shield" size={23} />
                <div>
                  <strong>Пока без заявок</strong>
                  <p>Попроси Buddy помочь с доступом к нужному инструменту.</p>
                </div>
                <button
                  className="text-button"
                  onClick={() => {
                    buddy.navigate("chat");
                    buddy.setQuestion("Как получить нужные доступы?");
                  }}
                >
                  Спросить Buddy <Icon name="arrow" size={14} />
                </button>
              </div>
            ) : (
              d.requests.map((r) => (
                <div className="request-item" key={r.id}>
                  <span className="prompt-icon blue">
                    <Icon name="shield" size={17} />
                  </span>
                  <div>
                    <strong>
                      {d.tools.find((t) => t.id === r.system)?.name || r.system}
                    </strong>
                    <small>
                      {r.externalId} · {dateLabel(r.createdAt)}
                    </small>
                  </div>
                  <div className="request-status">
                    <span>
                      <i />
                      Отправлена
                    </span>
                    <small>Ожидает согласования</small>
                  </div>
                </div>
              ))
            )}
          </section>
        </div>
        <aside className="profile-aside">
          <section className="buddy-panel">
            <span className="section-kicker">ТВОЙ БАДДИ</span>
            <div className={"avatar avatar-lg " + color}>
              {initials(team.buddy)}
            </div>
            <h3>{team.buddy}</h3>
            <p>{buddyPerson?.position || "Наставник команды"}</p>
            <div className="buddy-note">
              Поможет с первыми задачами, доступами и вопросами, которые лучше
              обсудить с человеком.
            </div>
            <button
              className="secondary-button"
              onClick={() => {
                buddy.navigate("chat");
                buddy.setQuestion("Как подготовиться к встрече с моим бадди?");
              }}
            >
              Подготовиться к встрече
              <Icon name="arrow" size={15} />
            </button>
          </section>
          <section className="next-step-panel">
            <span className="prompt-icon green">
              <Icon name="target" size={21} />
            </span>
            <span className="section-kicker">СЛЕДУЮЩИЙ ШАГ</span>
            <h3>{d.checklist.nextTask?.title || "Ты прошёл все этапы!"}</h3>
            <p>
              {d.checklist.nextTask?.description ||
                "Обсуди с бадди, какие цели взять дальше."}
            </p>
            <button
              className="text-button"
              onClick={() => {
                buddy.navigate("chat");
                buddy.setQuestion("Подскажи мой следующий шаг адаптации");
              }}
            >
              Разобраться с Buddy
              <Icon name="arrow" size={14} />
            </button>
          </section>
          <section className="mini-team-panel">
            <div className="section-heading">
              <h3>Твоя команда</h3>
              <span>{team.members.length}</span>
            </div>
            {team.members
              .filter((m) => m.id !== d.profile.userId)
              .slice(0, 5)
              .map((m) => (
                <div className="mini-member" key={m.id}>
                  <span className={"avatar avatar-sm " + color}>
                    {initials(m.fullName)}
                  </span>
                  <span>
                    <strong>{m.fullName}</strong>
                    <small>{roleLabel[m.role] || m.position}</small>
                  </span>
                </div>
              ))}
            <button
              className="text-button"
              onClick={() => buddy.navigate("teams")}
            >
              Все участники и сервисы
              <Icon name="arrow" size={14} />
            </button>
          </section>
        </aside>
      </div>
    </div>
  );
}

function SecuritySection({
  email,
  onNotice,
}: {
  email: string;
  onNotice: (text: string) => void;
}) {
  const [open, setOpen] = useState(false),
    [stage, setStage] = useState<"ask" | "code">("ask"),
    [code, setCode] = useState(""),
    [password, setPassword] = useState(""),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  async function request(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await api.request("/auth/password/reset", "POST", { email });
      setStage("code");
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  async function confirm(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await api.request("/auth/password/confirm", "POST", {
        code,
        newPassword: password,
      });
      setOpen(false);
      setStage("ask");
      setCode("");
      setPassword("");
      onNotice("Пароль обновлён");
    } catch (err) {
      setError(errorText(err));
    } finally {
      setBusy(false);
    }
  }
  if (!open)
    return (
      <button
        className="secondary-button security-toggle"
        onClick={() => {
          setOpen(true);
          setStage("ask");
        }}
      >
        <Icon name="shield" size={16} />
        Безопасность: сменить пароль
      </button>
    );
  return (
    <section className="security-section">
      <h2>Безопасность</h2>
      <p>Подтверди смену пароля кодом из письма на {email}.</p>
      {stage === "ask" ? (
        <form className="security-form" onSubmit={request}>
          <button className="primary-button" disabled={busy}>
            {busy ? "Отправляем…" : "Получить код"}
          </button>
          <button
            type="button"
            className="text-button"
            disabled={busy}
            onClick={() => setOpen(false)}
          >
            Отмена
          </button>
        </form>
      ) : (
        <form className="security-form" onSubmit={confirm}>
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
          <button className="primary-button" disabled={busy}>
            {busy ? "Меняем пароль…" : "Сменить пароль"}
          </button>
          <button
            type="button"
            className="text-button"
            disabled={busy}
            onClick={() => setOpen(false)}
          >
            Отмена
          </button>
        </form>
      )}
      {error && (
        <p className="inline-error" role="alert">
          <Icon name="shield" size={16} />
          {error}
        </p>
      )}
    </section>
  );
}
