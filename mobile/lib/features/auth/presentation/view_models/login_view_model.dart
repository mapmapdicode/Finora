import 'package:flutter/foundation.dart';
import 'package:mobile/core/network/api_exception.dart';
import 'package:mobile/features/auth/domain/entities/auth_credentials.dart';
import 'package:mobile/features/auth/domain/repositories/auth_repository.dart';

/// Owns login UI state and commands; the view never calls the repository.
class LoginViewModel extends ChangeNotifier {
  LoginViewModel(this._repository);

  final AuthRepository _repository;
  bool _isBusy = false;
  String? _error;

  bool get isBusy => _isBusy;
  String? get error => _error;

  Future<bool> authenticate({
    required bool registering,
    required String email,
    required String password,
    String? name,
    String? workspaceName,
  }) async {
    if (_isBusy) return false;
    _isBusy = true;
    _error = null;
    notifyListeners();
    try {
      final credentials = AuthCredentials(
        email: email.trim(),
        password: password,
        name: registering ? name?.trim() : null,
        workspaceName: registering ? workspaceName?.trim() : null,
      );
      if (registering) {
        await _repository.register(credentials);
      } else {
        await _repository.signIn(credentials);
      }
      return true;
    } on ApiException catch (error) {
      _error = error.message;
      return false;
    } catch (_) {
      _error = 'Không thể đăng nhập';
      return false;
    } finally {
      _isBusy = false;
      notifyListeners();
    }
  }

  void clearError() {
    if (_error == null) return;
    _error = null;
    notifyListeners();
  }
}
