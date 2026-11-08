import 'package:geolocator/geolocator.dart';

/// One GPS reading, in the shape the Go backend expects.
class Ping {
  final int? id; // SQLite row id (null until stored)
  final int busId;
  final double latitude;
  final double longitude;
  final double speed; // m/s
  final double heading; // degrees [0, 360)
  final DateTime timestamp; // UTC

  const Ping({
    this.id,
    required this.busId,
    required this.latitude,
    required this.longitude,
    required this.speed,
    required this.heading,
    required this.timestamp,
  });

  /// Sanitises values so the backend's validation never rejects a real reading
  /// (geolocator reports -1 for unknown speed/heading).
  factory Ping.fromPosition(int busId, Position p) => Ping(
        busId: busId,
        latitude: p.latitude,
        longitude: p.longitude,
        speed: p.speed < 0 ? 0 : p.speed,
        heading: (p.heading < 0 || p.heading >= 360) ? 0 : p.heading,
        timestamp: p.timestamp.toUtc(),
      );

  Map<String, dynamic> toJson() => {
        'bus_id': busId,
        'latitude': latitude,
        'longitude': longitude,
        'speed': speed,
        'heading': heading,
        'timestamp': timestamp.toIso8601String(), // RFC 3339 / UTC
      };

  Map<String, Object?> toRow() => {
        'bus_id': busId,
        'latitude': latitude,
        'longitude': longitude,
        'speed': speed,
        'heading': heading,
        'recorded_at': timestamp.millisecondsSinceEpoch,
      };

  factory Ping.fromRow(Map<String, Object?> r) => Ping(
        id: r['id'] as int,
        busId: r['bus_id'] as int,
        latitude: (r['latitude'] as num).toDouble(),
        longitude: (r['longitude'] as num).toDouble(),
        speed: (r['speed'] as num).toDouble(),
        heading: (r['heading'] as num).toDouble(),
        timestamp:
            DateTime.fromMillisecondsSinceEpoch(r['recorded_at'] as int, isUtc: true),
      );
}