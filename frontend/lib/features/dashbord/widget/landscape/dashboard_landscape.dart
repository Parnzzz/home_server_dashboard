import 'package:flutter/material.dart';

import '../../model/dashboard_summary.dart';
import '../../model/metric.dart';
import '../common/cpu_chart.dart';
import '../common/dashboard_metric.dart';
import '../common/disk_chart.dart';
import '../common/memory_chart.dart';
import '../common/server_status_card.dart';

class DashboardLandscape extends StatelessWidget {
  final DashboardSummary summary;
  final List<Metric> metrics;

  const DashboardLandscape({
    super.key,
    required this.summary,
    required this.metrics
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 280,
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: ServerStatusCard(
              server: summary.server,
              lastSeenAt: summary.lastSeenAt,
            ),
          ),
        ),

        const VerticalDivider(width: 1),

        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                DashboardMetrics(
                  summary: summary,
                  crossAxisCount: 3,
                ),
                const SizedBox(height: 16),

                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: CpuChart(metrics: metrics),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: MemoryChart(metrics: metrics),
                    ),
                  ],
                ),

                const SizedBox(height: 16),

                DiskChart(metrics: metrics),
              ],
            ),
          ),
        ),
      ],
    );
  }
}