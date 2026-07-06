import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:grux_poc_mobile/main.dart';
import 'package:grux_poc_mobile/models/item.dart';
import 'package:grux_poc_mobile/services/api_client.dart';

class FakeApiClient extends ApiClient {
  FakeApiClient() : super(client: http.Client());

  @override
  Future<List<Item>> listItems() async {
    return [
      Item(
        id: '1',
        name: 'Test Item 1',
        createdAt: DateTime(2024, 1, 15, 10, 30),
      ),
      Item(
        id: '2',
        name: 'Test Item 2',
        createdAt: DateTime(2024, 1, 16, 14, 45),
      ),
    ];
  }

  @override
  Future<Item> createItem(String name) async {
    return Item(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      name: name,
      createdAt: DateTime.now(),
    );
  }

  @override
  Future<Item> getItem(String id) async {
    return Item(
      id: id,
      name: 'Test Item',
      createdAt: DateTime.now(),
    );
  }
}

void main() {
  testWidgets('App shows Items title and loading indicator initially',
      (WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: ChangeNotifierProvider(
          create: (_) => ItemsProvider(FakeApiClient()),
          child: const ItemsScreen(),
        ),
      ),
    );

    expect(find.text('Items'), findsOneWidget);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);
  });
}
