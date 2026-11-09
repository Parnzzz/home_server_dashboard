import 'package:get/get.dart';
import '../../../core/network/api_client.dart';
import '../../auth/controller/auth_controller.dart';
import '../controller/profile_controller.dart';
import '../service/profile_service.dart';

class ProfileBinding extends Bindings {
  @override
  void dependencies() {
    Get.lazyPut<ProfileService>(
          () => ProfileService(Get.find<ApiClient>()),
    );

    Get.lazyPut<ProfileController>(
          () => ProfileController(
        profileService: Get.find<ProfileService>(),
        authController: Get.find<AuthController>(),
      ),
    );
  }
}