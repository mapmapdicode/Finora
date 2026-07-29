/// A user-safe error returned when an API request cannot be completed.
class ApiException implements Exception {
  const ApiException(this.message, {this.code});

  final String message;
  final String? code;

  @override
  String toString() => message;
}
