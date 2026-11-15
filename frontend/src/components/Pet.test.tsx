import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { Pet } from "./Pet";

afterEach(() => vi.useRealTimers());

it("lets the user pet the buddy repeatedly", () => {
  vi.useFakeTimers();
  render(<Pet percent={20} busy={false} />);
  act(() => vi.advanceTimersByTime(1400));
  fireEvent.click(screen.getByRole("button", { name: "Погладить бобрика" }));
  expect(screen.getByRole("status").textContent).toBe(
    "Спасибо, что погладил! 💚",
  );
  act(() => vi.advanceTimersByTime(1000));
  fireEvent.click(screen.getByRole("button", { name: "Погладить бобрика" }));
  act(() => vi.advanceTimersByTime(1000));
  expect(screen.getByRole("status").textContent).toBe(
    "Спасибо, что погладил! 💚",
  );
  act(() => vi.advanceTimersByTime(400));
  expect(screen.getByRole("status").textContent).toBe("Скоро прогуляюсь");
});

it("walks automatically, supports pausing, and cleans up timers", () => {
  vi.useFakeTimers();
  const view = render(<Pet percent={50} busy={false} />);
  act(() => vi.advanceTimersByTime(1400));
  act(() => vi.advanceTimersByTime(6000));
  expect(screen.getByRole("status").textContent).toBe("Гуляю");
  fireEvent.click(screen.getByRole("button", { name: "Остановить прогулку" }));
  act(() => vi.advanceTimersByTime(20000));
  expect(screen.getByRole("status").textContent).toBe("Отдыхаю");
  fireEvent.click(screen.getByRole("button", { name: "Продолжить прогулку" }));
  act(() => vi.advanceTimersByTime(6000));
  expect(screen.getByRole("status").textContent).toBe("Гуляю");
  view.unmount();
  expect(vi.getTimerCount()).toBe(0);
});

it("stays with the current task while busy and keeps progress accessories", () => {
  vi.useFakeTimers();
  const view = render(<Pet percent={100} busy={true} />);
  act(() => vi.advanceTimersByTime(1400));
  act(() => vi.advanceTimersByTime(20000));
  expect(screen.getByRole("status").textContent).toBe("Думаю…");
  expect(screen.getByText("Бобрик: Выпускник")).toBeTruthy();
  view.rerender(<Pet percent={100} busy={false} />);
  act(() => vi.advanceTimersByTime(6000));
  expect(screen.getByRole("status").textContent).toBe("Гуляю");
});
