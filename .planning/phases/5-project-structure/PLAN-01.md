---
phase: 5
plan: 01
type: IMPLEMENTATION
wave: 3
depends_on:
  - phase: 1
  - phase: 2
  - phase: 3
  - phase: 4
plans: ["01"]
files_modified:
  - mobile/pubspec.yaml
  - mobile/lib/main.dart
  - mobile/lib/screens/items_screen.dart
  - mobile/lib/services/api_client.dart
  - mobile/lib/models/item.dart
  - mobile/test/widget_test.dart
autonomous: true
requirements:
  - Flutter 3.x app
  - Single screen: items list + add form
  - HTTP client wrapping Go backend
  - Pubspec deps: http, provider
  - Widget test
---

# Phase 5 — Project Structure (Flutter mobile app)

## What this phase delivers

The Flutter app: a single screen that calls the Go backend, displays the
items list, and has an add form. Demonstrates cross-platform capability.

## Files to Create

| File | Purpose |
|---|---|
| `mobile/pubspec.yaml` | Flutter project metadata + dependencies |
| `mobile/pubspec.lock` | Pinned dependency versions (auto-generated) |
| `mobile/analysis_options.yaml` | Dart analyzer config (flutter_lints) |
| `mobile/lib/main.dart` | App entry: Material theme + routing + Provider setup |
| `mobile/lib/screens/items_screen.dart` | Main screen: ListView of items + add form |
| `mobile/lib/services/api_client.dart` | HTTP client wrapping Go API base URL |
| `mobile/lib/models/item.dart` | Item model with fromJson/toJson |
| `mobile/lib/providers/items_provider.dart` | State management: items list + loading/error |
| `mobile/test/widget_test.dart` | Widget test: app renders without errors |

## Implementation requirements

### `mobile/pubspec.yaml`
```yaml
name: grux_poc_mobile
description: Grux PoC Flutter app
publish_to: 'none'
version: 0.1.0+1

environment:
  sdk: '>=3.4.0 <4.0.0'
  flutter: '>=3.22.0'

dependencies:
  flutter:
    sdk: flutter
  http: ^1.2.0
  provider: ^6.1.1
  cupertino_icons: ^1.0.6

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^4.0.0
```

### `mobile/lib/main.dart`
- MaterialApp with title "Grux PoC"
- Theme: light blue accent
- Home: `ItemsScreen()`
- Provider: `ChangeNotifierProvider<ItemsProvider>`

### `mobile/lib/screens/items_screen.dart`
- StatefulWidget
- AppBar: title "Items", action refresh button
- Body: ListView of items (loading spinner if loading)
- FAB: opens add-item bottom sheet
- Add sheet: TextField + Submit button → calls `provider.createItem(name)`
- Error snackbar if API call fails

### `mobile/lib/services/api_client.dart`
- Base URL from const: `String.fromEnvironment('API_BASE_URL', defaultValue: 'http://localhost:8080')`
- Methods: `Future<List<Item>> listItems()`, `Future<Item> createItem(String name)`, `Future<Item> getItem(String id)`
- Uses `package:http`
- Throws `ApiException` on non-2xx status

### `mobile/lib/models/item.dart`
- `Item({required this.id, required this.name, required this.createdAt})`
- `factory Item.fromJson(Map<String, dynamic> json)`
- `Map<String, dynamic> toJson()`

### `mobile/lib/providers/items_provider.dart`
- `ChangeNotifier` subclass
- State: `List<Item> items`, `bool loading`, `String? error`
- Methods: `load()`, `createItem(String name)`
- On error: sets `error`, listeners show snackbar

### `mobile/test/widget_test.dart`
- Smoke test: pump `GruxPocApp`, verify `ItemsScreen` title appears

## Verification

```bash
cd mobile
flutter pub get
flutter analyze    # expect 0 errors
flutter test       # expect 1+ PASS
flutter build apk --debug  # expect SUCCESS (optional, takes time)
```

## Acceptance criteria

- [ ] `flutter pub get` succeeds
- [ ] `flutter analyze` returns 0 errors
- [ ] `flutter test` returns ≥1 PASS
- [ ] App structure: `main.dart` → `ItemsScreen` → `ApiClient` → Go backend
- [ ] Provider state management pattern (not setState directly)