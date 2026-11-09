import 'package:get/get.dart';

import '../../../core/network/api_client.dart';
import '../../../core/storage/token_storage.dart';
import '../controller/auth_controller.dart';
import '../service/auth_service.dart';

class AuthBinding extends Bindings {
  @override
  void dependencies() {
    Get.lazyPut<AuthService>(
      () => AuthService(Get.find<ApiClient>().dio),
    );

    Get.put<AuthController>(
      AuthController(
        authService: Get.find<AuthService>(),
        tokenStorage: Get.find<TokenStorage>(),
      ),
      permanent: true,
    );
  }
}
