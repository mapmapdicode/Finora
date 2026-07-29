class Loan {
  const Loan({
    required this.id,
    required this.borrower,
    required this.principalBalance,
    required this.dailyRatePerMillion,
    required this.status,
    this.days = 0,
    this.customerId = '',
    this.startDate = '',
    this.accruedInterest = '0',
    this.dailyRateText = '',
  });

  final String id;
  final String borrower;
  final String principalBalance;
  final String dailyRatePerMillion;
  final String status;
  final int days;
  final String customerId;
  final String startDate;
  final String accruedInterest;
  final String dailyRateText;

  factory Loan.fromJson(Map<String, dynamic> json) => Loan(
    id: json['id']?.toString() ?? '',
    borrower: json['counterparty']?.toString() ?? 'Chưa đặt tên',
    principalBalance:
        json['principalBalance']?.toString() ??
        json['principalInitial']?.toString() ??
        '0',
    dailyRatePerMillion: json['dailyRatePerMillion']?.toString() ?? '0',
    status: json['status']?.toString() ?? 'active',
    days: int.tryParse(json['days']?.toString() ?? '') ?? 0,
    customerId: json['customerId']?.toString() ?? '',
    startDate: json['startDate']?.toString() ?? '',
    accruedInterest: json['accruedInterest']?.toString() ?? '0',
    dailyRateText: json['dailyRateText']?.toString() ?? '',
  );
}

class BorrowerContact {
  const BorrowerContact({
    required this.name,
    required this.loanCount,
    required this.totalPrincipal,
    required this.dailyInterest,
    this.avatarUrl,
    this.avatarInitial,
    this.avatarBgColorHex,
    this.isPriority = false,
  });

  final String name;
  final int loanCount;
  final String totalPrincipal;
  final String dailyInterest;
  final String? avatarUrl;
  final String? avatarInitial;
  final String? avatarBgColorHex;
  final bool isPriority;
}

class LoanHistoryItem {
  const LoanHistoryItem({
    required this.date,
    required this.days,
    required this.amount,
    this.principalPart = '',
    this.interestPart = '',
    this.note,
    this.isCreation = false,
  });

  final String date;
  final int days;
  final String amount;
  final String principalPart;
  final String interestPart;
  final String? note;
  final bool isCreation;
}

class LoanPaymentRecord {
  const LoanPaymentRecord({
    required this.id,
    required this.principal,
    required this.interest,
    required this.fee,
    required this.occurredAt,
    this.interestDays,
  });

  final String id;
  final String principal;
  final String interest;
  // Nullable to keep hot-reloaded sessions and records created by older API
  // versions safe while the backend rollout is in progress.
  final int? interestDays;
  final String fee;
  final String occurredAt;

  factory LoanPaymentRecord.fromJson(Map<String, dynamic> json) =>
      LoanPaymentRecord(
        id: json['id']?.toString() ?? '',
        principal: json['principalAmount']?.toString() ?? '0',
        interest: json['interestAmount']?.toString() ?? '0',
        interestDays: int.tryParse(json['interestDays']?.toString() ?? '') ?? 0,
        fee: json['feeAmount']?.toString() ?? '0',
        occurredAt: json['occurredAt']?.toString() ?? '',
      );
}

class LoanSummary {
  const LoanSummary({
    required this.activePrincipal,
    required this.dailyInterest,
    required this.accruedInterest,
    required this.paidInterest,
    this.contactCount = 0,
    this.loanCount = 0,
  });

  final String activePrincipal;
  final String dailyInterest;
  final String accruedInterest;
  final String paidInterest;
  final int contactCount;
  final int loanCount;

  factory LoanSummary.fromJson(Map<String, dynamic> json) => LoanSummary(
    activePrincipal: json['activePrincipal']?.toString() ?? '0',
    dailyInterest: json['dailyInterest']?.toString() ?? '0',
    accruedInterest: json['accruedInterest']?.toString() ?? '0',
    paidInterest: json['paidInterest']?.toString() ?? '0',
    contactCount: int.tryParse(json['contactCount']?.toString() ?? '') ?? 0,
    loanCount: int.tryParse(json['loanCount']?.toString() ?? '') ?? 0,
  );
}

class LoanScheduleItem {
  const LoanScheduleItem({
    required this.borrower,
    required this.paymentDate,
    required this.expectedInterest,
    required this.cycleDays,
    required this.status,
  });

  final String borrower, paymentDate, expectedInterest, status;
  final int cycleDays;

  factory LoanScheduleItem.fromJson(Map<String, dynamic> json) =>
      LoanScheduleItem(
        borrower: json['borrower']?.toString() ?? 'Chưa đặt tên',
        paymentDate: json['paymentDate']?.toString() ?? '',
        expectedInterest: json['expectedInterest']?.toString() ?? '0',
        cycleDays: int.tryParse(json['cycleDays']?.toString() ?? '') ?? 0,
        status: json['status']?.toString() ?? 'upcoming',
      );
}

class LoanAccrual {
  const LoanAccrual({
    required this.loanId,
    required this.days,
    required this.accruedInterest,
    this.lastInterestPaidAt,
  });

  final String loanId;
  final int days;
  final String accruedInterest;
  // A running app can retain an older LoanAccrual instance across hot reload.
  // This must remain nullable until it is refreshed from the API.
  final String? lastInterestPaidAt;

  factory LoanAccrual.fromJson(Map<String, dynamic> json) {
    final periods = (json['accruals'] as List? ?? const []);
    final current = periods.isEmpty
        ? const <String, dynamic>{}
        : Map<String, dynamic>.from(periods.last as Map);
    return LoanAccrual(
      loanId: json['loanId']?.toString() ?? '',
      days: int.tryParse(current['days']?.toString() ?? '') ?? 0,
      accruedInterest:
          current['unpaidInterest']?.toString() ??
          json['unpaidInterest']?.toString() ??
          '0',
      lastInterestPaidAt: json['lastInterestPaidDate']?.toString() ?? '',
    );
  }
}
