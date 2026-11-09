import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

import '../../model/metric.dart';

class DiskChart extends StatelessWidget {
  final List<Metric> metrics;

  const DiskChart({
    super.key,
    required this.metrics,
  });

  double _toGB(int bytes) {
    return bytes / (1024 * 1024 * 1024);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    final sortedMetrics = metrics.reversed.toList();

    if (sortedMetrics.isEmpty) {
      return const Center(
        child: Text('No Disk data available'),
      );
    }

    final spots = sortedMetrics.map((metric) {
      return FlSpot(
        metric.recordedAt.millisecondsSinceEpoch.toDouble(),
        _toGB(metric.diskUsedBytes),
      );
    }).toList();

    final xInterval = spots.length > 1
        ? (spots.last.x - spots.first.x) / 3
        : 1.0;

    final showSeconds = spots.length > 1 &&
        spots.last.x - spots.first.x <
            const Duration(hours: 1).inMilliseconds;

    final maxDisk = sortedMetrics
        .map((metric) => _toGB(metric.diskTotalBytes))
        .reduce((a, b) => a > b ? a : b);

    final maxY = maxDisk <= 0 ? 1.0 : maxDisk * 1.1;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Disk Usage',
              style: theme.textTheme.titleLarge,
            ),
            const SizedBox(height: 20),
            SizedBox(
              height: 220,
              child: LineChart(
                LineChartData(
                  minY: 0,
                  maxY: maxY,
                  gridData: FlGridData(
                    show: true,
                    drawVerticalLine: false,
                    getDrawingHorizontalLine: (value) {
                      return FlLine(
                        color: colorScheme.outlineVariant,
                        strokeWidth: 0.5,
                      );
                    },
                  ),
                  lineTouchData: LineTouchData(
                    enabled: true,
                    touchTooltipData: LineTouchTooltipData(
                      getTooltipItems: (touchedSpots) {
                        return touchedSpots.map((spot) {
                          final date = DateTime.fromMillisecondsSinceEpoch(
                            spot.x.toInt(),
                          );

                          final time =
                              '${date.hour.toString().padLeft(2, '0')}:'
                              '${date.minute.toString().padLeft(2, '0')}:'
                              '${date.second.toString().padLeft(2, '0')}';

                          return LineTooltipItem(
                            'Disk: ${spot.y.toStringAsFixed(2)} GB\n$time',
                            TextStyle(
                              color: colorScheme.onInverseSurface,
                              fontWeight: FontWeight.w600,
                              fontSize: 12,
                            ),
                          );
                        }).toList();
                      },
                    ),
                  ),
                  titlesData: FlTitlesData(
                    topTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false),
                    ),
                    rightTitles: const AxisTitles(
                      sideTitles: SideTitles(showTitles: false),
                    ),
                    leftTitles: AxisTitles(
                      sideTitles: SideTitles(
                        showTitles: true,
                        reservedSize: 48,
                        interval: maxY / 4,
                        getTitlesWidget: (value, meta) {
                          return Text(
                            '${value.toStringAsFixed(1)} GB',
                            style: theme.textTheme.bodySmall,
                          );
                        },
                      ),
                    ),
                    bottomTitles:  AxisTitles(
                      sideTitles: SideTitles(
                        showTitles: true,
                        reservedSize: 32,
                        interval: xInterval,
                        getTitlesWidget: (value, meta) {
                          final date = DateTime.fromMillisecondsSinceEpoch(
                            value.toInt(),
                          );

                          final hour = date.hour.toString().padLeft(2, '0');
                          final minute = date.minute.toString().padLeft(2, '0');
                          final second = date.second.toString().padLeft(2, '0');

                          final label = showSeconds
                              ? '$hour:$minute:$second'
                              : '$hour:$minute';

                          return SideTitleWidget(
                            meta: meta,
                            child: Text(
                              label,
                              style: theme.textTheme.bodySmall,
                            ),
                          );
                        },
                      ),
                    )
                  ),
                  borderData: FlBorderData(show: false),
                  lineBarsData: [
                    LineChartBarData(
                      spots: spots,
                      isCurved: true,
                      color: colorScheme.secondary,
                      barWidth: 2.5,
                      dotData: const FlDotData(show: false),
                      belowBarData: BarAreaData(
                        show: true,
                        color: colorScheme.secondary.withValues(
                          alpha: 0.12,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}