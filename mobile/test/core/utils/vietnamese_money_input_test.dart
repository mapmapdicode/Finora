import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/utils/vietnamese_money_input.dart';

void main() {
  test('expands familiar million shorthand into a full VND value', () {
    expect(parseVietnameseMoneyInput('30tr'), 30000000);
    expect(parseVietnameseMoneyInput('30t'), 30000000);
    expect(parseVietnameseMoneyInput('30m'), 30000000);
  });

  test('accepts grouped and decimal Vietnamese money input', () {
    expect(parseVietnameseMoneyInput('30.000.000'), 30000000);
    expect(parseVietnameseMoneyInput('30,000,000'), 30000000);
    expect(parseVietnameseMoneyInput('1,5tr'), 1500000);
  });
}
