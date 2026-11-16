import { afterEach, expect, it, vi } from "vitest";
import { createUUID } from "./uuid";

afterEach(() => vi.unstubAllGlobals());

it("creates distinct valid UUIDs when randomUUID is unavailable on HTTP", () => {
  const crypto = globalThis.crypto;
  vi.stubGlobal("crypto", {
    getRandomValues: crypto.getRandomValues.bind(crypto),
  });
  const first = createUUID();
  const second = createUUID();
  expect(first).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  expect(second).not.toBe(first);
});
