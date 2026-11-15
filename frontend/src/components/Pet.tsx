import { useEffect, useState } from "react";
import idle from "../assets/pet/бобрик/idle.svg";
import think from "../assets/pet/бобрик/think.svg";
import jump from "../assets/pet/бобрик/jump.svg";
import run from "../assets/pet/бобрик/run.svg";
import belly from "../assets/pet/бобрик/belly.svg";
import hat from "../assets/pet/бобрик/hat.svg";
import tuxedo from "../assets/pet/бобрик/tuxedo.svg";
import cap from "../assets/pet/бобрик/cap.svg";

const moods = { idle, think, jump, run } as const;

type Mood = keyof typeof moods;

export const petStages = [
  "Новичок",
  "Освоился",
  "В шляпе",
  "Джентльмен",
  "Выпускник",
] as const;

// Питомец растёт вместе с чек-листом: 25% — животик, 50% — шляпа,
// 75% — смокинг, 100% — академическая шапочка. При смене стадии прыгает.
export function Pet({ percent, busy }: { percent: number; busy: boolean }) {
  const stage =
    percent >= 100
      ? 4
      : percent >= 75
        ? 3
        : percent >= 50
          ? 2
          : percent >= 25
            ? 1
            : 0;
  const [celebrate, setCelebrate] = useState(false);
  const [paused, setPaused] = useState(false);
  const [walking, setWalking] = useState(false);
  const [onRight, setOnRight] = useState(false);
  const [pets, setPets] = useState(0);
  useEffect(() => {
    setCelebrate(true);
    const timer = setTimeout(() => setCelebrate(false), 1400);
    return () => clearTimeout(timer);
  }, [stage, pets]);
  useEffect(() => {
    setWalking(false);
    if (paused || busy || celebrate) return;
    let resting: ReturnType<typeof setTimeout> | undefined;
    const stroll = setInterval(() => {
      setOnRight((right) => !right);
      setWalking(true);
      resting = setTimeout(() => setWalking(false), 1800);
    }, 6000);
    return () => {
      clearInterval(stroll);
      clearTimeout(resting);
    };
  }, [paused, busy, celebrate]);
  const mood: Mood = celebrate
    ? "jump"
    : busy
      ? "think"
      : walking
        ? "run"
        : "idle";
  const status = celebrate
    ? pets > 0
      ? "Спасибо, что погладил! 💚"
      : "Рад тебя видеть!"
    : busy
      ? "Думаю…"
      : paused
        ? "Отдыхаю"
        : walking
          ? "Гуляю"
          : "Скоро прогуляюсь";
  return (
    <figure className={"pet-figure pet-" + mood}>
      <button
        type="button"
        className={"pet-button" + (onRight ? " pet-at-right" : "")}
        aria-label="Погладить бобрика"
        title="Погладить бобрика"
        onClick={() => setPets((count) => count + 1)}
      >
        <div className="pet-frame" key={pets}>
          <img src={moods[mood]} alt="" width="120" height="132" />
          {stage >= 1 && (
            <img
              className={
                "pet-overlay" + (mood === "jump" ? " pet-overlay-jump" : "")
              }
              src={belly}
              alt=""
            />
          )}
          {stage >= 2 && stage < 4 && (
            <img className="pet-overlay" src={hat} alt="" />
          )}
          {stage >= 3 && <img className="pet-overlay" src={tuxedo} alt="" />}
          {stage >= 4 && <img className="pet-overlay" src={cap} alt="" />}
        </div>
      </button>
      <figcaption>Бобрик: {petStages[stage]}</figcaption>
      <p
        className="pet-status"
        role="status"
        aria-live={celebrate && pets > 0 ? "polite" : "off"}
      >
        {status}
      </p>
      <button
        type="button"
        className="pet-pause"
        aria-pressed={paused}
        onClick={() => setPaused((value) => !value)}
      >
        {paused ? "Продолжить прогулку" : "Остановить прогулку"}
      </button>
    </figure>
  );
}
