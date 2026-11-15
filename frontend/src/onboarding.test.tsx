import { beforeEach, expect, it, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { App } from "./App";
import { api } from "./api";
import { fakeFetch, mockState, type MockState } from "./test-fixtures";
import { safeExternalUrl } from "./model";

let state: MockState;
beforeEach(() => {
  vi.restoreAllMocks();
  api.token = "";
  state = mockState();
  state.resume = true;
  vi.spyOn(globalThis, "fetch").mockImplementation(fakeFetch(state));
});

it("shows real progress on home and updates the same checklist", async () => {
  const user = userEvent.setup();
  render(<App />);
  await screen.findByRole("heading", { name: "Привет, Егор 👋" });
  expect(screen.getByText("Выполнено 0 из 3 шагов.")).toBeTruthy();
  expect(screen.getByText("Бобрик: Новичок")).toBeTruthy();
  const month = new Date().toLocaleDateString("ru-RU", {
    month: "long",
    year: "numeric",
  });
  expect(screen.getByRole("heading", { name: month })).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "Следующий месяц" }));
  expect(screen.queryByRole("heading", { name: month })).toBeNull();
  await user.click(screen.getByRole("button", { name: "Предыдущий месяц" }));
  expect(screen.getByRole("heading", { name: month })).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "Открыть чек-лист" }));
  await screen.findByRole("heading", { name: "Чек-лист адаптации" });
  await user.click(
    screen.getByRole("checkbox", { name: "Прочитать гид новичка" }),
  );
  await screen.findByRole("heading", { name: "Твоя адаптация — 33%" });
  await user.click(screen.getByRole("button", { name: "Готово" }));
  expect(screen.getAllByRole("checkbox")).toHaveLength(1);
  await user.click(
    screen.getByRole("checkbox", { name: "Прочитать гид новичка" }),
  );
  await waitFor(() =>
    expect(screen.queryAllByRole("checkbox")).toHaveLength(0),
  );
  expect(screen.getByText("В этом списке пока нет шагов.")).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "Осталось" }));
  expect(screen.getAllByRole("checkbox")).toHaveLength(3);
  await user.click(
    screen.getAllByRole("button", { name: "Взять в работу" })[0],
  );
  await user.click(await screen.findByRole("button", { name: "В работе" }));
  await screen.findAllByRole("button", { name: "Взять в работу" });
  await user.click(screen.getAllByRole("button", { name: "Инструкция" })[0]);
  await screen.findByRole("dialog", { name: "Гид новичка" });
  await user.click(screen.getByRole("button", { name: "Закрыть окно" }));
  await user.click(screen.getByRole("button", { name: "Главная" }));
  await user.click(
    screen.getByRole("button", { name: "Расскажи о моей команде и бадди" }),
  );
  expect(
    (
      screen.getByRole("textbox", {
        name: "Сообщение для Buddy",
      }) as HTMLTextAreaElement
    ).value,
  ).toBe("Расскажи о моей команде и бадди");
});

it("shows saved notes on home and opens the working notebook", async () => {
  state.notes = [
    {
      id: "note-1",
      userId: "egor",
      owner: "egor",
      content: "Встреча с бадди\nОбсудить доступы",
      version: 1,
      createdAt: "2026-11-12T10:00:00Z",
      updatedAt: "2026-11-12T10:00:00Z",
    },
  ];
  render(<App />);
  const user = userEvent.setup();
  const note = await screen.findByRole("button", {
    name: /Встреча с бадди Обсудить доступы/,
  });
  await user.click(note);
  await screen.findByRole("heading", { name: "Мои заметки" });
  expect(screen.getByText("Встреча с бадди Обсудить доступы")).toBeTruthy();
});

it("does not let an unavailable notebook break home", async () => {
  state.fail = "GET /notes";
  render(<App />);
  await screen.findByRole("heading", { name: "Привет, Егор 👋" });
  await screen.findByText("Временно недоступно");
  expect(screen.getByRole("button", { name: "Открыть чек-лист" })).toBeTruthy();
});

it("allows only http and https external source links", () => {
  expect(safeExternalUrl("javascript:alert(1)")).toBeUndefined();
  expect(safeExternalUrl("data:text/html,abc")).toBeUndefined();
  expect(safeExternalUrl("invalid")).toBeUndefined();
  expect(safeExternalUrl("https://example.com/doc")).toBe(
    "https://example.com/doc",
  );
});
