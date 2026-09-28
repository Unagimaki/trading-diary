import { useQuery } from "@tanstack/react-query";
import { BarChart3 } from "lucide-react";
import {
  analyticsApi,
  type AnalyticsMetric,
} from "@/entities/analytics/api/analytics";

const reasonLabels: Record<NonNullable<AnalyticsMetric["reason"]>, string> = {
  role_missing: "Добавьте колонку с подходящей системной ролью",
  no_values: "Заполните значения в системной колонке",
  no_win_loss_values: "Добавьте значения Win или Loss",
  no_negative_pnl: "Для расчёта нужны отрицательные значения PnL",
};

export function JournalAnalytics({ journalId }: { journalId: string }) {
  const analytics = useQuery({
    queryKey: ["analytics", journalId],
    queryFn: () => analyticsApi.get(journalId),
  });

  if (analytics.isPending) {
    return <div className="list-status">Загрузка аналитики…</div>;
  }
  if (analytics.isError) {
    return <div className="list-status error">Не удалось загрузить аналитику</div>;
  }

  const report = analytics.data;
  return (
    <section className="analytics-view" aria-label="Аналитика журнала">
      <div className="analytics-metrics">
        <MetricCard
          label="Наблюдения"
          metric={{ available: true, value: report.observationCount, sampleSize: report.observationCount }}
        />
        <MetricCard label="Winrate" metric={report.metrics.winRate} suffix="%" />
        <MetricCard label="Total PnL" metric={report.metrics.totalPnl} />
        <MetricCard label="Average PnL" metric={report.metrics.averagePnl} />
        <MetricCard label="Profit Factor" metric={report.metrics.profitFactor} />
        <MetricCard label="Total R" metric={report.metrics.totalR} suffix="R" />
        <MetricCard label="Average R" metric={report.metrics.averageR} suffix="R" />
        <CountCard label="Прибыльные сделки" value={report.trades.wins} />
        <CountCard label="Убыточные сделки" value={report.trades.losses} />
      </div>

      {report.dataQuality.warningRows > 0 && (
        <div className="quality-notice">
          <span>Расхождения в строках: {report.dataQuality.warningRows}</span>
          <small>Расчёт выполнен по приоритету R → PnL → Result. Проверьте отметки в журнале.</small>
        </div>
      )}

      <div className="analytics-section-heading">
        <div>
          <h2>Распределения</h2>
          <p>Значения категориальных колонок журнала</p>
        </div>
      </div>

      {report.selectDistributions.length === 0 ? (
        <div className="analytics-empty">
          <BarChart3 size={20} />
          <span>Добавьте колонку типа «Выбор», чтобы увидеть распределение.</span>
        </div>
      ) : (
        <div className="distribution-list">
          {report.selectDistributions.map((distribution) => (
            <section className="distribution" key={distribution.columnId}>
              <div className="distribution-heading">
                <h3>{distribution.columnName}</h3>
                <span>{distribution.total} заполнено</span>
              </div>
              <div className="distribution-values">
                {distribution.values.map((item) => (
                  <div className="distribution-row" key={item.value}>
                    <div className="distribution-label">
                      <span>{item.value}</span>
                      <strong>{item.count} · {formatNumber(item.percentage)}%</strong>
                    </div>
                    <div className="distribution-track" aria-hidden="true">
                      <span style={{ width: `${item.percentage}%` }} />
                    </div>
                  </div>
                ))}
              </div>
            </section>
          ))}
        </div>
      )}
    </section>
  );
}

function MetricCard({ label, metric, suffix = "" }: { label: string; metric: AnalyticsMetric; suffix?: string }) {
  const reason = metric.reason ? reasonLabels[metric.reason] : "Недостаточно данных";
  return (
    <article className={`metric-card${metric.available ? "" : " unavailable"}`}>
      <span>{label}</span>
      <strong>{metric.available && metric.value !== undefined ? `${formatNumber(metric.value)}${suffix}` : "—"}</strong>
      <small>{metric.available ? `${metric.sampleSize} значений` : reason}</small>
    </article>
  );
}

function CountCard({ label, value }: { label: string; value: number }) {
  return (
    <article className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
      <small>по расчётному результату</small>
    </article>
  );
}

function formatNumber(value: number) {
  return new Intl.NumberFormat("ru-RU", { maximumFractionDigits: 2 }).format(value);
}
