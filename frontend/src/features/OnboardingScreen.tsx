import { useState } from "react";
import type { Buddy } from "../hooks/useBuddy";
import type { Document, Task } from "../api";
import { Icon } from "../components/Icon";
import { Pet } from "../components/Pet";
import { dateLabel } from "../model";
import { useNotes } from "../hooks/useNotes";

const prompts = [
  "Расскажи о моей команде и бадди",
  "Какие шаги адаптации у меня остались?",
  "Как получить нужные доступы?",
  "Какие инструкции изучить в моей команде?",
];
const calendarDate = (value: string) =>
  new Date(/^\d{4}-\d{2}-\d{2}$/.test(value) ? value + "T00:00:00" : value);

export function TaskCard({
  task,
  buddy,
  onDocument,
}: {
  task: Task;
  buddy: Buddy;
  onDocument: (doc: Document) => void;
}) {
  const doc = buddy.data!.documents.find((d) => d.id === task.documentId);
  return (
    <div
      className={
        "onboarding-task " + (task.status === "completed" ? "done" : "")
      }
    >
      <input
        type="checkbox"
        aria-label={task.title}
        checked={task.status === "completed"}
        disabled={buddy.updatingTask}
        onChange={() =>
          void buddy.updateTask(
            task.id,
            task.status === "completed" ? "pending" : "completed",
          )
        }
      />
      <div>
        <strong>{task.title}</strong>
        <p>{task.description}</p>
        <div className="onboarding-task-actions">
          <span>{task.period}</span>
          {doc && (
            <button className="text-button" onClick={() => onDocument(doc)}>
              Инструкция <Icon name="external" size={13} />
            </button>
          )}
          {task.status !== "completed" && (
            <button
              className="text-button"
              disabled={buddy.updatingTask}
              onClick={() =>
                void buddy.updateTask(
                  task.id,
                  task.status === "in_progress" ? "pending" : "in_progress",
                )
              }
            >
              {task.status === "in_progress" ? "В работе" : "Взять в работу"}
            </button>
          )}
        </div>
      </div>
    </div>
  );
}

export function Calendar({
  startDate,
  tasks,
}: {
  startDate: string;
  tasks: Task[];
}) {
  const [offset, setOffset] = useState(0);
  const today = new Date();
  const month = new Date(today.getFullYear(), today.getMonth() + offset, 1);
  const first = (month.getDay() + 6) % 7;
  const count = new Date(
    month.getFullYear(),
    month.getMonth() + 1,
    0,
  ).getDate();
  // The API has periods, not deadlines: only actual start/completion dates are shown.
  const events = [
    { date: startDate, title: "Начало стажировки" },
    ...tasks
      .filter((t) => t.completedAt)
      .map((t) => ({ date: t.completedAt!, title: t.title })),
  ];
  const key = (date: Date) =>
    `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
  return (
    <section className="dashboard-card calendar-card">
      <div className="dashboard-card-heading">
        <h2>
          {month.toLocaleDateString("ru-RU", {
            month: "long",
            year: "numeric",
          })}
        </h2>
        <div>
          <button
            className="icon-button"
            aria-label="Предыдущий месяц"
            onClick={() => setOffset((x) => x - 1)}
          >
            <Icon name="back" size={17} />
          </button>
          <button
            className="icon-button"
            aria-label="Следующий месяц"
            onClick={() => setOffset((x) => x + 1)}
          >
            <Icon name="arrow" size={17} />
          </button>
        </div>
      </div>
      <div className="onboarding-calendar">
        {["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"].map((day) => (
          <span className="calendar-weekday" key={day}>
            {day}
          </span>
        ))}
        {Array.from({ length: first }, (_, i) => (
          <span key={"blank-" + i} />
        ))}
        {Array.from({ length: count }, (_, i) => {
          const date = new Date(month.getFullYear(), month.getMonth(), i + 1);
          const titles = events
            .filter((e) => key(calendarDate(e.date)) === key(date))
            .map((e) => e.title);
          return (
            <div
              key={i}
              className={
                "calendar-day " + (key(date) === key(today) ? "today" : "")
              }
              title={titles.join(" · ")}
            >
              <span>{i + 1}</span>
              {titles.length > 0 && <i aria-label={titles.join(" · ")} />}
            </div>
          );
        })}
      </div>
      <p className="dashboard-help">
        Начало стажировки и выполненные шаги. Сроки задач уточняй у бадди.
      </p>
    </section>
  );
}

export function HomeScreen({
  buddy,
  onDocument,
}: {
  buddy: Buddy;
  onDocument: (doc: Document) => void;
}) {
  const d = buddy.data!,
    store = useNotes();
  const ask = (question: string) => {
    buddy.navigate("chat");
    buddy.setQuestion(question);
  };
  const next = d.checklist.tasks
    .filter((task) => task.status !== "completed")
    .slice(0, 3);
  return (
    <div className="page-content dashboard-page">
      <div className="dashboard-heading">
        <div>
          <h1>Привет, {d.profile.fullName.split(" ")[0]} 👋</h1>
          <p>
            {new Date().toLocaleDateString("ru-RU", {
              weekday: "long",
              day: "numeric",
              month: "long",
            })}
          </p>
        </div>
        <span className="team-label blue">
          {d.teams.find((team) => team.id === d.profile.teamId)?.name ||
            d.profile.teamId}
        </span>
      </div>
      <div className="dashboard-grid">
        <div className="dashboard-column">
          <section className="dashboard-card adaptation-card">
            <div className="dashboard-card-heading">
              <h2>Процесс адаптации</h2>
              <strong>{d.checklist.percent}%</strong>
            </div>
            <div
              className="adaptation-bar"
              role="progressbar"
              aria-label="Прогресс адаптации"
              aria-valuenow={d.checklist.percent}
              aria-valuemin={0}
              aria-valuemax={100}
            >
              <i style={{ width: `${d.checklist.percent}%` }} />
            </div>
            <p>
              Выполнено {d.checklist.completed} из {d.checklist.total} шагов.
            </p>
            <button
              className="dashboard-link"
              onClick={() => buddy.navigate("checklist")}
            >
              Открыть чек-лист <Icon name="arrow" size={16} />
            </button>
          </section>
          <section className="dashboard-card">
            <div className="dashboard-card-heading">
              <h2>Ближайшие шаги</h2>
              <Icon name="target" />
            </div>
            <div className="dashboard-task-list">
              {next.length ? (
                next.map((task) => (
                  <TaskCard
                    key={task.id}
                    task={task}
                    buddy={buddy}
                    onDocument={onDocument}
                  />
                ))
              ) : (
                <p className="dashboard-help">
                  Все шаги пройдены! Обсуди следующие цели со своим бадди.
                </p>
              )}
            </div>
          </section>
          <section className="dashboard-card">
            <h2>Быстрые вопросы</h2>
            <div className="dashboard-chips">
              {prompts.map((question) => (
                <button key={question} onClick={() => ask(question)}>
                  {question}
                  <Icon name="arrow" size={14} />
                </button>
              ))}
            </div>
          </section>
        </div>
        <div className="dashboard-column">
          <Calendar startDate={d.profile.startDate} tasks={d.checklist.tasks} />
          <section className="dashboard-card">
            <div className="dashboard-card-heading">
              <h2>Последние заметки</h2>
              <button
                className="dashboard-link"
                onClick={() => buddy.navigate("notes")}
              >
                Все <Icon name="arrow" size={15} />
              </button>
            </div>
            {store.error ? (
              <p className="inline-error" role="alert">
                {store.error}
              </p>
            ) : store.items === null ? (
              <p className="dashboard-help">Загружаем заметки…</p>
            ) : store.items.length ? (
              <div className="dashboard-notes">
                {store.items.slice(0, 3).map((note, i) => (
                  <button
                    key={note.id}
                    className={"dashboard-note note-color-" + i}
                    onClick={() => buddy.navigate("notes")}
                  >
                    <strong>{note.content.split("\n")[0]}</strong>
                    <span>
                      {note.content.split("\n").slice(1).join(" ") ||
                        "Открыть заметки"}
                    </span>
                    <small>{dateLabel(note.updatedAt)}</small>
                  </button>
                ))}
              </div>
            ) : (
              <button
                className="dashboard-empty-note"
                onClick={() => buddy.navigate("notes")}
              >
                <Icon name="plus" />
                Сохрани первую заметку
              </button>
            )}
          </section>
          <section className="dashboard-card dashboard-pet">
            <div>
              <h2>Твой напарник</h2>
              <p className="dashboard-help">
                Бобрик растёт вместе с твоим прогрессом.
              </p>
            </div>
            <Pet
              percent={d.checklist.percent}
              busy={buddy.updatingTask || buddy.generating}
            />
          </section>
        </div>
      </div>
    </div>
  );
}

export function ChecklistScreen({
  buddy,
  onDocument,
}: {
  buddy: Buddy;
  onDocument: (doc: Document) => void;
}) {
  const [filter, setFilter] = useState("all");
  const d = buddy.data!;
  const tasks = d.checklist.tasks.filter(
    (task) =>
      filter === "all" ||
      (filter === "done"
        ? task.status === "completed"
        : task.status !== "completed"),
  );
  return (
    <div className="page-content dashboard-page">
      <div className="page-heading">
        <span className="eyebrow">ТВОЙ ПУТЬ В КОМАНДЕ</span>
        <h1>Чек-лист адаптации</h1>
        <p>
          {d.checklist.completed} из {d.checklist.total} шагов выполнено.
          Прогресс сохраняется в твоём профиле.
        </p>
      </div>
      <div className="dashboard-grid">
        <section className="dashboard-card">
          <div className="segmented" aria-label="Фильтр задач">
            {[
              ["all", "Все шаги"],
              ["pending", "Осталось"],
              ["done", "Готово"],
            ].map(([id, title]) => (
              <button
                key={id}
                className={filter === id ? "selected" : ""}
                onClick={() => setFilter(id)}
              >
                {title}
              </button>
            ))}
          </div>
          <div className="dashboard-task-list">
            {tasks.map((task) => (
              <TaskCard
                key={task.id}
                task={task}
                buddy={buddy}
                onDocument={onDocument}
              />
            ))}
            {tasks.length === 0 && (
              <p className="dashboard-help">В этом списке пока нет шагов.</p>
            )}
          </div>
        </section>
        <div className="dashboard-column">
          <section className="dashboard-card dashboard-pet">
            <div>
              <h2>Твоя адаптация — {d.checklist.percent}%</h2>
              <p className="dashboard-help">
                Каждый выполненный шаг помогает освоиться.
              </p>
            </div>
            <Pet percent={d.checklist.percent} busy={buddy.updatingTask} />
          </section>
          <Calendar startDate={d.profile.startDate} tasks={d.checklist.tasks} />
        </div>
      </div>
    </div>
  );
}
