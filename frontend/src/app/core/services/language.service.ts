import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

export type SupportedLanguage = 'vi' | 'en';

export interface TranslationDictionary {
  [key: string]: string | TranslationDictionary;
}

const VI_DICTIONARY: TranslationDictionary = {
  common: {
    search: 'Tìm kiếm...',
    workspace: 'Workspace',
    logout: 'Đăng xuất',
    login: 'Đăng nhập',
    register: 'Đăng ký',
    save: 'Lưu',
    create: 'Tạo mới',
    cancel: 'Hủy',
    edit: 'Sửa',
    delete: 'Xóa',
    refresh: 'Làm mới',
    filter: 'Bộ lọc',
    actions: 'Thao tác',
    status: 'Trạng thái',
    amount: 'Số tiền',
    currency: 'Đơn vị',
    type: 'Phân loại',
    name: 'Tên',
    description: 'Ghi chú',
    date: 'Thời gian',
    account: 'Tài khoản',
    portfolio: 'Danh mục',
    category: 'Hạng mục',
    snapshot: 'Lấy Snapshot',
    reset: 'Đặt lại',
    loading: 'Đang tải...',
    readonly_warning: 'Chế độ chỉ xem: Các hành vi thay đổi dữ liệu bị vô hiệu hóa.',
    all_accounts: 'Tất cả tài khoản',
    all_portfolios: 'Tất cả danh mục',
    all_types: 'Tất cả loại',
    load_more: 'Xem thêm',
    details: 'Chi tiết',
    close: 'Đóng',
    yes: 'Có',
    no: 'Không',
  },
  nav: {
    dashboard: 'Tổng quan',
    accounts: 'Tài khoản',
    transactions: 'Giao dịch',
    loans: 'Khoản vay & Nợ',
    assets: 'Tài sản đầu tư',
    properties: 'Bất động sản',
    budgets: 'Ngân sách tháng',
    forecast: 'Dự báo tài chính',
    portfolios: 'Danh mục đầu tư',
    audit: 'Nhật ký kiểm toán',
    sepay: 'Kết nối SePay',
    automation: 'Quy tắc tự động',
    assistant: 'Trợ lý AI',
    main_section: 'QUẢN LÝ CHÍNH',
    tools_section: 'CÔNG CỤ & AI',
  },
  dashboard: {
    title: 'Tổng quan tài chính',
    subtitle: 'Theo dõi vị thế tài sản ròng, phân bổ tài sản và lịch sử tăng trưởng.',
    net_worth: 'Tổng tài sản ròng',
    liquid_cash: 'Tiền mặt khả dụng',
    liabilities: 'Tổng nợ phải trả',
    valuation_delta: 'Biến động tài sản',
    growth_chart: 'Đồ thị tăng trưởng tài sản ròng',
    chart_hint: 'Bấm vào điểm mốc để soi chi tiết lịch sử danh mục.',
    asset_distribution: 'Phân bổ lớp tài sản',
    attribution_quality: 'Nguồn gốc biến động & Độ tin cậy',
    snapshot_records: 'Lịch sử Snapshot tài sản',
    cash_deposits: 'Tiền gửi ngân hàng & Tiền mặt',
    receivables_loans: 'Khoản cho vay & Phải thu',
    properties: 'Bất động sản',
    invested_assets: 'Tài sản đầu tư',
    external_cashflow: 'Dòng tiền bên ngoài',
    accrued_interest: 'Lãi dồn tích',
    revaluation: 'Điều chỉnh định giá',
    reconciled_accounts: 'Tài khoản đối soát',
    stale_valuations: 'Định giá quá hạn',
  },
  accounts: {
    title: 'Quản lý tài khoản',
    subtitle: 'Quản lý ví tiền mặt, tài khoản ngân hàng và thẻ tín dụng.',
    create_title: 'Thêm tài khoản mới',
    create_desc: 'Thêm tài khoản tài chính để bắt đầu ghi nhận giao dịch.',
    name_label: 'Tên tài khoản',
    type_label: 'Loại tài khoản',
    currency_label: 'Đơn vị tiền tệ',
    portfolio_label: 'Danh mục chứa',
    table_name: 'Tên tài khoản',
    table_type: 'Loại hình',
    table_currency: 'Đơn vị',
    table_id: 'Mã tài khoản',
    cash_type: 'Ví tiền mặt',
    bank_type: 'Tài khoản ngân hàng',
    card_type: 'Thẻ tín dụng / Ghi nợ',
  },
  transactions: {
    title: 'Sổ nhật ký giao dịch',
    subtitle: 'Ghi nhận thu nhập, chi phí và chuyển tiền giữa các tài khoản.',
    standard_mode: 'Ghi nhận giao dịch',
    transfer_mode: 'Chuyển tiền nội bộ',
    new_entry: 'Tạo giao dịch mới',
    new_transfer: 'Tạo lệnh chuyển tiền nội bộ',
    source_account: 'Tài khoản nguồn (Rút)',
    destination_account: 'Tài khoản đích (Nhận)',
    type_expense: 'Chi phí',
    type_income: 'Thu nhập',
    type_valuation: 'Điều chỉnh định giá',
    note_label: 'Ghi chú / Diễn giải',
    record_button: 'Ghi nhận giao dịch',
    transfer_button: 'Thực hiện chuyển tiền',
  },
  loans: {
    title: 'Khoản vay & Hợp đồng nợ',
    subtitle: 'Theo dõi cho vay (phải thu), đi vay (phải trả) và tính lãi dồn tích.',
    new_contract: 'Tạo hợp đồng nợ mới',
    counterparty: 'Đối tác / Người vay / Người cho vay',
    direction: 'Chiều dòng tiền',
    receivable: 'Cho vay (Phải thu)',
    payable: 'Đi vay (Phải trả)',
    principal: 'Số tiền gốc',
    annual_rate: 'Lãi suất năm (%)',
    disbursement_account: 'Tài khoản giải ngân',
    register_button: 'Đăng ký hợp đồng nợ',
  },
  assets: {
    title: 'Tài sản đầu tư',
    subtitle: 'Theo dõi cổ phiếu, vàng, tiền mã hóa và tài sản giá trị.',
    register_title: 'Đăng ký tài sản đầu tư',
    asset_name: 'Tên tài sản / Mã chứng khoán',
    asset_category: 'Phân loại tài sản',
    quantity: 'Số lượng sở hữu',
    unit: 'Đơn vị tính',
    register_button: 'Đăng ký sở hữu',
    add_valuation: 'Thêm định giá',
  },
  properties: {
    title: 'Bất động sản',
    subtitle: 'Quản lý đất đai, nhà ở, căn hộ và định giá tài sản bất động sản.',
    add_title: 'Thêm bất động sản mới',
    address: 'Địa chỉ / Vị trí',
    area: 'Diện tích (m²)',
    create_button: 'Tạo bất động sản',
  },
  budgets: {
    title: 'Ngân sách chi tiêu tháng',
    subtitle: 'Thiết lập hạn mức chi tiêu tháng và theo dõi tình hình thực tế.',
    period: 'Kỳ ngân sách (YYYY-MM)',
    spending_limit: 'Hạn mức chi tiêu',
    save_button: 'Lưu hạn mức ngân sách',
  },
  forecast: {
    title: 'Dự báo & Mô phỏng tài chính',
    subtitle: 'Chạy mô phỏng tốc độ tăng trưởng tài sản và kịch bản tài chính.',
    create_title: 'Tạo kịch bản mô phỏng',
    assumptions: 'Giả định mô phỏng (Dạng JSON)',
    run_button: 'Chạy mô phỏng',
  },
  portfolios: {
    title: 'Danh mục đầu tư',
    subtitle: 'Gom nhóm tài khoản, tài sản và bất động sản vào các danh mục.',
    create_title: 'Tạo danh mục mới',
    base_currency: 'Đơn vị gốc báo cáo',
  },
  audit: {
    title: 'Nhật ký kiểm toán hệ thống',
    subtitle: 'Lịch sử thao tác an toàn, đối soát dữ liệu và phân tích người dùng.',
  },
  sepay: {
    title: 'Kết nối ngân hàng SePay',
    subtitle: 'Nhận dữ liệu webhook ngân hàng tự động và phân loại sổ sách.',
    connect_button: 'Kết nối SePay Ngân Hàng',
  },
  automation: {
    title: 'Quy tắc tự động hóa',
    subtitle: 'Cấu hình quy tắc khớp Regex và ưu tiên tự động phân loại giao dịch.',
  },
  assistant: {
    title: 'Trợ lý tài chính AI Hermes',
    subtitle: 'Gửi câu lệnh ngôn ngữ tự nhiên và duyệt kế hoạch thực thi.',
  },
};

const EN_DICTIONARY: TranslationDictionary = {
  common: {
    search: 'Search workspace...',
    workspace: 'Workspace',
    logout: 'Logout',
    login: 'Login',
    register: 'Register',
    save: 'Save',
    create: 'Create',
    cancel: 'Cancel',
    edit: 'Edit',
    delete: 'Delete',
    refresh: 'Refresh',
    filter: 'Filter',
    actions: 'Actions',
    status: 'Status',
    amount: 'Amount',
    currency: 'Currency',
    type: 'Type',
    name: 'Name',
    description: 'Description',
    date: 'Timestamp',
    account: 'Account',
    portfolio: 'Portfolio',
    category: 'Category',
    snapshot: 'Take Snapshot',
    reset: 'Reset',
    loading: 'Loading...',
    readonly_warning: 'Read-only workspace: Data modification actions are disabled.',
    all_accounts: 'All Accounts',
    all_portfolios: 'All Portfolios',
    all_types: 'All Types',
    load_more: 'Load More',
    details: 'Details',
    close: 'Close',
    yes: 'Yes',
    no: 'No',
  },
  nav: {
    dashboard: 'Dashboard',
    accounts: 'Accounts',
    transactions: 'Transactions',
    loans: 'Loans & Debt',
    assets: 'Investment Assets',
    properties: 'Properties',
    budgets: 'Monthly Budgets',
    forecast: 'Financial Forecast',
    portfolios: 'Portfolios',
    audit: 'Audit Logs',
    sepay: 'SePay Integration',
    automation: 'Automation Rules',
    assistant: 'AI Assistant',
    main_section: 'MAIN NAVIGATION',
    tools_section: 'TOOLS & AI',
  },
  dashboard: {
    title: 'Financial Overview',
    subtitle: 'Track real-time net worth, asset allocation, and portfolio growth trajectory.',
    net_worth: 'Total Net Worth',
    liquid_cash: 'Liquid Cash',
    liabilities: 'Total Liabilities',
    valuation_delta: 'Valuation Delta',
    growth_chart: 'Net Worth Growth Trajectory',
    chart_hint: 'Click on any node to drill down into historical portfolio composition.',
    asset_distribution: 'Asset Class Distribution',
    attribution_quality: 'Return Attribution & Data Quality',
    snapshot_records: 'Snapshot Records',
    cash_deposits: 'Bank Deposits & Cash',
    receivables_loans: 'Receivables & Loans',
    properties: 'Properties',
    invested_assets: 'Invested Assets',
    external_cashflow: 'External Cash Flow',
    accrued_interest: 'Accrued Interest',
    revaluation: 'Re-valuation Adjustment',
    reconciled_accounts: 'Reconciled Accounts',
    stale_valuations: 'Stale Valuations',
  },
  accounts: {
    title: 'Accounts Management',
    subtitle: 'Manage cash wallets, bank accounts, and credit cards.',
    create_title: 'Create New Account',
    create_desc: 'Add a financial account to start recording transactions.',
    name_label: 'Account Name',
    type_label: 'Account Type',
    currency_label: 'Currency',
    portfolio_label: 'Portfolio Container',
    table_name: 'Account Name',
    table_type: 'Type',
    table_currency: 'Currency',
    table_id: 'Account ID',
    cash_type: 'Cash Wallet',
    bank_type: 'Bank Account',
    card_type: 'Credit / Debit Card',
  },
  transactions: {
    title: 'Transactions Ledger',
    subtitle: 'Record income, expense, and inter-account transfers.',
    standard_mode: 'Standard Entry',
    transfer_mode: 'Account Transfer',
    new_entry: 'New Ledger Entry',
    new_transfer: 'New Account Transfer',
    source_account: 'Source Account (Outflow)',
    destination_account: 'Destination Account (Inflow)',
    type_expense: 'Expense',
    type_income: 'Income',
    type_valuation: 'Valuation Adjustment',
    note_label: 'Description / Note',
    record_button: 'Record Transaction',
    transfer_button: 'Execute Transfer',
  },
  loans: {
    title: 'Loans & Debt Agreements',
    subtitle: 'Track receivables, payables, and interest accruals.',
    new_contract: 'New Loan Agreement',
    counterparty: 'Counterparty Entity',
    direction: 'Flow Direction',
    receivable: 'Receivable (I lent money)',
    payable: 'Payable (I borrowed money)',
    principal: 'Principal Amount',
    annual_rate: 'Annual Rate (%)',
    disbursement_account: 'Disbursement Account',
    register_button: 'Register Loan Contract',
  },
  assets: {
    title: 'Investment Assets',
    subtitle: 'Track stocks, gold, crypto, luxury assets, and custom holdings.',
    register_title: 'Register Investment Asset',
    asset_name: 'Asset Name / Symbol',
    asset_category: 'Asset Category',
    quantity: 'Quantity',
    unit: 'Unit Code',
    register_button: 'Register Holding',
    add_valuation: 'Add Valuation',
  },
  properties: {
    title: 'Real Estate Properties',
    subtitle: 'Manage land, residential properties, and appraisals.',
    add_title: 'Add Property Listing',
    address: 'Full Address / Location',
    area: 'Area (m²)',
    create_button: 'Create Property',
  },
  budgets: {
    title: 'Monthly Budgets',
    subtitle: 'Set monthly spending targets and monitor actual expenses.',
    period: 'Budget Period (YYYY-MM)',
    spending_limit: 'Spending Limit',
    save_button: 'Save Budget Limit',
  },
  forecast: {
    title: 'Financial Forecast',
    subtitle: 'Run growth rate projections and scenario modeling.',
    create_title: 'Create Simulation Scenario',
    assumptions: 'Simulation Assumptions (JSON)',
    run_button: 'Run Simulation',
  },
  portfolios: {
    title: 'Investment Portfolios',
    subtitle: 'Group accounts, assets, and properties into portfolios.',
    create_title: 'Create New Portfolio',
    base_currency: 'Base Reporting Currency',
  },
  audit: {
    title: 'System Audit Logs',
    subtitle: 'Security, operational logs, and data integrity history.',
  },
  sepay: {
    title: 'SePay Bank Integration',
    subtitle: 'Realtime bank webhooks and automated ledger reconciliation.',
    connect_button: 'Connect SePay Bank',
  },
  automation: {
    title: 'Automation Rules',
    subtitle: 'Configure regex rules to automatically classify bank feeds.',
  },
  assistant: {
    title: 'AI Hermes Assistant',
    subtitle: 'Send natural language commands and manage execution plans.',
  },
};

@Injectable({
  providedIn: 'root',
})
export class LanguageService {
  private readonly STORAGE_KEY = 'wealthos_lang';
  private currentLangSubject = new BehaviorSubject<SupportedLanguage>(this.loadInitialLanguage());

  public currentLang$ = this.currentLangSubject.asObservable();

  private loadInitialLanguage(): SupportedLanguage {
    const saved = localStorage.getItem(this.STORAGE_KEY) as SupportedLanguage;
    if (saved === 'vi' || saved === 'en') {
      return saved;
    }
    return 'vi'; // Default to Vietnamese
  }

  public get currentLanguage(): SupportedLanguage {
    return this.currentLangSubject.value;
  }

  public setLanguage(lang: SupportedLanguage) {
    if (lang !== this.currentLanguage) {
      localStorage.setItem(this.STORAGE_KEY, lang);
      this.currentLangSubject.next(lang);
    }
  }

  public t(key: string): string {
    const lang = this.currentLanguage;
    const dict = lang === 'vi' ? VI_DICTIONARY : EN_DICTIONARY;
    
    const parts = key.split('.');
    let current: any = dict;
    
    for (const part of parts) {
      if (current && typeof current === 'object' && part in current) {
        current = current[part];
      } else {
        return key; // Fallback to key if translation missing
      }
    }

    return typeof current === 'string' ? current : key;
  }
}
