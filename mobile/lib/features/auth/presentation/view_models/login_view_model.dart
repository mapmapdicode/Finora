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
  String? _pendingVerificationEmail;

  bool get isBusy => _isBusy;
  String? get error => _error;
  String? get pendingVerificationEmail => _pendingVerificationEmail;

  Future<bool> authenticate({
    required bool registering,
    required String email,
    required String password,
    String? confirmPassword,
    String? name,
  }) async {
    if (_isBusy) return false;
    _isBusy = true;
    _error = null;
    notifyListeners();
    try {
      if (registering && (confirmPassword == null || confirmPassword.isEmpty)) {
        _error = 'Vui lòng nhập lại mật khẩu';
        return false;
      }
      if (registering && password != confirmPassword) {
        _error = 'Mật khẩu nhập lại không khớp';
        return false;
      }
      final credentials = AuthCredentials(
        email: email.trim(),
        password: password,
        confirmPassword: registering ? confirmPassword : null,
        name: registering ? name?.trim() : null,
      );
      if (registering) {
        final result = await _repository.register(credentials);
        _pendingVerificationEmail = result.email;
        return false;
      } else {
        await _repository.signIn(credentials);
      }
      _pendingVerificationEmail = null;
      return true;
    } on ApiException catch (error) {
      if (error.code == 'EMAIL_NOT_VERIFIED') {
        _pendingVerificationEmail = email.trim();
      }
      _error = _friendlyAuthError(error);
      return false;
    } catch (_) {
      _error =
          'Không thể đăng nhập lúc này. Vui lòng kiểm tra kết nối và thử lại.';
      return false;
    } finally {
      _isBusy = false;
      notifyListeners();
    }
  }

  Future<bool> verifyEmail(String code) async {
    final email = _pendingVerificationEmail;
    if (email == null || email.isEmpty) {
      _error = 'Không tìm thấy email cần xác thực';
      notifyListeners();
      return false;
    }
    if (_isBusy) return false;
    _isBusy = true;
    _error = null;
    notifyListeners();
    try {
      await _repository.verifyEmail(email: email, code: code.trim());
      _pendingVerificationEmail = null;
      return true;
    } on ApiException catch (error) {
      _error = error.message;
      return false;
    } catch (_) {
      _error = 'Không thể xác thực email';
      return false;
    } finally {
      _isBusy = false;
      notifyListeners();
    }
  }

  Future<bool> resendVerificationEmail() async {
    final email = _pendingVerificationEmail;
    if (email == null || email.isEmpty || _isBusy) return false;
    _isBusy = true;
    _error = null;
    notifyListeners();
    try {
      await _repository.resendVerificationEmail(email);
      return true;
    } on ApiException catch (error) {
      _error = error.message;
      return false;
    } catch (_) {
      _error = 'Không thể gửi lại mã xác thực';
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

  void cancelEmailVerification() {
    _pendingVerificationEmail = null;
    _error = null;
    notifyListeners();
  }

  String _friendlyAuthError(ApiException error) {
    switch (error.code) {
      case 'INVALID_CREDENTIALS':
        return 'Email hoặc mật khẩu chưa đúng. Hãy kiểm tra lại và thử lại.';
      case 'EMAIL_NOT_VERIFIED':
        return 'Email này chưa được xác thực. Hãy kiểm tra hộp thư để lấy mã.';
      case 'NETWORK_UNAVAILABLE':
        return 'Không thể kết nối tới Finora. Kiểm tra mạng hoặc thử lại sau ít phút.';
      default:
        return error.message;
    }
  }
}
