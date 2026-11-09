enum AppEnvironment {
  development,
  testing,
  production,
}

class AppConfig {
  static const String _environment = String.fromEnvironment(
    'APP_ENV',
    defaultValue: 'development',
  );

  static AppEnvironment get environment {
    return AppEnvironment.values.firstWhere(
          (e) => e.name == _environment,
      orElse: () => AppEnvironment.development,
    );
  }

  static String get apiBaseUrl {
    switch (environment) {
      case AppEnvironment.development:
        return 'http://10.0.2.2:8080';

      case AppEnvironment.testing:
        return 'http://192.168.1.33:8080';

      case AppEnvironment.production:
        return 'https://api.yourdomain.com';
    }
  }

  static bool get isDevelopment =>
      environment == AppEnvironment.development;

  static bool get isTesting =>
      environment == AppEnvironment.testing;

  static bool get isProduction =>
      environment == AppEnvironment.production;
}