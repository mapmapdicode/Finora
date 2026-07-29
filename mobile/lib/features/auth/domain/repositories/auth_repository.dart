import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';
import 'package:mobile/features/auth/domain/entities/auth_session.dart';
import 'package:mobile/features/auth/domain/entities/registration_result.dart';

/// Authentication source of truth for the feature.
abstract interface class AuthRepository {
  Future<AuthSession> signIn(AuthCredentials credentials);

  Future<RegistrationResult> register(AuthCredentials credentials);

  Future<AuthSession> verifyEmail({
    required String email,
    required String code,
  });

  Future<void> resendVerificationEmail(String email);
}
