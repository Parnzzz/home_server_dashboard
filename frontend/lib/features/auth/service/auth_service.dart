import 'package:dio/dio.dart';

import '../model/auth_response.dart';
import '../model/auth_user.dart';

class AuthService {
  final Dio _dio;

  AuthService(this._dio);

  Future<UserProfile> register({
    required String email,
    required String displayName,
    required String password,
  }) async {
    try {
      final response = await _dio.post(
        '/api/v1/auth/register',
        data: {
          'email': email.trim(),
          'display_name': displayName.trim(),
          'password': password,
        },
        options: Options(
          extra: {
            'requiresAuth': false,
          },
        ),
      );

      return UserProfile.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (e) {
      final data = e.response?.data;
      final error = data is Map ? data['error'] : null;

      if (error == 'email already exists') {
        throw Exception('An account with this email already exists.');
      }

      if (error == 'invalid registration data') {
        throw Exception('Please check your details and try again.');
      }

      if (error is String && error.isNotEmpty) {
        throw Exception(error);
      }

      if (e.response == null) {
        throw Exception(
          'Unable to connect to server. Please check your connection.',
        );
      }

      throw Exception('Registration failed. Please try again.');
    }
  }

  Future<AuthResponse> login({
    required String email,
    required String password,
  }) async {
    try {
      final response = await _dio.post(
        '/api/v1/auth/login',
        data: {
          'email': email,
          'password': password,
        },
        options: Options(
          extra: {
            'requiresAuth': false,
          },
        ),
      );

      return AuthResponse.fromJson(
        response.data as Map<String, dynamic>,
      );
    } on DioException catch (e) {
      final data = e.response?.data;

      if (data is Map<String, dynamic>) {
        final message = data['error'] ?? data['message'];

        if (message is String && message.isNotEmpty) {
          throw Exception(message);
        }
      }

      if (e.response == null) {
        throw Exception(
          'Unable to connect to server. Please check your connection.',
        );
      }

      throw Exception('Login failed. Please try again.');
    }
  }

  Future<void> logout({required String refreshToken}) async {
    await _dio.post(
      '/api/v1/auth/logout',
      data: {'refresh_token': refreshToken},
      options: Options(
        extra: {
          'requiresAuth': false,
        },
      ),
    );
  }

  Future<void> requestPasswordReset({required String email}) async {
    try {
      await _dio.post(
        '/api/v1/auth/password-reset/request',
        data: {'email': email.trim()},
        options: Options(extra: {'requiresAuth': false}),
      );
    } on DioException catch (e) {
      final data = e.response?.data;
      final message = data is Map ? data['error'] : null;
      if (message is String && message.isNotEmpty) {
        throw Exception(message);
      }
      if (e.response == null) {
        throw Exception(
          'Unable to connect to server. Please check your connection.',
        );
      }
      throw Exception('Unable to send a reset code. Please try again.');
    }
  }

  Future<void> resetPassword({
    required String email,
    required String code,
    required String newPassword,
  }) async {
    try {
      await _dio.post(
        '/api/v1/auth/password-reset/confirm',
        data: {
          'email': email.trim(),
          'code': code.trim(),
          'new_password': newPassword,
        },
        options: Options(extra: {'requiresAuth': false}),
      );
    } on DioException catch (e) {
      final data = e.response?.data;
      final message = data is Map ? data['error'] : null;
      if (message is String && message.isNotEmpty) {
        throw Exception(message);
      }
      if (e.response == null) {
        throw Exception(
          'Unable to connect to server. Please check your connection.',
        );
      }
      throw Exception('Unable to reset password. Please try again.');
    }
  }

  Future<UserProfile> getCurrentUser(String accessToken) async {
    final response = await _dio.get(
      '/api/v1/auth/me',
      options: Options(
        headers: {
          'Authorization': 'Bearer $accessToken',
        },
      ),
    );

    return UserProfile.fromJson(
      response.data as Map<String, dynamic>,
    );
  }
}

