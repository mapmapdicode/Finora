/// Input required to authenticate a Finora user.
class AuthCredentials {
  const AuthCredentials({
    required this.email,
    required this.password,
    this.name,
    this.workspaceName,
  });

  final String email;
  final String password;
  final String? name;
  final String? workspaceName;
}
