import { useEffect, useState } from "react";
import idle from "../assets/pet/бобрик/idle.svg";
import run from "../assets/pet/бобрик/run.svg";
import think from "../assets/pet/бобрик/think.svg";
import jump from "../assets/pet/бобрик/jump.svg";

const moods = { idle, run, think, jump } as const;

type Mood = keyof typeof moods;

// Цикл (~10с): идёт в сторону → стоит → прыжок. Каждый четвёртый круг
// вместо прыжка задумывается на 2 секунды. Пока бадди генерирует ответ — думает.
export function Pet({ busy }: { busy: boolean }) {
  const [mood, setMood] = useState<Mood>("idle");
  const [side, setSide] = useState<"a" | "b">("a");
  useEffect(() => {
    if (busy) {
      setMood("think");
      return;
    }
    let alive = true;
    let timer = 0;
    let round = 0;
    const later = (ms: number, fn: () => void) => {
      timer = window.setTimeout(() => {
        if (alive) fn();
      }, ms);
    };
    const step = () => {
      if (!alive) return;
      round++;
      setSide((s) => (s === "a" ? "b" : "a"));
      setMood("run");
      later(3500, () => {
        setMood("idle");
        later(1500, () => {
          if (round % 4 === 0) {
            setMood("think");
            later(2000, () => {
              setMood("idle");
              later(3500, step);
            });
          } else {
            setMood("jump");
            later(1000, () => {
              setMood("idle");
              later(4000, step);
            });
          }
        });
      });
    };
    later(1200, step);
    return () => {
      alive = false;
      clearTimeout(timer);
    };
  }, [busy]);
  return (
    <div className={"pet pet-" + mood + " pet-side-" + side} aria-hidden="true">
      <span className="pet-flip">
        <img src={moods[mood]} alt="" width="96" height="106" />
      </span>
    </div>
  );
}
