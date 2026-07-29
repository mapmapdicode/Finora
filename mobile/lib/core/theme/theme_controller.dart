import 'package:flutter/material.dart';

/// The refreshed product intentionally uses one accessible light appearance.
/// This controller remains in place to preserve the app composition contract.
class FinoraThemeController extends ChangeNotifier {
  final ThemeMode _mode = ThemeMode.light;
  ThemeMode get mode => _mode;

  Future<void> restore() async {}

  Future<void> setMode(ThemeMode mode) async {}
}

class FinoraThemeScope extends InheritedNotifier<FinoraThemeController> {
  const FinoraThemeScope({
    super.key,
    required FinoraThemeController controller,
    required super.child,
  }) : super(notifier: controller);

  static FinoraThemeController of(BuildContext context) {
    final scope = context
        .dependOnInheritedWidgetOfExactType<FinoraThemeScope>();
    assert(scope != null, 'FinoraThemeScope is not available in this context.');
    return scope!.notifier!;
  }
}
