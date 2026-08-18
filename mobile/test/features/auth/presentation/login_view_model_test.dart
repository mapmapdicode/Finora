import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/network/api_exception.dart';
import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';
import 'package:mobile/features/auth/domain/entities/auth_session.dart';
import 'package:mobile/features/auth/domain/entities/registration_result.dart';
import 'package:mobile/features/auth/domain/repositories/auth_repository.dart';
import 'package:mobile/features/auth/presentation/view_models/login_view_model.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  setUp(() => SharedPreferences.setMockInitialValues({}));

  test(
    'sign in delegates trimmed credentials and clears loading state',
    () async {
      final repository = _FakeAuthRepository();
      final viewModel = LoginViewModel(repository);

      final result = await viewModel.authenticate(
        registering: false,
        email: ' demo@finora.vn ',
        password: 'secret',
      );

      expect(result, isTrue);
      expect(repository.signInCredentials?.email, 'demo@finora.vn');
      expect(viewModel.isBusy, isFalse);
      expect(viewModel.error, isNull);
    },
  );

  test(
    'remembers the identifier from the last successful login only',
    () async {
      final viewModel = LoginViewModel(_FakeAuthRepository());

      await viewModel.authenticate(
        registering: false,
        email: ' latest@finora.vn ',
        password: 'secret',
      );

      expect(await viewModel.loadLastLoginIdentifier(), 'latest@finora.vn');
    },
  );

  test('surfaces a safe API error to the view', () async {
    final viewModel = LoginViewModel(_FakeAuthRepository(shouldFail: true));

    final result = await viewModel.authenticate(
      registering: true,
      email: 'demo@finora.vn',
      password: 'secret',
      confirmPassword: 'secret',
      name: 'Demo',
    );

    expect(result, isFalse);
    expect(viewModel.error, 'Không thể kết nối');
    expect(viewModel.isBusy, isFalse);
  });

  test('shows a friendly message when credentials are incorrect', () async {
    final viewModel = LoginViewModel(
      _FakeAuthRepository(
        shouldFail: true,
        failureCode: 'INVALID_CREDENTIALS',
        failureMessage: 'invalid credentials',
      ),
    );

    final result = await viewModel.authenticate(
      registering: false,
      email: 'demo@finora.vn',
      password: 'wrong-password',
    );

    expect(result, isFalse);
    expect(
      viewModel.error,
      'Email hoặc mật khẩu chưa đúng. Hãy kiểm tra lại và thử lại.',
    );
  });

  test('shows a friendly message when the backend cannot be reached', () async {
    final viewModel = LoginViewModel(
      _FakeAuthRepository(
        shouldFail: true,
        failureCode: 'NETWORK_UNAVAILABLE',
        failureMessage: 'socket exception',
      ),
    );

    final result = await viewModel.authenticate(
      registering: false,
      email: 'demo@finora.vn',
      password: 'secret',
    );

    expect(result, isFalse);
    expect(
      viewModel.error,
      'Không thể kết nối tới Finora. Kiểm tra mạng hoặc thử lại sau ít phút.',
    );
  });

  test('requires confirmation password before registration', () async {
    final viewModel = LoginViewModel(_FakeAuthRepository());

    final result = await viewModel.authenticate(
      registering: true,
      email: 'demo@finora.vn',
      password: 'secret',
      name: 'Demo',
    );

    expect(result, isFalse);
    expect(viewModel.error, 'Vui lòng nhập lại mật khẩu');
    expect(viewModel.isBusy, isFalse);
  });

  test('registration records the email for the follow-up login flow', () async {
    final viewModel = LoginViewModel(_FakeAuthRepository());

    final result = await viewModel.authenticate(
      registering: true,
      email: 'new@finora.vn',
      password: 'secret',
      confirmPassword: 'secret',
      name: 'New User',
    );

    expect(result, isFalse);
    expect(viewModel.pendingVerificationEmail, 'new@finora.vn');
    expect(viewModel.error, isNull);
  });

  test('verifies a pending email and clears the verification state', () async {
    final viewModel = LoginViewModel(_FakeAuthRepository());
    await viewModel.authenticate(
      registering: true,
      email: 'new@finora.vn',
      password: 'secret',
      confirmPassword: 'secret',
      name: 'New User',
    );

    final verified = await viewModel.verifyEmail('123456');

    expect(verified, isTrue);
    expect(viewModel.pendingVerificationEmail, isNull);
    expect(viewModel.error, isNull);
  });

  test('opens email verification after an unverified login attempt', () async {
    final viewModel = LoginViewModel(
      _FakeAuthRepository(shouldFail: true, failureCode: 'EMAIL_NOT_VERIFIED'),
    );

    await viewModel.authenticate(
      registering: false,
      email: 'pending@finora.vn',
      password: 'secret',
    );

    expect(viewModel.pendingVerificationEmail, 'pending@finora.vn');
  });
}

class _FakeAuthRepository implements AuthRepository {
  _FakeAuthRepository({
    this.shouldFail = false,
    this.failureCode,
    this.failureMessage = 'Không thể kết nối',
  });

  final bool shouldFail;
  final String? failureCode;
  final String failureMessage;
  AuthCredentials? signInCredentials;

  @override
  Future<RegistrationResult> register(AuthCredentials credentials) async {
    await _result();
    return RegistrationResult(email: credentials.email);
  }

  @override
  Future<AuthSession> signIn(AuthCredentials credentials) {
    signInCredentials = credentials;
    return _result();
  }

  @override
  Future<void> resendVerificationEmail(String email) async {}

  @override
  Future<AuthSession> verifyEmail({
    required String email,
    required String code,
  }) => _result();

  Future<AuthSession> _result() {
    if (shouldFail) {
      throw ApiException(failureMessage, code: failureCode);
    }
    return Future.value(const AuthSession(token: 'token'));
  }
}
