/// Input required to authenticate a Finora user.
class AuthCredentials {
  const AuthCredentials({
    required this.email,
    required this.password,
    this.name,
  });

  final String email;
  final String password;
  final String? name;
}
