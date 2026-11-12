import { Excalidraw } from "@excalidraw/excalidraw";
import "@excalidraw/excalidraw/index.css";
import { Icon } from "../components/Icon";
import type { BoardFull } from "../boards";
import { useSceneAutosave } from "../hooks/useSceneAutosave";

// Тип данных холста берём из пропсов самого Excalidraw: тянуть типы из его внутренних путей хрупко.
type InitialData = NonNullable<
  React.ComponentProps<typeof Excalidraw>["initialData"]
>;

// Холст чужой (Excalidraw), наше здесь — сохранение. Экран грузится отдельным куском
// через React.lazy: SDK весит больше мегабайта, и в основной бандл ему незачем.
export function BoardEditor({
  board,
  onSaved,
  onClose,
}: {
  board: BoardFull;
  onSaved: (board: BoardFull) => void;
  onClose: () => void;
}) {
  const autosave = useSceneAutosave(board, onSaved);
  const scene = autosave.initialScene;
  const status =
    autosave.status === "saving"
      ? "Сохраняю…"
      : autosave.status === "saved"
        ? autosave.savedAt
          ? `Сохранено в ${autosave.savedAt}`
          : "Изменений нет"
        : autosave.message;

  return (
    <div className="board-editor">
      <div className="board-toolbar">
        <button className="text-button" onClick={onClose}>
          <Icon name="back" size={16} />К списку
        </button>
        <input
          className="board-title"
          aria-label="Название доски"
          maxLength={200}
          value={autosave.title}
          onChange={(e) => autosave.setTitle(e.target.value)}
        />
        <span
          className={"save-badge save-" + autosave.status}
          role="status"
          aria-live="polite"
        >
          {status}
        </span>
      </div>
      {autosave.status === "conflict" && (
        <p className="inline-error" role="alert">
          Доску изменили в другой вкладке.{" "}
          <button className="text-button" onClick={() => void autosave.reload()}>
            Загрузить актуальную
          </button>
        </p>
      )}
      {/* key пересоздаёт холст, когда мы перечитали доску с сервера */}
      <div className="board-canvas" key={autosave.revision}>
        <Excalidraw
          initialData={
            {
              elements: scene.elements,
              appState: scene.appState,
              files: scene.files,
            } as unknown as InitialData
          }
          onChange={(elements, appState, files) =>
            autosave.onSceneChange(elements, appState, files)
          }
        />
      </div>
    </div>
  );
}
