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
  remove: vi.fn(), upload: vi.fn(), contentUrl: () => "/image.png",
} }));
vi.mock("@/widgets/journal-analytics", () => ({ JournalAnalytics: () => <div>Аналитика журнала</div> }));

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

  it("keeps cells and analytics accessible while an image uploads, including after returning to the table", async () => {
    const request = deferred();
    vi.mocked(attachmentsApi.upload).mockImplementation(async () => { await request.promise; return {} as never; });
    vi.mocked(columnsApi.list).mockResolvedValue([
      { id: "c1", journalId: "j1", name: "Скриншот", type: "image", role: null, options: [], position: 0, createdAt: "", updatedAt: "" },
      { id: "c2", journalId: "j1", name: "Заметка", type: "text", role: null, options: [], position: 1, createdAt: "", updatedAt: "" },
    ]);
    vi.mocked(rowsApi.list).mockResolvedValue([{ id: "r1", journalId: "j1", position: 0, warnings: [], createdAt: "", updatedAt: "", values: {} }]);
    const { container } = renderPage();
    await screen.findByRole("textbox");
    const file = new File(["image"], "screen.png", { type: "image/png" });
    fireEvent.change(container.querySelector('input[type="file"]')!, { target: { files: [file] } });
    await waitFor(() => expect(container.querySelector('input[type="file"]')).toBeDisabled());
    expect(screen.getByRole("textbox").closest("tr")).not.toHaveAttribute("inert");
    expect(screen.getByTitle("Удалить строку")).toBeDisabled();
    fireEvent.change(screen.getByRole("textbox"), { target: { value: "Заметка во время загрузки" } });
    expect(screen.getByRole("textbox")).toHaveValue("Заметка во время загрузки");
    fireEvent.click(screen.getByRole("tab", { name: "Аналитика" }));
    expect(screen.getByText("Аналитика журнала")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("tab", { name: "Журнал" }));
    expect(container.querySelector('input[type="file"]')).toBeDisabled();
    expect(screen.getByText("Загрузка…")).toBeInTheDocument();
    expect(attachmentsApi.upload).toHaveBeenCalledTimes(1);
    await act(async () => request.resolve());
    await waitFor(() => expect(container.querySelector('input[type="file"]')).toBeEnabled());
    expect(screen.getByTitle("Удалить строку")).toBeEnabled();
  });

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
    if (target === "row") expect(button.closest("tr")).toHaveAttribute("inert");
    else expect(button.closest("tr")).not.toHaveAttribute("inert");
    expect(screen.getAllByTitle("Удалить строку")[1]).toBeEnabled();
    fireEvent.click(button);
    expect(remove).toHaveBeenCalledTimes(1);
    await act(async () => request.reject(new Error("Ошибка запроса")));
    await waitFor(() => expect(button).toBeEnabled());
    expect(button.closest("tr")).not.toHaveAttribute("inert");
  });

  it("saves different cells independently, queues saves of the same cell, and preserves unfinished input", async () => {
    const first = deferred();
    const second = deferred();
    vi.mocked(columnsApi.list).mockResolvedValue(["c1", "c2"].map((id, position) => ({
      id, journalId: "j1", name: id, type: "text", role: null, options: [], position, createdAt: "", updatedAt: "",
    })));
    vi.mocked(rowsApi.list).mockResolvedValue([{ id: "r1", journalId: "j1", position: 0, warnings: [], createdAt: "", updatedAt: "", values: {} }]);
    vi.mocked(rowsApi.setCell).mockImplementation(async (_journal, _row, column, next) => {
      if (column === "c1" && next === "Первое") await first.promise;
      if (column === "c1" && next === "Второе") await second.promise;
    });
    renderPage();
    const [a, b] = await screen.findAllByRole("textbox");
    fireEvent.focus(a);
    fireEvent.change(a, { target: { value: "Первое" } });
    fireEvent.blur(a);
    await waitFor(() => expect(rowsApi.setCell).toHaveBeenCalledTimes(1));
    expect(a.closest("tr")).not.toHaveAttribute("inert");
    expect(a).toBeEnabled();
    expect(b).toBeEnabled();
    expect(screen.getByRole("status")).toHaveTextContent("Сохранение…");
    fireEvent.focus(a);
    fireEvent.change(a, { target: { value: "Второе" } });
    fireEvent.blur(a);
    fireEvent.focus(b);
    fireEvent.change(b, { target: { value: "Другая ячейка" } });
    fireEvent.blur(b);
    await waitFor(() => expect(rowsApi.setCell).toHaveBeenCalledWith("j1", "r1", "c2", "Другая ячейка"));
    expect(rowsApi.setCell).not.toHaveBeenCalledWith("j1", "r1", "c1", "Второе");
    fireEvent.focus(a);
    fireEvent.change(a, { target: { value: "Ещё не закончено" } });
    await act(async () => first.resolve());
    await waitFor(() => expect(rowsApi.setCell).toHaveBeenCalledWith("j1", "r1", "c1", "Второе"));
    expect(a).toHaveValue("Ещё не закончено");
    await act(async () => second.resolve());
    await waitFor(() => expect(screen.getByRole("status")).toBeEmptyDOMElement());
    expect(a).toHaveValue("Ещё не закончено");
  });
});
