import { useCallback, useEffect, useState } from "react";
import { notes, type Note } from "../notes";
import { errorText } from "./useBuddy";

// Заметки живут в отдельном сервисе и могут быть недоступны сами по себе,
// поэтому у них своё состояние: недоступный блокнот не ломает остальной экран.
export function useNotes() {
  const [items, setItems] = useState<Note[] | null>(null),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  const load = useCallback(async () => {
    setBusy(true);
    setError("");
    try {
      setItems((await notes.list()).items);
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
    create: (content: string) =>
      run(async () => {
        const note = await notes.create(content);
        setItems((list) => [note, ...(list || [])]);
      }),
    save: (note: Note, content: string) =>
      run(async () => {
        const saved = await notes.save(note.id, content, note.version);
        setItems((list) =>
          (list || []).map((item) => (item.id === saved.id ? saved : item)),
        );
      }),
    remove: (id: string) =>
      run(async () => {
        await notes.remove(id);
        setItems((list) => (list || []).filter((item) => item.id !== id));
      }),
  };
}
