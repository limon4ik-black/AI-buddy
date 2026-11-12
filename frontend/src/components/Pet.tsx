import { useEffect, useState } from "react";
import idle from "../assets/pet/бобрик/idle.svg";
import think from "../assets/pet/бобрик/think.svg";
import jump from "../assets/pet/бобрик/jump.svg";
import belly from "../assets/pet/бобрик/belly.svg";
import hat from "../assets/pet/бобрик/hat.svg";
import tuxedo from "../assets/pet/бобрик/tuxedo.svg";
import cap from "../assets/pet/бобрик/cap.svg";

const moods = { idle, think, jump } as const;

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
    percent >= 100 ? 4 : percent >= 75 ? 3 : percent >= 50 ? 2 : percent >= 25 ? 1 : 0;
  const [celebrate, setCelebrate] = useState(false);
  useEffect(() => {
    setCelebrate(true);
    const timer = setTimeout(() => setCelebrate(false), 1400);
    return () => clearTimeout(timer);
  }, [stage]);
  const mood: Mood = celebrate ? "jump" : busy ? "think" : "idle";
  return (
    <figure className={"pet-figure pet-" + mood} aria-hidden="true">
      <div className="pet-frame">
        <img src={moods[mood]} alt="" width="120" height="132" />
        {stage >= 1 && (
          <img
            className={"pet-overlay" + (mood === "jump" ? " pet-overlay-jump" : "")}
            src={belly}
            alt=""
          />
        )}
        {stage >= 2 && stage < 4 && <img className="pet-overlay" src={hat} alt="" />}
        {stage >= 3 && <img className="pet-overlay" src={tuxedo} alt="" />}
        {stage >= 4 && <img className="pet-overlay" src={cap} alt="" />}
      </div>
      <figcaption>Бобрик: {petStages[stage]}</figcaption>
    </figure>
  );
}
