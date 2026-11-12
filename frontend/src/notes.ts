import { api } from "./api";

export type Note = {
  id: string;
  userId: string;
  content: string;
  version: number;
  createdAt: string;
  updatedAt: string;
};

// Заметки отдаёт тот же origin: nginx фронта проксирует /notes-api/ в сервис заметок.
// Ходить напрямую на http://127.0.0.1:8079 нельзя — CSP фронта (connect-src 'self')
// отбрасывает запрос на чужой origin ещё до отправки.
// VITE_NOTES_URL переопределяет адрес целиком (например http://127.0.0.1:8079/api/v1).
const base =
  (import.meta.env.VITE_NOTES_URL as string | undefined) ?? "/notes-api";

async function call<T>(
  path: string,
  method = "GET",
  body?: unknown,
  retry = true,
): Promise<T> {
  const res = await fetch(base + path, {
    method,
    headers: {
      ...(body === undefined ? {} : { "Content-Type": "application/json" }),
      ...(api.token ? { Authorization: "Bearer " + api.token } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
    signal: AbortSignal.timeout(15000),
  });
  if (res.status === 401 && retry && (await api.refresh()))
    return call<T>(path, method, body, false);
  if (!res.ok) {
    const data = await res.json().catch(() => null);
    throw new Error(
      data?.error?.message || "Блокнот недоступен. Повтори попытку.",
    );
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const notes = {
  list: (limit = 50, offset = 0) =>
    call<{ items: Note[]; limit: number; offset: number }>(
      `/notes?limit=${limit}&offset=${offset}`,
    ),
  create: (content: string) => call<Note>("/notes", "POST", { content }),
  save: (id: string, content: string, version: number) =>
    call<Note>(`/notes/${id}`, "PATCH", { content, version }),
  remove: (id: string) => call<void>(`/notes/${id}`, "DELETE"),
};
