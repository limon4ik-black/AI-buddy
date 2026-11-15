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
  await screen.findByRole("heading", { name: /Привет,/ });
  return user;
}
it("ведёт заметки во внешнем сервисе: создать, изменить, удалить", async () => {
  const user = await login();
  await user.click(screen.getByRole("button", { name: "Заметки" }));
  await screen.findByRole("heading", { name: "Мои заметки" });
  expect(await screen.findByText("Пока пусто")).toBeTruthy();

  await user.click(screen.getAllByRole("button", { name: "Новая заметка" })[0]);
  await user.type(screen.getByLabelText("Текст новой заметки"), "купить чипсы");
  await user.click(screen.getByRole("button", { name: "Добавить заметку" }));
  expect(await screen.findByText("купить чипсы")).toBeTruthy();

  await user.click(
    screen.getByRole("button", { name: /купить чипсы Открыть заметку/ }),
  );
  const editor = screen.getByLabelText(/^Текст заметки /);
  await user.clear(editor);
  await user.type(editor, "купить чипсы и шоколад");
  await user.click(screen.getByRole("button", { name: "Сохранить" }));
  expect(
    await screen.findByRole("button", {
      name: /купить чипсы и шоколад Открыть заметку/,
    }),
  ).toBeTruthy();
  expect(
    await screen.findByText(/Все изменения сохранены · версия 2/),
  ).toBeTruthy();

  await user.click(screen.getByRole("button", { name: "Удалить" }));
  expect(await screen.findByText("Пока пусто")).toBeTruthy();

  const calls = vi
    .mocked(globalThis.fetch)
    .mock.calls.map((call) => String(call[0]));
  expect(calls.some((url) => url.startsWith("/notes-api/notes"))).toBe(true);
  // наружу не ходим: CSP фронта разрешает только свой origin
  expect(calls.some((url) => /^https?:\/\//.test(url))).toBe(false);
});

it("ищет заметки и сохраняет черновик при конфликте версий", async () => {
  state.resume = true;
  state.notes = [
    {
      id: "saved-note",
      userId: "egor",
      owner: "egor",
      content: "Встреча с бадди\nОбсудить доступы",
      version: 1,
      createdAt: "2026-11-12T10:00:00Z",
      updatedAt: "2026-11-12T10:00:00Z",
    },
  ];
  const user = userEvent.setup();
  render(<App />);
  await user.click(
    await screen.findByRole("button", { name: "Заметки" }),
  );
  const search = screen.getByRole("textbox", { name: "Поиск по заметкам" });
  await user.type(search, "несуществующая");
  expect(await screen.findByText("Ничего не найдено")).toBeTruthy();
  await user.clear(search);
  await user.click(
    screen.getByRole("button", { name: /Встреча с бадди Обсудить доступы/ }),
  );
  const editor = screen.getByLabelText("Текст заметки saved-note");
  await user.clear(editor);
  await user.type(editor, "Мой несохранённый черновик");
  state.notes[0].version = 2;
  await user.click(screen.getByRole("button", { name: "Сохранить" }));
  await screen.findByText("Заметка уже изменена, актуальная версия 2");
  expect((editor as HTMLTextAreaElement).value).toBe(
    "Мой несохранённый черновик",
  );
  expect(state.notes[0].content).toBe("Встреча с бадди\nОбсудить доступы");
  const confirm = vi.spyOn(window, "confirm").mockReturnValue(false);
  await user.click(screen.getByRole("button", { name: "К списку заметок" }));
  expect((editor as HTMLTextAreaElement).value).toBe(
    "Мой несохранённый черновик",
  );
  confirm.mockReturnValue(true);
  await user.click(screen.getByRole("button", { name: "К списку заметок" }));
  await user.click(screen.getByRole("button", { name: "Повторить загрузку" }));
  await user.click(
    await screen.findByRole("button", {
      name: /Встреча с бадди Обсудить доступы/,
    }),
  );
  expect(
    (screen.getByLabelText("Текст заметки saved-note") as HTMLTextAreaElement)
      .value,
  ).toBe("Встреча с бадди\nОбсудить доступы");
});
