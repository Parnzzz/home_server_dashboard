import 'package:get/get.dart';

import '../../../core/storage/token_storage.dart';
import '../model/auth_response.dart';
import '../model/auth_user.dart';
import '../service/auth_service.dart';

class AuthController extends GetxController {
  final AuthService _authService;
  final TokenStorage _tokenStorage;

  AuthController({
    required this._authService,
    required this._tokenStorage,
  });

  final RxBool isLoading = false.obs;
  final RxBool isAuthenticated = false.obs;
  final RxnString errorMessage = RxnString();
  final Rxn<UserProfile> user = Rxn<UserProfile>();

  Future<bool> register({
    required String email,
    required String displayName,
    required String password,
  }) async {
    if (isLoading.value) return false;

    isLoading.value = true;
    errorMessage.value = null;

    UserProfile? createdUser;
    try {
      createdUser = await _authService.register(
        email: email.trim(),
        displayName: displayName.trim(),
        password: password,
      );

      final AuthResponse result = await _authService.login(
        email: email.trim(),
        password: password,
      );

      await _tokenStorage.saveTokens(
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
      );

      user.value = result.user;
      isAuthenticated.value = true;
      return true;
    } catch (e) {
      errorMessage.value = createdUser == null
          ? e.toString().replaceFirst('Exception: ', '')
          : 'Account created, but automatic sign-in failed. Please sign in with your new account.';
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  Future<bool> login({
    required String email,
    required String password,
  }) async {
    if (isLoading.value) return false;

    isLoading.value = true;
    errorMessage.value = null;

    try {
      final AuthResponse result = await _authService.login(
        email: email.trim(),
        password: password,
      );

      await _tokenStorage.saveTokens(
        accessToken: result.accessToken,
        refreshToken: result.refreshToken,
      );

      user.value = result.user;
      isAuthenticated.value = true;

      return true;
    } catch (e) {
      errorMessage.value = e.toString().replaceFirst('Exception: ', '');
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  Future<bool> requestPasswordReset({required String email}) async {
    if (isLoading.value) return false;

    isLoading.value = true;
    errorMessage.value = null;

    try {
      await _authService.requestPasswordReset(email: email.trim());
      return true;
    } catch (e) {
      errorMessage.value = e.toString().replaceFirst('Exception: ', '');
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  Future<bool> resetPassword({
    required String email,
    required String code,
    required String newPassword,
  }) async {
    if (isLoading.value) return false;

    isLoading.value = true;
    errorMessage.value = null;

    try {
      await _authService.resetPassword(
        email: email.trim(),
        code: code.trim(),
        newPassword: newPassword,
      );
      try {
        await _tokenStorage.clearTokens();
      } catch (_) {
        // The password has already been changed on the server.
      }
      user.value = null;
      isAuthenticated.value = false;
      return true;
    } catch (e) {
      errorMessage.value = e.toString().replaceFirst('Exception: ', '');
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  Future<bool> restoreSession() async {
    try {
      final accessToken = await _tokenStorage.getAccessToken();
      final refreshToken = await _tokenStorage.getRefreshToken();

      final hasAccessToken = accessToken != null && accessToken.isNotEmpty;
      final hasRefreshToken = refreshToken != null && refreshToken.isNotEmpty;
      if (!hasAccessToken && !hasRefreshToken) {
        isAuthenticated.value = false;
        return false;
      }

      try {
        // A 401 from /me is refreshed by AuthInterceptor before this returns.
        user.value = await _authService.getCurrentUser(accessToken ?? '');
        isAuthenticated.value = true;
        return true;
      } catch (_) {
        // Keep the session on transient network/server errors. If refresh was
        // rejected, AuthInterceptor clears the stored tokens and this returns false.
        final storedAccessToken = await _tokenStorage.getAccessToken();
        final storedRefreshToken = await _tokenStorage.getRefreshToken();
        final hasStoredSession =
            (storedAccessToken != null && storedAccessToken.isNotEmpty) ||
            (storedRefreshToken != null && storedRefreshToken.isNotEmpty);

        if (!hasStoredSession) {
          user.value = null;
        }
        isAuthenticated.value = hasStoredSession;
        return hasStoredSession;
      }
    } catch (_) {
      isAuthenticated.value = false;
      return false;
    }
  }

  Future<void> logout() async {
    try {
      final refreshToken = await _tokenStorage.getRefreshToken();
      if (refreshToken != null && refreshToken.isNotEmpty) {
        await _authService.logout(refreshToken: refreshToken);
      }
    } catch (_) {
      // Always clear the local session even when the server cannot be reached.
    } finally {
      try {
        await _tokenStorage.clearTokens();
      } catch (_) {
        // Keep the UI signed out if secure storage reports an error.
      }
    }

    user.value = null;
    isAuthenticated.value = false;
    errorMessage.value = null;
  }

  void clearError() {
    errorMessage.value = null;
  }
}
