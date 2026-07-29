import 'package:mobile/features/loans/data/services/loan_remote_service.dart';
import 'package:mobile/features/loans/domain/entities/loan.dart';
import 'package:mobile/features/loans/domain/repositories/loan_repository.dart';

class LoanRepositoryImpl implements LoanRepository {
  const LoanRepositoryImpl(this._remote);
  final LoanRemoteService _remote;

  @override
  Future<List<Loan>> list() async {
    final data = await _remote.get('/loans') as List;
    return data
        .whereType<Map>()
        .map((e) => Loan.fromJson(Map<String, dynamic>.from(e)))
        .toList();
  }

  @override
  Future<LoanSummary> summary() async => LoanSummary.fromJson(
    Map<String, dynamic>.from(await _remote.get('/loans/summary') as Map),
  );

  @override
  Future<List<LoanScheduleItem>> schedule() async {
    final data = await _remote.get('/loans/schedule?months=3') as List;
    return data
        .whereType<Map>()
        .map((e) => LoanScheduleItem.fromJson(Map<String, dynamic>.from(e)))
        .toList();
  }

  @override
  Future<LoanAccrual> accrual(String loanId) async => LoanAccrual.fromJson(
    Map<String, dynamic>.from(
      await _remote.get('/loans/$loanId/accruals') as Map,
    ),
  );

  @override
  Future<List<LoanPaymentRecord>> payments(String loanId) async {
    final data = await _remote.get('/loans/$loanId/payments') as List;
    return data
        .whereType<Map>()
        .map(
          (item) => LoanPaymentRecord.fromJson(Map<String, dynamic>.from(item)),
        )
        .toList();
  }

  @override
  Future<void> create(Map<String, dynamic> input) async {
    await _remote.post('/loans', input);
  }

  @override
  Future<LoanPaymentRecord> receive(
    String loanId,
    Map<String, dynamic> input,
  ) async => LoanPaymentRecord.fromJson(
    Map<String, dynamic>.from(
      await _remote.post('/loans/$loanId/payments', input) as Map,
    ),
  );

  @override
  Future<void> delete(String loanId) async {
    await _remote.delete('/loans/$loanId');
  }
}
