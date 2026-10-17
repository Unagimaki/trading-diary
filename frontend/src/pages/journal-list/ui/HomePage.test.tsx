import { QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { journalsApi, type Journal } from "@/entities/journal/api/journals";
import { queryClient } from "@/shared/api/query-client";
import { rememberJournal } from "@/entities/journal/model/last-journal";
import { JournalListPage } from "./JournalListPage";

vi.mock("@/entities/user/api/auth", () => ({ authApi: {
  me: vi.fn().mockResolvedValue({ id: "u1", name: "Михаил" }), logout: vi.fn(),
} }));
vi.mock("@/entities/journal/api/journals", () => ({ journalsApi: { list: vi.fn(), create: vi.fn(), rename: vi.fn(), remove: vi.fn() } }));

const journal = (id: string, updatedAt: string): Journal => ({ id, name: `Журнал ${id}`, updatedAt, createdAt: updatedAt, initialDeposit: 1000, riskPercent: 1, defaultRR: 3 });
function renderHome() {
  render(<QueryClientProvider client={queryClient}><MemoryRouter><JournalListPage /></MemoryRouter></QueryClientProvider>);
}

describe("home page", () => {
  beforeEach(() => { queryClient.clear(); localStorage.clear(); vi.clearAllMocks(); });
  afterEach(() => { cleanup(); queryClient.clear(); localStorage.clear(); });

  it("offers the first journal, quick start, an example, and email feedback to a new user", async () => {
    vi.mocked(journalsApi.list).mockResolvedValue([]);
    renderHome();
    await screen.findByText("Здесь начнётся ваша история");
    expect(screen.getByRole("heading", { name: "Быстрый старт" })).toBeInTheDocument();
    expect(screen.getByText(/Пробои уровней/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Напишите мне" })).toHaveAttribute("href", expect.stringContaining("mailto:m.yark0v.2000@gmail.com"));
    fireEvent.click(screen.getByRole("button", { name: "Создать первый журнал" }));
    expect(screen.getByRole("dialog", { name: "Новый журнал" })).toBeInTheDocument();
  });

  it("continues in the last visited journal, even when another was modified later", async () => {
    rememberJournal("u1", "j1");
    vi.mocked(journalsApi.list).mockResolvedValue([journal("j1", "2026-10-01"), journal("j2", "2026-10-17")]);
    renderHome();
    expect(await screen.findByRole("link", { name: "Продолжить работу" })).toHaveAttribute("href", "/journals/j1");
  });

  it.each(["deleted", "other-user"])("falls back to an available journal when the saved journal belongs to $0", async (scenario) => {
    rememberJournal(scenario === "other-user" ? "u2" : "u1", "missing");
    vi.mocked(journalsApi.list).mockResolvedValue([journal("j1", "2026-10-01"), journal("j2", "2026-10-17")]);
    renderHome();
    expect(await screen.findByRole("link", { name: "Продолжить работу" })).toHaveAttribute("href", "/journals/j2");
  });
});
