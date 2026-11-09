import 'package:flutter/material.dart';
import 'package:get/get.dart';

import 'core/binding/app_binding.dart';
import 'core/theme/app_theme.dart';
import 'core/theme/theme_service.dart';
import 'features/auth/binding/auth_binding.dart';
import 'features/auth/screens/login_screen.dart';
import 'features/auth/screens/register_screen.dart';
import 'features/auth/screens/forgot_password_screen.dart';
import 'features/auth/screens/splash_screen.dart';
import 'features/dashbord/binding/dashboard_binding.dart';
import 'features/dashbord/screens/dashboard_screen.dart';
import 'features/dashbord/screens/server_management_screen.dart';
import 'features/profile/binding/profile_binding.dart';
import 'features/profile/screen/profile_screen.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final themeService = await ThemeService.load();
  runApp(HomeDashboardApp(themeService: themeService));
}

class HomeDashboardApp extends StatelessWidget {
  const HomeDashboardApp({super.key, required this.themeService});

  final ThemeService themeService;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: themeService,
      builder: (context, child) {
        return GetMaterialApp(
          title: 'Home Server Dashboard',
          initialBinding: AppBindings(),
          initialRoute: '/splash',
          getPages: [
            GetPage(
              name: '/splash',
              page: () => const SplashScreen(),
              binding: AuthBinding(),
            ),
            GetPage(
              name: '/login',
              page: () => const LoginScreen(),
              binding: AuthBinding(),
            ),
            GetPage(
              name: '/register',
              page: () => const RegisterScreen(),
              binding: AuthBinding(),
            ),
            GetPage(
              name: '/forgot-password',
              page: () => const ForgotPasswordScreen(),
              binding: AuthBinding(),
            ),
            GetPage(
              name: '/dashboard',
              page: () => DashboardScreen(themeService: themeService),
              binding: DashboardBinding(),
            ),
            GetPage(
              name: '/servers',
              page: () => const ServerManagementScreen(),
            ),
            GetPage(
              name: '/profile',
              page: () => const ProfileScreen(),
              binding: ProfileBinding(),
            ),
          ],
          debugShowCheckedModeBanner: false,
          theme: AppTheme.lightTheme,
          darkTheme: AppTheme.darkTheme,
          themeMode: themeService.themeMode,
        );
      },
    );
  }
}
