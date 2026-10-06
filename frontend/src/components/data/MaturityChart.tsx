import { MaturityChart, type MaturityBucket } from "../../lib/api";
import { formatMoney } from "../../lib/format";
import { useAsync } from "../../lib/hooks";
import { Card, CardBody, CardHead } from "../ui/Card";
import { EmptyState, ErrorState, Loading } from "../ui/States";

function bucketTitle(b: MaturityBucket): string {
  const month = new Date(`${b.period}-01T00:00:00Z`).toLocaleDateString(undefined, {
    month: "long",
    year: "numeric",
  });
  return `${month}: ${b.count} FD${b.count === 1 ? "" : "s"} · ${formatMoney(b.amount)}`;
}

/** Dashboard card charting active FD maturities over the next 12 months. */
export function MaturityChartCard() {
  const chart = useAsync(() => MaturityChart(), []);

  const buckets = chart.data ?? [];
  const max = Math.max(1, ...buckets.map((b) => b.amount));
  const hasData = buckets.some((b) => b.count > 0);

  return (
    <Card>
      <CardHead
        title="Maturities by Month"
        actions={<span className="u-text-xs u-muted">Next 12 months</span>}
      />
      <CardBody>
        {chart.loading ? (
          <Loading />
        ) : chart.error ? (
          <ErrorState message={chart.error} onRetry={chart.reload} />
        ) : !hasData ? (
          <EmptyState
            title="No maturities in the next 12 months"
            description="Active FDs maturing within the next year appear here automatically."
          />
        ) : (
          <>
            <div className="hs-chart" role="img" aria-label="Bar chart of active FD maturities by month">
              {buckets.map((b) => (
                <div key={b.period} className="hs-chart__col" title={bucketTitle(b)}>
                  <span className="hs-chart__count">{b.count > 0 ? b.count : ""}</span>
                  <div className="hs-chart__barslot">
                    <div
                      className={`hs-chart__bar${b.count === 0 ? " hs-chart__bar--empty" : ""}`}
                      style={{ height: `${Math.max(2, Math.round((b.amount / max) * 100))}%` }}
                    />
                  </div>
                  <span className="hs-chart__label">{b.label}</span>
                </div>
              ))}
            </div>
            <div className="hs-chart__legend">
              <span>Bar height = maturity amount</span>
              <span>Number above a bar = FDs maturing that month</span>
            </div>
          </>
        )}
      </CardBody>
    </Card>
  );
}
