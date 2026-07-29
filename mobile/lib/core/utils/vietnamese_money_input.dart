/// Parses familiar Vietnamese money shorthand into a full numeric VND value.
///
/// Examples: `30tr`, `30t`, and `30m` all become 30,000,000. Grouped input
/// such as `30.000.000` and `30,000,000` is supported as well.
double parseVietnameseMoneyInput(String rawInput) {
  final raw = rawInput
      .trim()
      .toLowerCase()
      .replaceAll(RegExp(r'\s+'), '')
      .replaceAll('vnd', '')
      .replaceAll('đ', '');
  if (raw.isEmpty) return 0;

  final match = RegExp(
    r'^([0-9.,]+)(k|tr|t|m|trieu|triệu|ty|tỷ|b)?$',
  ).firstMatch(raw);
  if (match == null) return 0;
  final value = _parseLocalizedNumber(match.group(1) ?? '');
  if (value == null || value < 0) return 0;
  return value *
      switch (match.group(2)) {
        'k' => 1000,
        'tr' || 't' || 'm' || 'trieu' || 'triệu' => 1000000,
        'ty' || 'tỷ' || 'b' => 1000000000,
        _ => 1,
      };
}

double? _parseLocalizedNumber(String raw) {
  final dots = '.'.allMatches(raw).length;
  final commas = ','.allMatches(raw).length;
  String normalized;
  if (dots > 0 && commas > 0) {
    final decimalAt = raw.lastIndexOf(RegExp(r'[.,]'));
    normalized =
        '${raw.substring(0, decimalAt).replaceAll(RegExp(r'[.,]'), '')}.${raw.substring(decimalAt + 1)}';
  } else {
    final separator = dots > 0 ? '.' : (commas > 0 ? ',' : '');
    if (separator.isEmpty) {
      normalized = raw;
    } else {
      final count = dots + commas;
      final lastGroup = raw.substring(raw.lastIndexOf(separator) + 1);
      normalized = count > 1 || lastGroup.length == 3
          ? raw.replaceAll(separator, '')
          : raw.replaceAll(separator, '.');
    }
  }
  return double.tryParse(normalized);
}
