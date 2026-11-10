import { beforeEach, describe, expect, it, vi } from "vitest";
import { API } from "./api";
const response = (data: unknown, status = 200) =>
  new Response(status === 204 ? null : JSON.stringify(data), {
    status,
    headers: { "Content-Type": "application/json" },
  });
beforeEach(() => vi.restoreAllMocks());
describe("API transport", () => {
  it("rotates an expired token and retries the original request", async () => {
    const api = new API();
    api.token = "old";
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(response({}, 401))
      .mockResolvedValueOnce(response({ accessToken: "new" }))
      .mockResolvedValueOnce(response({ fullName: "Евгений" }));
    expect(await api.request("/profile")).toEqual({ fullName: "Евгений" });
    expect(api.token).toBe("new");
    expect(fetchMock.mock.calls[2][1]?.headers).toMatchObject({
      Authorization: "Bearer new",
    });
  });
  it("shares concurrent refreshes and clears failed sessions", async () => {
    const api = new API();
    let finish!: (r: Response) => void;
    const pending = new Promise<Response>((resolve) => {
      finish = resolve;
    });
    const fetchMock = vi.spyOn(globalThis, "fetch").mockReturnValue(pending);
    const first = api.refresh(),
      second = api.refresh();
    finish(response({ accessToken: "rotated" }));
    expect(await first).toBe(true);
    expect(await second).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    fetchMock.mockResolvedValue(
      response({ error: { message: "Войдите снова" } }, 401),
    );
    const expired = vi.fn();
    api.onExpired = expired;
    await expect(api.request("/profile")).rejects.toThrow("Войдите снова");
    expect(expired).toHaveBeenCalled();
    expect(api.token).toBe("");
  });
  it("handles failed networks, invalid error JSON, mutation bodies and empty responses", async () => {
    const api = new API();
    const f = vi
      .spyOn(globalThis, "fetch")
      .mockRejectedValueOnce(new Error("offline"));
    expect(await api.refresh()).toBe(false);
    f.mockResolvedValueOnce(new Response("offline", { status: 502 }));
    await expect(api.request("/profile")).rejects.toThrow("Сервис недоступен");
    f.mockResolvedValueOnce(response(null, 204));
    expect(
      await api.request(
        "/auth/logout",
        "POST",
        { confirmed: true },
        { "Idempotency-Key": "abc" },
      ),
    ).toBeUndefined();
    f.mockResolvedValueOnce(response({ accessToken: "login" }));
    await api.request("/auth/login", "POST");
    expect(api.token).toBe("login");
  });
  it("loads every page without losing message order", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(
        response({ items: [{ id: "one" }], nextCursor: 100 }),
      )
      .mockResolvedValueOnce(
        response({ items: [{ id: "two" }], nextCursor: null }),
      );
    expect((await new API().allMessages("chat")).map((m) => m.id)).toEqual([
      "one",
      "two",
    ]);
  });
});
