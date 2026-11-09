import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../../auth/controller/auth_controller.dart';
import '../../../core/theme/theme_service.dart';
import '../../profile/controller/profile_controller.dart';
import '../controller/dashboard_controller.dart';
import '../model/dashboard_summary.dart';
import 'add_server_screen.dart';
import 'connect_agent_screen.dart';
import '../widget/common/time_range_selector.dart';
import '../widget/landscape/dashboard_landscape.dart';
import '../widget/portrait/dashboard_portrait.dart';

class DashboardScreen extends GetView<DashboardController> {
  final ThemeService themeService;

  const DashboardScreen({
    super.key,
    required this.themeService,
  });


  @override
  Widget build(BuildContext context) {

    return Obx(() {
      final summary = controller.summary.value;

      if (controller.needsAgentSetup.value) {
        return const ConnectAgentScreen();
      }

      if (controller.needsServerSetup.value) {
        return const AddServerScreen();
      }

      if (controller.isLoading.value && summary == null) {
        return const Scaffold(
          body: Center(
            child: CircularProgressIndicator(),
          ),
        );
      }

      if (controller.errorMessage.value != null &&
          summary == null) {
        return Scaffold(
          body: _buildErrorState(context),
        );
      }

      if (summary == null) {
        return const Scaffold(
          body: Center(
            child: Text('No dashboard data'),
          ),
        );
      }

      return Scaffold(
        drawer: _buildDrawer(context, summary),
        appBar: AppBar(
          title: const Text('Home Dashboard'),
          actions: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    controller.isBackendConnected.value
                        ? Icons.cloud_done
                        : Icons.cloud_off,
                    size: 18,
                    color: !controller.hasCheckedConnection.value
                        ? Colors.grey
                        : controller.isBackendConnected.value
                        ? Colors.green
                        : Colors.red,
                  ),
                  const SizedBox(width: 6),
                  Text(
                    !controller.hasCheckedConnection.value
                        ? 'Checking'
                        : controller.isBackendConnected.value
                        ? 'Connected'
                        : 'Disconnected',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
            ),
          ],
        ),
        body: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: TimeRangeSelector(
                selectedRange: controller.selectedRange.value,
                onChanged: controller.changeTimeRange,
              ),
            ),
            Expanded(
              child: LayoutBuilder(
                builder: (context, constraints) {
                  if (constraints.maxWidth > constraints.maxHeight) {
                    return DashboardLandscape(
                      summary: summary,
                      metrics: controller.metrics,
                    );
                  }

                  return DashboardPortrait(
                    summary: summary,
                    metrics: controller.metrics,
                  );
                },
              ),
            ),
          ],
        ),
      );
    });
  }

  Widget _buildDrawer(BuildContext context, DashboardSummary summary) {
    final isDarkMode = Theme.of(context).brightness == Brightness.dark;
    final user = Get.find<AuthController>().user.value;
    final avatarImage = _drawerAvatarImage(user?.avatarUrl);

    return Drawer(
      child: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            DrawerHeader(

              child: Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  CircleAvatar(
                    radius: 32,
                    backgroundColor:
                        Theme.of(context).colorScheme.surfaceContainerHighest,
                    backgroundImage: avatarImage,
                    child: avatarImage == null
                        ? Icon(
                            Icons.person,
                            size: 34,
                            color:
                                Theme.of(context).colorScheme.onSurfaceVariant,
                          )
                        : null,
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          user?.displayName ?? 'User',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 4),
                        Text(
                          user?.email ?? summary.server.name,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                padding: EdgeInsets.zero,
                children: [
                  ListTile(
                    leading: const Icon(Icons.dns_outlined),
                    title: const Text('Servers'),
                    onTap: () {
                      Navigator.pop(context);
                      Get.toNamed('/servers');
                    },
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.link_rounded),
                    title: const Text('Connect to computer'),
                    onTap: () {
                      Navigator.pop(context);
                      controller.openAgentSetup();
                    },
                  ),
                  const Divider(),
                  ListTile(
                    leading: const Icon(Icons.account_circle_outlined),
                    title: const Text('Profile'),
                    onTap: () {
                      Navigator.pop(context);
                      Get.toNamed('/profile');
                    },
                  ),
                  ListTile(
                    leading: Icon(
                      isDarkMode ? Icons.light_mode : Icons.dark_mode,
                    ),
                    title: Text(isDarkMode ? 'Light theme' : 'Dark theme'),
                    onTap: () {
                      Navigator.pop(context);
                      themeService.toggleTheme();
                    },
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            ListTile(
              leading: Icon(
                Icons.logout,
                color: Theme.of(context).colorScheme.error,
              ),
              title: Text(
                'Log out',
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
              onTap: _confirmLogout,
            ),
          ],
        ),
      ),
    );
  }

  ImageProvider? _drawerAvatarImage(String? avatarUrl) {
    if (avatarUrl == null || avatarUrl.isEmpty) return null;

    if (avatarUrl.startsWith('data:image/')) {
      final separator = avatarUrl.indexOf(',');
      if (separator < 0) return null;
      try {
        return MemoryImage(base64Decode(avatarUrl.substring(separator + 1)));
      } on FormatException {
        return null;
      }
    }

    return NetworkImage(avatarUrl);
  }

  Future<void> _confirmLogout() async {
    final confirmed = await Get.dialog<bool>(
      AlertDialog(
        title: const Text('Log out?'),
        content: const Text('You will need to sign in again to continue.'),
        actions: [
          TextButton(
            onPressed: () => Get.back(result: false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Get.back(result: true),
            child: const Text('Log out'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    await Get.find<AuthController>().logout();
    Get.offAllNamed('/login');
  }

  Widget _buildErrorState(BuildContext context) {
    final theme = Theme.of(context);

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.cloud_off_outlined,
              size: 56,
              color: theme.colorScheme.error,
            ),
            const SizedBox(height: 16),
            Text(
              'Connection Failed',
              style: theme.textTheme.titleLarge,
            ),
            const SizedBox(height: 8),
            Text(
              controller.errorMessage.value ??
                  'Unable to connect to Backend',
              textAlign: TextAlign.center,
              style: theme.textTheme.bodyMedium,
            ),
            const SizedBox(height: 24),
            FilledButton.icon(
              onPressed: controller.retry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}
