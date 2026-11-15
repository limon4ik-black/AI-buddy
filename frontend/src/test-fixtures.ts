import type {
  Profile,
  Checklist,
  Chat,
  Message,
  Document,
  TeamDetail,
  Service,
  AccessRequest,
  Tool,
} from "./api";
import type { Note } from "./notes";
import type { Board, Scene } from "./boards";
export const testPeople = [
  {
    email: "backend@buddy.test",
    id: "egor",
    name: "Егор Соколов",
    team: "backend",
    tool: "stash",
  },
  {
    email: "frontend@buddy.test",
    id: "anna",
    name: "Анна Лебедева",
    team: "frontend",
    tool: "figma",
  },
  {
    email: "analytics@buddy.test",
    id: "maxim",
    name: "Максим Орлов",
    team: "analytics",
    tool: "metabase",
  },
  {
    email: "qa@buddy.test",
    id: "sofia",
    name: "София Морозова",
    team: "qa",
    tool: "testrail",
  },
];
export const names: Record<string, string> = {
    backend: "Backend",
    frontend: "Frontend",
    analytics: "Analytics",
    qa: "QA",
  },
  toolNames: Record<string, string> = {
    stash: "Stash",
    figma: "Figma",
    metabase: "Metabase",
    testrail: "TestRail",
  };
const doc = (id: string, teamId: string | null, title: string): Document => ({
  id,
  teamId,
  title,
  content: "ДЕМО-ИНСТРУКЦИЯ. " + title + ". Вопросы можно обсудить с бадди.",
  category: "onboarding",
  sourceUrl: "https://example.test/source",
  synthetic: true,
  visibility: teamId ? "team" : "company",
});
export const documents = [
  doc("welcome", null, "Гид новичка"),
  doc("security", null, "Безопасная работа"),
  ...testPeople.map((p) =>
    doc(p.team + "-access", p.team, "Доступы " + names[p.team]),
  ),
];
export const services: Service[] = testPeople.map((p) => ({
  id: p.team + "-api",
  name: names[p.team] + " API",
  description: "Учебный сервис " + names[p.team],
  ownerTeamId: p.team,
  technicalOwnerId: p.team + "-buddy",
  lifecycle: p.team === "backend" ? "active" : "development",
  repositoryUrl: p.team === "qa" ? "" : "https://example.test/repo",
  documentationUrl: "",
  updatedAt: "2026-11-09T00:00:00Z",
}));
export const details: TeamDetail[] = testPeople.map((p) => ({
  id: p.team,
  name: names[p.team],
  description: "Команда " + names[p.team],
  buddy: "Бадди " + names[p.team],
  buddyId: p.team + "-buddy",
  members: [
    {
      id: p.team + "-buddy",
      fullName: "Бадди " + names[p.team],
      position: "Инженер команды",
      role: "buddy",
    },
    { id: p.id, fullName: p.name, position: "Стажёр", role: "intern" },
  ],
  services: services.filter((s) => s.ownerTeamId === p.team),
}));
export function initialChecklist(): Checklist {
  const tasks = [
    {
      id: "read-welcome",
      title: "Прочитать гид новичка",
      description: "Разобраться с первыми шагами.",
      period: "1 неделя",
      documentId: "welcome",
      sourceUrl: "https://example.test",
      status: "pending" as const,
      completedAt: null,
    },
    {
      id: "meet-team",
      title: "Познакомиться с командой",
      description: "Обсудить вопросы с бадди.",
      period: "2 неделя",
      documentId: null,
      sourceUrl: "https://example.test",
      status: "pending" as const,
      completedAt: null,
    },
    {
      id: "feedback",
      title: "Получить обратную связь",
      description: "Обсудить результаты.",
      period: "1 месяц",
      documentId: "welcome",
      sourceUrl: "https://example.test",
      status: "pending" as const,
      completedAt: null,
    },
  ];
  return { tasks, total: 3, completed: 0, percent: 0, nextTask: tasks[0] };
}
export type MockState = {
  resume: boolean;
  email: string;
  loginCode: string;
  fail: string;
  generateFails: boolean;
  actionFails: boolean;
  checklists: Record<string, Checklist>;
  profiles: Record<string, Profile>;
  requests: Record<string, AccessRequest[]>;
  chats: (Chat & { owner: string })[];
  messages: Record<string, Message[]>;
  notes: (Note & { owner: string })[];
  boards: (Board & { scene: Scene; owner: string })[];
  sequence: number;
};
export function mockState(): MockState {
  return {
    resume: false,
    email: testPeople[0].email,
    loginCode: "",
    fail: "",
    generateFails: false,
    actionFails: false,
    checklists: {},
    profiles: {},
    requests: {},
    chats: [],
    messages: {},
    notes: [],
    boards: [],
    sequence: 0,
  };
}
const response = (value: unknown, status = 200) =>
  new Response(status === 204 ? null : JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json" },
  });
export function fakeFetch(state: MockState) {
  return async (input: RequestInfo | URL, init?: RequestInit) => {
    const path = String(input)
        .replace(/^https?:\/\/[^/]+/, "") // адрес заметок может быть переопределён абсолютным
        .replace("/notes-api", "") // заметки ходят через nginx фронта: тот же origin
        .replace("/board-api", "") // и доски тоже
        .replace("/api/v1", "")
        .split("?")[0],
      method = init?.method || "GET",
      body = init?.body ? JSON.parse(String(init.body)) : {};
    const person = testPeople.find((p) => p.email === state.email) || {
      ...testPeople[0],
      email: state.email,
      id: "new",
      name: "Стажёр",
    };
    if (state.fail === method + " " + path)
      return response({ error: { message: "Временно недоступно" } }, 503);
    if (path === "/auth/refresh")
      return state.resume
        ? response({ accessToken: "test-token" })
        : response({}, 401);
    if (path === "/auth/login") {
      if (body.password === "WrongPassword123!")
        return response(
          { error: { message: "Неверный email или пароль" } },
          401,
        );
      state.email = body.email;
      state.loginCode = "123456";
      return response({
        message: "Код для входа отправлен на " + body.email,
      });
    }
    if (path === "/auth/confirm") {
      if (body.code !== state.loginCode)
        return response(
          { error: { message: "Код неверный или устарел" } },
          400,
        );
      state.resume = true;
      return response({ accessToken: "test-token", expiresIn: 900 });
    }
    if (path === "/auth/password/reset") {
      state.loginCode = "123456";
      return response({
        message: "Если аккаунт существует, код отправлен на почту",
      });
    }
    if (path === "/auth/password/confirm") {
      if (body.code !== state.loginCode)
        return response(
          { error: { message: "Код неверный или устарел" } },
          400,
        );
      state.resume = true;
      return response({ accessToken: "test-token", expiresIn: 900 });
    }
    if (path === "/auth/logout") {
      state.resume = false;
      return response(null, 204);
    }
    if (path === "/auth/me")
      return response({ id: person.id, email: person.email, role: "intern" });
    if (!state.profiles[person.id])
      state.profiles[person.id] = {
        userId: person.id,
        fullName: person.name,
        position: "Стажёр " + names[person.team],
        teamId: person.team,
        startDate: "2026-11-02T00:00:00Z",
      };
    if (!state.checklists[person.id])
      state.checklists[person.id] = initialChecklist();
    if (path === "/profile") {
      if (method === "PATCH")
        state.profiles[person.id] = { ...state.profiles[person.id], ...body };
      return response(state.profiles[person.id]);
    }
    if (path === "/checklist") return response(state.checklists[person.id]);
    if (path.startsWith("/checklist/tasks/")) {
      const list = state.checklists[person.id],
        id = path.split("/").at(-1),
        tasks = list.tasks.map((t) =>
          t.id === id
            ? {
                ...t,
                status: body.status,
                completedAt:
                  body.status === "completed" ? "2026-11-09T00:00:00Z" : null,
              }
            : t,
        ),
        completed = tasks.filter((t) => t.status === "completed").length;
      state.checklists[person.id] = {
        tasks,
        total: 3,
        completed,
        percent: Math.floor((completed * 100) / 3),
        nextTask: tasks.find((t) => t.status !== "completed") || null,
      };
      return response(state.checklists[person.id]);
    }
    if (path === "/teams")
      return response({
        items: details.map(({ members, services, ...t }) => t),
      });
    if (path.startsWith("/teams/"))
      return response(details.find((t) => t.id === path.split("/").at(-1)));
    if (path === "/services") return response({ items: services });
    if (path === "/information")
      return response({
        items: documents.filter(
          (d) => d.teamId === null || d.teamId === person.team,
        ),
      });
    const tools: Tool[] = [
      {
        id: person.tool,
        name: toolNames[person.tool],
        description: "Учебный инструмент",
        teamId: person.team,
        documentId: person.team + "-access",
      },
      {
        id: "confluence",
        name: "Confluence",
        description: "База знаний",
        teamId: null,
        documentId: "welcome",
      },
    ];
    if (path === "/tools") return response({ items: tools });
    if (path === "/access-requests") {
      if (method === "POST") {
        if (state.actionFails)
          return response(
            { error: { message: "Заявка временно не отправлена" } },
            502,
          );
        const request = {
          id: "request-" + person.id,
          system: body.system,
          status: "submitted",
          externalId: "DEMO-" + person.id.toUpperCase(),
          idempotencyKey: "key",
          createdAt: "2026-11-09T00:00:00Z",
        };
        state.requests[person.id] = [request];
        return response(request, 201);
      }
      return response({ items: state.requests[person.id] || [] });
    }
    if (path === "/chats") {
      if (method === "POST") {
        const chat = {
          id: "chat-" + ++state.sequence,
          title: body.title,
          userId: person.id,
          owner: person.id,
          createdAt: "",
          updatedAt: "",
        };
        state.chats.push(chat);
        return response(chat, 201);
      }
      return response({
        items: state.chats.filter((c) => c.owner === person.id),
      });
    }
    const id = path.split("/")[2];
    if (path === `/chats/${id}`) {
      if (method === "PATCH") {
        const chat = state.chats.find((c) => c.id === id)!;
        chat.title = body.title;
        return response(chat);
      }
      if (method === "DELETE") {
        state.chats = state.chats.filter((c) => c.id !== id);
        delete state.messages[id];
        return response(null, 204);
      }
    }
    if (path === `/chats/${id}/messages`) {
      if (method === "POST") {
        const message: Message = {
          id: "m-" + ++state.sequence,
          chatId: id,
          content: body.content,
          role: "user",
          sequence: state.sequence,
          sources: [],
          suggestedAction: null,
          replyTo: null,
          mode: "mock",
        };
        state.messages[id] = [...(state.messages[id] || []), message];
        return response(message, 201);
      }
      return response({ items: state.messages[id] || [], nextCursor: null });
    }
    if (path.endsWith("/answer")) {
      if (state.generateFails)
        return response(
          { error: { message: "Не удалось получить ответ" } },
          502,
        );
      const question = state.messages[id]
        .filter((m) => m.role === "user")
        .at(-1)!;
      const source = question.content.includes("команд")
        ? "team:" + person.team
        : person.team + "-access";
      const message: Message = {
        ...question,
        id: "a-" + ++state.sequence,
        sequence: state.sequence,
        role: "assistant",
        content:
          "Тестовый ответ для " +
          person.name +
          ". Инструмент " +
          toolNames[person.tool] +
          ".",
        sources: [
          {
            documentId: source,
            title: source.startsWith("team:")
              ? "Моя команда"
              : "Доступы " + names[person.team],
            section: "Демо",
          },
        ],
        suggestedAction: {
          type: "REQUEST_ACCESS",
          label: "Запросить доступ",
          system: person.tool,
        },
        replyTo: question.id,
      };
      state.messages[id].push(message);
      return response(message, 201);
    }
    if (path === "/notes") {
      if (method === "POST") {
        const note = {
          id: "note-" + ++state.sequence,
          userId: person.id,
          content: body.content,
          version: 1,
          createdAt: "2026-11-12T10:00:00Z",
          updatedAt: "2026-11-12T10:00:00Z",
          owner: person.id,
        };
        state.notes.unshift(note);
        return response(note, 201);
      }
      return response({
        items: state.notes.filter((n) => n.owner === person.id),
        limit: 20,
        offset: 0,
      });
    }
    if (path.startsWith("/notes/")) {
      const noteId = path.split("/").at(-1)!,
        note = state.notes.find(
          (n) => n.id === noteId && n.owner === person.id,
        );
      if (!note)
        return response({ error: { message: "Объект не найден" } }, 404);
      if (method === "PATCH") {
        if (body.version !== note.version)
          return response(
            {
              error: {
                message:
                  "Заметка уже изменена, актуальная версия " + note.version,
              },
            },
            409,
          );
        note.content = body.content;
        note.version += 1;
        note.updatedAt = "2026-11-12T11:00:00Z";
        return response(note);
      }
      if (method === "DELETE") {
        state.notes = state.notes.filter((n) => n.id !== noteId);
        return response(null, 204);
      }
    }
    if (path === "/boards") {
      if (method === "POST") {
        const board = {
          id: "board-" + ++state.sequence,
          userId: person.id,
          title: body.title,
          scene: body.scene,
          version: 1,
          createdAt: "2026-11-12T10:00:00Z",
          updatedAt: "2026-11-12T10:00:00Z",
          owner: person.id,
        };
        state.boards.unshift(board);
        return response(board, 201);
      }
      // Список без сцен — как в настоящем сервисе.
      return response({
        items: state.boards
          .filter((b) => b.owner === person.id)
          .map((b) => ({
            id: b.id,
            userId: b.userId,
            title: b.title,
            version: b.version,
            createdAt: b.createdAt,
            updatedAt: b.updatedAt,
          })),
        limit: 20,
        offset: 0,
      });
    }
    if (path.startsWith("/boards/")) {
      const boardId = path.split("/").at(-1)!,
        board = state.boards.find(
          (b) => b.id === boardId && b.owner === person.id,
        );
      if (!board)
        return response({ error: { message: "Объект не найден" } }, 404);
      if (method === "PATCH") {
        if (body.version !== board.version)
          return response(
            {
              error: {
                message:
                  "Доска уже изменена, актуальная версия " + board.version,
              },
            },
            409,
          );
        board.title = body.title;
        board.scene = body.scene;
        board.version += 1;
        board.updatedAt = "2026-11-12T11:00:00Z";
        return response(board);
      }
      if (method === "DELETE") {
        state.boards = state.boards.filter((b) => b.id !== boardId);
        return response(null, 204);
      }
      return response(board);
    }
    throw new Error("Unhandled " + method + " " + path);
  };
}
