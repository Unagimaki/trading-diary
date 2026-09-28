import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { BarChart3 } from "lucide-react";
import {
  analyticsApi,
  type AnalyticsMetric,
  type EquityPoint,
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
          <small>Проверьте отмеченные строки журнала и ручные значения.</small>
        </div>
      )}

      <EquityCurve points={report.equityCurve} />

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

function EquityCurve({ points }: { points: EquityPoint[] }) {
  const [activeIndex, setActiveIndex] = useState<number | null>(null);
  const width = 1000;
  const height = 280;
  const padding = { top: 20, right: 24, bottom: 38, left: 72 };

  if (points.length < 2) {
    return (
      <section className="equity-section">
        <div className="analytics-section-heading">
          <div><h2>Кривая депозита</h2><p>Изменение баланса после каждой рассчитанной сделки</p></div>
        </div>
        <div className="analytics-empty"><BarChart3 size={20} /><span>Добавьте сделки с рассчитанным PnL.</span></div>
      </section>
    );
  }

  const balances = points.map((point) => point.balance);
  const rawMin = Math.min(...balances);
  const rawMax = Math.max(...balances);
  const valuePadding = Math.max((rawMax - rawMin) * 0.12, Math.abs(rawMax || 1) * 0.01, 1);
  const min = rawMin - valuePadding;
  const max = rawMax + valuePadding;
  const chartWidth = width - padding.left - padding.right;
  const chartHeight = height - padding.top - padding.bottom;
  const x = (index: number) => padding.left + index * chartWidth / (points.length - 1);
  const y = (value: number) => padding.top + (max - value) * chartHeight / (max - min);
  const coordinates = points.map((point, index) => `${x(index)},${y(point.balance)}`).join(" ");
  const area = `${padding.left},${height - padding.bottom} ${coordinates} ${x(points.length - 1)},${height - padding.bottom}`;
  const ticks = Array.from({ length: 5 }, (_, index) => max - index * (max - min) / 4);
  const active = activeIndex === null ? null : points[activeIndex];
  const change = points.at(-1)!.balance - points[0].balance;

  return (
    <section className="equity-section">
      <div className="analytics-section-heading equity-heading">
        <div><h2>Кривая депозита</h2><p>Изменение баланса после каждой рассчитанной сделки</p></div>
        <div className={`equity-change ${change >= 0 ? "positive" : "negative"}`}>
          <span>Изменение</span><strong>{formatSigned(change)}</strong>
        </div>
      </div>
      <div className="equity-chart" onMouseLeave={() => setActiveIndex(null)}>
        <svg viewBox={`0 0 ${width} ${height}`} role="img" aria-label="График изменения депозита">
          {ticks.map((tick) => (
            <g key={tick}>
              <line className="equity-grid-line" x1={padding.left} x2={width - padding.right} y1={y(tick)} y2={y(tick)} />
              <text className="equity-axis-label" x={padding.left - 12} y={y(tick) + 4} textAnchor="end">{formatCompact(tick)}</text>
            </g>
          ))}
          <line className="equity-start-line" x1={padding.left} x2={width - padding.right} y1={y(points[0].balance)} y2={y(points[0].balance)} />
          <polygon className="equity-area" points={area} />
          <polyline className="equity-line" points={coordinates} />
          {points.map((point, index) => (
            <g key={`${point.tradeNumber}-${point.rowId ?? "start"}`} onMouseEnter={() => setActiveIndex(index)} onFocus={() => setActiveIndex(index)} tabIndex={0} aria-label={`Сделка ${point.tradeNumber}: баланс ${formatNumber(point.balance)}`}>
              <circle className="equity-hit" cx={x(index)} cy={y(point.balance)} r="13" />
              <circle className={`equity-point${activeIndex === index ? " active" : ""}`} cx={x(index)} cy={y(point.balance)} r={activeIndex === index ? 5 : 3.5} />
            </g>
          ))}
          <text className="equity-axis-label" x={padding.left} y={height - 10}>Старт</text>
          <text className="equity-axis-label" x={width - padding.right} y={height - 10} textAnchor="end">Сделка {points.length - 1}</text>
        </svg>
        {active && activeIndex !== null && (
          <div className="equity-tooltip" style={{ left: `${x(activeIndex) / width * 100}%`, top: `${y(active.balance) / height * 100}%` }}>
            <span>{active.tradeNumber === 0 ? "Старт" : `Сделка ${active.tradeNumber}`}</span>
            <strong>{formatNumber(active.balance)}</strong>
            {active.tradeNumber > 0 && <small className={active.pnl >= 0 ? "positive" : "negative"}>{formatSigned(active.pnl)}</small>}
          </div>
        )}
      </div>
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

function formatCompact(value: number) {
  return new Intl.NumberFormat("ru-RU", { notation: "compact", maximumFractionDigits: 1 }).format(value);
}

function formatSigned(value: number) {
  return `${value > 0 ? "+" : ""}${formatNumber(value)}`;
}
