import 'package:flutter/material.dart';

import '../../model/dashboard_summary.dart';
import 'metric_card.dart';

class DashboardMetrics extends StatelessWidget {
  final DashboardSummary summary;
  final int crossAxisCount;

  const DashboardMetrics({
    super.key,
    required this.summary,
    this.crossAxisCount = 2,
  });

  String _formatGB(int? bytes) {
    if (bytes == null) return '--';

    final gb = bytes / (1024 * 1024 * 1024);
    return gb.toStringAsFixed(2);
  }

  double? _getProgress(int? used, int? total) {
    if (used == null || total == null || total <= 0) {
      return null;
    }

    return (used / total).clamp(0.0, 1.0);
  }

  @override
  Widget build(BuildContext context) {
    final cpu = summary.cpuPercent;

    return GridView.count(
      crossAxisCount: crossAxisCount,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      crossAxisSpacing: 12,
      mainAxisSpacing: 12,
      childAspectRatio: 1.0,
      children: [
        MetricCard(
          title: 'CPU Usage',
          value: cpu?.toStringAsFixed(1) ?? '--',
          unit: '%',
          detail: 'Processor utilization',
          icon: Icons.memory,
          progress: cpu == null ? null : cpu / 100,
        ),
        MetricCard(
          title: 'Memory',
          value: _formatGB(summary.memoryUsedBytes),
          unit: 'GB',
          detail:
          'Total ${_formatGB(summary.memoryTotalBytes)} GB',
          icon: Icons.storage,
          progress: _getProgress(
            summary.memoryUsedBytes,
            summary.memoryTotalBytes,
          ),
        ),
        MetricCard(
          title: 'Disk',
          value: _formatGB(summary.diskUsedBytes),
          unit: 'GB',
          detail:
          'Total ${_formatGB(summary.diskTotalBytes)} GB',
          icon: Icons.disc_full,
          progress: _getProgress(
            summary.diskUsedBytes,
            summary.diskTotalBytes,
          ),
        ),
      ],
    );
  }
}