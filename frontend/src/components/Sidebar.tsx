import type { Buddy } from "../hooks/useBuddy";
import { initials, teamColors, teamNames } from "../model";
import { Brand, Icon } from "./Icon";
export function Sidebar({
  buddy,
  open,
  onClose,
}: {
  buddy: Buddy;
  open: boolean;
  onClose: () => void;
}) {
  const d = buddy.data!;
  function navigate(view: typeof buddy.view) {
    buddy.navigate(view);
    onClose();
  }
  return (
    <>
      <button
        className={"mobile-shade " + (open ? "visible" : "")}
        aria-label="Закрыть навигацию"
        onClick={onClose}
        tabIndex={open ? 0 : -1}
      />
      <aside className={"sidebar " + (open ? "sidebar-open" : "")}>
        <div className="sidebar-brand">
          <button
            className="brand-button"
            onClick={() => {
              buddy.navigate("home");
              onClose();
            }}
            aria-label="На главную"
          >
            <Brand small />
          </button>
          <button
            className="icon-button mobile-only"
            onClick={onClose}
            aria-label="Скрыть боковую панель"
          >
            <Icon name="close" />
          </button>
        </div>
        <button
          className={
            "profile-launcher " + (buddy.view === "profile" ? "active" : "")
          }
          aria-label="Открыть профиль"
          onClick={() => navigate("profile")}
        >
          <span className={"avatar " + teamColors[d.profile.teamId]}>
            {initials(d.profile.fullName)}
          </span>
          <span className="profile-launcher-copy">
            <strong>{d.profile.fullName}</strong>
            <span>
              {teamNames[d.profile.teamId] || d.profile.teamId} · стажёр
            </span>
            <span className="mini-progress">
              <i style={{ width: d.checklist.percent + "%" }} />
            </span>
          </span>
          <Icon name="chevron" size={16} />
        </button>
        <nav className="primary-navigation" aria-label="Разделы пространства">
          {(
            [
              ["home", "Главная", "home"],
              ["chat", "AI-buddy", "chat"],
              ["notes", "Заметки", "edit"],
              ["checklist", "Чек-лист", "check"],
              ["profile", "Профиль", "user"],
            ] as const
          ).map(([view, title, icon]) => (
            <button
              key={view}
              className={buddy.view === view ? "active" : ""}
              aria-current={buddy.view === view ? "page" : undefined}
              onClick={() => navigate(view)}
            >
              <Icon name={icon} size={20} />
              <span>{title}</span>
            </button>
          ))}
        </nav>
        <div className="sidebar-bottom">
          <nav aria-label="Справочники">
            <button
              className={buddy.view === "knowledge" ? "active" : ""}
              onClick={() => navigate("knowledge")}
            >
              <Icon name="book" size={19} />
              База знаний
              <Icon name="chevron" size={14} />
            </button>
            <button
              className={buddy.view === "teams" ? "active" : ""}
              onClick={() => navigate("teams")}
            >
              <Icon name="users" size={19} />
              Команды
              <Icon name="chevron" size={14} />
            </button>
            <button
              className={buddy.view === "services" ? "active" : ""}
              onClick={() => navigate("services")}
            >
              <Icon name="cube" size={19} />
              Сервисы
              <Icon name="chevron" size={14} />
            </button>
          </nav>
          <div className="sidebar-footer">
            <span>
              <i /> Тестовое пространство
            </span>
            <button
              className="icon-button"
              aria-label="Выйти из аккаунта"
              onClick={() => void buddy.logout()}
              disabled={buddy.authBusy || buddy.generating || buddy.sending}
            >
              <Icon name="logout" size={18} />
            </button>
          </div>
        </div>
      </aside>
    </>
  );
}
