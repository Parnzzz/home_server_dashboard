import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:get/get.dart';

import '../../../core/config/app_config.dart';
import '../controller/dashboard_controller.dart';

class ConnectAgentScreen extends StatefulWidget {
  const ConnectAgentScreen({super.key});

  @override
  State<ConnectAgentScreen> createState() => _ConnectAgentScreenState();
}

class _ConnectAgentScreenState extends State<ConnectAgentScreen> {
  final _dashboardController = Get.find<DashboardController>();
  late final TextEditingController _backendUrlController;

  @override
  void initState() {
    super.initState();
    _backendUrlController = TextEditingController(
      text: _defaultAgentBackendUrl(),
    );
  }

  String _defaultAgentBackendUrl() {
    final configuredUrl = AppConfig.apiBaseUrl.trim().replaceFirst(
      RegExp(r'/+$'),
      '',
    );

    if (Uri.tryParse(configuredUrl)?.host == '10.0.2.2') {
      return 'http://localhost:8080';
    }

    return configuredUrl;
  }

  String _setupCommands(String token) {
    final backendUrl = _backendUrlController.text.trim();
    return "\$env:API_BASE_URL = '$backendUrl'\n"
        "\$env:AGENT_TOKEN = '$token'\n"
        'go run .';
  }

  Future<void> _copyText(String value, String label) async {
    await Clipboard.setData(ClipboardData(text: value));
    Get.snackbar(
      'Copied',
      '$label copied to clipboard.',
      snackPosition: SnackPosition.BOTTOM,
    );
  }

  @override
  void dispose() {
    _backendUrlController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Connect to Computer'),
        leading: IconButton(
          onPressed: _dashboardController.finishAgentSetup,
          tooltip: 'Back to dashboard',
          icon: const Icon(Icons.arrow_back),
        ),
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 560),
              child: Obx(() {
                final token = _dashboardController.agentToken.value;
                final isGenerating =
                    _dashboardController.isGeneratingAgentToken.value;
                final error = _dashboardController.agentSetupError.value;

                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Icon(
                      Icons.link_rounded,
                      size: 56,
                      color: colorScheme.primary,
                    ),
                    const SizedBox(height: 16),
                    Text(
                      'Connect ${_dashboardController.agentServerName.value}',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Run the Home Server Agent on the computer you want to monitor. It will send metrics every 5 seconds.',
                      textAlign: TextAlign.center,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: colorScheme.onSurfaceVariant,
                      ),
                    ),
                    const SizedBox(height: 28),
                    TextFormField(
                      controller: _backendUrlController,
                      keyboardType: TextInputType.url,
                      decoration: const InputDecoration(
                        labelText: 'Backend URL reachable from the computer',
                        hintText: 'http://192.168.1.33:8080',
                        prefixIcon: Icon(Icons.dns_outlined),
                        border: OutlineInputBorder(),
                      ),
                      onChanged: (_) => setState(() {}),
                    ),
                    const SizedBox(height: 24),
                    Text(
                      'Agent token',
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    const SizedBox(height: 8),
                    if (token != null) ...[
                      Container(
                        padding: const EdgeInsets.all(14),
                        decoration: BoxDecoration(
                          color: colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: SelectableText(
                          token,
                          style: theme.textTheme.bodyMedium?.copyWith(
                            fontFamily: 'monospace',
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      OutlinedButton.icon(
                        onPressed: () => _copyText(token, 'Agent token'),
                        icon: const Icon(Icons.copy),
                        label: const Text('Copy token'),
                      ),
                      const SizedBox(height: 20),
                      Text(
                        'Open PowerShell in the Agent folder on that computer and run:',
                        style: theme.textTheme.bodyMedium,
                      ),
                      const SizedBox(height: 8),
                      Container(
                        padding: const EdgeInsets.all(14),
                        decoration: BoxDecoration(
                          color: colorScheme.surfaceContainerHighest,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: SelectableText(
                          _setupCommands(token),
                          style: theme.textTheme.bodyMedium?.copyWith(
                            fontFamily: 'monospace',
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      FilledButton.tonalIcon(
                        onPressed: () =>
                            _copyText(_setupCommands(token), 'Setup commands'),
                        icon: const Icon(Icons.copy_all),
                        label: const Text('Copy setup commands'),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        'Keep this token private. It is tied to this server; generating a replacement will invalidate the previous token.',
                        style: theme.textTheme.bodySmall?.copyWith(
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ] else ...[
                      if (error != null) ...[
                        Text(
                          error,
                          textAlign: TextAlign.center,
                          style: TextStyle(color: colorScheme.error),
                        ),
                        const SizedBox(height: 12),
                      ],
                      FilledButton.icon(
                        onPressed: isGenerating
                            ? null
                            : _dashboardController.generateAgentToken,
                        icon: isGenerating
                            ? const SizedBox(
                                width: 18,
                                height: 18,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                ),
                              )
                            : const Icon(Icons.key),
                        label: Text(
                          isGenerating
                              ? 'Generating token...'
                              : 'Generate agent token',
                        ),
                      ),
                    ],
                    const SizedBox(height: 24),
                    OutlinedButton(
                      onPressed: token == null || isGenerating
                          ? null
                          : _dashboardController.finishAgentSetup,
                      child: const Text('Continue to dashboard'),
                    ),
                  ],
                );
              }),
            ),
          ),
        ),
      ),
    );
  }
}
