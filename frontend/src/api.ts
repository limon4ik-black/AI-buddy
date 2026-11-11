export type Account = { id: string; email: string; role: string };
export type Tool = {
  id: string;
  name: string;
  description: string;
  teamId: string | null;
  documentId: string;
};
export type Profile = {
  userId: string;
  fullName: string;
  teamId: string;
  position: string;
  startDate: string;
};
export type Team = {
  id: string;
  name: string;
  description: string;
  buddy: string;
  buddyId: string;
};
export type Member = {
  id: string;
  fullName: string;
  position: string;
  role: string;
};
export type Service = {
  id: string;
  name: string;
  description: string;
  ownerTeamId: string;
  technicalOwnerId: string;
  lifecycle: string;
  repositoryUrl: string;
  documentationUrl: string;
  updatedAt: string;
};
export type TeamDetail = Team & { members: Member[]; services: Service[] };
export type Task = {
  id: string;
  title: string;
  description: string;
  period: string;
  documentId: string | null;
  sourceUrl: string;
  status: "pending" | "in_progress" | "completed";
  completedAt: string | null;
};
export type Checklist = {
  tasks: Task[];
  total: number;
  completed: number;
  percent: number;
  nextTask: Task | null;
};
export type Chat = {
  id: string;
  title: string;
  userId: string;
  createdAt: string;
  updatedAt: string;
};
export type Action = { type: "REQUEST_ACCESS"; label: string; system: string };
export type Message = {
  id: string;
  chatId: string;
  role: "user" | "assistant";
  content: string;
  sequence: number;
  sources: { documentId: string; title: string; section: string }[];
  suggestedAction: Action | null;
  replyTo: string | null;
  mode: "mock" | "rag";
};
export type Document = {
  id: string;
  title: string;
  content: string;
  category: string;
  sourceUrl: string;
  synthetic: boolean;
  visibility: string;
  teamId: string | null;
};
export type AccessRequest = {
  id: string;
  system: string;
  status: string;
  externalId: string;
  idempotencyKey: string;
  createdAt: string;
};
export class API {
  token = "";
  onExpired: () => void = () => {};
  private refreshing: Promise<boolean> | null = null;
  async refresh(): Promise<boolean> {
    if (this.refreshing) return this.refreshing;
    this.refreshing = (async () => {
      try {
        const res = await fetch("/api/v1/auth/refresh", {
          method: "POST",
          credentials: "include",
          signal: AbortSignal.timeout(30000),
        });
        if (!res.ok) return false;
        this.token = (await res.json()).accessToken;
        return true;
      } catch {
        return false;
      }
    })();
    try {
      return await this.refreshing;
    } finally {
      this.refreshing = null;
    }
  }
  async request<T>(
    path: string,
    method = "GET",
    body?: unknown,
    extra: Record<string, string> = {},
    retry = true,
  ): Promise<T> {
    const res = await fetch("/api/v1" + path, {
      method,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        ...(this.token ? { Authorization: "Bearer " + this.token } : {}),
        ...extra,
      },
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: AbortSignal.timeout(40000),
    });
    if (res.status === 401 && retry && !path.startsWith("/auth/")) {
      if (await this.refresh())
        return this.request<T>(path, method, body, extra, false);
      this.token = "";
      this.onExpired();
    }
    if (!res.ok) {
      const data = await res.json().catch(() => null);
      throw new Error(
        data?.error?.message || "Сервис недоступен. Повторите попытку.",
      );
    }
    if (res.status === 204) return undefined as T;
    const data = await res.json();
    if (data.accessToken) this.token = data.accessToken;
    return data;
  }
  async allMessages(id: string): Promise<Message[]> {
    const messages: Message[] = [];
    let cursor: number | null = 0;
    do {
      const page: { items: Message[]; nextCursor: number | null } =
        await this.request<{ items: Message[]; nextCursor: number | null }>(
          `/chats/${id}/messages?after=${cursor}`,
        );
      messages.push(...page.items);
      cursor = page.nextCursor;
    } while (cursor !== null);
    return messages;
  }
}
export const api = new API();

export type Source = Message["sources"][number];
