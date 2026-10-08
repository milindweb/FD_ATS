import { useMemo } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { GetMemberProfile, type FD } from "../lib/api";
import { formatDate, formatMoney } from "../lib/format";
import { useAsync } from "../lib/hooks";
import { StatusBadge } from "../components/ui/Badge";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { DataTable, type Column } from "../components/data/DataTable";
import { KpiCard } from "../components/data/KpiCard";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";

function InfoItem({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="hs-info">
      <dt className="hs-info__label">{label}</dt>
      <dd className={`hs-info__value ${mono ? "hs-mono" : ""}`}>{value || "—"}</dd>
    </div>
  );
}

/** Member profile with §50 field groups and the member-wise FD summary (SRS §51.3). */
export function MemberProfilePage() {
  const { memberId = "" } = useParams();
  const navigate = useNavigate();
  const profile = useAsync(() => GetMemberProfile(Number(memberId)), [memberId]);
  const member = profile.data?.member ?? null;
  const fds = profile.data?.fds ?? [];

  const fdColumns = useMemo<Array<Column<FD>>>(
    () => [
      {
        key: "fdNumber",
        header: "FD Number",
        mono: true,
        render: (r) => <Link to={`/fd/${r.fdNumber}`}>{r.fdNumber}</Link>,
      },
      { key: "principal", header: "FD Amount", numeric: true, render: (r) => formatMoney(r.principal) },
      { key: "startDate", header: "Start Date", render: (r) => formatDate(r.startDate) },
      { key: "maturityDate", header: "Maturity Date", render: (r) => formatDate(r.maturityDate) },
      { key: "status", header: "Status", render: (r) => <StatusBadge status={r.status} detail={r.closureType} /> },
    ],
    [],
  );

  if (profile.loading) return <Loading />;
  if (profile.error || !member) {
    return <ErrorState message={profile.error || "Member not found."} onRetry={profile.reload} />;
  }

  return (
    <>
      <PageHeader
        title={member.name}
        context={<span className="hs-mono">{member.genNo}</span>}
        actions={
          <>
            <Button icon="arrow-left" onClick={() => navigate("/member")}>
              Back to Members
            </Button>
            <Button variant="primary" onClick={() => navigate(`/member/${member.id}/edit`)}>
              Edit
            </Button>
          </>
        }
      />

      <div className="hs-kpi-row">
        <KpiCard
          tone="info"
          icon="grid"
          label="Total Active FDs"
          value={String(member.activeFdCount)}
          meta={`${fds.length} FD${fds.length === 1 ? "" : "s"} linked`}
        />
        <KpiCard
          tone="success"
          icon="banknote"
          label="Total Active FD Amount"
          value={formatMoney(member.activeFdAmount)}
          meta="held in active FDs"
        />
      </div>

      <div className="hs-two-col">
        <Card>
          <CardHead title="Identification" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="GEN No." value={member.genNo} mono />
              <InfoItem label="System Member ID" value={String(member.id)} mono />
            </dl>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Personal Information" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Name" value={member.name} />
              <InfoItem label="Date of Birth" value={member.dob ? formatDate(member.dob) : ""} />
              <InfoItem label="Mobile No." value={member.mobile} />
              <InfoItem label="Email ID" value={member.email} />
            </dl>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5 hs-two-col">
        <Card>
          <CardHead title="Address" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Present Address" value={member.presentAddress} />
              <InfoItem label="Permanent Address" value={member.permanentAddress} />
            </dl>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Employment Information" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Employer Name" value={member.employerName} />
              <InfoItem label="Department" value={member.department} />
              <InfoItem label="Designation" value={member.designation} />
              <InfoItem label="Token No." value={member.tokenNo} mono />
            </dl>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5 hs-two-col">
        <Card>
          <CardHead title="Nominee Information" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Nominee Name" value={member.nomineeName} />
              <InfoItem label="Nominee Relationship" value={member.nomineeRelationship} />
            </dl>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="Government ID" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Aadhaar No." value={member.aadhaar} mono />
              <InfoItem label="PAN No." value={member.pan} mono />
            </dl>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5 hs-two-col">
        <Card>
          <CardHead title="Banking Information" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Bank Name" value={member.bankName} />
              <InfoItem label="Account No." value={member.accountNo} mono />
              <InfoItem label="IFSC Code" value={member.ifsc} mono />
            </dl>
          </CardBody>
        </Card>

        <Card>
          <CardHead title="System Information" />
          <CardBody>
            <dl className="hs-info-grid">
              <InfoItem label="Created" value={member.createdAt} />
              <InfoItem label="Updated" value={member.updatedAt} />
              <InfoItem label="Profile Remarks" value={member.profileRemarks} />
            </dl>
          </CardBody>
        </Card>
      </div>

      <Card className="u-mt-5">
        <CardHead
          title="Fixed Deposits"
          actions={<span className="u-text-xs u-muted">{fds.length} linked</span>}
        />
        <CardBody flush>
          <DataTable
            columns={fdColumns}
            rows={fds}
            rowKey={(r) => r.fdNumber}
            onRowClick={(r) => navigate(`/fd/${r.fdNumber}`)}
            emptyTitle="No FDs linked yet"
            emptyDescription="FDs created for this member appear here."
          />
        </CardBody>
      </Card>
    </>
  );
}
