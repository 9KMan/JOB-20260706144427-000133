import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/item.dart';

class ApiException implements Exception {
  final int statusCode;
  final String message;

  ApiException(this.statusCode, this.message);

  @override
  String toString() => 'ApiException($statusCode): $message';
}

class ApiClient {
  static const String baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://localhost:8080',
  );

  final http.Client _client;

  ApiClient({http.Client? client}) : _client = client ?? http.Client();

  Future<List<Item>> listItems() async {
    try {
      final response = await _client
          .get(Uri.parse('$baseUrl/items'))
          .timeout(const Duration(seconds: 10));

      if (response.statusCode >= 200 && response.statusCode < 300) {
        final List<dynamic> jsonList = json.decode(response.body) as List<dynamic>;
        return jsonList
            .map((e) => Item.fromJson(e as Map<String, dynamic>))
            .toList();
      } else {
        throw ApiException(response.statusCode, response.body);
      }
    } on http.ClientException catch (e) {
      throw ApiException(0, e.message);
    }
  }

  Future<Item> createItem(String name) async {
    try {
      final response = await _client
          .post(
            Uri.parse('$baseUrl/items'),
            headers: {'Content-Type': 'application/json'},
            body: json.encode({'name': name}),
          )
          .timeout(const Duration(seconds: 10));

      if (response.statusCode >= 200 && response.statusCode < 300) {
        return Item.fromJson(json.decode(response.body) as Map<String, dynamic>);
      } else {
        throw ApiException(response.statusCode, response.body);
      }
    } on http.ClientException catch (e) {
      throw ApiException(0, e.message);
    }
  }

  Future<Item> getItem(String id) async {
    try {
      final response = await _client
          .get(Uri.parse('$baseUrl/items/$id'))
          .timeout(const Duration(seconds: 10));

      if (response.statusCode >= 200 && response.statusCode < 300) {
        return Item.fromJson(json.decode(response.body) as Map<String, dynamic>);
      } else {
        throw ApiException(response.statusCode, response.body);
      }
    } on http.ClientException catch (e) {
      throw ApiException(0, e.message);
    }
  }
}
