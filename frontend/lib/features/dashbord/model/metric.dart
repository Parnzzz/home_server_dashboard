class Metric {
  final int id;
  final String serverId;
  final DateTime recordedAt;

  final double cpuPercent;

  final int memoryUsedBytes;
  final int memoryTotalBytes;

  final int diskUsedBytes;
  final int diskTotalBytes;

  const Metric({
    required this.id,
    required this.serverId,
    required this.recordedAt,
    required this.cpuPercent,
    required this.memoryUsedBytes,
    required this.memoryTotalBytes,
    required this.diskUsedBytes,
    required this.diskTotalBytes,
  });

  factory Metric.fromJson(Map<String, dynamic> json) {
    return Metric(
      id: (json['id'] as num).toInt(),
      serverId: json['server_id'] as String,
      recordedAt: DateTime.parse(json['recorded_at'] as String),

      cpuPercent: (json['cpu_percent'] as num).toDouble(),

      memoryUsedBytes: (json['memory_used_bytes'] as num).toInt(),
      memoryTotalBytes: (json['memory_total_bytes'] as num).toInt(),

      diskUsedBytes: (json['disk_used_bytes'] as num).toInt(),
      diskTotalBytes: (json['disk_total_bytes'] as num).toInt(),
    );
  }
}