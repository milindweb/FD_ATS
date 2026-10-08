import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  GetMemberProfile,
  SaveMember,
  errorMessage,
  type Member,
} from "../lib/api";
import { useAsync } from "../lib/hooks";
import { Alert } from "../components/ui/Alert";
import { Button } from "../components/ui/Button";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { FormField, Input, Textarea } from "../components/ui/FormField";
import { ErrorState, Loading } from "../components/ui/States";
import { PageHeader } from "../components/ui/PageHeader";

export interface MemberFormPageProps {
  mode: "create" | "edit";
}

interface FormState {
  genNo: string;
  name: string;
  dob: string;
  mobile: string;
  email: string;
  presentAddress: string;
  permanentAddress: string;
  employerName: string;
  department: string;
  designation: string;
  tokenNo: string;
  nomineeName: string;
  nomineeRelationship: string;
  aadhaar: string;
  pan: string;
  bankName: string;
  accountNo: string;
  ifsc: string;
  profileRemarks: string;
}

const EMPTY_FORM: FormState = {
  genNo: "",
  name: "",
  dob: "",
  mobile: "",
  email: "",
  presentAddress: "",
  permanentAddress: "",
  employerName: "",
  department: "",
  designation: "",
  tokenNo: "",
  nomineeName: "",
  nomineeRelationship: "",
  aadhaar: "",
  pan: "",
  bankName: "",
  accountNo: "",
  ifsc: "",
  profileRemarks: "",
};

function formFromMember(m: Member): FormState {
  return {
    genNo: m.genNo,
    name: m.name,
    dob: m.dob,
    mobile: m.mobile,
    email: m.email,
    presentAddress: m.presentAddress,
    permanentAddress: m.permanentAddress,
    employerName: m.employerName,
    department: m.department,
    designation: m.designation,
    tokenNo: m.tokenNo,
    nomineeName: m.nomineeName,
    nomineeRelationship: m.nomineeRelationship,
    aadhaar: m.aadhaar,
    pan: m.pan,
    bankName: m.bankName,
    accountNo: m.accountNo,
    ifsc: m.ifsc,
    profileRemarks: m.profileRemarks,
  };
}

/** Add / Edit Member — field groups per SRS §50. */
export function MemberFormPage({ mode }: MemberFormPageProps) {
  const { memberId = "" } = useParams();
  const navigate = useNavigate();
  const profile = useAsync(
    () => (mode === "edit" ? GetMemberProfile(Number(memberId)) : Promise.resolve(null)),
    [mode, memberId],
  );
  const member = mode === "edit" ? (profile.data?.member ?? null) : null;

  const [form, setForm] = useState<FormState | null>(mode === "create" ? EMPTY_FORM : null);
  const [saving, setSaving] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof FormState, string>>>({});

  useEffect(() => {
    if (!member) return;
    setForm(formFromMember(member));
  }, [member]);

  const set = (key: keyof FormState) => (e: { target: { value: string } }) => {
    setForm((f) => (f ? { ...f, [key]: e.target.value } : f));
    setFieldErrors((fe) => ({ ...fe, [key]: undefined }));
  };

  const validate = (): boolean => {
    if (!form) return false;
    const errors: Partial<Record<keyof FormState, string>> = {};
    if (!form.genNo.trim()) errors.genNo = "GEN No. is required.";
    if (!form.name.trim()) errors.name = "Name is required.";
    setFieldErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const submit = async () => {
    setSubmitError(null);
    if (!form || !validate()) return;
    setSaving(true);
    try {
      const saved = await SaveMember({ ...form, id: mode === "edit" ? Number(memberId) : 0 });
      navigate(`/member/${saved.id}`);
    } catch (err: unknown) {
      setSubmitError(errorMessage(err));
      setSaving(false);
    }
  };

  const cancel = () => navigate(mode === "edit" ? `/member/${memberId}` : "/member");

  if (mode === "edit" && profile.loading) return <Loading />;
  if (mode === "edit" && (profile.error || !member)) {
    return <ErrorState message={profile.error || "Member not found."} onRetry={profile.reload} />;
  }
  if (!form) return <Loading />;

  return (
    <>
      <PageHeader
        title={mode === "edit" ? "Edit Member" : "Add Member"}
        context={mode === "edit" ? `${member!.genNo} — ${member!.name}` : "Create a new member record"}
        actions={<Button onClick={cancel}>Cancel</Button>}
      />

      {submitError && (
        <Alert tone="danger" title="Could not save the member">
          {submitError}
        </Alert>
      )}

      <Card>
        <CardHead title="Identification" />
        <CardBody>
          <div className="hs-field-grid">
            <FormField label="GEN No." required error={fieldErrors.genNo} htmlFor="member-gen" hint="Unique member ID">
              <Input
                id="member-gen"
                value={form.genNo}
                onChange={set("genNo")}
                invalid={!!fieldErrors.genNo}
                placeholder="e.g. GEN-001"
              />
            </FormField>
            <FormField label="Name" required error={fieldErrors.name} htmlFor="member-name">
              <Input
                id="member-name"
                value={form.name}
                onChange={set("name")}
                invalid={!!fieldErrors.name}
                placeholder="Full name"
              />
            </FormField>
            <FormField label="Date of Birth" htmlFor="member-dob">
              <Input id="member-dob" type="date" value={form.dob} onChange={set("dob")} />
            </FormField>
          </div>
        </CardBody>
      </Card>

      <div className="u-mt-5">
        <Card>
          <CardHead title="Personal Information" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Mobile No." htmlFor="member-mobile">
                <Input id="member-mobile" value={form.mobile} onChange={set("mobile")} placeholder="Mobile number" />
              </FormField>
              <FormField label="Email ID" htmlFor="member-email">
                <Input id="member-email" type="email" value={form.email} onChange={set("email")} placeholder="email@example.com" />
              </FormField>
            </div>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5">
        <Card>
          <CardHead title="Address" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Present Address" htmlFor="member-present">
                <Textarea id="member-present" rows={2} value={form.presentAddress} onChange={set("presentAddress")} />
              </FormField>
              <FormField label="Permanent Address" htmlFor="member-permanent">
                <Textarea id="member-permanent" rows={2} value={form.permanentAddress} onChange={set("permanentAddress")} />
              </FormField>
            </div>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5">
        <Card>
          <CardHead title="Employment Information" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Employer Name" htmlFor="member-employer">
                <Input id="member-employer" value={form.employerName} onChange={set("employerName")} />
              </FormField>
              <FormField label="Department" htmlFor="member-department">
                <Input id="member-department" value={form.department} onChange={set("department")} />
              </FormField>
              <FormField label="Designation" htmlFor="member-designation">
                <Input id="member-designation" value={form.designation} onChange={set("designation")} />
              </FormField>
              <FormField label="Token No." htmlFor="member-token">
                <Input id="member-token" value={form.tokenNo} onChange={set("tokenNo")} />
              </FormField>
            </div>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5">
        <Card>
          <CardHead title="Nominee Information" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Nominee Name" htmlFor="member-nominee">
                <Input id="member-nominee" value={form.nomineeName} onChange={set("nomineeName")} />
              </FormField>
              <FormField label="Nominee Relationship" htmlFor="member-nominee-rel">
                <Input id="member-nominee-rel" value={form.nomineeRelationship} onChange={set("nomineeRelationship")} />
              </FormField>
            </div>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5">
        <Card>
          <CardHead title="Government ID" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Aadhaar No." htmlFor="member-aadhaar">
                <Input id="member-aadhaar" value={form.aadhaar} onChange={set("aadhaar")} placeholder="12-digit Aadhaar" />
              </FormField>
              <FormField label="PAN No." htmlFor="member-pan">
                <Input id="member-pan" value={form.pan} onChange={set("pan")} placeholder="PAN number" />
              </FormField>
            </div>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5">
        <Card>
          <CardHead title="Banking Information" />
          <CardBody>
            <div className="hs-field-grid">
              <FormField label="Bank Name" htmlFor="member-bank">
                <Input id="member-bank" value={form.bankName} onChange={set("bankName")} />
              </FormField>
              <FormField label="Account No." htmlFor="member-account">
                <Input id="member-account" value={form.accountNo} onChange={set("accountNo")} />
              </FormField>
              <FormField label="IFSC Code" htmlFor="member-ifsc">
                <Input id="member-ifsc" value={form.ifsc} onChange={set("ifsc")} />
              </FormField>
            </div>
          </CardBody>
        </Card>
      </div>

      <div className="u-mt-5">
        <Card>
          <CardHead title="Profile Remarks" />
          <CardBody>
            <FormField label="Profile Remarks" htmlFor="member-remarks">
              <Textarea id="member-remarks" rows={3} value={form.profileRemarks} onChange={set("profileRemarks")} />
            </FormField>
          </CardBody>
        </Card>
      </div>

      {mode === "edit" && member && (
        <div className="u-mt-5">
          <Card>
            <CardHead title="System Information" />
            <CardBody>
              <dl className="hs-info-grid">
                <div className="hs-info">
                  <dt className="hs-info__label">System Member ID</dt>
                  <dd className="hs-info__value hs-mono">{member.id}</dd>
                </div>
                <div className="hs-info">
                  <dt className="hs-info__label">Created</dt>
                  <dd className="hs-info__value">{member.createdAt || "—"}</dd>
                </div>
                <div className="hs-info">
                  <dt className="hs-info__label">Updated</dt>
                  <dd className="hs-info__value">{member.updatedAt || "—"}</dd>
                </div>
              </dl>
            </CardBody>
          </Card>
        </div>
      )}

      <div className="u-mt-5 u-row u-gap-2">
        <Button variant="primary" onClick={submit} loading={saving}>
          Save Member
        </Button>
        <Button onClick={cancel}>Cancel</Button>
      </div>
    </>
  );
}
