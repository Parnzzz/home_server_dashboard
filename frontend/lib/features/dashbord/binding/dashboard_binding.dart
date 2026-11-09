import 'package:get/get.dart';

import '../../../core/network/api_client.dart';
import '../controller/dashboard_controller.dart';
import '../service/dashboard_service.dart';
import '../service/metric_service.dart';

class DashboardBinding extends Bindings {
  @override
  void dependencies() {
    Get.lazyPut<DashboardService>(
          () => DashboardService(Get.find<ApiClient>()),
    );

    Get.lazyPut<MetricsService>(
          () => MetricsService(Get.find<ApiClient>()),
    );

    Get.lazyPut<DashboardController>(
          () => DashboardController(
        dashboardService: Get.find<DashboardService>(),
        metricsService: Get.find<MetricsService>(),
      ),
    );
  }
}