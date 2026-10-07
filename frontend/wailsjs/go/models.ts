export namespace api {
	
	export class AuthInfo {
	    username: string;
	    recoveryCode: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.recoveryCode = source["recoveryCode"];
	    }
	}
	export class Calculation {
	    principal: number;
	    startDate: string;
	    tenureDays: number;
	    days: number;
	    ratePercent: number;
	    interest: number;
	    maturityDate: string;
	    maturityAmount: number;
	
	    static createFrom(source: any = {}) {
	        return new Calculation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.principal = source["principal"];
	        this.startDate = source["startDate"];
	        this.tenureDays = source["tenureDays"];
	        this.days = source["days"];
	        this.ratePercent = source["ratePercent"];
	        this.interest = source["interest"];
	        this.maturityDate = source["maturityDate"];
	        this.maturityAmount = source["maturityAmount"];
	    }
	}
	export class ChangeCredentialsRequest {
	    currentPassword: string;
	    newUsername: string;
	    newPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new ChangeCredentialsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentPassword = source["currentPassword"];
	        this.newUsername = source["newUsername"];
	        this.newPassword = source["newPassword"];
	    }
	}
	export class CloseRequest {
	    fdNumber: string;
	    closureDate: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new CloseRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.closureDate = source["closureDate"];
	        this.remark = source["remark"];
	    }
	}
	export class ClosurePreview {
	    fdNumber: string;
	    closureDate: string;
	    isPremature: boolean;
	    closureType: string;
	    daysHeld: number;
	    ratePercent: number;
	    interest: number;
	    payable: number;
	    maturityDate: string;
	
	    static createFrom(source: any = {}) {
	        return new ClosurePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.closureDate = source["closureDate"];
	        this.isPremature = source["isPremature"];
	        this.closureType = source["closureType"];
	        this.daysHeld = source["daysHeld"];
	        this.ratePercent = source["ratePercent"];
	        this.interest = source["interest"];
	        this.payable = source["payable"];
	        this.maturityDate = source["maturityDate"];
	    }
	}
	export class DashboardStats {
	    totalFds: number;
	    activeFds: number;
	    activePrincipal: number;
	    totalInterest: number;
	    fyLabel: string;
	    fyDeposits: number;
	    maturingToday: number;
	    maturing7: number;
	    maturing30: number;
	    maturing90: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalFds = source["totalFds"];
	        this.activeFds = source["activeFds"];
	        this.activePrincipal = source["activePrincipal"];
	        this.totalInterest = source["totalInterest"];
	        this.fyLabel = source["fyLabel"];
	        this.fyDeposits = source["fyDeposits"];
	        this.maturingToday = source["maturingToday"];
	        this.maturing7 = source["maturing7"];
	        this.maturing30 = source["maturing30"];
	        this.maturing90 = source["maturing90"];
	    }
	}
	export class EditFDRequest {
	    fdNumber: string;
	    customerName: string;
	    customerNumber: string;
	    principal: number;
	    startDate: string;
	    tenureDays: number;
	
	    static createFrom(source: any = {}) {
	        return new EditFDRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.customerName = source["customerName"];
	        this.customerNumber = source["customerNumber"];
	        this.principal = source["principal"];
	        this.startDate = source["startDate"];
	        this.tenureDays = source["tenureDays"];
	    }
	}
	export class FD {
	    fdNumber: string;
	    customerName: string;
	    customerNumber: string;
	    principal: number;
	    startDate: string;
	    tenureDays: number;
	    interestRate: number;
	    maturityDate: string;
	    interestAmount: number;
	    maturityAmount: number;
	    status: string;
	    closureDate?: string;
	    closureType?: string;
	    closureRemark?: string;
	    closureRate?: number;
	    closureDays?: number;
	    closureInterest?: number;
	    closurePayable?: number;
	    renewedFrom?: string;
	    renewedTo?: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FD(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.customerName = source["customerName"];
	        this.customerNumber = source["customerNumber"];
	        this.principal = source["principal"];
	        this.startDate = source["startDate"];
	        this.tenureDays = source["tenureDays"];
	        this.interestRate = source["interestRate"];
	        this.maturityDate = source["maturityDate"];
	        this.interestAmount = source["interestAmount"];
	        this.maturityAmount = source["maturityAmount"];
	        this.status = source["status"];
	        this.closureDate = source["closureDate"];
	        this.closureType = source["closureType"];
	        this.closureRemark = source["closureRemark"];
	        this.closureRate = source["closureRate"];
	        this.closureDays = source["closureDays"];
	        this.closureInterest = source["closureInterest"];
	        this.closurePayable = source["closurePayable"];
	        this.renewedFrom = source["renewedFrom"];
	        this.renewedTo = source["renewedTo"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class FDDetail {
	    fd: FD;
	    history: domain.HistoryEntry[];
	
	    static createFrom(source: any = {}) {
	        return new FDDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fd = this.convertValues(source["fd"], FD);
	        this.history = this.convertValues(source["history"], domain.HistoryEntry);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ListRequest {
	    search: string;
	    filter: string;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ListRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.search = source["search"];
	        this.filter = source["filter"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	}
	export class ListResponse {
	    items: FD[];
	    total: number;
	    page: number;
	    pageSize: number;
	
	    static createFrom(source: any = {}) {
	        return new ListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], FD);
	        this.total = source["total"];
	        this.page = source["page"];
	        this.pageSize = source["pageSize"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LoginRequest {
	    username: string;
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.username = source["username"];
	        this.password = source["password"];
	    }
	}
	export class MaturityBucket {
	    period: string;
	    label: string;
	    count: number;
	    amount: number;
	
	    static createFrom(source: any = {}) {
	        return new MaturityBucket(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.period = source["period"];
	        this.label = source["label"];
	        this.count = source["count"];
	        this.amount = source["amount"];
	    }
	}
	export class PreviewRequest {
	    customerName: string;
	    customerNumber: string;
	    principal: number;
	    startDate: string;
	    tenureDays: number;
	
	    static createFrom(source: any = {}) {
	        return new PreviewRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.customerName = source["customerName"];
	        this.customerNumber = source["customerNumber"];
	        this.principal = source["principal"];
	        this.startDate = source["startDate"];
	        this.tenureDays = source["tenureDays"];
	    }
	}
	export class RenewRequest {
	    fdNumber: string;
	    mode: string;
	    startDate: string;
	    tenureDays: number;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new RenewRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.mode = source["mode"];
	        this.startDate = source["startDate"];
	        this.tenureDays = source["tenureDays"];
	        this.remark = source["remark"];
	    }
	}
	export class RenewResult {
	    previousFd: FD;
	    newFd: FD;
	
	    static createFrom(source: any = {}) {
	        return new RenewResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.previousFd = this.convertValues(source["previousFd"], FD);
	        this.newFd = this.convertValues(source["newFd"], FD);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ReopenRequest {
	    fdNumber: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new ReopenRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.remark = source["remark"];
	    }
	}
	export class ReportPreview {
	    kind: string;
	    headers: string[];
	    rows: string[][];
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new ReportPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.headers = source["headers"];
	        this.rows = source["rows"];
	        this.total = source["total"];
	    }
	}
	export class ReportRequest {
	    kind: string;
	    fromDate: string;
	    toDate: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new ReportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.fromDate = source["fromDate"];
	        this.toDate = source["toDate"];
	        this.path = source["path"];
	    }
	}
	export class ReportResult {
	    path: string;
	    rowCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ReportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rowCount = source["rowCount"];
	    }
	}
	export class ResetCredentialsRequest {
	    recoveryCode: string;
	    newUsername: string;
	    newPassword: string;
	
	    static createFrom(source: any = {}) {
	        return new ResetCredentialsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.recoveryCode = source["recoveryCode"];
	        this.newUsername = source["newUsername"];
	        this.newPassword = source["newPassword"];
	    }
	}
	export class ReverseRenewalRequest {
	    fdNumber: string;
	    remark: string;
	
	    static createFrom(source: any = {}) {
	        return new ReverseRenewalRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.remark = source["remark"];
	    }
	}
	export class SaveSlabsRequest {
	    slabs: domain.RateSlab[];
	
	    static createFrom(source: any = {}) {
	        return new SaveSlabsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slabs = this.convertValues(source["slabs"], domain.RateSlab);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SystemStatus {
	    ready: boolean;
	    error: string;
	    authed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SystemStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ready = source["ready"];
	        this.error = source["error"];
	        this.authed = source["authed"];
	    }
	}
	export class UpcomingFD {
	    fdNumber: string;
	    customerName: string;
	    principal: number;
	    maturityDate: string;
	    maturityAmount: number;
	    daysRemaining: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new UpcomingFD(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fdNumber = source["fdNumber"];
	        this.customerName = source["customerName"];
	        this.principal = source["principal"];
	        this.maturityDate = source["maturityDate"];
	        this.maturityAmount = source["maturityAmount"];
	        this.daysRemaining = source["daysRemaining"];
	        this.status = source["status"];
	    }
	}
	export class UpcomingRequest {
	    fromDate: string;
	    toDate: string;
	
	    static createFrom(source: any = {}) {
	        return new UpcomingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fromDate = source["fromDate"];
	        this.toDate = source["toDate"];
	    }
	}

}

export namespace brand {
	
	export class Info {
	    product: string;
	    version: string;
	    developer: string;
	    services: string;
	    website: string;
	    email: string;
	    phone: string;
	    copyright: string;
	    description: string;
	    developerBlurb: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.product = source["product"];
	        this.version = source["version"];
	        this.developer = source["developer"];
	        this.services = source["services"];
	        this.website = source["website"];
	        this.email = source["email"];
	        this.phone = source["phone"];
	        this.copyright = source["copyright"];
	        this.description = source["description"];
	        this.developerBlurb = source["developerBlurb"];
	    }
	}

}

export namespace domain {
	
	export class HistoryEntry {
	    id: number;
	    fdNumber: string;
	    eventDate: string;
	    eventType: string;
	    amount: number;
	    interest: number;
	    referenceFd: string;
	    remarks: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fdNumber = source["fdNumber"];
	        this.eventDate = source["eventDate"];
	        this.eventType = source["eventType"];
	        this.amount = source["amount"];
	        this.interest = source["interest"];
	        this.referenceFd = source["referenceFd"];
	        this.remarks = source["remarks"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class RateSlab {
	    id: number;
	    sortOrder: number;
	    minDays: number;
	    maxDays: number;
	    ratePercent: number;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new RateSlab(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sortOrder = source["sortOrder"];
	        this.minDays = source["minDays"];
	        this.maxDays = source["maxDays"];
	        this.ratePercent = source["ratePercent"];
	        this.label = source["label"];
	    }
	}

}

