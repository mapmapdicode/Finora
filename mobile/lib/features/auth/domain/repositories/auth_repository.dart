import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';
import 'package:mobile/features/auth/domain/entities/auth_session.dart';

/// Authentication source of truth for the feature.
abstract interface class AuthRepository {
  Future<AuthSession> signIn(AuthCredentials credentials);

  Future<AuthSession> register(AuthCredentials credentials);
}
