import { useEffect, useState, type FormEvent } from "react";
import type { Chat, Document, Source } from "./api";
import { useBuddy, errorText } from "./hooks/useBuddy";
import { Brand, Icon } from "./components/Icon";
import { Sidebar } from "./components/Sidebar";
import { Dialog } from "./components/Dialog";
import { AuthScreen } from "./features/AuthScreen";
import { ChatScreen } from "./features/ChatScreen";
import { ProfileScreen } from "./features/ProfileScreen";
import {
  TeamsScreen,
  ServicesScreen,
  KnowledgeScreen,
} from "./features/DirectoryScreens";
import { NotesScreen } from "./features/NotesScreen";
import { BoardsScreen } from "./features/BoardsScreen";
export function App() {
  const buddy = useBuddy(),
    [sidebar, setSidebar] = useState(false),
    [document, setDocument] = useState<Document | null>(null),
    [selectedTeam, setSelectedTeam] = useState(""),
    [chatAction, setChatAction] = useState<{
      kind: "rename" | "delete";
      chat: Chat;
    } | null>(null),
    [title, setTitle] = useState(""),
    [dialogError, setDialogError] = useState(""),
    [dialogBusy, setDialogBusy] = useState(false);
  const d = buddy.data;
  useEffect(() => {
    setDocument(null);
    setChatAction(null);
    setSelectedTeam("");
    setSidebar(false);
  }, [d?.profile.userId]);
  if (buddy.booting)
    return (
      <main className="boot-screen">
        <Brand />
        <span className="spinner" />
        <p role="status">Открываем твоё пространство…</p>
      </main>
    );
  if (!d)
    return (
      <AuthScreen
        busy={buddy.authBusy}
        error={buddy.error}
        notice={buddy.notice}
        codeSent={buddy.codeSent}
        onLogin={buddy.login}
        onConfirm={buddy.confirmCode}
        onForgot={buddy.forgotPassword}
        onResetConfirm={buddy.resetPassword}
      />
    );
  function openSource(source: Source) {
    if (source.documentId.startsWith("team:")) {
      setSelectedTeam(source.documentId.slice(5));
      buddy.navigate("teams");
      return;
    }
    if (source.documentId.startsWith("service:")) {
      buddy.navigate("services");
      return;
    }
    const doc = d!.documents.find((x) => x.id === source.documentId);
    if (doc) setDocument(doc);
    else buddy.setError("Этот источник сейчас недоступен. Обнови страницу.");
  }
  function changeChat(kind: "rename" | "delete", chat: Chat) {
    setChatAction({ kind, chat });
    setTitle(chat.title);
    setDialogError("");
  }
  async function submitChatAction(e: FormEvent) {
    e.preventDefault();
    if (!chatAction) return;
    setDialogBusy(true);
    setDialogError("");
    try {
      if (chatAction.kind === "rename")
        await buddy.renameChat(chatAction.chat.id, title.trim());
      else await buddy.deleteChat(chatAction.chat.id);
      setChatAction(null);
    } catch (e) {
      setDialogError(errorText(e));
    } finally {
      setDialogBusy(false);
    }
  }
  const activeTitle = d.chats.find((c) => c.id === buddy.chatId)?.title;
  const titles = {
    chat: activeTitle || "Твой AI Buddy",
    profile: "Мой профиль",
    teams: "Команды",
    services: "Сервисы",
    knowledge: "База знаний",
    notes: "Заметки",
    boards: "Доски",
  };
  const latestAssistant = [...buddy.messages]
    .reverse()
    .find((m) => m.role === "assistant");
  return (
    <div className="app-shell">
      <Sidebar
        buddy={buddy}
        open={sidebar}
        onClose={() => setSidebar(false)}
        onChatAction={changeChat}
      />
      <main className="main-panel">
        <header className="workspace-header">
          <div className="header-leading">
            <button
              className="icon-button mobile-only"
              aria-label="Открыть навигацию"
              aria-expanded={sidebar}
              onClick={() => setSidebar(true)}
            >
              <Icon name="menu" />
            </button>
            {buddy.view !== "chat" && (
              <button
                className="icon-button back-to-chat"
                aria-label="Вернуться к чату"
                onClick={() => buddy.navigate("chat")}
              >
                <Icon name="back" size={19} />
              </button>
            )}
            <span className="header-title">{titles[buddy.view]}</span>
            {buddy.view === "chat" && (
              <span className="mode-label">
                <i />
                {latestAssistant?.mode === "rag"
                  ? "RAG-помощник"
                  : "Демо-помощник"}
              </span>
            )}
          </div>
          <button
            className="header-progress"
            onClick={() => buddy.navigate("profile")}
          >
            <span className="header-progress-icon">
              <Icon name="target" size={16} />
            </span>
            <span>Моя адаптация</span>
            <strong>{d.checklist.percent}%</strong>
            <Icon name="chevron" size={14} />
          </button>
        </header>
        {buddy.error && (
          <div className="app-alert error" role="alert">
            <Icon name="shield" size={18} />
            <span>{buddy.error}</span>
            <button
              className="icon-button"
              aria-label="Скрыть ошибку"
              onClick={() => buddy.setError("")}
            >
              <Icon name="close" size={16} />
            </button>
          </div>
        )}
        {buddy.notice && (
          <div className="app-alert success" role="status">
            <Icon name="check" size={18} />
            <span>{buddy.notice}</span>
            <button
              className="icon-button"
              aria-label="Скрыть уведомление"
              onClick={() => buddy.setNotice("")}
            >
              <Icon name="close" size={16} />
            </button>
          </div>
        )}
        {buddy.view === "chat" && (
          <ChatScreen buddy={buddy} onSource={openSource} />
        )}
        <div
          className={
            "screen-scroll " + (buddy.view === "chat" ? "screen-hidden" : "")
          }
        >
          {buddy.view === "profile" && (
            <ProfileScreen buddy={buddy} onDocument={setDocument} />
          )}{" "}
          {buddy.view === "teams" && (
            <TeamsScreen
              data={d}
              selected={selectedTeam || d.profile.teamId}
              onSelect={setSelectedTeam}
              onServices={() => buddy.navigate("services")}
            />
          )}{" "}
          {buddy.view === "services" && (
            <ServicesScreen
              data={d}
              onTeam={(id) => {
                setSelectedTeam(id);
                buddy.navigate("teams");
              }}
            />
          )}{" "}
          {buddy.view === "knowledge" && (
            <KnowledgeScreen data={d} onDocument={setDocument} />
          )}{" "}
          {buddy.view === "notes" && <NotesScreen />}{" "}
          {buddy.view === "boards" && <BoardsScreen />}
        </div>
      </main>
      {buddy.pendingAction && (
        <Dialog
          title="Отправить заявку?"
          onClose={() => buddy.setPendingAction(null)}
          busy={buddy.requestBusy}
        >
          <div className="request-confirm-icon">
            <Icon name="shield" size={28} />
          </div>
          <p className="dialog-lead">
            Запросим доступ к{" "}
            <strong>
              {d.tools.find((t) => t.id === buddy.pendingAction?.action.system)
                ?.name || buddy.pendingAction.action.system}
            </strong>
            .
          </p>
          <div className="request-summary">
            <span>От имени</span>
            <strong>{d.profile.fullName}</strong>
            <span>Команда</span>
            <strong>
              {d.teams.find((t) => t.id === d.profile.teamId)?.name}
            </strong>
          </div>
          <p className="dialog-help">
            Тестовый сервис зарегистрирует заявку и вернёт номер. Сам доступ
            потребует отдельного согласования с бадди.
          </p>
          {buddy.actionError && (
            <p className="inline-error" role="alert">
              {buddy.actionError}
            </p>
          )}
          <div className="dialog-actions">
            <button
              className="secondary-button"
              disabled={buddy.requestBusy}
              onClick={() => buddy.setPendingAction(null)}
            >
              Отмена
            </button>
            <button
              className="primary-button"
              disabled={buddy.requestBusy}
              onClick={() => void buddy.confirmAction()}
            >
              {buddy.requestBusy ? "Отправляем…" : "Отправить заявку"}
              <Icon name="arrow" size={16} />
            </button>
          </div>
        </Dialog>
      )}
      {document && (
        <Dialog title={document.title} onClose={() => setDocument(null)} wide>
          <div className="document-modal-meta">
            <span className="team-label blue">
              {document.visibility === "team"
                ? "Материал твоей команды"
                : "Общий материал"}
            </span>
            <span>Демонстрационная инструкция</span>
          </div>
          <div className="document-content">{document.content}</div>
          <a
            className="document-reference"
            href={document.sourceUrl}
            target="_blank"
            rel="noreferrer"
          >
            Открытый материал по теме
            <Icon name="external" size={14} />
          </a>
        </Dialog>
      )}
      {chatAction && (
        <Dialog
          title={
            chatAction.kind === "rename" ? "Переименовать чат" : "Удалить чат?"
          }
          onClose={() => setChatAction(null)}
          busy={dialogBusy}
        >
          <form onSubmit={submitChatAction}>
            {chatAction.kind === "rename" ? (
              <label>
                Название чата
                <input
                  required
                  maxLength={80}
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                />
              </label>
            ) : (
              <p className="dialog-help">
                Разговор «{chatAction.chat.title}» и его сообщения будут
                удалены.
              </p>
            )}
            {dialogError && (
              <p className="inline-error" role="alert">
                {dialogError}
              </p>
            )}
            <div className="dialog-actions">
              <button
                type="button"
                className="secondary-button"
                disabled={dialogBusy}
                onClick={() => setChatAction(null)}
              >
                Отмена
              </button>
              <button
                className={
                  chatAction.kind === "delete"
                    ? "danger-button"
                    : "primary-button"
                }
                disabled={dialogBusy}
              >
                {dialogBusy
                  ? "Сохраняем…"
                  : chatAction.kind === "rename"
                    ? "Сохранить название"
                    : "Удалить чат"}
              </button>
            </div>
          </form>
        </Dialog>
      )}
    </div>
  );
}
