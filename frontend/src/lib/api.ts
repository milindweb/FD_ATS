// api.ts — the single import surface for backend calls.
// Pages import from here only; generated Wails bindings stay an implementation
// detail, which keeps future API changes to a single file.

import {
  AppInfo,
  BackupDatabase,
  ChangeCredentials,
  CloseFD,
  CreateFD,
  DashboardStats,
  ExportReport,
  GetAuthInfo,
  GetFD,
  GetRateSlabs,
  ListFDs,
  LoadSampleData,
  Login,
  Logout,
  MaturityChart,
  PickBackupPath,
  PickReportPath,
  PickRestorePath,
  PreviewClosure,
  PreviewFD,
  PreviewReport,
  RenewFD,
  ResetCredentials,
  RestoreDatabase,
  SaveRateSlabs,
  SystemStatus,
  UpcomingMaturities,
} from "../../wailsjs/go/main/App";
import type { api, brand, domain } from "../../wailsjs/go/models";
import { api as apiModels } from "../../wailsjs/go/models";

export type Calculation = api.Calculation;
export type CloseRequest = api.CloseRequest;
export type ClosurePreview = api.ClosurePreview;
export type DashboardStats = api.DashboardStats;
export type FD = api.FD;
export type FDDetail = api.FDDetail;
export type ListRequest = api.ListRequest;
export type ListResponse = api.ListResponse;
export type PreviewRequest = api.PreviewRequest;
export type RenewRequest = api.RenewRequest;
export type RenewResult = api.RenewResult;
export type ReportRequest = api.ReportRequest;
export type ReportResult = api.ReportResult;
export type ReportPreview = api.ReportPreview;
export type MaturityBucket = api.MaturityBucket;
export type LoginRequest = api.LoginRequest;
export type ChangeCredentialsRequest = api.ChangeCredentialsRequest;
export type ResetCredentialsRequest = api.ResetCredentialsRequest;
export type AuthInfo = api.AuthInfo;
export type SystemStatus = api.SystemStatus;
export type UpcomingFD = api.UpcomingFD;
export type UpcomingRequest = api.UpcomingRequest;
export type RateSlab = domain.RateSlab;
export type HistoryEntry = domain.HistoryEntry;
export type AppInfo = brand.Info;

export {
  AppInfo,
  BackupDatabase,
  ChangeCredentials,
  CloseFD,
  CreateFD,
  DashboardStats,
  ExportReport,
  GetAuthInfo,
  GetFD,
  GetRateSlabs,
  ListFDs,
  LoadSampleData,
  Login,
  Logout,
  MaturityChart,
  PickBackupPath,
  PickReportPath,
  PickRestorePath,
  PreviewClosure,
  PreviewFD,
  PreviewReport,
  RenewFD,
  ResetCredentials,
  RestoreDatabase,
  SaveRateSlabs,
  SystemStatus,
  UpcomingMaturities,
};

/** Extracts a user-facing message from any thrown value (SRS §40). */
export function errorMessage(err: unknown): string {
  if (typeof err === "string") return err;
  if (err instanceof Error && err.message) return err.message;
  if (err && typeof err === "object" && "message" in err) {
    const msg = (err as { message?: unknown }).message;
    if (typeof msg === "string" && msg) return msg;
  }
  return "Something went wrong. Please try again.";
}

/** Replaces the rate slab configuration (SRS §30). */
export function SaveSlabs(slabs: RateSlab[]): Promise<void> {
  return SaveRateSlabs(apiModels.SaveSlabsRequest.createFrom({ slabs }));
}

/** Report kinds with their default file names. */export const reportKinds = [
  { kind: "REGISTER" as const, label: "FD Register", file: "FD_Register" },
  { kind: "MATURITY" as const, label: "Maturity Report", file: "Maturity_Report" },
  { kind: "ACTIVE" as const, label: "Active FD Report", file: "Active_FD_Report" },
  { kind: "CLOSED" as const, label: "Closed FD Report", file: "Closed_FD_Report" },
];

/** FD list filters (SRS §9.2). */
export const fdFilters = [
  { value: "ALL" as const, label: "All" },
  { value: "ACTIVE" as const, label: "Active" },
  { value: "CLOSED" as const, label: "Closed" },
  { value: "MATURING" as const, label: "Maturing" },
];
