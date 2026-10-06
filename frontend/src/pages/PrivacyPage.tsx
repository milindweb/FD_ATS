import { brand } from "../lib/brandInfo";
import { Card, CardBody, CardHead } from "../components/ui/Card";
import { PageHeader } from "../components/ui/PageHeader";

const collectedData = [
  { what: "Customer or member name and number", why: "Identify who the deposit belongs to" },
  { what: "Deposit amount and interest rate", why: "Calculate interest and maturity values" },
  { what: "Start, tenure and maturity dates", why: "Track schedules and report maturities" },
  { what: "Renewal and closure details", why: "Maintain the deposit's history" },
  { what: "Interest rate slab configuration", why: "Apply the rates you configure" },
];

const sections: Array<{ title: string; body: string }> = [
  {
    title: "Overview",
    body: "This application is a desktop tool that runs entirely on your computer. It is designed so that the deposit records you manage never leave your machine.",
  },
  {
    title: "How your information is used",
    body: "Your information is used only to perform the actions you request — creating deposits, calculating interest, renewing or closing them, and producing reports. Nothing is used for advertising, profiling or analytics of any kind.",
  },
  {
    title: "No internet transfer",
    body: "The application works fully offline. It does not upload data to any server, does not include telemetry or crash reporting, and makes no background network requests. The only time your browser or system opens an internet link is when you click a website, email or phone link yourself.",
  },
  {
    title: "Reports and exports",
    body: "Excel reports are written only to the folder you choose in the save dialog. The application never shares, uploads or auto-sends report files.",
  },
  {
    title: "Storage and location of data",
    body: "All records are stored in a local database file inside your user configuration folder (or the folder set through the FD_ATS_DATA_DIR environment variable). Only someone with access to your computer user account can read it.",
  },
  {
    title: "Security",
    body: "Data is protected by your operating system's user account controls. For additional protection, enable full-disk encryption on your computer and keep your account password strong.",
  },
  {
    title: "Backups",
    body: "You can create a backup of the database from Settings → Backup & Restore, and restore it later on this or another computer. Backups contain the same information as the live database — store them somewhere safe and private.",
  },
  {
    title: "Retention",
    body: "Data is kept for as long as the application and its database exist on your computer. Deleting the database file, or uninstalling the application and removing its data folder, permanently removes all records.",
  },
  {
    title: "Your rights",
    body: "Because the data is stored locally under your control, you can review, correct or delete any record at any time using the application, or remove the data entirely by deleting the database file.",
  },
  {
    title: "Changes to this policy",
    body: "Any future change to how this application handles data will be reflected on this page, with an updated date below.",
  },
];

export function PrivacyPage() {
  const updated = "Last updated: October 2026";

  return (
    <>
      <PageHeader title="Privacy Policy" context={`How this application handles your data · ${updated}`} />

      <Card>
        <CardHead title="Information collected and why" />
        <CardBody>
          <div className="hs-table-wrap is-responsive">
            <table className="hs-table">
              <thead>
                <tr>
                  <th scope="col">Information stored</th>
                  <th scope="col">Why it is needed</th>
                </tr>
              </thead>
              <tbody>
                {collectedData.map((row) => (
                  <tr key={row.what}>
                    <td>{row.what}</td>
                    <td>{row.why}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardBody>
      </Card>

      <Card className="u-mt-5">
        <CardBody>
          {sections.map((s) => (
            <section key={s.title} className="u-mb-4">
              <CardHead title={s.title} />
              <p>{s.body}</p>
            </section>
          ))}
          <section>
            <CardHead title="Contact" />
            <p>
              Questions about this policy or your data:{" "}
              <a href={`mailto:${brand.email}`}>{brand.email}</a> ·{" "}
              <a href={`tel:${brand.phone.replace(/\s/g, "")}`}>{brand.phone}</a>
            </p>
          </section>
        </CardBody>
      </Card>
    </>
  );
}
