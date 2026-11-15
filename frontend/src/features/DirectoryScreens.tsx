import { useState } from "react";
import type { Document } from "../api";
import type { Workspace } from "../model";
import { initials, roleLabel, teamColors, safeExternalUrl } from "../model";
import { Icon } from "../components/Icon";
export function TeamsScreen({
  data,
  selected,
  onSelect,
  onServices,
}: {
  data: Workspace;
  selected: string;
  onSelect: (id: string) => void;
  onServices: () => void;
}) {
  const team =
    data.teamDetails[selected] || data.teamDetails[data.profile.teamId];
  return (
    <div className="page-content directory-page">
      <div className="page-heading">
        <span className="eyebrow">ЛЮДИ РЯДОМ</span>
        <h1>Знакомься с командами</h1>
        <p>
          Кто чем занимается, к кому обратиться и за какие сервисы отвечает.
        </p>
      </div>
      <div className="team-grid">
        {data.teams.map((t) => (
          <button
            key={t.id}
            className={"team-card " + (team.id === t.id ? "selected" : "")}
            onClick={() => onSelect(t.id)}
          >
            <span className={"team-card-symbol " + teamColors[t.id]}>
              <Icon name="users" size={22} />
            </span>
            <span className="team-card-heading">
              <strong>{t.name}</strong>
              {t.id === data.profile.teamId && <small>Твоя команда</small>}
            </span>
            <p>{t.description}</p>
            <span className="team-card-footer">
              Бадди: {t.buddy}
              <Icon name="arrow" size={16} />
            </span>
          </button>
        ))}
      </div>
      <section className="team-detail">
        <div className="section-heading">
          <div>
            <span className="section-kicker">
              КОМАНДА {team.name.toUpperCase()}
            </span>
            <h2>У каждого есть свой человек</h2>
          </div>
          <span className="count-label">{team.members.length} участников</span>
        </div>
        <div className="member-grid">
          {team.members.map((m) => (
            <div className="member-card" key={m.id}>
              <span className={"avatar " + teamColors[team.id]}>
                {initials(m.fullName)}
              </span>
              <div>
                <strong>
                  {m.fullName}
                  {m.id === data.profile.userId && (
                    <small className="you-tag">ты</small>
                  )}
                </strong>
                <p>{m.position}</p>
                <span
                  className={
                    "member-role " + (m.role === "buddy" ? "is-buddy" : "")
                  }
                >
                  {roleLabel[m.role] || m.role}
                </span>
              </div>
            </div>
          ))}
        </div>
        <div className="section-heading">
          <h3>За что отвечает команда</h3>
          <button className="text-button" onClick={onServices}>
            Весь каталог
            <Icon name="arrow" size={14} />
          </button>
        </div>
        <div className="team-services">
          {team.services.map((s) => (
            <div key={s.id}>
              <Icon name="cube" size={19} />
              <span>
                <strong>{s.name}</strong>
                <p>{s.description}</p>
                <small>
                  Ответственный:{" "}
                  {team.members.find((m) => m.id === s.technicalOwnerId)
                    ?.fullName || s.technicalOwnerId}
                </small>
              </span>
              <span className={"lifecycle " + s.lifecycle}>
                {s.lifecycle === "active" ? "Работает" : "В разработке"}
              </span>
            </div>
          ))}
        </div>
      </section>
    </div>
  );
}
export function ServicesScreen({
  data,
  onTeam,
}: {
  data: Workspace;
  onTeam: (id: string) => void;
}) {
  const [query, setQuery] = useState(""),
    [teamId, setTeamId] = useState("all");
  const list = data.services.filter(
    (s) =>
      (teamId === "all" || s.ownerTeamId === teamId) &&
      (s.name + " " + s.description)
        .toLowerCase()
        .includes(query.toLowerCase()),
  );
  return (
    <div className="page-content">
      <div className="page-heading">
        <span className="eyebrow">ЗОНЫ ОТВЕТСТВЕННОСТИ</span>
        <h1>Кто отвечает за сервис</h1>
        <p>
          Технические сервисы, их команды-владельцы и конкретные ответственные.
        </p>
      </div>
      <div className="catalog-toolbar">
        <label className="search-field">
          <Icon name="search" size={18} />
          <input
            aria-label="Поиск сервиса"
            placeholder="Название или назначение сервиса"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </label>
        <select
          aria-label="Команда-владелец"
          value={teamId}
          onChange={(e) => setTeamId(e.target.value)}
        >
          <option value="all">Все команды</option>
          {data.teams.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name}
            </option>
          ))}
        </select>
        <span className="count-label">Найдено: {list.length}</span>
      </div>
      {list.length === 0 ? (
        <div className="empty-panel">
          <Icon name="search" size={30} />
          <h3>Такого сервиса не нашли</h3>
          <p>Попробуй другое название или выбери все команды.</p>
        </div>
      ) : (
        <div className="service-catalog">
          <div className="service-table-heading">
            <span>Сервис</span>
            <span>Команда</span>
            <span>Ответственный</span>
            <span>Статус</span>
          </div>
          {list.map((s) => {
            const team = data.teamDetails[s.ownerTeamId],
              person = team?.members.find((m) => m.id === s.technicalOwnerId);
            return (
              <article className="service-catalog-row" key={s.id}>
                <div className="service-name">
                  <span
                    className={"service-symbol " + teamColors[s.ownerTeamId]}
                  >
                    <Icon name="cube" size={19} />
                  </span>
                  <div>
                    <strong>{s.name}</strong>
                    <p>{s.description}</p>
                    {s.repositoryUrl && (
                      <a
                        href={safeExternalUrl(s.repositoryUrl)}
                        target="_blank"
                        rel="noreferrer"
                      >
                        Репозиторий
                        <Icon name="external" size={11} />
                      </a>
                    )}
                  </div>
                </div>
                <button
                  className={"team-label " + teamColors[s.ownerTeamId]}
                  onClick={() => onTeam(s.ownerTeamId)}
                >
                  {team?.name || s.ownerTeamId}
                </button>
                <span className="service-owner">
                  <strong>{person?.fullName || s.technicalOwnerId}</strong>
                  <small>{roleLabel[person?.role || "engineer"]}</small>
                </span>
                <span className={"lifecycle " + s.lifecycle}>
                  {s.lifecycle === "active" ? "Работает" : "В разработке"}
                </span>
              </article>
            );
          })}
        </div>
      )}
    </div>
  );
}
const categoryNames: Record<string, string> = {
  onboarding: "Адаптация",
  company: "Компания",
  access: "Доступы",
  engineering: "Разработка",
  process: "Процессы",
  learning: "Обучение",
  security: "Безопасность",
  analytics: "Аналитика",
  quality: "Тестирование",
};
export function KnowledgeScreen({
  data,
  onDocument,
}: {
  data: Workspace;
  onDocument: (doc: Document) => void;
}) {
  const [query, setQuery] = useState(""),
    [filter, setFilter] = useState("all");
  const list = data.documents.filter(
    (d) =>
      (filter === "all" || d.visibility === filter) &&
      (d.title + " " + d.content).toLowerCase().includes(query.toLowerCase()),
  );
  return (
    <div className="page-content">
      <div className="page-heading">
        <span className="eyebrow">ЗНАНИЯ В НУЖНЫЙ МОМЕНТ</span>
        <h1>Всё начинается с вопроса</h1>
        <p>
          Общие материалы и инструкции твоей команды. Сохраняй время на главное.
        </p>
      </div>
      <div className="knowledge-hero">
        <span className="knowledge-hero-icon">
          <Icon name="book" size={30} />
        </span>
        <div>
          <strong>
            Материалы для{" "}
            {data.teams.find((t) => t.id === data.profile.teamId)?.name}
          </strong>
          <p>Это тестовые инструкции, а не внутренние регламенты Авито.</p>
        </div>
        <span>
          {data.documents.length}
          <small>материалов</small>
        </span>
      </div>
      <div className="catalog-toolbar">
        <label className="search-field">
          <Icon name="search" size={18} />
          <input
            aria-label="Поиск по базе знаний"
            placeholder="Например, доступы или первая задача"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
        </label>
        <div className="segmented">
          {[
            ["all", "Все"],
            ["company", "Общие"],
            ["team", "Моя команда"],
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
      </div>
      {list.length === 0 ? (
        <div className="empty-panel">
          <Icon name="search" size={30} />
          <h3>По этому запросу ничего нет</h3>
          <p>Попробуй другие слова или спроси у своего бадди.</p>
        </div>
      ) : (
        <div className="document-grid">
          {list.map((doc) => (
            <button
              className="document-card"
              key={doc.id}
              onClick={() => onDocument(doc)}
            >
              <div>
                <span
                  className={
                    "document-icon " +
                    (doc.visibility === "team" ? "purple" : "blue")
                  }
                >
                  <Icon name="book" size={19} />
                </span>
                <span className="document-category">
                  {categoryNames[doc.category] || doc.category}
                </span>
                <Icon name="arrow" size={16} />
              </div>
              <h3>{doc.title}</h3>
              <p>
                {doc.content
                  .replace(
                    "ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.",
                    "",
                  )
                  .trim()}
              </p>
              <span className="document-audience">
                {doc.visibility === "team"
                  ? "Для твоей команды"
                  : "Для всех стажёров"}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
