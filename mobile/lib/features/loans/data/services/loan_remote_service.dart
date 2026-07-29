import 'package:mobile/core/network/api_client.dart';

class LoanRemoteService {
  const LoanRemoteService(this._api);
  final ApiClient _api;

  Future<dynamic> get(String path) => _api.request('GET', path);
  Future<dynamic> post(String path, Map<String, dynamic> body) =>
      _api.request('POST', path, body);
  Future<dynamic> delete(String path) => _api.request('DELETE', path);
}
