import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controller/dashboard_controller.dart';
import '../model/dashboard_summary.dart';
import 'add_server_screen.dart';

class ServerManagementScreen extends GetView<DashboardController> {
  const ServerManagementScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Servers'),
        actions: [
          IconButton(
            onPressed: _addServer,
            tooltip: 'Add server',
            icon: const Icon(Icons.add),
          ),
        ],
      ),
      body: Obx(() {
        if (controller.servers.isEmpty) {
          return Center(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.dns_outlined,
                    size: 56,
                    color: Theme.of(context).colorScheme.primary,
                  ),
                  const SizedBox(height: 12),
                  Text(
                    'No servers yet',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  const SizedBox(height: 8),
                  const Text('Add a server to start monitoring it.'),
                  const SizedBox(height: 20),
                  FilledButton.icon(
                    onPressed: _addServer,
                    icon: const Icon(Icons.add),
                    label: const Text('Add server'),
                  ),
                ],
              ),
            ),
          );
        }

        return ListView.separated(
          padding: const EdgeInsets.all(16),
          itemCount: controller.servers.length,
          separatorBuilder: (_, __) => const SizedBox(height: 8),
          itemBuilder: (context, index) {
            final server = controller.servers[index];
            final isSelected =
                controller.selectedServerId.value == server.id;
            final details = [
              if (server.hostname.isNotEmpty) server.hostname,
              server.status.toUpperCase(),
              if (isSelected) 'Selected',
            ].join(' • ');

            return Card(
              clipBehavior: Clip.antiAlias,
              child: ListTile(
                selected: isSelected,
                leading: Icon(
                  isSelected ? Icons.dns : Icons.dns_outlined,
                  color: isSelected
                      ? Theme.of(context).colorScheme.primary
                      : null,
                ),
                title: Text(server.name),
                subtitle: Text(details),
                onTap: () => _selectServer(server),
                trailing: IconButton(
                  onPressed: controller.isDeletingServer.value
                      ? null
                      : () => _confirmDeleteServer(context, server),
                  tooltip: 'Delete ${server.name}',
                  icon: Icon(
                    Icons.delete_outline,
                    color: Theme.of(context).colorScheme.error,
                  ),
                ),
              ),
            );
          },
        );
      }),
    );
  }

  void _addServer() {
    Get.to(
      () => const AddServerScreen(closeToDashboardOnSuccess: true),
    );
  }

  Future<void> _selectServer(ServerInfo server) async {
    if (controller.selectedServerId.value != server.id) {
      await controller.selectServer(server.id);
    }

    if (controller.summary.value != null) {
      Get.back();
      return;
    }

    final error = controller.errorMessage.value;
    if (error != null) {
      Get.snackbar(
        'Unable to load server',
        error,
        snackPosition: SnackPosition.BOTTOM,
      );
    }
  }

  Future<void> _confirmDeleteServer(
    BuildContext context,
    ServerInfo server,
  ) async {
    final confirmed = await Get.dialog<bool>(
      AlertDialog(
        title: const Text('Delete server?'),
        content: Text(
          'Delete ${server.name}? Its metrics and agent token will also be removed.',
        ),
        actions: [
          TextButton(
            onPressed: () => Get.back(result: false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Get.back(result: true),
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
              foregroundColor: Theme.of(context).colorScheme.onError,
            ),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    final deleted = await controller.deleteServer(server.id);
    if (deleted) {
      Get.snackbar(
        'Server deleted',
        '${server.name} was removed.',
        snackPosition: SnackPosition.BOTTOM,
      );
      return;
    }

    Get.snackbar(
      'Unable to delete server',
      controller.serverManagementError.value ?? 'Please try again.',
      snackPosition: SnackPosition.BOTTOM,
    );
  }
}
