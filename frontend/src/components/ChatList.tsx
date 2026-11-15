import { useState } from "react";
import type { Buddy } from "../hooks/useBuddy";
import type { Chat } from "../api";
import { Icon } from "./Icon";

export function ChatList({
  buddy,
  onChatAction,
  onOpen,
}: {
  buddy: Buddy;
  onChatAction: (kind: "rename" | "delete", chat: Chat) => void;
  onOpen: () => void;
}) {
  const [search, setSearch] = useState(""),
    [menu, setMenu] = useState<string | null>(null);
  const d = buddy.data!;
  const chats = d.chats.filter((c) =>
    c.title.toLowerCase().includes(search.toLowerCase()),
  );
  return (
    <aside className="conversation-list" aria-label="Список чатов">
      <h2>Чаты</h2>
      <button
        className="new-chat-button"
        onClick={() => {
          buddy.newChat();
          onOpen();
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
                  onOpen();
                }}
                disabled={buddy.generating || buddy.sending}
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
    </aside>
  );
}
