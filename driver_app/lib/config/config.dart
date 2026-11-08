class AppConfig {
  /// Android emulator -> host machine is 10.0.2.2
  /// iOS simulator   -> http://localhost:8080
  /// Real phone      -> your computer's LAN IP, e.g. http://192.168.1.20:8080
  /// Override at run time:  flutter run --dart-define=API_BASE_URL=http://192.168.1.20:8080
  static const baseUrl =
      String.fromEnvironment('API_BASE_URL', defaultValue: 'http://10.0.2.2:8080');

  static const defaultBusId = 1;
  static const batchSize = 500; // backend accepts at most 500 points per request
  static const retryInterval = Duration(seconds: 15);
  static const minPingGap = Duration(seconds: 2);
  static const maxBufferedPings = 50000; // oldest are dropped beyond this
}