import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import type { BoardFull, Scene } from "../boards";
import { cleanAppState, useSceneAutosave } from "./useSceneAutosave";

const scene: Scene = {
  type: "excalidraw",
  version: 2,
  elements: [],
  appState: {},
  files: {},
};
const board: BoardFull = {
  id: "board-1",
  userId: "anton",
  title: "Доска",
  version: 1,
  createdAt: "2026-11-12T10:00:00Z",
  updatedAt: "2026-11-12T10:00:00Z",
  scene,
};

let calls: { url: string; method: string; body: Record<string, unknown> }[];
let version: number;
const reply = (body: unknown, status = 200) =>
  ({ ok: status < 400, status, json: async () => body }) as Response;

beforeEach(() => {
  calls = [];
  version = 1;
  vi.restoreAllMocks();
  vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
    const body = init?.body ? JSON.parse(String(init.body)) : {};
    calls.push({ url: String(input), method: init?.method || "GET", body });
    if (init?.method === "PATCH") {
      if (body.version !== version)
        return reply(
          {
            error: {
              message: `Доска уже изменена, актуальная версия ${version}`,
            },
          },
          409,
        );
      version += 1;
      return reply({ ...board, scene: body.scene, version });
    }
    if (init?.method === "GET") return reply({ ...board, title: "с сервера", version: 7 });
    return reply({ ...board });
  });
});

it("выкидывает состояние вида, которого нет на доске", () => {
  expect(
    cleanAppState({
      zoom: { value: 2 },
      selectedElementIds: { el1: true },
      viewBackgroundColor: "#fff",
      gridSize: null,
    }),
  ).toEqual({ viewBackgroundColor: "#fff", gridSize: null });
});

it("после паузы сохраняет сцену с ожидаемой версией", async () => {
  const saved: BoardFull[] = [];
  const { result } = renderHook(() =>
    useSceneAutosave(board, (b) => saved.push(b), 10),
  );
  act(() =>
    result.current.onSceneChange(
      [{ id: "el1" }],
      { zoom: { value: 3 }, viewBackgroundColor: "#fff" },
      {},
    ),
  );
  // ждём именно запрос: статус "saved" стоит и до сохранения
  await waitFor(() =>
    expect(calls.some((call) => call.method === "PATCH")).toBe(true),
  );
  const patch = calls.find((call) => call.method === "PATCH");
  expect(patch?.url).toBe("/board-api/boards/board-1");
  expect(patch?.body.version).toBe(1);
  expect(patch?.body.title).toBe("Доска");
  expect((patch?.body.scene as Scene).appState).toEqual({
    viewBackgroundColor: "#fff",
  });
  await waitFor(() => expect(saved.length).toBe(1));
  expect(result.current.savedAt).not.toBe("");
});

it("не дёргает сервис, когда холст не менялся", async () => {
  const { result } = renderHook(() => useSceneAutosave(board, () => {}, 10));
  act(() => result.current.onSceneChange([{ id: "el1" }], {}, {}));
  await waitFor(() =>
    expect(calls.some((call) => call.method === "PATCH")).toBe(true),
  );
  act(() => result.current.onSceneChange([{ id: "el1" }], {}, {}));
  await new Promise((r) => setTimeout(r, 60));
  expect(calls.filter((call) => call.method === "PATCH").length).toBe(1);
});

it("называет конфликт версий, а не теряет правку молча", async () => {
  version = 5; // кто-то сохранил доску раньше нас
  const { result } = renderHook(() => useSceneAutosave(board, () => {}, 10));
  act(() => result.current.onSceneChange([{ id: "el1" }], {}, {}));
  await waitFor(() => expect(result.current.status).toBe("conflict"));
  expect(result.current.message).toContain("актуальная версия 5");
});

it("перезагрузка забирает доску с сервера", async () => {
  const { result } = renderHook(() => useSceneAutosave(board, () => {}, 10));
  act(() => void result.current.reload());
  await waitFor(() => expect(result.current.title).toBe("с сервера"));
  expect(result.current.revision).toBe(1);
  expect(calls.some((call) => call.method === "GET")).toBe(true);
});
