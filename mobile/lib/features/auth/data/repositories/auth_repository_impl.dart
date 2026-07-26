import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/network/api_exception.dart';
import 'package:mobile/features/auth/data/services/auth_remote_service.dart';
import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';
import 'package:mobile/features/auth/domain/entities/auth_session.dart';
import 'package:mobile/features/auth/domain/repositories/auth_repository.dart';

/// Maps remote authentication data to domain objects and updates the API session.
class AuthRepositoryImpl implements AuthRepository {
  AuthRepositoryImpl(this._apiClient, this._remoteService);

  final ApiClient _apiClient;
  final AuthRemoteService _remoteService;

  @override
  Future<AuthSession> signIn(AuthCredentials credentials) async =>
      _saveSession(await _remoteService.signIn(credentials));

  @override
  Future<AuthSession> register(AuthCredentials credentials) async =>
      _saveSession(await _remoteService.register(credentials));

  AuthSession _saveSession(Map<String, dynamic> response) {
    final token = response['token']?.toString();
    if (token == null || token.isEmpty) {
      throw const ApiException('Phản hồi đăng nhập không có token');
    }
    final workspace = response['workspace'];
    final workspaceId = workspace is Map ? workspace['id']?.toString() : null;
    final session = AuthSession(token: token, workspaceId: workspaceId);
    _apiClient
      ..token = session.token
      ..workspaceId = session.workspaceId;
    return session;
  }
}
