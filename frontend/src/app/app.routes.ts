import { Routes } from '@angular/router';
import { DashboardComponent } from './features/dashboard/dashboard.component';
import { AccountListComponent } from './features/accounts/account-list.component';
import { TransactionListComponent } from './features/transactions/transaction-list.component';
import { LoanListComponent } from './features/loans/loan-list.component';
import { AssetListComponent } from './features/assets/asset-list.component';
import { ForecastPageComponent } from './features/forecast/forecast-page.component';
import { SePayComponent } from './features/sepay/sepay.component';
import { LoginComponent } from './features/auth/login.component';
import { RegisterComponent } from './features/auth/register.component';
import { authGuard } from './core/guards/auth.guard';
import { BudgetPageComponent } from './features/budgets/budget-page.component';
import { PropertyListComponent } from './features/properties/property-list.component';
import { AutomationRulesComponent } from './features/automation/automation-rules.component';
import { AssistantCommandsComponent } from './features/assistant/assistant-commands.component';
import { ForbiddenComponent } from './features/forbidden/forbidden.component';
import { NotFoundComponent } from './features/not-found/not-found.component';
import { PortfolioListComponent } from './features/portfolios/portfolio-list.component';
import { AuditLogsComponent } from './features/audit/audit-logs.component';
import { ProfileComponent } from './features/profile/profile.component';
import { DepositComponent } from './features/transactions/deposit.component';
import { ReportsComponent } from './features/reports/reports.component';
import { TransactionCreateComponent } from './features/transactions/transaction-create.component';
import { LoanCreateComponent } from './features/loans/loan-create.component';

export const appRoutes: Routes = [
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'dashboard', component: DashboardComponent, canActivate: [authGuard] },
  { path: 'accounts', component: AccountListComponent, canActivate: [authGuard] },
  { path: 'transactions', component: TransactionListComponent, canActivate: [authGuard] },
  { path: 'transactions/create', component: TransactionCreateComponent, canActivate: [authGuard] },
  { path: 'deposit', component: DepositComponent, canActivate: [authGuard] },
  { path: 'loans', component: LoanListComponent, canActivate: [authGuard] },
  { path: 'loans/create', component: LoanCreateComponent, canActivate: [authGuard] },
  { path: 'assets', component: AssetListComponent, canActivate: [authGuard] },
  { path: 'properties', component: PropertyListComponent, canActivate: [authGuard] },
  { path: 'forecast', component: ForecastPageComponent, canActivate: [authGuard] },
  { path: 'reports', component: ReportsComponent, canActivate: [authGuard] },
  { path: 'budgets', component: BudgetPageComponent, canActivate: [authGuard] },
  { path: 'budgets/:period', component: BudgetPageComponent, canActivate: [authGuard] },
  { path: 'sepay', component: SePayComponent, canActivate: [authGuard] },
  { path: 'automation', component: AutomationRulesComponent, canActivate: [authGuard] },
  { path: 'assistant', component: AssistantCommandsComponent, canActivate: [authGuard] },
  { path: 'audit-logs', component: AuditLogsComponent, canActivate: [authGuard] },
  { path: 'profile', component: ProfileComponent, canActivate: [authGuard] },
  { path: 'forbidden', component: ForbiddenComponent, canActivate: [authGuard] },
  { path: 'portfolios', component: PortfolioListComponent, canActivate: [authGuard] },
  { path: '**', component: NotFoundComponent },
];
