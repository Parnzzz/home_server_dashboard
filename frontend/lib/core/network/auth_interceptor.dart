import 'package:dio/dio.dart';

import '../storage/token_storage.dart';

class AuthInterceptor extends Interceptor {
  final TokenStorage _tokenStorage;
  final Dio _dio;
  final String _baseUrl;

  Future<String?>? _refreshFuture;

  AuthInterceptor({
    required TokenStorage tokenStorage,
    required Dio dio,
    required String baseUrl,
  })  : _tokenStorage = tokenStorage,
        _dio = dio,
        _baseUrl = baseUrl;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    if (options.extra['requiresAuth'] == false) {
      handler.next(options);
      return;
    }

    final accessToken = await _tokenStorage.getAccessToken();

    if (accessToken != null && accessToken.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $accessToken';
    }

    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final request = err.requestOptions;
    final isAuthRequest = request.extra['requiresAuth'] == false ||
        request.path.endsWith('/api/v1/auth/login') ||
        request.path.endsWith('/api/v1/auth/register') ||
        request.path.endsWith('/api/v1/auth/refresh');

    if (err.response?.statusCode != 401 ||
        isAuthRequest ||
        request.extra['authRetry'] == true) {
      handler.next(err);
      return;
    }

    try {
      final accessToken = await _refreshAccessToken();
      if (accessToken == null || accessToken.isEmpty) {
        await _tokenStorage.clearTokens();
        handler.next(err);
        return;
      }

      request.headers['Authorization'] = 'Bearer $accessToken';
      request.extra['authRetry'] = true;

      final response = await _dio.fetch<dynamic>(request);
      handler.resolve(response);
    } on DioException catch (refreshError) {
      if (refreshError.response?.statusCode == 401) {
        await _tokenStorage.clearTokens();
        handler.next(err);
      } else {
        handler.next(refreshError);
      }
    } catch (_) {
      handler.next(err);
    }
  }

  Future<String?> _refreshAccessToken() async {
    final currentRefresh = _refreshFuture;
    if (currentRefresh != null) return currentRefresh;

    final refreshFuture = _requestNewTokens();
    _refreshFuture = refreshFuture;

    try {
      return await refreshFuture;
    } finally {
      if (identical(_refreshFuture, refreshFuture)) {
        _refreshFuture = null;
      }
    }
  }

  Future<String?> _requestNewTokens() async {
    final refreshToken = await _tokenStorage.getRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) return null;

    final refreshDio = Dio(
      BaseOptions(
        baseUrl: _baseUrl,
        connectTimeout: const Duration(seconds: 10),
        receiveTimeout: const Duration(seconds: 10),
        sendTimeout: const Duration(seconds: 10),
        headers: const {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
      ),
    );

    final response = await refreshDio.post<Map<String, dynamic>>(
      '/api/v1/auth/refresh',
      data: {'refresh_token': refreshToken},
    );

    final data = response.data;
    final accessToken = data?['access_token'] as String?;
    final newRefreshToken = data?['refresh_token'] as String?;
    if (accessToken == null ||
        accessToken.isEmpty ||
        newRefreshToken == null ||
        newRefreshToken.isEmpty) {
      throw const FormatException('Invalid token refresh response');
    }

    await _tokenStorage.saveTokens(
      accessToken: accessToken,
      refreshToken: newRefreshToken,
    );

    return accessToken;
  }
}
