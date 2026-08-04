// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:mobile/app/app.dart';
import 'package:mobile/features/finora/presentation/finora_pages.dart';

void main() {
  testWidgets('renders sign in form', (WidgetTester tester) async {
    await tester.binding.setSurfaceSize(const Size(1440, 1200));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await tester.pumpWidget(const FinoraApp());
    expect(find.text('Đăng nhập'), findsOneWidget);
  });

  testWidgets('starts sign in with blank customer credentials', (
    WidgetTester tester,
  ) async {
    await tester.binding.setSurfaceSize(const Size(390, 844));
    addTearDown(() => tester.binding.setSurfaceSize(null));

    await tester.pumpWidget(const FinoraApp());
    final fields = tester
        .widgetList<TextField>(find.byType(TextField))
        .toList();

    expect(fields, hasLength(greaterThanOrEqualTo(2)));
    expect(fields[0].controller?.text, isEmpty);
    expect(fields[1].controller?.text, isEmpty);
  });

  test('normalizes technical and network errors for users', () {
    expect(
      presentableError('ApiException: Không thể lưu dữ liệu'),
      'Không thể lưu dữ liệu',
    );
    expect(
      presentableError('SocketException: Failed host lookup'),
      'Không thể kết nối máy chủ. Kiểm tra mạng rồi thử lại.',
    );
  });
}
