import 'package:flutter/material.dart';

import '../../model/dashboard_summary.dart';
import '../../model/metric.dart';
import '../common/cpu_chart.dart';
import '../common/dashboard_metric.dart';
import '../common/disk_chart.dart';
import '../common/memory_chart.dart';
import '../common/server_status_card.dart';

class DashboardPortrait extends StatelessWidget {
  final DashboardSummary summary;
  final List<Metric> metrics;

  const DashboardPortrait({
    super.key,
    required this.summary,
    required this.metrics
  });

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ServerStatusCard(
            server: summary.server,
            lastSeenAt: summary.lastSeenAt,
          ),
          const SizedBox(height: 16),

          DashboardMetrics(
            summary: summary,
            crossAxisCount: 2,
          ),
          const SizedBox(height: 16),

          CpuChart(metrics: metrics),
          const SizedBox(height: 16),

          MemoryChart(metrics: metrics),
          const SizedBox(height: 16),

          DiskChart(metrics: metrics),
        ],
      ),
    );
  }
}