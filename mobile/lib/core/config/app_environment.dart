/// Runtime configuration. Override [apiBase] with `--dart-define=API_BASE=...`.
abstract final class AppEnvironment {
  static const _configuredApiBase = String.fromEnvironment(
    'API_BASE',
    defaultValue: '',
  );

  // The public Finora VPS proxies the mobile API at `/api/v1`. Keep this as
  // the default so TestFlight builds never depend on a developer's Mac or LAN.
  static const _productionApiBase = 'http://110.172.29.117:2001';

  static String get apiBase {
    if (_configuredApiBase.isNotEmpty) {
      return _configuredApiBase;
    }
    return _productionApiBase;
  }
}
