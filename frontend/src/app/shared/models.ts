export type Role = 'owner' | 'editor' | 'viewer';

export interface User {
  id: string;
  email: string;
  name: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface AuthResponse {
  token: string;
  user?: User;
  workspace?: Workspace;
}

export interface Workspace {
  id: string;
  name: string;
  baseCurrency: string;
  role?: Role;
  fiscalYearEnd?: string;
}

export interface NetWorthSummary {
  asOfAt: string;
  baseCurrency: string;
  netWorth: string;
  cash: string;
  liabilities: string;
  netWorthChange?: string;
  snapshotVersion?: number;
  assets?: {
    cash: string;
    receivables: string;
    property: string;
    otherAssets: string;
    accruedInterest: string;
  };
  dataQuality?: {
    reconciledAccounts: number;
    staleValuations: number;
    asOfSource: string;
  };
  attribution?: {
    externalCashFlow: string;
    accruedInterest: string;
    valuationChange: string;
    accruedFee?: string;
  };
}

export interface PortfolioSnapshotPage {
  items: NetWorthSummary[];
  nextCursor: string;
}

export type BankDirection = 'in' | 'out';
export type BankPostingState = 'pending_review' | 'auto_ready' | 'posted' | 'ignored';

export interface BankConnection {
  id: string;
  workspaceId: string;
  provider: string;
  externalId?: string;
  status: string;
  scope?: string;
  callbackState?: string;
  syncStatus?: string;
  lastSyncedAt?: string;
  lastSyncRequestedAt?: string;
}

export interface SePayConnectResponse {
  connectionId: string;
  provider: string;
  scope: string;
  externalId?: string;
  callbackState: string;
  connectUrl: string;
}

export interface BankFeedTransaction {
  id: string;
  workspaceId: string;
  connectionId: string;
  accountId: string;
  amount: string;
  currency: string;
  direction: BankDirection;
  counterparty?: string;
  description?: string;
  reference?: string;
  occurredAt: string;
  classificationEvidence?: string;
  classificationConfidence?: number;
  postingState: BankPostingState;
  postedTransactionId?: string;
  autoClassified?: boolean;
}

export interface Portfolio {
  id: string;
  workspaceId: string;
  name: string;
  baseCurrency: string;
}

export interface AuditLog {
  id: string;
  workspaceId: string;
  actorId: string;
  actorRole: string;
  action: string;
  targetType: string;
  targetId: string;
  requestId: string;
  path: string;
  method: string;
  policyDecision: string;
  result: string;
  reason?: string;
  correlationId?: string;
  beforeJson?: string;
  afterJson?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Account {
  id: string;
  workspaceId: string;
  portfolioId: string;
  name: string;
  type: string;
  currency: string;
}

export interface Transaction {
  id: string;
  workspaceId: string;
  accountId: string;
  categoryId?: string;
  portfolioId?: string;
  name?: string;
  type: 'income' | 'expense' | 'transfer' | 'investment_funding' | 'loan_disbursement' | 'loan_payment' | 'valuation_adjustment';
  amount: string;
  currency: string;
  note?: string;
  occurredAt: string;
  status?: 'draft' | 'pending' | 'posted' | 'reconciled' | 'voided' | string;
}

export interface TransactionListPage {
  items: Transaction[];
  nextCursor: string;
}

export interface Loan {
  id: string;
  workspaceId: string;
  direction: 'receivable' | 'payable';
  counterparty: string;
  principalInitial: string;
  principalBalance?: string;
  portfolioId?: string;
  startAt?: string;
  dueAt?: string;
  annualRate: string;
	  dailyRatePerMillion?: string;
	  settlementAccountId?: string;
  dayCountBasis?: string;
  status: string;
}

export interface LoanPayment {
  id: string;
  workspaceId: string;
  loanId: string;
  accountId?: string;
  transactionId?: string;
  principalAmount: string;
  interestAmount: string;
  feeAmount: string;
  waivedAmount: string;
  occurredAt: string;
}

export interface LoanAccruals {
  loanId: string;
  workspaceId: string;
  asOfAt: string;
  status: string;
  currency: string;
  principalInitial: string;
  principalBalance: string;
  annualRate: string;
	  dailyRatePerMillion?: string;
	  dailyInterest?: string;
	  lastInterestPaidDate?: string;
	  nextPaymentDate?: string;
  dayCountBasis: string;
  totalAccruedInterest: string;
  totalPaidInterest: string;
  unpaidInterest: string;
  accruals: Array<{
    periodStart: string;
    periodEnd: string;
    principal: string;
    accruedInterest: string;
    paidInterest: string;
    unpaidInterest: string;
    remainingPrincipal: string;
    days: number;
  }>;
}

export interface LoanPortfolioSummary {
  activePrincipal: string;
  dailyInterest: string;
  accruedInterest: string;
  paidInterest: string;
}

export interface LoanScheduleItem {
  loanId: string;
  borrower: string;
  paymentDate: string;
  cycleDays: number;
  expectedInterest: string;
  status: string;
}

export interface BudgetRow {
  categoryId: string;
  limit: string;
  spent: string;
  currency: string;
}

export interface BudgetPeriod {
  period: string;
  workspaceId: string;
  asOfAt: string;
  rows: BudgetRow[];
}

export interface Asset {
  id: string;
  workspaceId: string;
  portfolioId: string;
  name: string;
  assetType: string;
}

export interface ForecastScenario {
  id: string;
  workspaceId: string;
  name: string;
  status: string;
  assumptions: string;
  result?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface LoanPaymentRequest {
  id: string;
  loanId: string;
  paymentCode: string;
  amount: string;
  currency: string;
  expiresAt: string;
  status: string;
  qr?: string;
}

export interface BankAutomationRule {
  id: string;
  workspaceId: string;
  accountId: string;
  name: string;
  predicate?: string;
  actionType: string;
  direction?: string;
  type: string;
  categoryId?: string;
  priority: number;
  enabled: boolean;
  contentPattern?: string;
  referencePattern?: string;
  minAmount?: string;
  maxAmount?: string;
}

export interface AssistantCommand {
  id: string;
  workspaceId: string;
  userId: string;
  command: string;
  status: string;
  plan?: string;
  approvalId?: string;
  approvalExpiresAt?: string;
  approvalUsedAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface AssetValuation {
  id: string;
  assetId: string;
  valuationAmount: string;
  currency: string;
  source: string;
  effectiveAt: string;
}

export interface Property {
  id: string;
  workspaceId: string;
  portfolioId: string;
  name: string;
  address: string;
  areaM2: string;
  purchaseAt?: string;
}

export interface PropertyValuation {
  id: string;
  propertyId: string;
  valuationAmount: string;
  currency: string;
  source: string;
  effectiveAt: string;
}
