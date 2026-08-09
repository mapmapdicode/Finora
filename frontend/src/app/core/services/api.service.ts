import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import {
  Account,
  AssetValuation,
  Asset,
  BankConnection,
  AuditLog,
  BudgetPeriod,
  BankFeedTransaction,
  BankAutomationRule,
  AssistantCommand,
  Loan,
  LoanAccruals,
  LoanPayment,
  LoanPaymentRequest,
  LoanPortfolioSummary,
  LoanScheduleItem,
  NetWorthSummary,
  Portfolio,
  PortfolioSnapshotPage,
  SePayConnectResponse,
  Property,
  PropertyValuation,
  ForecastScenario,
  Transaction,
  TransactionListPage,
  Workspace,
  UserSettings,
  Customer,
  MySePaySummary,
} from '../../shared/models';

@Injectable({ providedIn: 'root' })
export class ApiService {
  private base = environment.apiBase;
  private readonly idempotencyNamespace = 'web-' + Math.floor(Math.random() * 1_000_000);

  private nextIdempotencyKey(prefix: string): string {
    if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
      return `${this.idempotencyNamespace}-${prefix}-${crypto.randomUUID()}`;
    }
    return `${this.idempotencyNamespace}-${prefix}-${Date.now().toString(36)}-${Math.floor(Math.random() * 1_000_000)}`;
  }

  private postWithIdempotency<T>(path: string, body: any, prefix: string): Observable<T> {
    const headers = new HttpHeaders().set('Idempotency-Key', this.nextIdempotencyKey(prefix));
    return this.http.post<T>(this.baseURL(path), body, { headers });
  }

  private putWithIdempotency<T>(path: string, body: any, prefix: string): Observable<T> {
    const headers = new HttpHeaders().set('Idempotency-Key', this.nextIdempotencyKey(prefix));
    return this.http.put<T>(this.baseURL(path), body, { headers });
  }

  constructor(private http: HttpClient) {}

  private baseURL(path: string) {
    return `${this.base}/api/v1${path}`;
  }

  getWorkspaces(): Observable<Workspace[]> {
    return this.http.get<Workspace[]>(this.baseURL('/workspaces'));
  }

  getPortfolios(): Observable<Portfolio[]> {
    return this.http.get<Portfolio[]>(this.baseURL('/portfolios'));
  }

  createPortfolio(payload: { name: string; baseCurrency: string }) {
    return this.postWithIdempotency<Portfolio>('/portfolios', payload, 'create-portfolio');
  }

  getPortfolioSnapshots(portfolioId: string, limit?: number, cursor?: string): Observable<PortfolioSnapshotPage> {
    const query = new URLSearchParams();
    if (typeof limit === 'number' && limit > 0) {
      query.set('limit', String(limit));
    }
    if (cursor) {
      query.set('cursor', cursor);
    }
    const qs = query.toString();
    return this.http.get<PortfolioSnapshotPage>(`${this.baseURL(`/portfolios/${portfolioId}/snapshots`)}${qs ? `?${qs}` : ''}`);
  }

  getNetWorth(portfolioId: string, asOfAt?: string): Observable<NetWorthSummary> {
    const query = new URLSearchParams();
    if (asOfAt && asOfAt.trim()) {
      query.set('asOf', asOfAt.trim());
    }
    const qs = query.toString();
    return this.http.get<NetWorthSummary>(`${this.baseURL(`/portfolios/${portfolioId}/net-worth`)}${qs ? `?${qs}` : ''}`);
  }

  getAccounts(): Observable<Account[]> {
    return this.http.get<Account[]>(this.baseURL('/accounts'));
  }

  createAccount(payload: Partial<Account>) {
    return this.postWithIdempotency<Account>('/accounts', payload, 'create-account');
  }

  deleteAccount(id: string) {
    return this.http.delete(this.baseURL(`/accounts/${id}`));
  }

  createTransfer(payload: {
    fromAccountId: string;
    toAccountId: string;
    amount: string;
    currency: string;
    note?: string;
    occurredAt?: string;
  }) {
    return this.postWithIdempotency('/transfers', payload, 'create-transfer');
  }

  getTransactions(options?: {
    accountId?: string;
    type?: string;
    status?: string;
    categoryId?: string;
    search?: string;
    from?: string;
    to?: string;
    cursor?: string;
    limit?: number;
  }): Observable<TransactionListPage> {
    const query = new URLSearchParams();
    if (options?.accountId?.trim()) {
      query.set('accountId', options.accountId.trim());
    }
    if (options?.type?.trim()) {
      query.set('type', options.type.trim());
    }
    if (options?.status?.trim()) {
      query.set('status', options.status.trim());
    }
    if (options?.categoryId?.trim()) {
      query.set('categoryId', options.categoryId.trim());
    }
    if (options?.search?.trim()) {
      query.set('search', options.search.trim());
    }
    if (options?.from?.trim()) {
      query.set('from', options.from.trim());
    }
    if (options?.to?.trim()) {
      query.set('to', options.to.trim());
    }
    if (options?.cursor?.trim()) {
      query.set('cursor', options.cursor.trim());
    }
    if (typeof options?.limit === 'number' && options.limit > 0) {
      query.set('limit', String(options.limit));
    }
    const qs = query.toString();
    return this.http.get<TransactionListPage>(`${this.baseURL('/transactions')}${qs ? `?${qs}` : ''}`);
  }

  createTransaction(payload: Partial<Transaction>) {
    return this.postWithIdempotency<Transaction>('/transactions', payload, 'create-transaction');
  }

  previewMarkdownImport(payload: { markdown: string; month: string; overwrite: boolean }) {
    return this.http.post<{
      month: string; overwrite: boolean; canCommit: boolean;
      summary: { accounts: number; transactions: number; loans: number; payments: number };
      issues: Array<{ line: number; section: string; message: string }>;
    }>(this.baseURL('/imports/markdown/preview'), payload);
  }

  commitMarkdownImport(payload: { markdown: string; month: string; overwrite: boolean }) {
    return this.http.post<{
      month: string;
      result: { accountsCreated: number; transactionsCreated: number; loansCreated: number; paymentsCreated: number; rowsSkipped: number };
    }>(this.baseURL('/imports/markdown/commit'), payload);
  }

  getLoans(): Observable<Loan[]> {
    return this.http.get<Loan[]>(this.baseURL('/loans'));
  }

  getCustomers(search = '', limit = 50): Observable<Customer[]> {
    const query = new URLSearchParams();
    if (search.trim()) {
      query.set('q', search.trim());
    }
    if (limit > 0) {
      query.set('limit', String(limit));
    }
    const qs = query.toString();
    return this.http.get<Customer[]>(`${this.baseURL('/customers')}${qs ? `?${qs}` : ''}`);
  }

  createCustomer(payload: { name: string; phone?: string }): Observable<Customer> {
    return this.postWithIdempotency<Customer>('/customers', payload, 'create-customer');
  }

  getLoanSummary(): Observable<LoanPortfolioSummary> {
    return this.http.get<LoanPortfolioSummary>(this.baseURL('/loans/summary'));
  }

  getLoanSchedule(months = 3): Observable<LoanScheduleItem[]> {
    return this.http.get<LoanScheduleItem[]>(this.baseURL(`/loans/schedule?months=${months}`));
  }

  getLoanAccruals(loanId: string): Observable<LoanAccruals> {
    return this.http.get<LoanAccruals>(this.baseURL(`/loans/${loanId}/accruals`));
  }

  getLoanPayments(loanId: string): Observable<LoanPayment[]> {
    return this.http.get<LoanPayment[]>(this.baseURL(`/loans/${loanId}/payments`));
  }

  createLoan(payload: Partial<Loan>) {
    return this.postWithIdempotency<Loan>('/loans', payload, 'create-loan');
  }

  deleteLoan(id: string) {
    return this.http.delete(this.baseURL(`/loans/${id}`));
  }

  createLoanPayment(loanId: string, payload: {
    principalAmount: string;
    interestAmount: string;
    feeAmount: string;
    waivedAmount: string;
    accountId?: string;
    occurredAt?: string;
  }) {
    return this.postWithIdempotency<LoanPayment>(`/loans/${loanId}/payments`, payload, 'create-loan-payment');
  }

  listProperties(): Observable<Property[]> {
    return this.http.get<Property[]>(this.baseURL('/properties'));
  }

  createProperty(payload: Partial<Property>) {
    return this.postWithIdempotency<Property>('/properties', payload, 'create-property');
  }

  deleteProperty(id: string) {
    return this.http.delete(this.baseURL(`/properties/${id}`));
  }

  addPropertyValuation(propertyId: string, payload: {
    valuationAmount: string;
    currency: string;
    source: string;
    effectiveAt?: string;
  }): Observable<PropertyValuation> {
    return this.postWithIdempotency<PropertyValuation>(
      `/properties/${propertyId}/valuations`,
      payload,
      'add-property-valuation',
    );
  }

  getAssets(): Observable<Asset[]> {
    return this.http.get<Asset[]>(this.baseURL('/assets'));
  }

  createAsset(payload: Partial<Asset>) {
    return this.postWithIdempotency<Asset>('/assets', payload, 'create-asset');
  }

  deleteAsset(id: string) {
    return this.http.delete(this.baseURL(`/assets/${id}`));
  }

  addAssetValuation(assetId: string, payload: {
    assetId?: string;
    valuationAmount: string;
    currency: string;
    source: string;
    effectiveAt?: string;
  }): Observable<AssetValuation> {
    return this.postWithIdempotency<AssetValuation>(
      `/assets/${assetId}/valuations`,
      payload,
      'add-asset-valuation',
    );
  }

  getBudget(period: string): Observable<BudgetPeriod> {
    return this.http.get<BudgetPeriod>(this.baseURL(`/budgets/${encodeURIComponent(period)}`));
  }

  upsertBudget(period: string, payload: { period?: string; categoryId?: string; limit: string }) {
    return this.putWithIdempotency<unknown>(`/budgets/${encodeURIComponent(period)}`, payload, 'upsert-budget');
  }

  createForecastScenario(payload: { name: string; assumptions: string }): Observable<ForecastScenario> {
    return this.postWithIdempotency<ForecastScenario>('/forecast-scenarios', payload, 'create-forecast');
  }

  listSePayConnections(): Observable<unknown[]> {
    return this.http.get<unknown[]>(this.baseURL('/bank-connections'));
  }

  getUserSettings(): Observable<UserSettings> {
    return this.http.get<UserSettings>(this.baseURL('/user/settings'));
  }

  updateUserSettings(payload: Pick<UserSettings, 'amountDisplayMode'>): Observable<UserSettings> {
    return this.http.put<UserSettings>(this.baseURL('/user/settings'), payload);
  }

  getMySePaySummary(): Observable<MySePaySummary> {
    return this.http.get<MySePaySummary>(this.baseURL('/me/sepay'));
  }

  listBankConnections(): Observable<BankConnection[]> {
    return this.http.get<BankConnection[]>(this.baseURL('/bank-connections'));
  }

  syncBankConnection(connectionId: string) {
    return this.postWithIdempotency(`/bank-connections/${connectionId}/sync`, {}, 'sync-bank-connection');
  }

  revokeBankConnection(connectionId: string) {
    return this.http.post(this.baseURL(`/bank-connections/${connectionId}/revoke`), {});
  }

  connectSePay(): Observable<SePayConnectResponse> {
    return this.postWithIdempotency<SePayConnectResponse>('/integrations/sepay/connect', {
      provider: 'sepay',
      scope: 'read_transactions',
    }, 'connect-sepay');
  }

  listBankFeed(): Observable<unknown[]> {
    return this.http.get<unknown[]>(this.baseURL('/bank-feed-transactions'));
  }

  listBankFeedTransactions(options?: { state?: string; postingState?: string; accountId?: string }): Observable<BankFeedTransaction[]> {
    const query = new URLSearchParams();
    const state = options?.state || options?.postingState;
    if (state) {
      query.set('state', state);
    }
    if (options?.accountId) {
      query.set('accountId', options.accountId);
    }
    const qs = query.toString();
    return this.http.get<BankFeedTransaction[]>(`${this.baseURL('/bank-feed-transactions')}${qs ? `?${qs}` : ''}`);
  }

  approveBankFeed(id: string) {
    return this.postWithIdempotency(`/bank-feed-transactions/${id}/approve`, {}, 'approve-bank-feed');
  }

  reclassifyBankFeed(id: string, payload: { type: string; accountId: string; categoryId: string; reason: string }) {
    return this.postWithIdempotency(`/bank-feed-transactions/${id}/reclassify`, payload, 'reclassify-bank-feed');
  }

  ignoreBankFeed(id: string) {
    return this.postWithIdempotency(`/bank-feed-transactions/${id}/ignore`, {}, 'ignore-bank-feed');
  }

  previewRule(payload: Record<string, unknown>): Observable<unknown> {
    return this.http.post<unknown>(this.baseURL('/bank-automation-rules/preview'), payload);
  }

  createAutomationRule(payload: Partial<BankAutomationRule>) {
    return this.postWithIdempotency<BankAutomationRule>('/bank-automation-rules', payload, 'create-bank-rule');
  }

  listAutomationRules(): Observable<BankAutomationRule[]> {
    return this.http.get<BankAutomationRule[]>(this.baseURL('/bank-automation-rules'));
  }

  updateAutomationRule(ruleId: string, payload: Partial<BankAutomationRule>) {
    return this.http.patch<BankAutomationRule>(this.baseURL(`/bank-automation-rules/${ruleId}`), payload);
  }

  deleteAutomationRule(ruleId: string) {
    return this.http.delete(this.baseURL(`/bank-automation-rules/${ruleId}`));
  }

  createLoanPaymentRequest(loanId: string, payload: {
    amount?: string;
    currency?: string;
    expiresAt?: string;
    note?: string;
  }): Observable<LoanPaymentRequest> {
    return this.postWithIdempotency<LoanPaymentRequest>(
      `/loans/${loanId}/payment-requests`,
      payload,
      'create-payment-request',
    );
  }

  createAssistantCommand(payload: { command: string; plan?: string }) {
    return this.http.post<AssistantCommand>(this.baseURL('/assistant/commands'), payload);
  }

  listAssistantCommands(): Observable<AssistantCommand[]> {
    return this.http.get<AssistantCommand[]>(this.baseURL('/assistant/commands'));
  }

  getAssistantCommand(id: string): Observable<AssistantCommand> {
    return this.http.get<AssistantCommand>(this.baseURL(`/assistant/commands/${id}`));
  }

  approveAssistantCommand(id: string, approvalId?: string) {
    const path = approvalId
      ? `/assistant/commands/${id}/approve?approvalId=${encodeURIComponent(approvalId)}`
      : `/assistant/commands/${id}/approve`;
    return this.http.post<AssistantCommand>(this.baseURL(path), {});
  }

  cancelAssistantCommand(id: string) {
    return this.http.post<AssistantCommand>(this.baseURL(`/assistant/commands/${id}/cancel`), {});
  }

  getForecastScenarios(): Observable<ForecastScenario[]> {
    return this.http.get<ForecastScenario[]>(this.baseURL('/forecast-scenarios'));
  }

  runForecastScenario(id: string) {
    return this.http.post<ForecastScenario>(this.baseURL(`/forecast-scenarios/${id}/run`), {});
  }

  listAuditLogs(): Observable<AuditLog[]> {
    return this.http.get<AuditLog[]>(this.baseURL('/audit-logs'));
  }

  deletePortfolio(id: string) {
    return this.http.delete(this.baseURL(`/portfolios/${id}`));
  }
}
