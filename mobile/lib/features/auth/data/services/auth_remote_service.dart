import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/network/api_exception.dart';
import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';
import 'package:mobile/features/auth/domain/entities/registration_result.dart';

/// Stateless wrapper around the remote authentication endpoints.
class AuthRemoteService {
  AuthRemoteService(this._apiClient);

  final ApiClient _apiClient;

  Future<Map<String, dynamic>> signIn(AuthCredentials credentials) =>
      _post('/auth/login', credentials);

  Future<RegistrationResult> register(AuthCredentials credentials) async {
    final response = await _post('/auth/register', credentials);
    final email = response['user'] is Map
        ? (response['user'] as Map)['email']?.toString()
        : null;
    if (response['emailVerificationRequired'] != true ||
        email == null ||
        email.isEmpty) {
      throw const ApiException('Phản hồi đăng ký không hợp lệ');
    }
    return RegistrationResult(email: email);
  }

  Future<Map<String, dynamic>> verifyEmail(String email, String code) =>
      _apiClient
          .request('POST', '/auth/verify-email', {'email': email, 'code': code})
          .then((response) {
            if (response is! Map) {
              throw const ApiException('Phản hồi xác thực email không hợp lệ');
            }
            return Map<String, dynamic>.from(response);
          });

  Future<void> resendVerificationEmail(String email) async {
    await _apiClient.request('POST', '/auth/resend-verification-email', {
      'email': email,
    });
  }

  Future<Map<String, dynamic>> _post(
    String path,
    AuthCredentials credentials,
  ) async {
    final response = await _apiClient.request('POST', path, {
      'email': credentials.email,
      'password': credentials.password,
      if (credentials.confirmPassword != null)
        'confirmPassword': credentials.confirmPassword,
      if (credentials.name != null) 'name': credentials.name,
    });
    if (response is! Map) {
      throw const ApiException('Phản hồi xác thực không hợp lệ');
    }
    return Map<String, dynamic>.from(response);
  }
}
