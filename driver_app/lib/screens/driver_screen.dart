import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

import '../config/config.dart';
import '../controllers/tracking_controller.dart';

class DriverScreen extends StatefulWidget {
  const DriverScreen({super.key});
  @override
  State<DriverScreen> createState() => _DriverScreenState();
}

class _DriverScreenState extends State<DriverScreen> {
  final _busIdCtl = TextEditingController(
    text: AppConfig.defaultBusId.toString(),
  );

  @override
  void dispose() {
    _busIdCtl.dispose();
    super.dispose();
  }

  void _toggle(TrackingController c) {
    if (c.status == TrackingStatus.tracking) {
      c.stop();
      return;
    }
    final id = int.tryParse(_busIdCtl.text.trim());
    if (id == null || id <= 0) {
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('Enter a valid Bus ID')));
      return;
    }
    c.busId = id;
    c.start();
  }

  @override
  Widget build(BuildContext context) {
    final c = context.watch<TrackingController>();
    final tracking = c.status == TrackingStatus.tracking;
    final starting = c.status == TrackingStatus.starting;
    final speedKmh = ((c.last?.speed ?? 0) * 3.6).toStringAsFixed(0);

    return Scaffold(
      appBar: AppBar(title: const Text('Driver Tracking')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextField(
                controller: _busIdCtl,
                enabled: !tracking && !starting,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(
                  labelText: 'Bus ID',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              _StatusBanner(c: c),
              const SizedBox(height: 16),
              Expanded(
                child: GridView.count(
                  crossAxisCount: 2,
                  childAspectRatio: 1.6,
                  mainAxisSpacing: 12,
                  crossAxisSpacing: 12,
                  children: [
                    _Stat(
                      'Connection',
                      !tracking ? '—' : (c.online ? 'Online' : 'Offline'),
                    ),
                    _Stat('Pending sync', '${c.pending}'),
                    _Stat('Speed', '$speedKmh km/h'),
                    _Stat('Last sync', _time(c.lastSyncAt)),
                  ],
                ),
              ),
              if (c.last != null)
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: Text(
                    '${c.last!.latitude.toStringAsFixed(5)}, ${c.last!.longitude.toStringAsFixed(5)}',
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                ),
              SizedBox(
                height: 56,
                child: FilledButton(
                  style: tracking
                      ? FilledButton.styleFrom(
                          backgroundColor: Colors.red.shade700,
                        )
                      : null,
                  onPressed: starting ? null : () => _toggle(c),
                  child: Text(
                    tracking
                        ? 'Stop trip'
                        : (starting ? 'Starting…' : 'Start trip'),
                    style: const TextStyle(fontSize: 18),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _time(DateTime? t) {
    if (t == null) return '—';
    String two(int n) => n.toString().padLeft(2, '0');
    return '${two(t.hour)}:${two(t.minute)}:${two(t.second)}';
  }
}

class _StatusBanner extends StatelessWidget {
  const _StatusBanner({required this.c});
  final TrackingController c;

  @override
  Widget build(BuildContext context) {
    late final String text;
    late final Color color;
    switch (c.status) {
      case TrackingStatus.tracking:
        text = 'Tracking — location is being shared';
        color = Colors.green.shade700;
      case TrackingStatus.starting:
        text = 'Starting…';
        color = Colors.orange.shade800;
      case TrackingStatus.error:
        text = c.error ?? 'Something went wrong';
        color = Colors.red.shade700;
      case TrackingStatus.stopped:
        text = 'Trip not started';
        color = Colors.blueGrey;
    }
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        border: Border.all(color: color),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        text,
        style: TextStyle(color: color, fontWeight: FontWeight.w600),
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat(this.label, this.value);
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        side: BorderSide(color: Theme.of(context).dividerColor),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Padding(
        padding: const EdgeInsets.all(14),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label, style: Theme.of(context).textTheme.labelMedium),
            const SizedBox(height: 4),
            Text(value, style: Theme.of(context).textTheme.titleLarge),
          ],
        ),
      ),
    );
  }
}
