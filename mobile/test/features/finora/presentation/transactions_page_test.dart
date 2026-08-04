import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/features/finora/presentation/finora_pages.dart';

class _TransactionsApi extends ApiClient {
  @override
  Future<dynamic> request(
    String method,
    String path, [
    Map<String, dynamic>? body,
  ]) async {
    if (path.startsWith('/transactions')) {
      return {
        'items': [
          {
            'id': 'coffee',
            'name': 'Cà phê sáng',
            'categoryId': 'Ăn uống',
            'type': 'expense',
            'amount': '35000',
            'occurredAt': '2026-07-30T08:00:00Z',
          },
          {
            'id': 'salary',
            'name': 'Lương tháng',
            'categoryId': 'Thu nhập',
            'type': 'income',
            'amount': '20000000',
            'occurredAt': '2026-07-29T08:00:00Z',
          },
        ],
      };
    }
    return [];
  }
}

void main() {
  testWidgets('finds transactions by name and category', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(body: TransactionsPage(api: _TransactionsApi())),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Cà phê sáng'), findsOneWidget);
    expect(find.text('Lương tháng'), findsOneWidget);

    await tester.enterText(find.byType(TextField), 'thu nhập');
    await tester.pump();

    expect(find.text('Cà phê sáng'), findsNothing);
    expect(find.text('Lương tháng'), findsOneWidget);
  });

  testWidgets('opens a date range picker from transaction filters', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(body: TransactionsPage(api: _TransactionsApi())),
      ),
    );
    await tester.pumpAndSettle();

    final dateFilter = find.text('Khoảng ngày');
    await tester.ensureVisible(dateFilter);
    await tester.tap(dateFilter);
    await tester.pumpAndSettle();

    expect(find.byType(DateRangePickerDialog), findsOneWidget);
  });
}
