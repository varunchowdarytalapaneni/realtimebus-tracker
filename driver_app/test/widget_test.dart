// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';

import 'package:driver_app/main.dart';
import 'package:driver_app/controllers/tracking_controller.dart';
import 'package:driver_app/services/api_client.dart';
import 'package:driver_app/services/local_buffer.dart';
import 'package:driver_app/services/location_service.dart';
import 'package:driver_app/services/sync_service.dart';

void main() {
  testWidgets('renders the driver tracking screen', (
    WidgetTester tester,
  ) async {
    final buffer = LocalBuffer();
    await tester.pumpWidget(
      ChangeNotifierProvider(
        create: (_) => TrackingController(
          location: LocationService(),
          buffer: buffer,
          sync: SyncService(buffer, ApiClient()),
        ),
        child: const BusDriverApp(),
      ),
    );

    expect(find.text('Driver Tracking'), findsOneWidget);
    expect(find.text('Start trip'), findsOneWidget);
  });
}
