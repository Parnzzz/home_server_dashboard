
import '../../../core/network/api_client.dart';
import '../model/dashboard_summary.dart';

class DashboardService {
  final ApiClient _apiClient;

  DashboardService(this._apiClient);

  Future<List<ServerInfo>> getServers() async {
    final response = await _apiClient.request<List<dynamic>>(
      '/api/v1/servers',
      method: 'GET',
    );

    return (response.data ?? [])
        .map(
          (item) => ServerInfo.fromJson(item as Map<String, dynamic>),
        )
        .toList();
  }

  Future<ServerInfo> createServer({
    required String name,
    String hostname = '',
    String os = '',
  }) async {
    final response = await _apiClient.request<Map<String, dynamic>>(
      '/api/v1/servers',
      method: 'POST',
      data: {
        'name': name,
        'hostname': hostname,
        'os': os,
      },
    );

    return ServerInfo.fromJson(response.data!);
  }

  Future<void> deleteServer(String serverId) async {
    await _apiClient.request<dynamic>(
      '/api/v1/servers/$serverId',
      method: 'DELETE',
    );
  }

  Future<String> generateAgentToken(String serverId) async {
    final response = await _apiClient.request<Map<String, dynamic>>(
      '/api/v1/servers/$serverId/agent-token',
      method: 'POST',
    );

    final token = response.data?['token'] as String?;
    if (token == null || token.isEmpty) {
      throw const FormatException('Backend returned an invalid agent token');
    }

    return token;
  }

  Future<DashboardSummary> getSummary(String serverId) async {
    final response = await _apiClient.request<Map<String, dynamic>>(
      '/api/v1/servers/$serverId/summary',
      method: 'GET',
    );

    return DashboardSummary.fromJson(response.data!);
  }
}
