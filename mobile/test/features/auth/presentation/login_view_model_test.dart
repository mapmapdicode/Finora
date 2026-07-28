import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/network/api_exception.dart';
import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';
import 'package:mobile/features/auth/domain/entities/auth_session.dart';
import 'package:mobile/features/auth/domain/repositories/auth_repository.dart';
import 'package:mobile/features/auth/presentation/view_models/login_view_model.dart';

void main() {
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

  test('surfaces a safe API error to the view', () async {
    final viewModel = LoginViewModel(_FakeAuthRepository(shouldFail: true));

    final result = await viewModel.authenticate(
      registering: true,
      email: 'demo@finora.vn',
      password: 'secret',
      name: 'Demo',
    );

    expect(result, isFalse);
    expect(viewModel.error, 'Không thể kết nối');
    expect(viewModel.isBusy, isFalse);
  });
}

class _FakeAuthRepository implements AuthRepository {
  _FakeAuthRepository({this.shouldFail = false});

  final bool shouldFail;
  AuthCredentials? signInCredentials;

  @override
  Future<AuthSession> register(AuthCredentials credentials) => _result();

  @override
  Future<AuthSession> signIn(AuthCredentials credentials) {
    signInCredentials = credentials;
    return _result();
  }

  Future<AuthSession> _result() {
    if (shouldFail) {
      throw const ApiException('Không thể kết nối');
    }
    return Future.value(const AuthSession(token: 'token'));
  }
}
