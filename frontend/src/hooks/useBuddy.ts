import { useCallback, useEffect, useRef, useState } from "react";
import {
  api,
  type Profile,
  type Checklist,
  type Chat,
  type Message,
  type Team,
  type TeamDetail,
  type Service,
  type Document,
  type Tool,
  type AccessRequest,
  type Action,
  type Account,
} from "../api";
import type { View, Workspace } from "../model";
export const errorText = (e: unknown) =>
  e instanceof Error
    ? e.message
    : "Не удалось выполнить действие. Попробуй ещё раз.";
export function useBuddy() {
  const [data, setData] = useState<Workspace | null>(null),
    [booting, setBooting] = useState(true),
    [authBusy, setAuthBusy] = useState(false),
    [error, setError] = useState(""),
    [notice, setNotice] = useState(""),
    [codeSent, setCodeSent] = useState(false);
  const [view, setView] = useState<View>("home"),
    [chatId, setChatId] = useState<string | null>(null),
    [messages, setMessages] = useState<Message[]>([]),
    [question, setQuestion] = useState(""),
    [sending, setSending] = useState(false),
    [generating, setGenerating] = useState(false),
    [loadingChat, setLoadingChat] = useState(false),
    [retryId, setRetryId] = useState<string | null>(null),
    [savingProfile, setSavingProfile] = useState(false),
    [updatingTask, setUpdatingTask] = useState(false),
    [requestBusy, setRequestBusy] = useState(false),
    [actionError, setActionError] = useState("");
  const [pendingAction, setPendingAction] = useState<{
    action: Action;
    key: string;
  } | null>(null);
  const epoch = useRef(0),
    activeChat = useRef<string | null>(null),
    actionKeys = useRef(new Map<string, string>()),
    pendingSend = useRef<{
      id: string;
      content: string;
      chatId: string;
    } | null>(null),
    taskLock = useRef(false),
    sendLock = useRef(false),
    generationLock = useRef(false);
  const clearSession = useCallback(() => {
    epoch.current++;
    setData(null);
    setView("home");
    setChatId(null);
    activeChat.current = null;
    setMessages([]);
    setQuestion("");
    setRetryId(null);
    setLoadingChat(false);
    setSending(false);
    setGenerating(false);
    setPendingAction(null);
    actionKeys.current.clear();
    pendingSend.current = null;
    setError("");
    setNotice("");
  }, []);
  const loadWorkspace = useCallback(async () => {
    const current = epoch.current;
    const [
      account,
      profile,
      checklist,
      chats,
      teams,
      services,
      documents,
      tools,
      requests,
    ] = await Promise.all([
      api.request<Account>("/auth/me"),
      api.request<Profile>("/profile"),
      api.request<Checklist>("/checklist"),
      api.request<{ items: Chat[] }>("/chats"),
      api.request<{ items: Team[] }>("/teams"),
      api.request<{ items: Service[] }>("/services"),
      api.request<{ items: Document[] }>("/information"),
      api.request<{ items: Tool[] }>("/tools"),
      api.request<{ items: AccessRequest[] }>("/access-requests"),
    ]);
    const details = await Promise.all(
      teams.items.map((t) => api.request<TeamDetail>("/teams/" + t.id)),
    );
    if (current !== epoch.current) return;
    setData({
      account,
      profile,
      checklist,
      chats: chats.items,
      teams: teams.items,
      teamDetails: Object.fromEntries(details.map((d) => [d.id, d])),
      services: services.items,
      documents: documents.items,
      tools: tools.items,
      requests: requests.items,
    });
  }, []);
  useEffect(() => {
    api.onExpired = () => {
      clearSession();
      setError("Сессия истекла. Войди снова.");
    };
    void (async () => {
      if (await api.refresh()) {
        try {
          await loadWorkspace();
        } catch (e) {
          setError(errorText(e));
        }
      }
      setBooting(false);
    })();
    return () => {
      api.onExpired = () => {};
    };
  }, [loadWorkspace, clearSession]);
  async function login(email: string, password: string, register = false) {
    setAuthBusy(true);
    setError("");
    setNotice("");
    try {
      const res = await api.request<{ message?: string }>(
        "/auth/" + (register ? "register" : "login"),
        "POST",
        { email, password },
      );
      setCodeSent(true);
      setNotice(
        res?.message ||
          "Код для входа отправлен на " + email + ". Проверь почту.",
      );
    } catch (e) {
      setError(errorText(e));
    } finally {
      setAuthBusy(false);
    }
  }
  async function confirmCode(code: string) {
    setAuthBusy(true);
    setError("");
    try {
      await api.request("/auth/confirm", "POST", { code });
      setCodeSent(false);
      setNotice("");
      epoch.current++;
      await loadWorkspace();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setAuthBusy(false);
    }
  }
  async function forgotPassword(email: string) {
    setAuthBusy(true);
    setError("");
    setNotice("");
    try {
      await api.request("/auth/password/reset", "POST", { email });
      setNotice("Если аккаунт существует, код отправлен на почту");
      return true;
    } catch (e) {
      setError(errorText(e));
      return false;
    } finally {
      setAuthBusy(false);
    }
  }
  async function resetPassword(code: string, password: string) {
    setAuthBusy(true);
    setError("");
    try {
      await api.request("/auth/password/confirm", "POST", {
        code,
        newPassword: password,
      });
      setCodeSent(false);
      setNotice("");
      epoch.current++;
      await loadWorkspace();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setAuthBusy(false);
    }
  }
  async function logout() {
    setAuthBusy(true);
    try {
      await api.request("/auth/logout", "POST");
      api.token = "";
      clearSession();
    } catch (e) {
      setError(errorText(e));
    } finally {
      setAuthBusy(false);
    }
  }
  function navigate(next: View) {
    setView(next);
    setError("");
    setNotice("");
  }
  function newChat() {
    if (sendLock.current || generationLock.current) return;
    activeChat.current = null;
    setChatId(null);
    setMessages([]);
    setRetryId(null);
    setQuestion("");
    pendingSend.current = null;
    navigate("chat");
  }
  async function selectChat(id: string) {
    if (sendLock.current || generationLock.current) return;
    const current = epoch.current;
    activeChat.current = id;
    setChatId(id);
    setMessages([]);
    setRetryId(null);
    setLoadingChat(true);
    navigate("chat");
    try {
      const list = await api.allMessages(id);
      if (activeChat.current === id && current === epoch.current) {
        setMessages(list);
        setRetryId(
          list
            .filter(
              (m) => m.role === "user" && !list.some((a) => a.replyTo === m.id),
            )
            .at(-1)?.id || null,
        );
      }
    } catch (e) {
      if (activeChat.current === id) setError(errorText(e));
    } finally {
      if (activeChat.current === id) setLoadingChat(false);
    }
  }
  async function generate(id: string, messageId: string) {
    if (generationLock.current) return;
    generationLock.current = true;
    const current = epoch.current;
    setGenerating(true);
    setRetryId(messageId);
    try {
      const answer = await api.request<Message>(
        `/chats/${id}/messages/${messageId}/answer`,
        "POST",
      );
      if (activeChat.current === id && current === epoch.current) {
        setMessages((list) =>
          list.some((m) => m.id === answer.id) ? list : [...list, answer],
        );
        setRetryId(null);
      }
    } catch (e) {
      if (current === epoch.current) setError(errorText(e));
    } finally {
      generationLock.current = false;
      if (current === epoch.current) setGenerating(false);
    }
  }
  async function send() {
    const content = question.trim();
    if (
      !content ||
      sendLock.current ||
      generationLock.current ||
      loadingChat ||
      !data
    )
      return;
    sendLock.current = true;
    const current = epoch.current;
    setSending(true);
    setError("");
    try {
      let id = chatId;
      if (!id) {
        const chat = await api.request<Chat>("/chats", "POST", {
          title: Array.from(content).slice(0, 40).join(""),
        });
        if (current !== epoch.current) return;
        id = chat.id;
        activeChat.current = id;
        setChatId(id);
        setData((d) =>
          d && current === epoch.current
            ? { ...d, chats: [chat, ...d.chats] }
            : d,
        );
      }
      let pending = pendingSend.current;
      if (!pending || pending.content !== content || pending.chatId !== id) {
        pending = { id: crypto.randomUUID(), content, chatId: id };
        pendingSend.current = pending;
      }
      const message = await api.request<Message>(
        `/chats/${id}/messages`,
        "POST",
        { content, clientMessageId: pending.id },
      );
      if (current !== epoch.current) return;
      setMessages((list) =>
        list.some((m) => m.id === message.id) ? list : [...list, message],
      );
      setQuestion("");
      pendingSend.current = null;
      setSending(false);
      await generate(id, message.id);
    } catch (e) {
      if (current === epoch.current) setError(errorText(e));
    } finally {
      sendLock.current = false;
      if (current === epoch.current) setSending(false);
    }
  }
  async function renameChat(id: string, title: string) {
    const chat = await api.request<Chat>("/chats/" + id, "PATCH", { title });
    setData((d) =>
      d ? { ...d, chats: d.chats.map((c) => (c.id === id ? chat : c)) } : d,
    );
  }
  async function deleteChat(id: string) {
    await api.request("/chats/" + id, "DELETE");
    setData((d) =>
      d ? { ...d, chats: d.chats.filter((c) => c.id !== id) } : d,
    );
    if (chatId === id) newChat();
  }
  async function saveProfile(fullName: string, position: string) {
    const current = epoch.current;
    setSavingProfile(true);
    setError("");
    try {
      const profile = await api.request<Profile>("/profile", "PATCH", {
        fullName,
        position,
      });
      if (current === epoch.current) {
        setData((d) => (d ? { ...d, profile } : d));
        setNotice("Профиль сохранён");
        return true;
      }
    } catch (e) {
      setError(errorText(e));
    } finally {
      setSavingProfile(false);
    }
  }
  async function updateTask(id: string, status: string) {
    if (!data || taskLock.current) return;
    taskLock.current = true;
    setUpdatingTask(true);
    setError("");
    const previous = data.checklist,
      current = epoch.current;
    const tasks = previous.tasks.map((t) =>
        t.id === id
          ? {
              ...t,
              status: status as typeof t.status,
              completedAt:
                status === "completed" ? new Date().toISOString() : null,
            }
          : t,
      ),
      completed = tasks.filter((t) => t.status === "completed").length;
    setData((d) =>
      d
        ? {
            ...d,
            checklist: {
              tasks,
              total: tasks.length,
              completed,
              percent: tasks.length
                ? Math.floor((completed * 100) / tasks.length)
                : 0,
              nextTask: tasks.find((t) => t.status !== "completed") || null,
            },
          }
        : d,
    );
    try {
      const checklist = await api.request<Checklist>(
        "/checklist/tasks/" + id,
        "PATCH",
        { status },
      );
      if (current === epoch.current)
        setData((d) => (d ? { ...d, checklist } : d));
    } catch (e) {
      if (current === epoch.current) {
        setData((d) => (d ? { ...d, checklist: previous } : d));
        setError(errorText(e));
      }
    } finally {
      taskLock.current = false;
      setUpdatingTask(false);
    }
  }
  function proposeAction(action: Action, messageId: string) {
    const keyId = messageId + ":" + action.system;
    let key = actionKeys.current.get(keyId);
    if (!key) {
      key = crypto.randomUUID();
      actionKeys.current.set(keyId, key);
    }
    setPendingAction({ action, key });
    setActionError("");
  }
  async function confirmAction() {
    if (!pendingAction) return;
    setRequestBusy(true);
    setActionError("");
    const current = epoch.current;
    try {
      const result = await api.request<AccessRequest>(
        "/access-requests",
        "POST",
        { system: pendingAction.action.system, confirmed: true },
        { "Idempotency-Key": pendingAction.key },
      );
      if (current === epoch.current) {
        setData((d) =>
          d
            ? {
                ...d,
                requests: [
                  result,
                  ...d.requests.filter((r) => r.id !== result.id),
                ],
              }
            : d,
        );
        setNotice(
          `Заявка ${result.externalId} отправлена. Статус можно посмотреть в профиле.`,
        );
        setPendingAction(null);
      }
    } catch (e) {
      setActionError(errorText(e));
    } finally {
      setRequestBusy(false);
    }
  }
  return {
    data,
    booting,
    authBusy,
    error,
    notice,
    codeSent,
    confirmCode,
    view,
    chatId,
    messages,
    question,
    sending,
    generating,
    loadingChat,
    retryId,
    savingProfile,
    updatingTask,
    pendingAction,
    requestBusy,
    actionError,
    setQuestion,
    setError,
    setNotice,
    setPendingAction,
    login,
    forgotPassword,
    resetPassword,
    logout,
    navigate,
    newChat,
    selectChat,
    send,
    generate,
    renameChat,
    deleteChat,
    saveProfile,
    updateTask,
    proposeAction,
    confirmAction,
  };
}
export type Buddy = ReturnType<typeof useBuddy>;
