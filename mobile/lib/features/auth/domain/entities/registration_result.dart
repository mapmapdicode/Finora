/// Result of creating an account. A session is issued only after email ownership
/// has been confirmed.
class RegistrationResult {
  const RegistrationResult({required this.email});

  final String email;
}
