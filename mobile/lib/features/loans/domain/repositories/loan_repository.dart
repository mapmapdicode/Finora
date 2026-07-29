import 'package:mobile/features/loans/domain/entities/loan.dart';

abstract interface class LoanRepository {
  Future<List<Loan>> list();
  Future<LoanSummary> summary();
  Future<List<LoanScheduleItem>> schedule();

  Future<LoanAccrual> accrual(String loanId);
  Future<List<LoanPaymentRecord>> payments(String loanId);
  Future<void> create(Map<String, dynamic> input);

  /// Returns the payment persisted by the server so callers can update the
  /// collection timeline without waiting for another read request.
  Future<LoanPaymentRecord> receive(String loanId, Map<String, dynamic> input);
  Future<void> delete(String loanId);
}
