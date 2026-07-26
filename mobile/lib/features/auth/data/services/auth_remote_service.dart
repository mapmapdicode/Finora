import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/network/api_exception.dart';
import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';

/// Stateless wrapper around the remote authentication endpoints.
class AuthRemoteService {
  AuthRemoteService(this._apiClient);

  final ApiClient _apiClient;

  Future<Map<String, dynamic>> signIn(AuthCredentials credentials) =>
      _post('/auth/login', credentials);

  Future<Map<String, dynamic>> register(AuthCredentials credentials) =>
      _post('/auth/register', credentials);

  Future<Map<String, dynamic>> _post(
    String path,
    AuthCredentials credentials,
  ) async {
    final response = await _apiClient.request('POST', path, {
      'email': credentials.email,
      'password': credentials.password,
      if (credentials.name != null) 'name': credentials.name,
      if (credentials.workspaceName != null)
        'workspaceName': credentials.workspaceName,
    });
    if (response is! Map) {
      throw const ApiException('Phản hồi xác thực không hợp lệ');
    }
    return Map<String, dynamic>.from(response);
  }
}
