import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/items_provider.dart';
import 'services/api_client.dart';
import 'screens/items_screen.dart';

void main() {
  runApp(const GruxPocApp());
}

class GruxPocApp extends StatelessWidget {
  const GruxPocApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => ItemsProvider(ApiClient()),
      child: MaterialApp(
        title: 'Grux PoC',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(
          colorSchemeSeed: Colors.lightBlue,
          useMaterial3: true,
          scaffoldBackgroundColor: Colors.blueGrey.shade50,
        ),
        home: const ItemsScreen(),
      ),
    );
  }
}
