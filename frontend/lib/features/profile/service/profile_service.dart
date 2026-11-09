
import '../../../core/network/api_client.dart';
import '../../auth/model/auth_user.dart';

class ProfileService {
  final ApiClient _apiClient;

  ProfileService(this._apiClient);

  Future<UserProfile> updateProfile({
    required String displayName,
    String? avatarUrl,
  }) async {
    final response = await _apiClient.request<Map<String, dynamic>>(
      '/api/v1/users/me',
      method: 'PATCH',
      data: {
        'display_name': displayName,
        if (avatarUrl != null) 'avatar_url': avatarUrl,
      },
    );
    return UserProfile.fromJson(response.data!);
  }
}