import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/items_provider.dart';

class ItemsScreen extends StatefulWidget {
  const ItemsScreen({super.key});

  @override
  State<ItemsScreen> createState() => _ItemsScreenState();
}

class _ItemsScreenState extends State<ItemsScreen> {
  final GlobalKey<ScaffoldMessengerState> _scaffoldMessengerKey =
      GlobalKey<ScaffoldMessengerState>();
  String? _previousError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      Provider.of<ItemsProvider>(context, listen: false).load();
    });
  }

  void _showSnackBar(String message, {VoidCallback? onRetry}) {
    final scaffold = _scaffoldMessengerKey.currentState;
    if (scaffold == null) return;

    scaffold.showSnackBar(
      SnackBar(
        content: Text(message),
        action: onRetry != null
            ? SnackBarAction(
                label: 'Retry',
                onPressed: onRetry,
              )
            : null,
      ),
    );
  }

  Future<void> _showAddItemSheet() async {
    final textController = TextEditingController();
    final provider = Provider.of<ItemsProvider>(context, listen: false);

    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      builder: (context) {
        return Padding(
          padding: EdgeInsets.only(
            left: 16,
            right: 16,
            top: 16,
            bottom: MediaQuery.of(context).viewInsets.bottom + 16,
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextField(
                controller: textController,
                decoration: const InputDecoration(
                  labelText: 'Item Name',
                  border: OutlineInputBorder(),
                ),
                autofocus: true,
              ),
              const SizedBox(height: 16),
              StatefulBuilder(
                builder: (context, setButtonState) {
                  return ElevatedButton(
                    onPressed: provider.loading
                        ? null
                        : () async {
                            final name = textController.text.trim();
                            if (name.isEmpty) return;

                            setButtonState(() {});
                            try {
                              await provider.createItem(name);
                              if (mounted) {
                                Navigator.of(context).pop();
                                _showSnackBar('Item added successfully');
                              }
                            } catch (e) {
                              setButtonState(() {});
                              if (mounted) {
                                _showSnackBar('Failed to add item');
                              }
                            }
                          },
                    child: provider.loading
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('Add'),
                  );
                },
              ),
            ],
          ),
        );
      },
    );

    textController.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ScaffoldMessenger(
      key: _scaffoldMessengerKey,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Items'),
          actions: [
            Consumer<ItemsProvider>(
              builder: (context, provider, _) {
                return IconButton(
                  icon: const Icon(Icons.refresh),
                  onPressed: provider.loading ? null : () => provider.load(),
                );
              },
            ),
          ],
        ),
        body: Consumer<ItemsProvider>(
          builder: (context, provider, _) {
            final error = provider.error;
            if (error != null && error != _previousError) {
              _previousError = error;
              WidgetsBinding.instance.addPostFrameCallback((_) {
                _showSnackBar(error, onRetry: () => provider.load());
              });
            } else if (error == null) {
              _previousError = null;
            }

            if (provider.loading && provider.items.isEmpty) {
              return const Center(
                child: CircularProgressIndicator(),
              );
            }

            if (provider.items.isEmpty) {
              return const Center(
                child: Text('No items yet'),
              );
            }

            return ListView.builder(
              itemCount: provider.items.length,
              itemBuilder: (context, index) {
                final item = provider.items[index];
                return ListTile(
                  title: Text(item.name),
                  subtitle: Text(item.createdAt.toLocal().toString()),
                );
              },
            );
          },
        ),
        floatingActionButton: FloatingActionButton(
          onPressed: _showAddItemSheet,
          child: const Icon(Icons.add),
        ),
      ),
    );
  }
}
