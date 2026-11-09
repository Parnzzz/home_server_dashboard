import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';

import '../../model/metric.dart';

class CpuChart extends StatelessWidget {
  final List<Metric> metrics;

  const CpuChart({
    super.key,
    required this.metrics,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    final sortedMetrics = metrics.reversed.toList();

    if (sortedMetrics.isEmpty) {
      return const Center(
        child: Text('No CPU data available'),
      );
    }

    final spots = sortedMetrics
        .map(
          (metric) => FlSpot(
        metric.recordedAt.millisecondsSinceEpoch.toDouble(),
        metric.cpuPercent,
      ),
    )
        .toList();

    final xInterval = spots.length > 1
        ? (spots.last.x - spots.first.x) / 3
        : 1.0;

    final showSeconds = spots.length > 1 &&
        spots.last.x - spots.first.x <
            const Duration(hours: 1).inMilliseconds;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'CPU Usage',
              style: theme.textTheme.titleLarge,
            ),
            const SizedBox(height: 20),
            SizedBox(
              height: 220,
              child: LineChart(
                LineChartData(
                  minY: 0,
                  maxY: 100,
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
                            'CPU: ${spot.y.toStringAsFixed(1)}%\n$time',
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
                        reservedSize: 42,
                        interval: 25,
                        getTitlesWidget: (value, meta) {
                          return Text(
                            '${value.toInt()}%',
                            style: theme.textTheme.bodySmall,
                          );
                        },
                      ),
                    ),
                    bottomTitles: AxisTitles(
                      sideTitles: SideTitles(
                        showTitles: true,
                        reservedSize: 32,
                        interval: xInterval,
                        getTitlesWidget: (value, meta) {
                          if (value >= spots.last.x - xInterval * 0.5) {
                            return const SizedBox.shrink();
                          }

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
                      color: colorScheme.primary,
                      barWidth: 2.5,
                      dotData: const FlDotData(show: false),
                      belowBarData: BarAreaData(
                        show: true,
                        color: colorScheme.primary.withValues(
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