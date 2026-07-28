import 'dart:convert';
import 'dart:io';

import 'package:mobile/core/config/app_environment.dart';
import 'package:mobile/core/network/api_exception.dart';

/// HTTP boundary for the Finora backend.
class ApiClient {
  String? token;

  Future<dynamic> request(
    String method,
    String path, [
    Map<String, dynamic>? body,
  ]) async {
    final client = HttpClient();
    try {
      final request = await client.openUrl(
        method,
        Uri.parse('${AppEnvironment.apiBase}/api/v1$path'),
      );
      request.headers.contentType = ContentType.json;
      if (token != null) {
        request.headers.set(HttpHeaders.authorizationHeader, 'Bearer $token');
      }
      if (body != null) {
        request.headers.set(
          'Idempotency-Key',
          'mobile-${DateTime.now().microsecondsSinceEpoch}',
        );
        request.write(jsonEncode(body));
      }
      final response = await request.close();
      final raw = await utf8.decodeStream(response);
      final data = raw.isEmpty ? null : jsonDecode(raw);
      if (response.statusCode < 200 || response.statusCode >= 300) {
        final message = data is Map
            ? data['message'] ?? data['error'] ?? raw
            : raw;
        throw ApiException(
          message.toString().isEmpty
              ? 'Yêu cầu thất bại (${response.statusCode})'
              : message.toString(),
        );
      }
      return data;
    } on SocketException {
      throw ApiException(
        'Không kết nối được backend tại ${AppEnvironment.apiBase}. '
        'Kiểm tra API_BASE và server :8080.',
      );
    } finally {
      client.close(force: true);
    }
  }
}
