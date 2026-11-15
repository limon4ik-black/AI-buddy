import { useEffect, useState, type FormEvent } from "react";
import { Icon } from "../components/Icon";
import { useNotes } from "../hooks/useNotes";

const stamp = (iso: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(iso));

export function NotesScreen({
  onBoards,
  initialNoteId,
}: {
  onBoards: () => void;
  initialNoteId?: string | null;
}) {
  const store = useNotes();
  const [selected, setSelected] = useState<string | "new" | null>(
    initialNoteId || null,
  );
  const [content, setContent] = useState("");
  const [search, setSearch] = useState("");
  const note = store.items?.find((n) => n.id === selected);
  const dirty = content !== (note?.content || "");
  useEffect(() => {
    // Load a note opened from Home once the server list arrives.
    if (initialNoteId && note && !content) setContent(note.content);
  }, [initialNoteId, note]);
  useEffect(() => {
    if (!dirty) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [dirty]);
  function select(id: string | "new" | null) {
    if (
      dirty &&
      !window.confirm("Изменения не сохранены. Закрыть заметку без сохранения?")
    )
      return;
    setSelected(id);
    setContent(store.items?.find((n) => n.id === id)?.content || "");
  }
  async function save(e: FormEvent) {
    e.preventDefault();
    if (!content.trim()) return;
    if (note) await store.save(note, content);
    else if (await store.create(content)) {
      setSelected(null);
      setContent("");
    }
  }
  async function remove() {
    if (note && (await store.remove(note.id))) {
      setSelected(null);
      setContent("");
    }
  }
  const filtered = store.items?.filter((n) =>
    n.content.toLocaleLowerCase().includes(search.toLocaleLowerCase()),
  );
  return (
    <div className={"notebook " + (selected ? "note-open" : "note-closed")}>
      <aside className="notebook-list" aria-label="Список заметок">
        <div className="notebook-heading">
          <h1>Мои заметки</h1>
          <button
            className="icon-button"
            aria-label="Обновить заметки"
            disabled={store.busy || dirty}
            onClick={() => void store.reload()}
          >
            <Icon name="refresh" />
          </button>
        </div>
        <div className="notebook-board">
          <div>
            <strong>Доска</strong>
            <p>Стикеры, схемы и идеи</p>
          </div>
          <button className="secondary-button" onClick={onBoards}>
            Открыть доски
          </button>
        </div>
        <div className="notebook-search">
          <input
            aria-label="Поиск по заметкам"
            placeholder="Поиск по заметкам"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <button
            className="primary-button"
            aria-label="Новая заметка"
            disabled={store.busy}
            onClick={() => select("new")}
          >
            <Icon name="plus" />
          </button>
        </div>
        {store.error && (
          <div>
            <p className="inline-error" role="alert">
              {store.error}
            </p>
            <button
              className="text-button"
              disabled={store.busy || dirty}
              onClick={() => void store.reload()}
            >
              Повторить загрузку
            </button>
          </div>
        )}
        {store.items === null ? (
          <p role="status">Загружаем заметки…</p>
        ) : !store.items.length ? (
          <div className="notebook-empty">
            <Icon name="book" size={32} />
            <h2>Пока пусто</h2>
            <p>Сохрани первую мысль, ссылку или договорённость.</p>
          </div>
        ) : !filtered?.length ? (
          <p className="dashboard-help">Ничего не найдено</p>
        ) : (
          <ul className="notebook-cards">
            {filtered.map((n, i) => (
              <li key={n.id}>
                <button
                  disabled={store.busy}
                  className={
                    "notebook-card note-color-" +
                    (i % 3) +
                    (selected === n.id ? " selected" : "")
                  }
                  onClick={() => select(n.id)}
                >
                  <strong>{n.content.split("\n")[0] || "Без названия"}</strong>
                  <span>
                    {n.content.split("\n").slice(1).join(" ") ||
                      "Открыть заметку"}
                  </span>
                  <small>{stamp(n.updatedAt)}</small>
                </button>
              </li>
            ))}
          </ul>
        )}
        {store.items && (
          <span className="count-label">
            Всего заметок: {store.items.length}
          </span>
        )}
      </aside>
      <section className="notebook-editor" aria-label="Редактор заметки">
        {selected ? (
          <form onSubmit={save}>
            <div className="notebook-toolbar">
              <button
                type="button"
                className="icon-button"
                aria-label="К списку заметок"
                disabled={store.busy}
                onClick={() => select(null)}
              >
                <Icon name="back" />
              </button>
              <h2>{note ? "Заметка" : "Новая заметка"}</h2>
              {note && (
                <button
                  type="button"
                  className="text-button danger-text"
                  disabled={store.busy}
                  onClick={() => void remove()}
                >
                  <Icon name="trash" size={16} />
                  Удалить
                </button>
              )}
            </div>
            <p className="dashboard-help">
              Первая строка — заголовок. Остальное — твои мысли и
              договорённости.
            </p>
            <textarea
              aria-label={
                note ? "Текст заметки " + note.id : "Текст новой заметки"
              }
              placeholder="Заголовок заметки&#10;Начни писать…"
              value={content}
              maxLength={20000}
              onChange={(e) => setContent(e.target.value)}
            />
            <div className="notebook-save">
              <span role="status">
                {store.busy
                  ? "Сохраняем…"
                  : dirty
                    ? "Есть несохранённые изменения"
                    : note
                      ? "Все изменения сохранены · версия " + note.version
                      : "Личная заметка"}
              </span>
              <button
                className="primary-button"
                disabled={store.busy || !content.trim() || (!dirty && !!note)}
              >
                {note ? "Сохранить" : "Добавить заметку"}
              </button>
            </div>
          </form>
        ) : (
          <div className="notebook-placeholder">
            <span aria-hidden="true">📝</span>
            <h2>Выбери заметку или создай новую</h2>
            <p>Мысли, ссылки и договорённости в одном месте.</p>
            <button
              className="primary-button"
              disabled={store.busy}
              onClick={() => select("new")}
            >
              Новая заметка
            </button>
          </div>
        )}
      </section>
    </div>
  );
}
