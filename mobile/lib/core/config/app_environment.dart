import 'dart:io';

/// Runtime configuration. Override [apiBase] with `--dart-define=API_BASE=...`.
abstract final class AppEnvironment {
  static const _configuredApiBase = String.fromEnvironment(
    'API_BASE',
    defaultValue: '',
  );

  static String get apiBase {
    if (_configuredApiBase.isNotEmpty) {
      return _configuredApiBase;
    }
    return Platform.isAndroid
        ? 'http://10.0.2.2:8080'
        : 'http://127.0.0.1:8080';
  }
}
