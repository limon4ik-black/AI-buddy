import { useState, type FormEvent } from "react";
import { Icon } from "../components/Icon";
import { useNotes } from "../hooks/useNotes";

const stamp = (iso: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(iso));

export function NotesScreen() {
  const store = useNotes(),
    [draft, setDraft] = useState(""),
    [editing, setEditing] = useState<{ id: string; content: string } | null>(
      null,
    );
  async function add(e: FormEvent) {
    e.preventDefault();
    const content = draft.trim();
    if (!content) return;
    if (await store.create(content)) setDraft("");
  }
  async function saveEdit(e: FormEvent) {
    e.preventDefault();
    if (!editing) return;
    const note = store.items?.find((n) => n.id === editing.id);
    if (!note || !editing.content.trim()) return;
    if (await store.save(note, editing.content.trim())) setEditing(null);
  }
  return (
    <div className="page-content">
      <div className="page-heading">
        <span className="eyebrow">ЛИЧНЫЙ БЛОКНОТ</span>
        <h1>Мои заметки</h1>
        <p>
          Мысли, ссылки и договорённости, которые хочется сохранить под рукой.
          Эти заметки видишь только ты.
        </p>
      </div>
      <form className="notes-composer" onSubmit={add}>
        <label className="notes-field">
          <span className="quiet-label">Новая заметка</span>
          <textarea
            aria-label="Текст новой заметки"
            placeholder="Что стоит запомнить?"
            rows={3}
            maxLength={20000}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
          />
        </label>
        <button
          className="primary-button"
          type="submit"
          disabled={store.busy || !draft.trim()}
        >
          <Icon name="plus" size={17} />
          Добавить заметку
        </button>
      </form>
      {store.error && (
        <div className="notes-error">
          <p className="inline-error" role="alert">
            {store.error}
          </p>
          <button
            className="text-button"
            disabled={store.busy}
            onClick={() => void store.reload()}
          >
            Повторить загрузку
          </button>
        </div>
      )}
      {store.items === null ? (
        <div className="empty-panel">
          <span className="spinner" />
          <p role="status">Загружаем заметки…</p>
        </div>
      ) : store.items.length === 0 ? (
        <div className="empty-panel">
          <Icon name="book" size={30} />
          <h3>Пока пусто</h3>
          <p>Запиши первое, что не хочется забыть: мысль, ссылку, задачу.</p>
        </div>
      ) : (
        <ul className="notes-list">
          {store.items.map((note) => (
            <li className="note-card" key={note.id}>
              {editing?.id === note.id ? (
                <form className="note-editor" onSubmit={saveEdit}>
                  <textarea
                    aria-label={"Текст заметки " + note.id}
                    rows={3}
                    maxLength={20000}
                    value={editing.content}
                    onChange={(e) =>
                      setEditing({ id: note.id, content: e.target.value })
                    }
                  />
                  <div className="note-actions">
                    <button
                      className="primary-button"
                      type="submit"
                      disabled={store.busy || !editing.content.trim()}
                    >
                      Сохранить
                    </button>
                    <button
                      className="secondary-button"
                      type="button"
                      onClick={() => setEditing(null)}
                    >
                      Отмена
                    </button>
                  </div>
                </form>
              ) : (
                <>
                  <p className="note-text">{note.content}</p>
                  <div className="note-meta">
                    <span>{stamp(note.updatedAt)}</span>
                    <span className="status-chip">версия {note.version}</span>
                    <div className="note-actions">
                      <button
                        className="text-button"
                        onClick={() =>
                          setEditing({ id: note.id, content: note.content })
                        }
                        disabled={store.busy}
                      >
                        <Icon name="edit" size={16} />
                        Изменить
                      </button>
                      <button
                        className="text-button danger-text"
                        onClick={() => void store.remove(note.id)}
                        disabled={store.busy}
                      >
                        <Icon name="trash" size={16} />
                        Удалить
                      </button>
                    </div>
                  </div>
                </>
              )}
            </li>
          ))}
        </ul>
      )}
      {store.items && store.items.length > 0 && (
        <div className="notes-footer">
          <span className="count-label">
            Всего заметок: {store.items.length}
          </span>
          <button className="text-button" onClick={() => void store.reload()}>
            <Icon name="refresh" size={16} />
            Обновить
          </button>
        </div>
      )}
    </div>
  );
}
