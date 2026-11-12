import { useState } from "react";
import type { Buddy } from "../hooks/useBuddy";
import type { Chat } from "../api";
import { initials, teamColors, teamNames } from "../model";
import { Brand, Icon } from "./Icon";
export function Sidebar({
  buddy,
  open,
  onClose,
  onChatAction,
}: {
  buddy: Buddy;
  open: boolean;
  onClose: () => void;
  onChatAction: (kind: "rename" | "delete", chat: Chat) => void;
}) {
  const [search, setSearch] = useState(""),
    [menu, setMenu] = useState<string | null>(null);
  const d = buddy.data!;
  const chats = d.chats.filter((c) =>
    c.title.toLowerCase().includes(search.toLowerCase()),
  );
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
              buddy.newChat();
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
        <button
          className="new-chat-button"
          onClick={() => {
            buddy.newChat();
            onClose();
          }}
          disabled={buddy.sending || buddy.generating}
        >
          <Icon name="plus" size={18} />
          Новый чат
        </button>
        <div className="history-search">
          <Icon name="search" size={16} />
          <input
            aria-label="Поиск по чатам"
            placeholder="Поиск по чатам"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="history-heading">
          Твои разговоры<span>{d.chats.length}</span>
        </div>
        <div className="history-list">
          {chats.length === 0 ? (
            <div className="history-empty">
              <Icon name="chat" size={23} />
              <p>
                {search ? "Ничего не найдено" : "Здесь появятся твои вопросы"}
              </p>
            </div>
          ) : (
            chats.map((c) => (
              <div
                key={c.id}
                className={
                  "history-item " +
                  (buddy.view === "chat" && buddy.chatId === c.id
                    ? "selected"
                    : "")
                }
              >
                <button
                  className="history-chat"
                  onClick={() => {
                    void buddy.selectChat(c.id);
                    onClose();
                  }}
                  disabled={buddy.generating}
                >
                  <Icon name="chat" size={16} />
                  <span>{c.title}</span>
                </button>
                <button
                  className="history-more icon-button"
                  aria-label={"Действия с чатом " + c.title}
                  onClick={() => setMenu(menu === c.id ? null : c.id)}
                  disabled={buddy.generating || buddy.sending}
                >
                  <Icon name="more" size={16} />
                </button>
                {menu === c.id && (
                  <div className="chat-menu">
                    <button
                      onClick={() => {
                        setMenu(null);
                        onChatAction("rename", c);
                      }}
                    >
                      <Icon name="edit" size={16} />
                      Переименовать
                    </button>
                    <button
                      className="danger-text"
                      onClick={() => {
                        setMenu(null);
                        onChatAction("delete", c);
                      }}
                    >
                      <Icon name="trash" size={16} />
                      Удалить
                    </button>
                  </div>
                )}
              </div>
            ))
          )}
        </div>
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
            <button
              className={buddy.view === "notes" ? "active" : ""}
              onClick={() => navigate("notes")}
            >
              <Icon name="edit" size={19} />
              Заметки
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
