import 'package:get/get.dart';

import '../../auth/controller/auth_controller.dart';
import '../../auth/model/auth_user.dart';
import '../service/profile_service.dart';

class ProfileController extends GetxController {
  final ProfileService _profileService;
  final AuthController _authController;

  ProfileController({
    required this._profileService,
    required this._authController,
  });

  final RxBool isLoading = false.obs;
  final RxnString errorMessage = RxnString();

  Future<bool> updateProfile({
    required String displayName,
    String? avatarUrl,
  }) async {
    if (isLoading.value) return false;

    isLoading.value = true;
    errorMessage.value = null;

    try {
      final UserProfile updatedUser =
      await _profileService.updateProfile(
        displayName: displayName,
        avatarUrl: avatarUrl,
      );

      // Update Profile กลางของระบบ
      _authController.user.value = updatedUser;

      return true;
    } catch (e) {
      errorMessage.value =
          e.toString().replaceFirst('Exception: ', '');

      return false;
    } finally {
      isLoading.value = false;
    }
  }
}