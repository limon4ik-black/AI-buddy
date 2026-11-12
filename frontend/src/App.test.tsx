import { beforeEach, expect, it, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
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
async function login(person?: string) {
  const user = userEvent.setup();
  render(<App />);
  if (person)
    await user.click(await screen.findByRole("button", { name: person }));
  await user.click(
    await screen.findByRole("button", { name: "Войти в пространство" }),
  );
  await user.type(await screen.findByPlaceholderText("000000"), "123456");
  await user.click(await screen.findByRole("button", { name: "Подтвердить вход" }));
  await screen.findByRole("heading", { name: /С чего начнём/ });
  return user;
}
async function openProfile(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: "Открыть профиль" }));
  await screen.findByRole("heading", { name: "Профиль и адаптация" });
}
it("recovers the password with an email code", async () => {
  render(<App />);
  const user = userEvent.setup();
  await user.click(await screen.findByRole("button", { name: "Забыли пароль?" }));
  await user.click(screen.getByRole("button", { name: "Получить код" }));
  await user.type(await screen.findByPlaceholderText("000000"), "123456");
  await user.type(screen.getByLabelText("Новый пароль"), "NewPassword123!");
  await user.click(screen.getByRole("button", { name: "Сменить пароль" }));
  await screen.findByRole("heading", { name: /С чего начнём/ });
});
it("shows a team-aware chat, confirms a request and saves progress in the profile", async () => {
  const user = await login("Анна Frontend");
  await user.click(
    screen.getByRole("button", { name: /Получить нужные доступы/ }),
  );
  expect(
    (
      screen.getByRole("textbox", {
        name: "Сообщение для Buddy",
      }) as HTMLTextAreaElement
    ).value,
  ).toContain("Figma");
  await user.click(screen.getByRole("button", { name: "Отправить сообщение" }));
  await screen.findByText(
    "Тестовый ответ для Анна Лебедева. Инструмент Figma.",
  );
  await user.click(screen.getByRole("button", { name: /Запросить доступ/ }));
  await screen.findByRole("dialog", { name: "Отправить заявку?" });
  await user.click(screen.getByRole("button", { name: "Отмена" }));
  expect(screen.queryByRole("dialog")).toBeNull();
  await user.click(screen.getByRole("button", { name: /Запросить доступ/ }));
  await user.click(screen.getByRole("button", { name: "Отправить заявку" }));
  await screen.findByText(/Заявка DEMO-ANNA отправлена/);
  await user.click(screen.getByRole("button", { name: "Скрыть уведомление" }));
  await openProfile(user);
  expect(screen.getByText("DEMO-ANNA · 9 ноября")).toBeTruthy();
  const check = screen.getByRole("checkbox", { name: "Прочитать гид новичка" });
  expect(screen.getByText("Бобрик: Новичок")).toBeTruthy();
  await user.click(check);
  await waitFor(() => expect((check as HTMLInputElement).checked).toBe(true));
  await waitFor(() =>
    expect(screen.getByText("Бобрик: Освоился")).toBeTruthy(),
  );
  expect(screen.getByRole("progressbar").getAttribute("aria-valuenow")).toBe(
    "33",
  );
  await user.click(screen.getByRole("button", { name: "Готово" }));
  expect(screen.getAllByRole("checkbox")).toHaveLength(1);
  await user.click(screen.getByRole("button", { name: "Осталось" }));
  expect(screen.getAllByRole("checkbox")).toHaveLength(2);
  await user.click(screen.getByRole("button", { name: "Все шаги" }));
  await user.click(
    screen.getAllByRole("button", { name: "Взять в работу" })[0],
  );
  await screen.findByRole("button", { name: "В работе" });
  await user.click(screen.getByRole("button", { name: "Изменить профиль" }));
  await user.clear(screen.getByLabelText("Имя и фамилия"));
  await user.type(screen.getByLabelText("Имя и фамилия"), "Анна Тестовая");
  await user.click(screen.getByRole("button", { name: "Сохранить" }));
  await screen.findByRole("heading", { name: "Анна Тестовая" });
  await user.click(screen.getAllByRole("button", { name: "Инструкция" })[0]);
  await screen.findByRole("dialog", { name: "Гид новичка" });
  await user.keyboard("{Escape}");
  expect(screen.queryByRole("dialog")).toBeNull();
  await user.click(
    screen.getByRole("button", { name: "Подготовиться к встрече" }),
  );
  await screen.findByRole("textbox", { name: "Сообщение для Buddy" });
  expect(
    (
      screen.getByRole("textbox", {
        name: "Сообщение для Buddy",
      }) as HTMLTextAreaElement
    ).value,
  ).toContain("бадди");
});
it("supports directory search, team ownership and readable knowledge sources", async () => {
  const user = await login();
  await user.click(screen.getByRole("button", { name: "Команды" }));
  await screen.findByRole("heading", { name: "Знакомься с командами" });
  await user.click(
    screen.getByRole("button", { name: /Frontend Команда Frontend/ }),
  );
  expect(screen.getByText("Анна Лебедева")).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "Весь каталог" }));
  await screen.findByRole("heading", { name: "Кто отвечает за сервис" });
  await user.type(
    screen.getByRole("textbox", { name: "Поиск сервиса" }),
    "nonexistent",
  );
  await screen.findByText("Такого сервиса не нашли");
  await user.clear(screen.getByRole("textbox", { name: "Поиск сервиса" }));
  await user.selectOptions(
    screen.getByRole("combobox", { name: "Команда-владелец" }),
    "frontend",
  );
  expect(screen.getByText("Найдено: 1")).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "Frontend" }));
  await screen.findByRole("heading", { name: "Знакомься с командами" });
  await user.click(screen.getByRole("button", { name: "База знаний" }));
  await screen.findByRole("heading", { name: "Всё начинается с вопроса" });
  await user.click(screen.getByRole("button", { name: "Моя команда" }));
  expect(screen.getByRole("button", { name: /Доступы Backend/ })).toBeTruthy();
  expect(screen.queryByRole("button", { name: /Доступы Frontend/ })).toBeNull();
  await user.click(screen.getByRole("button", { name: "Общие" }));
  expect(screen.getByRole("button", { name: /Гид новичка/ })).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "Все" }));
  await user.type(
    screen.getByRole("textbox", { name: "Поиск по базе знаний" }),
    "zzz",
  );
  await screen.findByText("По этому запросу ничего нет");
  await user.clear(
    screen.getByRole("textbox", { name: "Поиск по базе знаний" }),
  );
  await user.click(screen.getByRole("button", { name: /Гид новичка/ }));
  const dialog = await screen.findByRole("dialog", { name: "Гид новичка" });
  expect(
    within(dialog).getByRole("link", { name: "Открытый материал по теме" }),
  ).toBeTruthy();
  await user.click(screen.getByRole("button", { name: "Закрыть окно" }));
});
it("renames, finds and deletes chats without mixing accounts", async () => {
  const user = await login();
  await user.type(
    screen.getByRole("textbox", { name: "Сообщение для Buddy" }),
    "Расскажи о моей команде{Enter}",
  );
  await screen.findByText(/Тестовый ответ для Егор/);
  await user.click(screen.getByRole("button", { name: /Действия с чатом/ }));
  await user.click(screen.getByRole("button", { name: "Переименовать" }));
  await user.clear(screen.getByRole("textbox", { name: "Название чата" }));
  await user.type(
    screen.getByRole("textbox", { name: "Название чата" }),
    "Мой первый день",
  );
  await user.click(screen.getByRole("button", { name: "Сохранить название" }));
  await screen.findByRole("button", { name: "Мой первый день" });
  await user.type(
    screen.getByRole("textbox", { name: "Поиск по чатам" }),
    "zzz",
  );
  await screen.findByText("Ничего не найдено");
  await user.clear(screen.getByRole("textbox", { name: "Поиск по чатам" }));
  await user.click(screen.getByRole("button", { name: "Новый чат" }));
  await user.click(screen.getByRole("button", { name: "Мой первый день" }));
  await screen.findByText(/Тестовый ответ для Егор/);
  await user.click(screen.getByRole("button", { name: /Моя команда/ }));
  await screen.findByRole("heading", { name: "Знакомься с командами" });
  await user.click(screen.getByRole("button", { name: "Вернуться к чату" }));
  await user.click(screen.getByRole("button", { name: /Действия с чатом/ }));
  await user.click(screen.getByRole("button", { name: "Удалить" }));
  await user.click(screen.getByRole("button", { name: "Удалить чат" }));
  await screen.findByRole("heading", { name: /С чего начнём/ });
  await user.click(screen.getByRole("button", { name: "Выйти из аккаунта" }));
  await screen.findByRole("button", { name: "Войти в пространство" });
  await user.click(screen.getByRole("button", { name: "София QA" }));
  await user.click(
    screen.getByRole("button", { name: "Войти в пространство" }),
  );
  await user.type(await screen.findByPlaceholderText("000000"), "123456");
  await user.click(await screen.findByRole("button", { name: "Подтвердить вход" }));
  await screen.findByRole("heading", { name: /Привет, София/ });
  expect(screen.queryByRole("button", { name: "Мой первый день" })).toBeNull();
  await openProfile(user);
  expect(screen.getByText("qa@buddy.test")).toBeTruthy();
  expect(screen.queryByText("backend@buddy.test")).toBeNull();
});
it("retries generation and corporate errors with stable idempotency keys", async () => {
  state.generateFails = true;
  const user = await login();
  await user.type(
    screen.getByRole("textbox", { name: "Сообщение для Buddy" }),
    "Доступ к Stash{Enter}",
  );
  await user.click(
    await screen.findByRole("button", { name: "Повторить ответ" }),
  );
  await screen.findByRole("button", { name: "Повторить ответ" });
  state.generateFails = false;
  await user.click(screen.getByRole("button", { name: "Скрыть ошибку" }));
  await user.click(screen.getByRole("button", { name: "Повторить ответ" }));
  await screen.findByText(/Тестовый ответ для Егор/);
  expect(
    vi
      .mocked(fetch)
      .mock.calls.filter(
        ([url, opt]) =>
          String(url).endsWith("/messages") && opt?.method === "POST",
      ),
  ).toHaveLength(1);
  state.actionFails = true;
  await user.click(screen.getByRole("button", { name: /Запросить доступ/ }));
  await user.click(screen.getByRole("button", { name: "Отправить заявку" }));
  await screen.findByText("Заявка временно не отправлена");
  state.actionFails = false;
  await user.click(screen.getByRole("button", { name: "Отправить заявку" }));
  await screen.findByText(/Заявка DEMO-EGOR отправлена/);
  const calls = vi
    .mocked(fetch)
    .mock.calls.filter(
      ([url, opt]) =>
        String(url).endsWith("/access-requests") && opt?.method === "POST",
    );
  expect(calls[0][1]?.headers).toMatchObject(calls[1][1]?.headers || {});
});
it("retains profile drafts and rolls back a failed checklist update", async () => {
  const user = await login();
  await openProfile(user);
  state.fail = "PATCH /checklist/tasks/read-welcome";
  const checkbox = screen.getByRole("checkbox", {
    name: "Прочитать гид новичка",
  });
  await user.click(checkbox);
  await screen.findByText("Временно недоступно");
  await waitFor(() =>
    expect((checkbox as HTMLInputElement).checked).toBe(false),
  );
  state.fail = "PATCH /profile";
  await user.click(screen.getByRole("button", { name: "Изменить профиль" }));
  await user.clear(screen.getByLabelText("Имя и фамилия"));
  await user.type(
    screen.getByLabelText("Имя и фамилия"),
    "Сохранённый черновик",
  );
  await user.click(screen.getByRole("button", { name: "Сохранить" }));
  expect(
    (screen.getByLabelText("Имя и фамилия") as HTMLInputElement).value,
  ).toBe("Сохранённый черновик");
  await user.click(screen.getByRole("button", { name: "Отмена" }));
});
it("handles registration, wrong credentials and mobile navigation", async () => {
  const user = userEvent.setup();
  render(<App />);
  await user.click(
    await screen.findByRole("button", { name: "Создать свой аккаунт" }),
  );
  await user.clear(screen.getByLabelText("Email"));
  await user.type(screen.getByLabelText("Email"), "new@buddy.test");
  await user.clear(screen.getByLabelText("Пароль"));
  await user.type(screen.getByLabelText("Пароль"), "WrongPassword123!");
  await user.click(await screen.findByRole("button", { name: "Создать аккаунт" }));
  await screen.findByText("Неверный email или пароль");
  await user.clear(screen.getByLabelText("Пароль"));
  await user.type(screen.getByLabelText("Пароль"), "NewPassword123!");
  await user.click(
    screen.getByRole("button", { name: "Уже есть аккаунт? Войти" }),
  );
  await user.click(
    screen.getByRole("button", { name: "Войти в пространство" }),
  );
  await user.type(await screen.findByPlaceholderText("000000"), "123456");
  await user.click(await screen.findByRole("button", { name: "Подтвердить вход" }));
  await screen.findByRole("heading", { name: /С чего начнём/ });
  await user.click(screen.getByRole("button", { name: "Открыть навигацию" }));
  expect(
    screen
      .getByRole("button", { name: "Открыть навигацию" })
      .getAttribute("aria-expanded"),
  ).toBe("true");
  await user.click(screen.getByRole("button", { name: "Закрыть навигацию" }));
  await user.click(screen.getByRole("button", { name: "Открыть навигацию" }));
  await user.click(
    screen.getByRole("button", { name: "Скрыть боковую панель" }),
  );
});
it("restores sessions and unfinished conversations", async () => {
  state.resume = true;
  state.chats = [
    {
      id: "saved",
      title: "Сохранённый вопрос",
      owner: "egor",
      userId: "egor",
      createdAt: "",
      updatedAt: "",
    },
  ];
  state.messages.saved = [
    {
      id: "old",
      chatId: "saved",
      role: "user",
      content: "Доступы",
      sequence: 1,
      sources: [],
      suggestedAction: null,
      replyTo: null,
      mode: "mock",
    },
  ];
  const user = userEvent.setup();
  render(<App />);
  await user.click(
    await screen.findByRole("button", {
      name: "Сохранённый вопрос",
    }),
  );
  await user.click(
    await screen.findByRole("button", { name: "Повторить ответ" }),
  );
  await screen.findByText(/Тестовый ответ для Егор/);
});
it("shows a recoverable workspace load failure", async () => {
  state.resume = true;
  state.fail = "GET /profile";
  render(<App />);
  await screen.findByText("Временно недоступно");
  expect(
    screen.getByRole("button", { name: "Войти в пространство" }),
  ).toBeTruthy();
});
