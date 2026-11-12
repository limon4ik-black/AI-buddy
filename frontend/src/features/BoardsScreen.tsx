import { lazy, Suspense, useState } from "react";
import { Icon } from "../components/Icon";
import { boards, type BoardFull } from "../boards";
import { useBoards } from "../hooks/useBoards";
import { errorText } from "../hooks/useBuddy";

// Холст грузим отдельным куском: Excalidraw весит больше мегабайта.
const BoardEditor = lazy(() =>
  import("./BoardEditor").then((m) => ({ default: m.BoardEditor })),
);

const stamp = (iso: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "long",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(iso));

export function BoardsScreen() {
  const store = useBoards(),
    [open, setOpen] = useState<BoardFull | null>(null),
    [loading, setLoading] = useState(""),
    [error, setError] = useState("");

  async function openBoard(id: string) {
    setLoading(id);
    setError("");
    try {
      setOpen(await boards.get(id));
    } catch (e) {
      setError(errorText(e));
    } finally {
      setLoading("");
    }
  }

  async function addBoard() {
    const created = await store.create("Новая доска");
    if (created) setOpen(created);
  }

  if (open)
    return (
      <div className="page-content">
        <Suspense
          fallback={
            <div className="empty-panel">
              <span className="spinner" />
              <p role="status">Загружаем холст…</p>
            </div>
          }
        >
          <BoardEditor
            board={open}
            onSaved={store.applySaved}
            onClose={() => setOpen(null)}
          />
        </Suspense>
      </div>
    );

  return (
    <div className="page-content">
      <div className="page-heading">
        <span className="eyebrow">ХОЛСТЫ КОМАНДЫ</span>
        <h1>Доски</h1>
        <p>
          Рисуй схемы и раскладывай стикеры. Доска сохраняется сама, а видишь её
          только ты — владелец берётся из токена.
        </p>
      </div>
      <div className="catalog-toolbar">
        <button className="primary-button" onClick={() => void addBoard()} disabled={store.busy}>
          <Icon name="plus" size={17} />
          Создать доску
        </button>
        <button className="text-button" onClick={() => void store.reload()}>
          <Icon name="refresh" size={16} />
          Обновить
        </button>
      </div>
      {store.error && (
        <p className="inline-error" role="alert">
          {store.error}
        </p>
      )}
      {error && (
        <p className="inline-error" role="alert">
          {error}
        </p>
      )}
      {store.items === null ? (
        <div className="empty-panel">
          <span className="spinner" />
          <p role="status">Загружаем доски…</p>
        </div>
      ) : store.items.length === 0 ? (
        <div className="empty-panel">
          <Icon name="board" size={30} />
          <h3>Пока ни одной доски</h3>
          <p>
            Создай первую — пригодится для схем, планов и разборов на стикерах.
          </p>
        </div>
      ) : (
        <ul className="boards-grid">
          {store.items.map((board) => (
            <li className="board-card" key={board.id}>
              <div className="board-card-head">
                <Icon name="board" size={20} />
                <strong>{board.title || "Без названия"}</strong>
              </div>
              <div className="board-meta">
                <span>{stamp(board.updatedAt)}</span>
                <span className="status-chip">версия {board.version}</span>
              </div>
              <div className="note-actions">
                <button
                  className="text-button"
                  onClick={() => void openBoard(board.id)}
                  disabled={store.busy || loading === board.id}
                >
                  <Icon name="edit" size={16} />
                  {loading === board.id ? "Открываю…" : "Открыть"}
                </button>
                <button
                  className="text-button danger-text"
                  onClick={() => void store.remove(board.id)}
                  disabled={store.busy}
                >
                  <Icon name="trash" size={16} />
                  Удалить
                </button>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
