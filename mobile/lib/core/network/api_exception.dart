/// A user-safe error returned when an API request cannot be completed.
class ApiException implements Exception {
  const ApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
