import 'dart:io';

/// Runtime configuration. Override [apiBase] with `--dart-define=API_BASE=...`.
abstract final class AppEnvironment {
  static const _configuredApiBase = String.fromEnvironment(
    'API_BASE',
    defaultValue: '',
  );

  // A physical iPhone cannot reach the Mac through 127.0.0.1. Bonjour keeps
  // this local development endpoint stable when the router changes its IP.
  // Production and every non-local environment must override this through
  // `--dart-define=API_BASE=https://...`.
  static const _iosDeviceDevelopmentBase = 'http://Hoangs-Mac-mini.local:8080';

  static String get apiBase {
    if (_configuredApiBase.isNotEmpty) {
      return _configuredApiBase;
    }
    if (Platform.isAndroid) return 'http://10.0.2.2:8080';
    if (Platform.isIOS) return _iosDeviceDevelopmentBase;
    return 'http://127.0.0.1:8080';
  }
}
