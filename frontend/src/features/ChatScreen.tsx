import { useEffect, useRef, useState } from "react";
import type { Buddy } from "../hooks/useBuddy";
import type { Source } from "../api";
import { Brand, Icon, type IconName } from "../components/Icon";
export function ChatScreen({
  buddy,
  onSource,
}: {
  buddy: Buddy;
  onSource: (source: Source) => void;
}) {
  const { data, question, messages, generating, sending, loadingChat } = buddy;
  const end = useRef<HTMLDivElement>(null),
    input = useRef<HTMLTextAreaElement>(null);
  const [copied, setCopied] = useState<string | null>(null);
  useEffect(() => {
    end.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, generating]);
  useEffect(() => {
    if (input.current) {
      input.current.style.height = "auto";
      input.current.style.height =
        Math.min(input.current.scrollHeight, 180) + "px";
    }
  }, [question]);
  const ownTool =
    data!.tools.find((t) => t.teamId === data!.profile.teamId) ||
    data!.tools[0];
  const prompts: {
    icon: IconName;
    title: string;
    subtitle: string;
    question: string;
    color: string;
  }[] = [
    {
      icon: "users",
      title: "Познакомиться с командой",
      subtitle: "Кто рядом и к кому обратиться",
      question: "Расскажи о моей команде и бадди",
      color: "blue",
    },
    {
      icon: "shield",
      title: "Получить нужные доступы",
      subtitle: ownTool ? "Инструмент: " + ownTool.name : "Инструменты для работы",
      question:
        "Как получить доступ к " + (ownTool?.name || "Confluence") + "?",
      color: "purple",
    },
    {
      icon: "target",
      title: "Понять следующий шаг",
      subtitle: "Что осталось в моей адаптации",
      question: "Какие шаги адаптации у меня остались?",
      color: "green",
    },
    {
      icon: "book",
      title: "Найти первую инструкцию",
      subtitle: "Материалы именно для моей команды",
      question: "Какие инструкции изучить в моей команде?",
      color: "red",
    },
  ];
  async function copy(id: string, text: string) {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(id);
      setTimeout(() => setCopied(null), 2000);
    } catch {
      buddy.setError(
        "Не удалось скопировать ответ. Выдели нужный текст вручную.",
      );
    }
  }
  return (
    <section className="chat-screen">
      <div className="chat-scroll" aria-live="polite">
        {loadingChat ? (
          <div className="chat-loading" role="status">
            <span className="spinner" />
            Загружаем разговор…
          </div>
        ) : messages.length === 0 ? (
          <div className="chat-welcome">
            <div className="welcome-orbit" aria-hidden="true">
              <i />
              <i />
              <i />
              <i />
            </div>
            <div className="welcome-kicker">
              ТВОЙ BUDDY В КОМАНДЕ{" "}
              {data!.teams
                .find((t) => t.id === data!.profile.teamId)
                ?.name.toUpperCase()}
            </div>
            <h1>
              Привет, {data!.profile.fullName.split(" ")[0]}
              <span className="wave">.</span>
              <br />С чего начнём?
            </h1>
            <p>
              Помогу разобраться в команде, найти нужное
              <br className="desktop-break" /> и сделать следующий шаг.
              Спрашивай как есть.
            </p>
            <div className="prompt-grid">
              {prompts.map((p) => (
                <button
                  key={p.title}
                  onClick={() => {
                    buddy.setQuestion(p.question);
                    input.current?.focus();
                  }}
                >
                  <span className={"prompt-icon " + p.color}>
                    <Icon name={p.icon} size={20} />
                  </span>
                  <span>
                    <strong>{p.title}</strong>
                    <small>{p.subtitle}</small>
                  </span>
                  <Icon name="arrow" size={16} />
                </button>
              ))}
            </div>
          </div>
        ) : (
          <div className="messages">
            {messages.map((m) => (
              <article key={m.id} className={"message " + m.role}>
                {m.role === "assistant" && (
                  <div className="message-author">
                    <span className="mini-brand">
                      <Brand small />
                    </span>
                    <span className="message-mode">
                      {m.mode === "rag" ? "RAG" : "Демо-ответ"}
                    </span>
                  </div>
                )}
                <div className="message-body">{m.content}</div>
                {m.role === "assistant" && (
                  <>
                    <div className="source-chips">
                      {m.sources.map((s) => (
                        <button key={s.documentId} onClick={() => onSource(s)}>
                          <Icon name="book" size={13} />
                          {s.title}
                          <Icon name="external" size={12} />
                        </button>
                      ))}
                    </div>
                    {m.suggestedAction && (
                      <button
                        className="action-card"
                        onClick={() =>
                          buddy.proposeAction(m.suggestedAction!, m.id)
                        }
                      >
                        <span className="prompt-icon blue">
                          <Icon name="shield" size={19} />
                        </span>
                        <span>
                          <strong>{m.suggestedAction.label}</strong>
                          <small>
                            {data!.tools.find(
                              (t) => t.id === m.suggestedAction?.system,
                            )?.name || m.suggestedAction.system}{" "}
                            · потребуется твоё подтверждение
                          </small>
                        </span>
                        <Icon name="arrow" size={18} />
                      </button>
                    )}
                    <div className="message-actions">
                      <button
                        className="icon-button"
                        aria-label="Скопировать ответ"
                        onClick={() => void copy(m.id, m.content)}
                      >
                        <Icon
                          name={copied === m.id ? "check" : "copy"}
                          size={16}
                        />
                      </button>
                      {copied === m.id && (
                        <span role="status">Скопировано</span>
                      )}
                    </div>
                  </>
                )}
              </article>
            ))}
          </div>
        )}
        {generating && (
          <div className="thinking" role="status">
            <span className="thinking-dots">
              <i />
              <i />
              <i />
            </span>
            Buddy готовит ответ
          </div>
        )}
        {buddy.retryId && !generating && buddy.chatId && (
          <div className="retry-line">
            <p>Вопрос сохранён. Можно повторить получение ответа.</p>
            <button
              className="secondary-button"
              onClick={() => {
                buddy.setError("");
                void buddy.generate(buddy.chatId!, buddy.retryId!);
              }}
            >
              <Icon name="refresh" size={16} />
              Повторить ответ
            </button>
          </div>
        )}
        <div ref={end} />
      </div>
      <div className="composer-wrap">
        <form
          className="composer"
          onSubmit={(e) => {
            e.preventDefault();
            void buddy.send();
          }}
        >
          <textarea
            ref={input}
            aria-label="Сообщение для Buddy"
            placeholder="Спроси Buddy о команде, доступах или своих задачах…"
            value={question}
            maxLength={2500}
            rows={1}
            onChange={(e) => buddy.setQuestion(e.target.value)}
            onKeyDown={(e) => {
              if (
                e.key === "Enter" &&
                !e.shiftKey &&
                !e.nativeEvent.isComposing
              ) {
                e.preventDefault();
                void buddy.send();
              }
            }}
          />
          <div className="composer-bottom">
            <span>
              <Icon name="shield" size={14} /> Контекст твоей команды
            </span>
            <button
              type="submit"
              className="send-button"
              aria-label="Отправить сообщение"
              disabled={
                !question.trim() || sending || generating || loadingChat
              }
            >
              {sending ? (
                <span className="spinner" />
              ) : (
                <Icon name="send" size={22} />
              )}
            </button>
          </div>
        </form>
        <p className="composer-note">
          Buddy помогает с адаптацией. Важные договорённости уточняй у бадди.
          <span>Enter ↵</span>
        </p>
      </div>
    </section>
  );
}
