import 'package:dio/dio.dart';

import 'api_exception.dart';

class ApiExceptionMapper {
  static ApiException fromDioException(DioException exception) {
    switch (exception.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return const ApiException(
          message: 'Connection timeout',
          code: 'TIMEOUT',
        );

      case DioExceptionType.connectionError:
        return const ApiException(
          message: 'Unable to connect to server',
          code: 'CONNECTION_ERROR',
        );

      case DioExceptionType.badResponse:
        return _fromResponse(exception);

      case DioExceptionType.cancel:
        return const ApiException(
          message: 'Request cancelled',
          code: 'CANCELLED',
        );

      case DioExceptionType.badCertificate:
        return const ApiException(
          message: 'Invalid server certificate',
          code: 'BAD_CERTIFICATE',
        );

      case DioExceptionType.unknown:
        return ApiException(
          message: exception.message ?? 'Unexpected error',
          code: 'UNKNOWN',
        );
      case DioExceptionType.transformTimeout:
        return ApiException(
          message: exception.message ?? 'Unexpected error',
          code: 'transform_Timeout',
        );
    }
  }

  static ApiException _fromResponse(DioException exception) {
    final statusCode = exception.response?.statusCode;

    final message = switch (statusCode) {
      400 => 'Bad request',
      401 => 'Unauthorized',
      403 => 'Forbidden',
      404 => 'Resource not found',
      500 => 'Internal server error',
      502 => 'Bad gateway',
      503 => 'Service unavailable',
      _ => 'Something went wrong',
    };

    return ApiException(
      message: message,
      statusCode: statusCode,
      code: 'HTTP_$statusCode',
    );
  }
}