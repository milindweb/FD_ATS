import { AppInfo, errorMessage, type AppInfo as AppInfoData } from "../lib/api";
import { brand } from "../lib/brandInfo";
import { useAsync } from "../lib/hooks";
import { Alert } from "../components/ui/Alert";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";

function LinkRow({ label, value, href }: { label: string; value: string; href?: string }) {
  return (
    <div className="hs-info">
      <dt className="hs-info__label">{label}</dt>
      <dd className="hs-info__value">
        {href ? (
          <a href={href} target="_blank" rel="noreferrer">
            {value}
          </a>
        ) : (
          value
        )}
      </dd>
    </div>
  );
}

export function AboutPage() {
  const info = useAsync(() => AppInfo(), []);
  const data: AppInfoData | null = info.data;

  if (info.loading) return <Loading label="Loading…" />;

  return (
    <>
      <PageHeader title="About" context="Product and developer information" />

      {info.error && <Alert tone="warning">{errorMessage(info.error)}</Alert>}

      <div className="hs-two-col">
        <Card>
          <CardHead title="Product" />
          <CardBody>
            <dl className="hs-info-grid">
              <LinkRow label="Application" value={data?.product || brand.product} />
              <LinkRow label="Version" value={data?.version || brand.version} />
              <LinkRow label="Description" value={data?.description || brand.description} />
            </dl>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Developer" />
          <CardBody>
            <dl className="hs-info-grid">
              <LinkRow label="Name" value={data?.developer || brand.developer} />
              <LinkRow label="Services" value={data?.services || brand.services} />
              <LinkRow label="Website" value={data?.website || brand.website} href={data?.website || brand.website} />
              <LinkRow label="Email" value={data?.email || brand.email} href={`mailto:${data?.email || brand.email}`} />
              <LinkRow label="Phone" value={data?.phone || brand.phone} href={`tel:${data?.phone || brand.phone.replace(/\s/g, "")}`} />
            </dl>
            <p className="u-mt-3">{brand.developerBlurb}</p>
          </CardBody>
        </Card>
      </div>
    </>
  );
}
