import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import 'services/api_client.dart';
import 'services/local_buffer.dart';
import 'services/location_service.dart';
import 'services/sync_service.dart';
import 'controllers/tracking_controller.dart';
import 'screens/driver_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();

  final buffer = LocalBuffer();
  final sync = SyncService(buffer, ApiClient());

  runApp(
    ChangeNotifierProvider(
      create: (_) => TrackingController(
        location: LocationService(),
        buffer: buffer,
        sync: sync,
      ),
      child: const BusDriverApp(),
    ),
  );
}

class BusDriverApp extends StatelessWidget {
  const BusDriverApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Bus Driver',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF1F3A5F)),
      ),
      home: const DriverScreen(),
    );
  }
}
