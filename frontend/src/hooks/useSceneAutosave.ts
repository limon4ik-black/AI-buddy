import { useCallback, useEffect, useRef, useState } from "react";
import { boards, BoardError, type BoardFull, type Scene } from "../boards";
import { errorText } from "./useBuddy";

// Пауза перед сохранением: холст меняется на каждое движение мыши.
const saveDelay = 1500;

// От состояния вида сохраняем только то, что относится к самой доске: зум, выделение и
// курсоры у каждого свои, и хранить их — значит переписывать доску на каждое движение мыши.
const keepAppState = ["viewBackgroundColor", "gridSize", "gridModeEnabled"] as const;

export function cleanAppState(appState: Record<string, unknown>) {
  const out: Record<string, unknown> = {};
  for (const key of keepAppState) if (key in appState) out[key] = appState[key];
  return out;
}

export type SaveStatus = "saved" | "saving" | "conflict" | "error";

const clock = () =>
  new Intl.DateTimeFormat("ru-RU", { hour: "2-digit", minute: "2-digit" }).format(
    new Date(),
  );

// useSceneAutosave копит изменения холста и отправляет сцену целиком с ожидаемой версией.
// delay вынесен в параметр, чтобы тесты не ждали полторы секунды.
export function useSceneAutosave(
  board: BoardFull,
  onSaved: (board: BoardFull) => void,
  delay: number = saveDelay,
) {
  const [title, setTitleState] = useState(board.title),
    [status, setStatus] = useState<SaveStatus>("saved"),
    [message, setMessage] = useState(""),
    [savedAt, setSavedAt] = useState(""),
    [revision, setRevision] = useState(0);
  const scene = useRef<Scene>(board.scene),
    version = useRef(board.version),
    currentTitle = useRef(board.title),
    lastSent = useRef(""),
    timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const save = useCallback(async () => {
    if (timer.current) {
      clearTimeout(timer.current);
      timer.current = null;
    }
    const body = JSON.stringify({
      title: currentTitle.current,
      scene: scene.current,
    });
    if (body === lastSent.current) return; // ничего не изменилось — не дёргаем сервис
    setStatus("saving");
    try {
      const updated = await boards.save(
        board.id,
        currentTitle.current,
        scene.current,
        version.current,
      );
      version.current = updated.version;
      lastSent.current = body;
      setStatus("saved");
      setMessage("");
      setSavedAt(clock());
      onSaved(updated);
    } catch (e) {
      if (e instanceof BoardError && e.status === 409) {
        setStatus("conflict");
        setMessage(e.message);
        return;
      }
      setStatus("error");
      setMessage(errorText(e));
    }
  }, [board.id, onSaved]);

  const schedule = useCallback(() => {
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void save(), delay);
  }, [delay, save]);

  const onSceneChange = useCallback(
    (elements: readonly unknown[], appState: unknown, files: unknown) => {
      // Данные холста идут от Excalidraw как есть: разбирать их нам незачем.
      scene.current = {
        type: "excalidraw",
        version: 2,
        source: "ai-buddy",
        elements,
        appState: cleanAppState(appState as Record<string, unknown>),
        files: files as Record<string, unknown>,
      };
      schedule();
    },
    [schedule],
  );

  const setTitle = useCallback(
    (value: string) => {
      currentTitle.current = value;
      setTitleState(value);
      schedule();
    },
    [schedule],
  );

  // Досылаем при уходе со страницы и при размонтировании: закрытая вкладка не должна съедать работу.
  useEffect(() => {
    const onHide = () => {
      if (document.visibilityState === "hidden") void save();
    };
    document.addEventListener("visibilitychange", onHide);
    return () => {
      document.removeEventListener("visibilitychange", onHide);
      void save();
    };
  }, [save]);

  const reload = useCallback(async () => {
    try {
      const fresh = await boards.get(board.id);
      scene.current = fresh.scene;
      version.current = fresh.version;
      currentTitle.current = fresh.title;
      lastSent.current = "";
      setTitleState(fresh.title);
      setStatus("saved");
      setMessage("");
      setRevision((r) => r + 1); // холст пересоздаём: так в него попадают данные с сервера
      onSaved(fresh);
    } catch (e) {
      setStatus("error");
      setMessage(errorText(e));
    }
  }, [board.id, onSaved]);

  return {
    title,
    setTitle,
    status,
    message,
    savedAt,
    onSceneChange,
    reload,
    revision,
    initialScene: board.scene,
  };
}
