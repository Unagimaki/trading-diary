import { apiRequest } from "@/shared/api/http";

export type AnalyticsMetric = {
  available: boolean;
  value?: number;
  sampleSize: number;
  reason?: "role_missing" | "no_values" | "no_win_loss_values" | "no_negative_pnl";
};

export type DistributionValue = {
  value: string;
  count: number;
  percentage: number;
};

export type SelectDistribution = {
  columnId: string;
  columnName: string;
  total: number;
  values: DistributionValue[];
};

export type JournalAnalytics = {
  observationCount: number;
  selectDistributions: SelectDistribution[];
  metrics: {
    winRate: AnalyticsMetric;
    totalPnl: AnalyticsMetric;
    averagePnl: AnalyticsMetric;
    profitFactor: AnalyticsMetric;
    totalR: AnalyticsMetric;
    averageR: AnalyticsMetric;
  };
  dataQuality: {
    cleanRows: number;
    warningRows: number;
    issues: { rowId: string; codes: string[] }[];
  };
  trades: { wins: number; losses: number; breakeven: number };
};

export const analyticsApi = {
  get: (journalId: string) =>
    apiRequest<JournalAnalytics>(`/journals/${journalId}/analytics`),
};
