import { useCallback, useEffect, useState } from "react";
import { boards, emptyScene, type Board, type BoardFull } from "../boards";
import { errorText } from "./useBuddy";

// Доски живут в отдельном сервисе, поэтому у них своё состояние: упавший
// блокнот досок не должен ломать остальной экран.
export function useBoards() {
  const [items, setItems] = useState<Board[] | null>(null),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const load = useCallback(async () => {
    setBusy(true);
    setError("");
    try {
      setItems((await boards.list()).items);
    } catch (e) {
      setItems([]);
      setError(errorText(e));
    } finally {
      setBusy(false);
    }
  }, []);
  useEffect(() => {
    void load();
  }, [load]);
  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    try {
      await action();
      return true;
    } catch (e) {
      setError(errorText(e));
      return false;
    } finally {
      setBusy(false);
    }
  }
  return {
    items,
    busy,
    error,
    reload: load,
    // Новая доска создаётся пустой и сразу открывается в редакторе.
    create: async (title: string): Promise<BoardFull | null> => {
      setBusy(true);
      setError("");
      try {
        const created = await boards.create(title, emptyScene);
        setItems((list) => [created, ...(list || [])]);
        return created;
      } catch (e) {
        setError(errorText(e));
        return null;
      } finally {
        setBusy(false);
      }
    },
    remove: (id: string) =>
      run(async () => {
        await boards.remove(id);
        setItems((list) => (list || []).filter((item) => item.id !== id));
      }),
    // Редактор сохранил доску — обновляем её строку в списке.
    applySaved: (board: Board) =>
      setItems((list) =>
        (list || []).map((item) =>
          item.id === board.id ? { ...item, ...board } : item,
        ),
      ),
  };
}
