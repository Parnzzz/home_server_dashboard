class DashboardSummary {
  final ServerInfo server;
  final double? cpuPercent;
  final int? memoryUsedBytes;
  final int? memoryTotalBytes;
  final int? diskUsedBytes;
  final int? diskTotalBytes;
  final DateTime? lastSeenAt;

  DashboardSummary({
    required this.server,
    this.cpuPercent,
    this.memoryUsedBytes,
    this.memoryTotalBytes,
    this.diskUsedBytes,
    this.diskTotalBytes,
    this.lastSeenAt,
  });

  factory DashboardSummary.fromJson(Map<String, dynamic> json) {
    return DashboardSummary(
      server: ServerInfo.fromJson(json['server']),
      cpuPercent: (json['cpu_percent'] as num?)?.toDouble(),
      memoryUsedBytes: json['memory_used_bytes'] as int?,
      memoryTotalBytes: json['memory_total_bytes'] as int?,
      diskUsedBytes: json['disk_used_bytes'] as int?,
      diskTotalBytes: json['disk_total_bytes'] as int?,
      lastSeenAt: json['last_seen_at'] != null
          ? DateTime.parse(json['last_seen_at'])
          : null,
    );
  }
}

class ServerInfo {
  final String id;
  final String name;
  final String hostname;
  final String os;
  final String status;
  final DateTime createdAt;

  ServerInfo({
    required this.id,
    required this.name,
    required this.hostname,
    required this.os,
    required this.status,
    required this.createdAt,
  });

  factory ServerInfo.fromJson(Map<String, dynamic> json) {
    return ServerInfo(
      id: json['id'],
      name: json['name'],
      hostname: json['hostname'],
      os: json['os'],
      status: json['status'],
      createdAt: DateTime.parse(json['created_at']),
    );
  }
}