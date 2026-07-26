/// Authenticated session used by the networking boundary.
class AuthSession {
  const AuthSession({required this.token, required this.workspaceId});

  final String token;
  final String? workspaceId;
}
