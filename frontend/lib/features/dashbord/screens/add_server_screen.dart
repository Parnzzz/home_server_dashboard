import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../controller/dashboard_controller.dart';

class AddServerScreen extends StatefulWidget {
  const AddServerScreen({
    super.key,
    this.closeToDashboardOnSuccess = false,
  });

  final bool closeToDashboardOnSuccess;

  @override
  State<AddServerScreen> createState() => _AddServerScreenState();
}

class _AddServerScreenState extends State<AddServerScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _hostnameController = TextEditingController();
  final _osController = TextEditingController();
  final _dashboardController = Get.find<DashboardController>();

  @override
  void dispose() {
    _nameController.dispose();
    _hostnameController.dispose();
    _osController.dispose();
    super.dispose();
  }

  Future<void> _createServer() async {
    if (!_formKey.currentState!.validate()) return;

    final created = await _dashboardController.createServer(
      name: _nameController.text.trim(),
      hostname: _hostnameController.text.trim(),
      os: _osController.text.trim(),
    );

    if (created && widget.closeToDashboardOnSuccess && mounted) {
      Get.until((route) => route.settings.name == '/dashboard');
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('Add Server')),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 460),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Icon(
                      Icons.dns_rounded,
                      size: 64,
                      color: colorScheme.primary,
                    ),
                    const SizedBox(height: 20),
                    Text(
                      widget.closeToDashboardOnSuccess
                          ? 'Add another server'
                          : 'Add your first server',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Register a server to start monitoring it. You can connect its agent afterward.',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 32),
                    TextFormField(
                      controller: _nameController,
                      textCapitalization: TextCapitalization.words,
                      textInputAction: TextInputAction.next,
                      maxLength: 100,
                      decoration: const InputDecoration(
                        labelText: 'Server name',
                        hintText: 'e.g. Home PC',
                        prefixIcon: Icon(Icons.computer_outlined),
                        border: OutlineInputBorder(),
                        counterText: '',
                      ),
                      validator: (value) {
                        if (value == null || value.trim().isEmpty) {
                          return 'Please enter a server name';
                        }
                        if (value.trim().length > 100) {
                          return 'Name must be 100 characters or fewer';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    TextFormField(
                      controller: _hostnameController,
                      textInputAction: TextInputAction.next,
                      decoration: const InputDecoration(
                        labelText: 'Hostname (optional)',
                        hintText: 'e.g. home-pc',
                        prefixIcon: Icon(Icons.language_outlined),
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 16),
                    TextFormField(
                      controller: _osController,
                      textInputAction: TextInputAction.done,
                      onFieldSubmitted: (_) => _createServer(),
                      decoration: const InputDecoration(
                        labelText: 'Operating system (optional)',
                        hintText: 'e.g. Windows 11',
                        prefixIcon: Icon(Icons.devices_outlined),
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 16),
                    Obx(() {
                      final error =
                          _dashboardController.serverSetupError.value;
                      if (error == null) return const SizedBox.shrink();

                      return Padding(
                        padding: const EdgeInsets.only(bottom: 16),
                        child: Text(
                          error,
                          textAlign: TextAlign.center,
                          style: TextStyle(color: colorScheme.error),
                        ),
                      );
                    }),
                    Obx(() {
                      final isCreating =
                          _dashboardController.isCreatingServer.value;

                      return FilledButton.icon(
                        onPressed: isCreating ? null : _createServer,
                        icon: isCreating
                            ? const SizedBox(
                                width: 18,
                                height: 18,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.add),
                        label: Text(
                          isCreating ? 'Adding server...' : 'Add server',
                        ),
                        style: FilledButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 16),
                        ),
                      );
                    }),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
