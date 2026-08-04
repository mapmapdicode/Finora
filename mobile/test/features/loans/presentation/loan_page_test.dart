import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/features/loans/domain/entities/loan.dart';
import 'package:mobile/features/loans/domain/repositories/loan_repository.dart';
import 'package:mobile/features/loans/presentation/screens/loan_page.dart';
import 'package:mobile/features/loans/presentation/view_models/loan_view_model.dart';

void main() {
  test('normalizes loan transport errors for user-facing copy', () {
    expect(
      loanErrorMessage('ApiException: Không thể lưu khoản vay'),
      'Không thể lưu khoản vay',
    );
    expect(
      loanErrorMessage('SocketException: Failed host lookup'),
      'Không thể kết nối máy chủ. Kiểm tra mạng rồi thử lại.',
    );
  });

  testWidgets(
    'groups the live loans by borrower instead of rendering samples',
    (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: LoanPage(viewModel: LoanViewModel(_LoanRepository())),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Chị Mai'), findsOneWidget);
      expect(find.text('2 khoản vay'), findsOneWidget);
      expect(find.text('Anh Minh'), findsNothing);
    },
  );

  testWidgets('shows a clear status for active and settled loans', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(home: LoanPage(viewModel: LoanViewModel(_LoanRepository()))),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Chị Mai'));
    await tester.pumpAndSettle();

    expect(find.text('Hoạt động'), findsOneWidget);
    await tester.scrollUntilVisible(find.text('Đã quyết toán'), 180);
    expect(find.text('Đã quyết toán'), findsOneWidget);
  });

  testWidgets('summarizes lending, recovery, interest, and forecast', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: LoanPage(viewModel: LoanViewModel(_MetricsLoanRepository())),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Chị Mai'));
    await tester.pumpAndSettle();

    expect(find.text('150,000,000 VND'), findsOneWidget);
    expect(find.text('90,000,000 VND'), findsOneWidget);
    expect(find.text('9,000,000 VND'), findsOneWidget);
    expect(find.text('50%'), findsOneWidget);
    expect(find.text('1/2 khoản đã quyết toán'), findsOneWidget);
    expect(find.text('1,800,000 VND'), findsOneWidget);
  });

  testWidgets('opens the familiar interest-rate converter from loans', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(home: LoanPage(viewModel: LoanViewModel(_LoanRepository()))),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Quy đổi lãi suất'));
    await tester.pumpAndSettle();

    expect(find.text('MỨC LÃI TƯƠNG ĐƯƠNG'), findsOneWidget);
    expect(find.textContaining('3000 đồng / đầu triệu / ngày'), findsOneWidget);
    expect(find.text('Ví dụ tiền lãi phải trả'), findsOneWidget);
  });

  testWidgets('selects a saved customer before creating a loan', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: LoanPage(
          api: _CustomerApiClient(),
          autoOpenCreate: true,
          viewModel: LoanViewModel(_LoanRepository()),
        ),
      ),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Tìm hoặc thêm người vay'));
    await tester.pumpAndSettle();

    expect(find.text('Chọn khách hàng'), findsOneWidget);
    expect(find.text('Nguyễn Minh Anh'), findsOneWidget);

    await tester.tap(find.text('Nguyễn Minh Anh'));
    await tester.pumpAndSettle();

    expect(find.text('Đổi'), findsOneWidget);
    expect(find.text('090 123 4567'), findsOneWidget);
  });

  testWidgets('shows a saved collection in history without reloading it', (
    tester,
  ) async {
    final repository = _LoanRepository();
    await tester.pumpWidget(
      MaterialApp(home: LoanPage(viewModel: LoanViewModel(repository))),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('Chị Mai'));
    await tester.pumpAndSettle();
    await tester.drag(find.byType(ListView).last, const Offset(0, -320));
    await tester.pumpAndSettle();
    await tester.tap(find.text('50,000,000 VND'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Ghi nhận thu'));
    await tester.pumpAndSettle();
    final amountInput = tester.widget<TextField>(
      find.byWidgetPredicate(
        (widget) =>
            widget is TextField &&
            widget.decoration?.labelText == 'Số tiền thực nhận',
      ),
    );
    expect(amountInput.keyboardType, TextInputType.number);
    expect(amountInput.controller!.text, '50000');

    await tester.tap(find.text('3 ngày'));
    await tester.pump();
    expect(amountInput.controller!.text, '150000');

    await tester.tap(find.byIcon(Icons.add_circle_outline));
    await tester.pump();
    expect(amountInput.controller!.text, '200000');
    await tester.tap(find.text('Lưu khoản thu'));
    await tester.pumpAndSettle();

    expect(find.text('Thu lãi · 4 ngày · 200,000 VND'), findsOneWidget);
    // The borrower overview loads two histories for its aggregate; the detail
    // then loads its own timeline once and keeps the saved receipt locally.
    expect(repository.paymentHistoryReads, 3);
  });
}

class _CustomerApiClient extends ApiClient {
  @override
  Future<dynamic> request(
    String method,
    String path, [
    Map<String, dynamic>? body,
  ]) async {
    if (method == 'GET' && path.startsWith('/customers')) {
      return [
        {
          'id': 'customer-1',
          'name': 'Nguyễn Minh Anh',
          'phone': '090 123 4567',
        },
      ];
    }
    return [];
  }
}

class _LoanRepository implements LoanRepository {
  int paymentHistoryReads = 0;

  @override
  Future<LoanAccrual> accrual(String loanId) => Future.value(
    LoanAccrual(loanId: loanId, days: 9, accruedInterest: '450000'),
  );

  @override
  Future<List<LoanPaymentRecord>> payments(String loanId) {
    paymentHistoryReads++;
    return Future.value(const []);
  }

  @override
  Future<void> create(Map<String, dynamic> input) async {}

  @override
  Future<void> delete(String loanId) async {}

  @override
  Future<List<Loan>> list() => Future.value(const [
    Loan(
      id: 'loan-1',
      borrower: 'Chị Mai',
      principalBalance: '50000000',
      dailyRatePerMillion: '1000',
      status: 'active',
    ),
    Loan(
      id: 'loan-2',
      borrower: 'Chị Mai',
      principalBalance: '25000000',
      dailyRatePerMillion: '1000',
      status: 'closed',
    ),
  ]);

  @override
  Future<LoanPaymentRecord> receive(
    String loanId,
    Map<String, dynamic> input,
  ) async => LoanPaymentRecord(
    id: 'payment-1',
    principal: '0',
    interest: input['interestAmount']?.toString() ?? '0',
    interestDays: input['interestDays'] as int? ?? 0,
    fee: '0',
    occurredAt: '2026-07-29T00:00:00.000Z',
  );

  @override
  Future<List<LoanScheduleItem>> schedule() => Future.value(const []);

  @override
  Future<LoanSummary> summary() => Future.value(
    const LoanSummary(
      activePrincipal: '75000000',
      dailyInterest: '75000',
      accruedInterest: '450000',
      paidInterest: '0',
    ),
  );
}

class _MetricsLoanRepository extends _LoanRepository {
  @override
  Future<List<Loan>> list() => Future.value(const [
    Loan(
      id: 'active-loan',
      borrower: 'Chị Mai',
      principalInitial: '100000000',
      principalBalance: '60000000',
      dailyRatePerMillion: '1000',
      status: 'active',
    ),
    Loan(
      id: 'settled-loan',
      borrower: 'Chị Mai',
      principalInitial: '50000000',
      principalBalance: '0',
      dailyRatePerMillion: '1000',
      status: 'closed',
    ),
  ]);

  @override
  Future<List<LoanPaymentRecord>> payments(String loanId) {
    if (loanId == 'active-loan') {
      return Future.value(const [
        LoanPaymentRecord(
          id: 'active-payment',
          principal: '40000000',
          interest: '4000000',
          fee: '0',
          occurredAt: '2026-07-01T00:00:00.000Z',
        ),
      ]);
    }
    return Future.value(const [
      LoanPaymentRecord(
        id: 'settled-payment',
        principal: '50000000',
        interest: '5000000',
        fee: '0',
        occurredAt: '2026-07-15T00:00:00.000Z',
      ),
    ]);
  }
}
