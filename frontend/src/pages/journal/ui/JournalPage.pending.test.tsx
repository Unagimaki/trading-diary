import { QueryClientProvider } from "@tanstack/react-query";
import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { queryClient } from "@/shared/api/query-client";
import { rowsApi } from "@/entities/observation/api/rows";
import { attachmentsApi } from "@/entities/attachment/api/attachments";
import { columnsApi, type ColumnType } from "@/entities/column/api/columns";
import { JournalPage } from "./JournalPage";

vi.mock("@/entities/journal/api/journals", () => ({ journalsApi: {
  get: vi.fn().mockResolvedValue({ id: "j1", name: "Журнал" }),
} }));
vi.mock("@/entities/column/api/columns", () => ({ columnsApi: {
  list: vi.fn().mockResolvedValue([{ id: "c1", name: "Скриншот", type: "image", role: null, options: [] }]),
} }));
vi.mock("@/entities/observation/api/rows", () => ({ rowsApi: {
  list: vi.fn(), create: vi.fn(), remove: vi.fn(), setCell: vi.fn(),
} }));
vi.mock("@/entities/attachment/api/attachments", () => ({ attachmentsApi: {
  remove: vi.fn(), contentUrl: () => "/image.png",
} }));

function deferred() {
  let resolve!: () => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<void>((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

function renderPage() {
  return render(<QueryClientProvider client={queryClient}><MemoryRouter initialEntries={["/journals/j1"]}>
    <Routes><Route path="/journals/:id" element={<JournalPage />} /></Routes>
  </MemoryRouter></QueryClientProvider>);
}

describe("journal pending actions", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    queryClient.clear();
    vi.mocked(rowsApi.list).mockResolvedValue([]);
    vi.mocked(columnsApi.list).mockResolvedValue([{ id: "c1", journalId: "j1", name: "Скриншот", type: "image", role: null, options: [], position: 0, createdAt: "", updatedAt: "" }]);
    vi.stubGlobal("ResizeObserver", class {
      observe() {}
      disconnect() {}
    });
  });
  afterEach(() => { cleanup(); queryClient.clear(); vi.unstubAllGlobals(); });

  const cellCases = (["text", "number", "date"] as ColumnType[]).flatMap((type) =>
    [undefined, null, ""].map((value) => ({ type, value })),
  );
  it.each(cellCases)("does not save untouched empty $type cells with value $value", async ({ type, value }) => {
    vi.mocked(columnsApi.list).mockResolvedValue([{ id: "c1", journalId: "j1", name: "Поле", type, role: null, options: [], position: 0, createdAt: "", updatedAt: "" }]);
    vi.mocked(rowsApi.list).mockResolvedValue([{ id: "r1", journalId: "j1", position: 0, warnings: [], createdAt: "", updatedAt: "", values: value === undefined ? {} : { c1: value } }]);
    const { container } = renderPage();
    await screen.findByTitle("Удалить строку");
    const input = container.querySelector<HTMLInputElement | HTMLTextAreaElement>(".cell-input")!;
    fireEvent.focus(input);
    fireEvent.blur(input);
    await act(async () => {});
    expect(rowsApi.setCell).not.toHaveBeenCalled();
    expect(rowsApi.list).toHaveBeenCalledTimes(1);
  });

  it.each([
    { value: undefined, draft: "Заметка", expected: "Заметка" },
    { value: "Заметка", draft: "", expected: null },
    { value: "Заметка", draft: "Заметка", expected: undefined },
  ])("saves text only when its content changes: $value → $draft", async ({ value, draft, expected }) => {
    vi.mocked(columnsApi.list).mockResolvedValue([{ id: "c1", journalId: "j1", name: "Текст", type: "text", role: null, options: [], position: 0, createdAt: "", updatedAt: "" }]);
    vi.mocked(rowsApi.list).mockResolvedValue([{ id: "r1", journalId: "j1", position: 0, warnings: [], createdAt: "", updatedAt: "", values: { c1: value } }]);
    vi.mocked(rowsApi.setCell).mockResolvedValue(undefined);
    renderPage();
    const input = await screen.findByRole("textbox");
    fireEvent.focus(input);
    fireEvent.change(input, { target: { value: draft } });
    fireEvent.blur(input);
    if (expected === undefined) {
      await act(async () => {});
      expect(rowsApi.setCell).not.toHaveBeenCalled();
    } else {
      await waitFor(() => expect(rowsApi.setCell).toHaveBeenCalledExactlyOnceWith("j1", "r1", "c1", expected));
    }
  });

  it("blocks both create buttons until the row list refresh completes", async () => {
    const request = deferred();
    const refresh = deferred();
    vi.mocked(rowsApi.create).mockImplementation(async () => { await request.promise; return {} as never; });
    renderPage();
    const emptyButton = await screen.findByRole("button", { name: "Добавить наблюдение" });
    vi.mocked(rowsApi.list).mockImplementation(async () => { await refresh.promise; return []; });
    fireEvent.click(emptyButton);
    await waitFor(() => expect(emptyButton).toBeDisabled());
    const headerButton = screen.getByRole("button", { name: "Наблюдение" });
    expect(headerButton).toBeDisabled();
    fireEvent.click(emptyButton);
    fireEvent.click(headerButton);
    expect(rowsApi.create).toHaveBeenCalledTimes(1);
    await act(async () => request.resolve());
    expect(emptyButton).toBeDisabled();
    await act(async () => refresh.resolve());
    await waitFor(() => expect(emptyButton).toBeEnabled());
  });

  it.each(["row", "image"])("blocks repeated %s deletion and unlocks after failure", async (target) => {
    const request = deferred();
    vi.mocked(rowsApi.list).mockResolvedValue([
      { id: "r1", journalId: "j1", position: 0, warnings: [], createdAt: "", updatedAt: "", values: { c1: { id: "a1", name: "Скрин.png" } } },
      { id: "r2", journalId: "j1", position: 1, warnings: [], createdAt: "", updatedAt: "", values: {} },
    ]);
    const remove = target === "row" ? rowsApi.remove : attachmentsApi.remove;
    vi.mocked(remove).mockReturnValue(request.promise);
    renderPage();
    await screen.findByTitle("Скрин.png");
    const button = target === "row" ? screen.getAllByTitle("Удалить строку")[0] : screen.getByRole("button", { name: "Удалить изображение" });
    fireEvent.click(button);
    await waitFor(() => expect(button).toBeDisabled());
    expect(button.closest("tr")).toHaveAttribute("inert");
    expect(screen.getAllByTitle("Удалить строку")[1]).toBeEnabled();
    fireEvent.click(button);
    expect(remove).toHaveBeenCalledTimes(1);
    await act(async () => request.reject(new Error("Ошибка запроса")));
    await waitFor(() => expect(button).toBeEnabled());
    expect(button.closest("tr")).not.toHaveAttribute("inert");
  });
});
