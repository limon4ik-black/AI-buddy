import { beforeEach, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App } from "./App";
import { api } from "./api";
import { fakeFetch, mockState, type MockState } from "./test-fixtures";

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
  await screen.findByRole("heading", { name: /С чего начнём/ });
  return user;
}
it("ведёт заметки во внешнем сервисе: создать, изменить, удалить", async () => {
  const user = await login();
  await user.click(screen.getByRole("button", { name: /Заметки/ }));
  await screen.findByRole("heading", { name: "Мои заметки" });
  expect(await screen.findByText("Пока пусто")).toBeTruthy();

  await user.type(screen.getByLabelText("Текст новой заметки"), "купить чипсы");
  await user.click(screen.getByRole("button", { name: "Добавить заметку" }));
  expect(await screen.findByText("купить чипсы")).toBeTruthy();

  await user.click(screen.getByRole("button", { name: "Изменить" }));
  const editor = screen.getByLabelText(/^Текст заметки /);
  await user.clear(editor);
  await user.type(editor, "купить чипсы и шоколад");
  await user.click(screen.getByRole("button", { name: "Сохранить" }));
  expect(await screen.findByText("купить чипсы и шоколад")).toBeTruthy();
  expect(await screen.findByText("версия 2")).toBeTruthy();

  await user.click(screen.getByRole("button", { name: "Удалить" }));
  expect(await screen.findByText("Пока пусто")).toBeTruthy();

  const calls = vi.mocked(globalThis.fetch).mock.calls.map((call) =>
    String(call[0]),
  );
  expect(calls.some((url) => url.startsWith("/notes-api/notes"))).toBe(true);
  // наружу не ходим: CSP фронта разрешает только свой origin
  expect(calls.some((url) => /^https?:\/\//.test(url))).toBe(false);
});
