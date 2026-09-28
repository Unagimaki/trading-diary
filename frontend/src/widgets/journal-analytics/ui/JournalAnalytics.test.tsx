import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { JournalAnalytics } from "./JournalAnalytics";

describe("JournalAnalytics", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("renders metrics and select distributions", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            observationCount: 3,
            selectDistributions: [
              {
                columnId: "setup",
                columnName: "Сетап",
                total: 3,
                values: [{ value: "Breakout", count: 2, percentage: 66.67 }],
              },
            ],
            metrics: {
              winRate: { available: true, value: 66.67, sampleSize: 3 },
              totalPnl: { available: true, value: 110, sampleSize: 3 },
              averagePnl: { available: true, value: 36.67, sampleSize: 3 },
              profitFactor: { available: false, sampleSize: 3, reason: "no_negative_pnl" },
              totalR: { available: true, value: 3, sampleSize: 3 },
              averageR: { available: true, value: 1, sampleSize: 3 },
            },
            dataQuality: { cleanRows: 3, warningRows: 0, issues: [] },
            trades: { wins: 2, losses: 1, breakeven: 0 },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
      ),
    );
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });

    render(
      <QueryClientProvider client={client}>
        <JournalAnalytics journalId="journal-1" />
      </QueryClientProvider>,
    );

    expect(await screen.findByText("66,67%")).toBeInTheDocument();
    expect(screen.getByText("Сетап")).toBeInTheDocument();
    expect(screen.getByText("Для расчёта нужны отрицательные значения PnL")).toBeInTheDocument();
  });
});
