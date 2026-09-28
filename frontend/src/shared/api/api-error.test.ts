import { describe, expect, it } from "vitest";
import { responseError } from "./api-error";

describe("responseError", () => {
  it("preserves the request id and maps a public message", async () => {
    const response = new Response(
      JSON.stringify({
        code: "INTERNAL_ERROR",
        message: "internal error",
        requestId: "req-123",
      }),
      { status: 500, headers: { "Content-Type": "application/json" } },
    );
    const error = await responseError(response);
    expect(error.message).toBe(
      "Не удалось выполнить операцию. Повторите попытку",
    );
    expect(error.requestId).toBe("req-123");
  });
});
