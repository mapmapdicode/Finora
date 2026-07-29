import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:mobile/features/loans/domain/entities/loan.dart';
import 'package:mobile/features/loans/domain/repositories/loan_repository.dart';

class LoanViewModel extends ChangeNotifier {
  LoanViewModel(this._repository);
  final LoanRepository _repository;

  List<Loan> loans = const [];
  List<LoanScheduleItem> schedule = const [];
  Map<String, LoanAccrual> accruals = const {};
  LoanSummary summary = const LoanSummary(
    activePrincipal: '0',
    dailyInterest: '0',
    accruedInterest: '0',
    paidInterest: '0',
  );
  bool loading = true;
  String? error;

  Future<void> load() async {
    loading = true;
    error = null;
    notifyListeners();
    try {
      final results = await Future.wait([
        _repository.list(),
        _repository.summary(),
        _repository.schedule(),
      ]);
      loans = results[0] as List<Loan>;
      summary = results[1] as LoanSummary;
      schedule = results[2] as List<LoanScheduleItem>;
      final details = await Future.wait(
        loans.map((loan) => _repository.accrual(loan.id)),
      );
      accruals = {for (final item in details) item.loanId: item};
    } catch (e) {
      error = e.toString();
    } finally {
      loading = false;
      notifyListeners();
    }
  }

  Future<void> create(Map<String, dynamic> input) async {
    await _repository.create(input);
    await load();
  }

  Future<LoanPaymentRecord> receive(
    String loanId,
    Map<String, dynamic> input,
  ) async {
    final payment = await _repository.receive(loanId, input);
    // Keep list and summary data in sync, but do not delay the detail screen
    // from rendering the payment that has just been saved.
    unawaited(load());
    return payment;
  }

  Future<List<LoanPaymentRecord>> payments(String loanId) =>
      _repository.payments(loanId);

  Future<void> delete(String loanId) async {
    await _repository.delete(loanId);
    await load();
  }
}
