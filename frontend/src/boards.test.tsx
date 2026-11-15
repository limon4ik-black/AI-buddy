import { beforeEach, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App } from "./App";
import { api } from "./api";
import { fakeFetch, mockState, type MockState } from "./test-fixtures";

// Холст подменяем заглушкой: Excalidraw в jsdom не живёт, а проверяем мы вкладку
// и то, какими запросами ходим в сервис.
vi.mock("./features/BoardEditor", () => ({
  BoardEditor: ({ onClose }: { onClose: () => void }) => (
    <div>
      <p role="status">холст-заглушка</p>
      <button onClick={onClose}>К списку</button>
    </div>
  ),
}));

let state: MockState;
beforeEach(() => {
  vi.restoreAllMocks();
  api.token = "";
  state = mockState();
  vi.spyOn(globalThis, "fetch").mockImplementation(fakeFetch(state));
});
async function login() {
  const user = userEvent.setup();
  render(<App />);
  await user.click(
    await screen.findByRole("button", { name: "Войти в пространство" }),
  );
  await user.type(await screen.findByPlaceholderText("000000"), "123456");
  await user.click(
    await screen.findByRole("button", { name: "Подтвердить вход" }),
  );
  await screen.findByRole("heading", { name: /Привет,/ });
  return user;
}

it("создаёт, открывает и удаляет доску через свой origin", async () => {
  const user = await login();
  await user.click(screen.getByRole("button", { name: "Заметки" }));
  await user.click(screen.getByRole("button", { name: "Открыть доски" }));
  await screen.findByRole("heading", { name: "Доски" });
  expect(await screen.findByText("Пока ни одной доски")).toBeTruthy();

  await user.click(screen.getByRole("button", { name: "Создать доску" }));
  // новая доска сразу открывается в редакторе
  expect(await screen.findByText("холст-заглушка")).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "К списку" }));
  expect(await screen.findByText("Новая доска")).toBeTruthy();
  expect(await screen.findByText("версия 1")).toBeTruthy();

  await user.click(screen.getByRole("button", { name: "Открыть" }));
  expect(await screen.findByText("холст-заглушка")).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "К списку" }));

  await user.click(await screen.findByRole("button", { name: "Удалить" }));
  expect(await screen.findByText("Пока ни одной доски")).toBeTruthy();

  const calls = vi
    .mocked(globalThis.fetch)
    .mock.calls.map((call) => String(call[0]));
  expect(calls.some((url) => url.startsWith("/board-api/boards"))).toBe(true);
  // наружу не ходим: CSP фронта разрешает только свой origin
  expect(calls.some((url) => /^https?:\/\//.test(url))).toBe(false);
});
