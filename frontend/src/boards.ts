import { api } from "./api";

export type Board = {
  id: string;
  userId: string;
  title: string;
  version: number;
  createdAt: string;
  updatedAt: string;
};

// Scene — то, что отдаёт Excalidraw: элементы, состояние вида и вложенные файлы (картинки).
export type Scene = {
  type: "excalidraw";
  version: number;
  source?: string;
  elements: readonly unknown[];
  appState: Record<string, unknown>;
  files: Record<string, unknown>;
};

export type BoardFull = Board & { scene: Scene };

// emptyScene — с доски начинают пустой: сервис требует объект, а не null.
export const emptyScene: Scene = {
  type: "excalidraw",
  version: 2,
  source: "ai-buddy",
  elements: [],
  appState: {},
  files: {},
};

// Доски живут в том же origin, что и остальное приложение: nginx (а в dev — прокси vite)
// ведёт /board-api/ в сервис заметок. Прямой адрес сервиса браузер не пустит —
// CSP фронта разрешает запросы только на свой origin.
const base =
  (import.meta.env.VITE_BOARDS_URL as string | undefined) ?? "/board-api";

// Сцена весит мегабайты, поэтому таймаут больше, чем у заметок.
const timeout = 30000;

// BoardError несёт HTTP-статус: по нему автосохранение отличает конфликт версий от прочих бед.
export class BoardError extends Error {
  constructor(
    message: string,
    readonly status: number,
  ) {
    super(message);
  }
}

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
    signal: AbortSignal.timeout(timeout),
  });
  if (res.status === 401 && retry && (await api.refresh()))
    return call<T>(path, method, body, false);
  if (!res.ok) {
    const data = await res.json().catch(() => null);
    throw new BoardError(
      data?.error?.message || "Доски недоступны. Повтори попытку.",
      res.status,
    );
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

export const boards = {
  list: (limit = 50, offset = 0) =>
    call<{ items: Board[]; limit: number; offset: number }>(
      `/boards?limit=${limit}&offset=${offset}`,
    ),
  get: (id: string) => call<BoardFull>(`/boards/${id}`),
  create: (title: string, scene: Scene) =>
    call<BoardFull>("/boards", "POST", { title, scene }),
  save: (id: string, title: string, scene: Scene, version: number) =>
    call<BoardFull>(`/boards/${id}`, "PATCH", { title, scene, version }),
  remove: (id: string) => call<void>(`/boards/${id}`, "DELETE"),
};
