class LoanReminder {
  const LoanReminder({
    required this.loanId,
    required this.borrower,
    required this.scheduledAt,
    required this.enabled,
  });

  final String loanId;
  final String borrower;
  final DateTime scheduledAt;
  final bool enabled;

  LoanReminder copyWith({
    String? borrower,
    DateTime? scheduledAt,
    bool? enabled,
  }) => LoanReminder(
    loanId: loanId,
    borrower: borrower ?? this.borrower,
    scheduledAt: scheduledAt ?? this.scheduledAt,
    enabled: enabled ?? this.enabled,
  );

  Map<String, dynamic> toJson() => {
    'loanId': loanId,
    'borrower': borrower,
    'scheduledAt': scheduledAt.toUtc().toIso8601String(),
    'enabled': enabled,
  };

  factory LoanReminder.fromJson(Map<String, dynamic> json) => LoanReminder(
    loanId: json['loanId']?.toString() ?? '',
    borrower: json['borrower']?.toString() ?? 'Khoản vay',
    scheduledAt:
        DateTime.tryParse(json['scheduledAt']?.toString() ?? '')?.toLocal() ??
        DateTime.now(),
    enabled: json['enabled'] == true,
  );
}
