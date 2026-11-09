import 'package:get/get.dart';

import '../config/app_config.dart';
import '../network/api_client.dart';
import '../storage/token_storage.dart';

class AppBindings extends Bindings {
  @override
  void dependencies() {

    Get.put<TokenStorage>(
      TokenStorage(),
      permanent: true,
    );

    Get.put<ApiClient>(
          ApiClient(
            baseUrl: AppConfig.apiBaseUrl,
            tokenStorage: Get.find<TokenStorage>(),
          ),
      permanent: true,
    );
  }
}