import 'dart:async';

import 'package:flutter/material.dart';
import 'package:get/get.dart';

import '../../../core/network/api_exception.dart';
import '../model/dashboard_summary.dart';
import '../model/metric.dart';
import '../service/dashboard_service.dart';
import '../service/metric_service.dart';

class DashboardController extends GetxController
    with WidgetsBindingObserver {
  final DashboardService _dashboardService;
  final MetricsService _metricsService;

  DashboardController({
    required this._dashboardService,
    required this._metricsService,
  });

  final servers = <ServerInfo>[].obs;
  final selectedServerId = RxnString();

  final summary = Rxn<DashboardSummary>();
  final metrics = <Metric>[].obs;

  final isLoading = true.obs;
  final isCreatingServer = false.obs;
  final isDeletingServer = false.obs;
  final isGeneratingAgentToken = false.obs;
  final needsServerSetup = false.obs;
  final needsAgentSetup = false.obs;
  final errorMessage = RxnString();
  final serverSetupError = RxnString();
  final serverManagementError = RxnString();
  final agentSetupError = RxnString();
  final agentToken = RxnString();
  final agentServerName = ''.obs;

  final isBackendConnected = false.obs;
  final hasCheckedConnection = false.obs;

  final selectedRange = '5m'.obs;

  Timer? _refreshTimer;
  bool _isRefreshing = false;
  bool _refreshRequested = false;
  bool _disposed = false;

  @override
  void onInit() {
    super.onInit();

    WidgetsBinding.instance.addObserver(this);

    loadDashboard();
    startRefreshTimer();
  }

  Future<void> loadDashboard() async {
    if (_disposed || needsServerSetup.value || needsAgentSetup.value) return;

    if (_isRefreshing) {
      _refreshRequested = true;
      return;
    }

    _isRefreshing = true;

    try {
      do {
        _refreshRequested = false;

        try {
          await _loadOnce();
        } on ApiException catch (error) {
          if (error.statusCode != 404) rethrow;

          // The cached server may have been removed. Resolve an owned server
          // again, or show setup if this account has none.
          selectedServerId.value = null;
          await _loadOnce();
        }

        if (_disposed || needsServerSetup.value) return;
      } while (_refreshRequested && !_disposed);
    } catch (error) {
      if (_disposed) return;

      isBackendConnected.value = false;
      hasCheckedConnection.value = true;

      if (summary.value == null) {
        errorMessage.value = _getErrorMessage(error);
      }

      isLoading.value = false;
    } finally {
      _isRefreshing = false;
    }
  }

  Future<String?> _resolveServerId() async {
    if (selectedServerId.value != null) return selectedServerId.value;

    final ownedServers = await _dashboardService.getServers();
    servers.assignAll(ownedServers);
    if (ownedServers.isEmpty) return null;

    selectedServerId.value = ownedServers.first.id;
    return selectedServerId.value;
  }

  Future<void> _loadOnce() async {
    final serverId = await _resolveServerId();
    if (serverId == null) {
      _showAddServer();
      return;
    }

    final results = await Future.wait([
      _dashboardService.getSummary(serverId),
      _metricsService.getMetrics(
        serverId,
        from: selectedRange.value,
        limit: 1000,
      ),
    ]);

    if (_disposed) return;

    summary.value = results[0] as DashboardSummary;
    metrics.assignAll(results[1] as List<Metric>);
    needsServerSetup.value = false;
    serverSetupError.value = null;
    errorMessage.value = null;
    isLoading.value = false;
    isBackendConnected.value = true;
    hasCheckedConnection.value = true;
  }

  void _showAddServer() {
    selectedServerId.value = null;
    summary.value = null;
    metrics.clear();
    needsServerSetup.value = true;
    needsAgentSetup.value = false;
    agentToken.value = null;
    errorMessage.value = null;
    serverSetupError.value = null;
    isLoading.value = false;
    isBackendConnected.value = true;
    hasCheckedConnection.value = true;
    stopRefreshTimer();
  }

  String _getErrorMessage(Object error) {
    if (error is ApiException) return error.message;
    return error
        .toString()
        .replaceFirst('Exception: ', '');
  }

  Future<bool> createServer({
    required String name,
    String hostname = '',
    String os = '',
  }) async {
    if (isCreatingServer.value) return false;

    isCreatingServer.value = true;
    serverSetupError.value = null;

    try {
      final server = await _dashboardService.createServer(
        name: name,
        hostname: hostname,
        os: os,
      );

      servers.removeWhere((item) => item.id == server.id);
      servers.insert(0, server);
      selectedServerId.value = server.id;
      summary.value = null;
      metrics.clear();
      needsServerSetup.value = false;
      agentServerName.value = server.name;
      agentSetupError.value = null;
      needsAgentSetup.value = true;
      isLoading.value = false;
      await generateAgentToken();
      startRefreshTimer();
      return true;
    } catch (error) {
      serverSetupError.value = _getErrorMessage(error);
      return false;
    } finally {
      isCreatingServer.value = false;
    }
  }

  Future<bool> generateAgentToken() async {
    final serverId = selectedServerId.value;
    if (serverId == null || isGeneratingAgentToken.value) return false;

    isGeneratingAgentToken.value = true;
    agentSetupError.value = null;

    try {
      agentToken.value = await _dashboardService.generateAgentToken(serverId);
      return true;
    } catch (error) {
      agentSetupError.value = _getErrorMessage(error);
      return false;
    } finally {
      isGeneratingAgentToken.value = false;
    }
  }

  void openAgentSetup() {
    final server = summary.value?.server;
    if (selectedServerId.value == null || server == null) return;

    agentServerName.value = server.name;
    agentSetupError.value = null;
    agentToken.value = null;
    needsAgentSetup.value = true;
    stopRefreshTimer();
  }

  Future<void> finishAgentSetup() async {
    needsAgentSetup.value = false;
    agentToken.value = null;
    errorMessage.value = null;
    isLoading.value = true;
    await loadDashboard();
    startRefreshTimer();
  }

  Future<void> selectServer(String serverId) async {
    if (selectedServerId.value == serverId && summary.value != null) return;

    selectedServerId.value = serverId;
    needsServerSetup.value = false;
    needsAgentSetup.value = false;
    agentToken.value = null;
    summary.value = null;
    metrics.clear();
    errorMessage.value = null;
    isLoading.value = true;

    await loadDashboard();
    startRefreshTimer();
  }

  Future<bool> deleteServer(String serverId) async {
    if (isDeletingServer.value) return false;

    isDeletingServer.value = true;
    serverManagementError.value = null;
    final wasSelected = selectedServerId.value == serverId;

    try {
      await _dashboardService.deleteServer(serverId);
      servers.removeWhere((server) => server.id == serverId);

      if (!wasSelected) return true;

      summary.value = null;
      metrics.clear();
      errorMessage.value = null;
      needsAgentSetup.value = false;
      agentToken.value = null;

      if (servers.isEmpty) {
        _showAddServer();
        return true;
      }

      selectedServerId.value = servers.first.id;
      isLoading.value = true;
      await loadDashboard();
      startRefreshTimer();
      return true;
    } catch (error) {
      serverManagementError.value = _getErrorMessage(error);
      return false;
    } finally {
      isDeletingServer.value = false;
    }
  }

  void changeTimeRange(String range) {
    if (selectedRange.value == range) return;

    selectedRange.value = range;
    loadDashboard();
  }

  void retry() {
    selectedServerId.value = null;
    isLoading.value = true;
    errorMessage.value = null;
    needsServerSetup.value = false;
    loadDashboard();
  }

  void startRefreshTimer() {
    stopRefreshTimer();
    if (needsServerSetup.value || needsAgentSetup.value) return;

    _refreshTimer = Timer.periodic(
      const Duration(seconds: 5),
      (_) => loadDashboard(),
    );
  }

  void stopRefreshTimer() {
    _refreshTimer?.cancel();
    _refreshTimer = null;
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    switch (state) {
      case AppLifecycleState.resumed:
        loadDashboard();
        startRefreshTimer();
        break;

      case AppLifecycleState.inactive:
      case AppLifecycleState.paused:
      case AppLifecycleState.detached:
      case AppLifecycleState.hidden:
        stopRefreshTimer();
        break;
    }
  }

  @override
  void onClose() {
    _disposed = true;

    stopRefreshTimer();
    WidgetsBinding.instance.removeObserver(this);

    super.onClose();
  }
}
