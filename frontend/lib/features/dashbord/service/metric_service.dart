
import '../../../core/network/api_client.dart';
import '../model/metric.dart';

class MetricsService {
  final ApiClient _apiClient;

  MetricsService(this._apiClient);

  Future<List<Metric>> getMetrics(
      String serverId, {
        int limit = 20,
        String? from,
      }) async {
    final response = await _apiClient.request<List<dynamic>>(
      '/api/v1/servers/$serverId/metrics',
      method: 'GET',
      queryParameters: {
        'limit': limit,
        'from': ?from,
      },
    );

    final data = response.data ?? [];

    return data
        .map(
          (item) => Metric.fromJson(
        item as Map<String, dynamic>,
      ),
    )
        .toList();
  }
}