import 'dart:convert';

import 'package:http/http.dart' as http;

import '../config/config.dart';
import '../models/ping.dart';

enum SendResult {
  ok, // server accepted: safe to delete locally
  rejected, // server says the data is invalid (4xx): retrying won't help
  retry, // network error / server error: keep and try later
}

class ApiClient {
  ApiClient({http.Client? client}) : _client = client ?? http.Client();
  final http.Client _client;

  Future<SendResult> sendBatch(List<Ping> pings) async {
    try {
      final res = await _client
          .post(
            Uri.parse('${AppConfig.baseUrl}/api/v1/telemetry/batch'),
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({'points': pings.map((p) => p.toJson()).toList()}),
          )
          .timeout(const Duration(seconds: 10));

      if (res.statusCode >= 200 && res.statusCode < 300) {
        return SendResult.ok;
      }
      if (res.statusCode == 408 || res.statusCode == 429) {
        return SendResult.retry;
      }
      if (res.statusCode >= 400 && res.statusCode < 500) {
        return SendResult.rejected;
      }
      return SendResult.retry;
    } catch (_) {
      return SendResult.retry; // offline, DNS failure, timeout...
    }
  }
}
