import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/features/auth/data/repositories/auth_repository_impl.dart';
import 'package:mobile/features/auth/data/services/auth_remote_service.dart';
import 'package:mobile/features/auth/domain/repositories/auth_repository.dart';

/// Manual dependency composition keeps framework-specific setup out of views.
class AppDependencies {
  AppDependencies({required this.apiClient, required this.authRepository});

  factory AppDependencies.production() {
    final apiClient = ApiClient();
    return AppDependencies(
      apiClient: apiClient,
      authRepository: AuthRepositoryImpl(
        apiClient,
        AuthRemoteService(apiClient),
      ),
    );
  }

  final ApiClient apiClient;
  final AuthRepository authRepository;
}
