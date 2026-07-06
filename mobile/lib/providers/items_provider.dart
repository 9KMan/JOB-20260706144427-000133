import 'package:flutter/foundation.dart';
import '../models/item.dart';
import '../services/api_client.dart';

class ItemsProvider extends ChangeNotifier {
  List<Item> _items = [];
  bool _loading = false;
  String? _error;

  final ApiClient _apiClient;

  ItemsProvider(this._apiClient);

  List<Item> get items => List.unmodifiable(_items);
  bool get loading => _loading;
  String? get error => _error;

  Future<void> load() async {
    _loading = true;
    _error = null;
    notifyListeners();

    try {
      _items = await _apiClient.listItems();
      _error = null;
    } catch (e) {
      _items = [];
      _error = e.toString();
    } finally {
      _loading = false;
      notifyListeners();
    }
  }

  Future<void> createItem(String name) async {
    try {
      final item = await _apiClient.createItem(name);
      _items.add(item);
      notifyListeners();
    } catch (e) {
      _error = e.toString();
      notifyListeners();
      rethrow;
    }
  }
}
