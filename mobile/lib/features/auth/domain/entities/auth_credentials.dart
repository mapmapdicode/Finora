/// Input required to authenticate a Finora user.
class AuthCredentials {
  const AuthCredentials({
    required this.email,
    required this.password,
    this.confirmPassword,
    this.name,
  });

  final String email;
  final String password;
  final String? confirmPassword;
  final String? name;
}
