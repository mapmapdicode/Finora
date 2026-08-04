import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/widgets/finora_core_widgets.dart';

void main() {
  testWidgets('formats VND with Vietnamese grouping separators', (
    tester,
  ) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: FinoraMoney(1250000))),
    );

    expect(find.text('1.250.000 VND'), findsOneWidget);
  });

  testWidgets('keeps the sign when displaying a negative amount', (
    tester,
  ) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: FinoraMoney(-35000))),
    );

    expect(find.text('-35.000 VND'), findsOneWidget);
  });

  testWidgets('uses the familiar million abbreviation for VND', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(body: FinoraMoney(2500000, compact: true)),
      ),
    );

    expect(find.text('2.5tr VND'), findsOneWidget);
  });
}
