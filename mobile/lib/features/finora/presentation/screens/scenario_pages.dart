part of '../finora_pages.dart';

/// Lightweight resource-backed screens for forecasts, automation and AI.
///
/// This stays in the same Dart library as the shared resource page while the
/// presentation folder is progressively split into screen-focused files.
class ForecastPage extends StatelessWidget {
  const ForecastPage({super.key, required this.api});

  final ApiClient api;

  @override
  Widget build(BuildContext context) => ScenarioPage(
    api: api,
    title: 'Dự báo',
    path: '/forecast-scenarios',
    fields: const [
      FieldSpec('name', 'Tên kịch bản'),
      FieldSpec('assumptions', 'Giả định (JSON)'),
    ],
  );
}

class AutomationPage extends StatelessWidget {
  const AutomationPage({super.key, required this.api});

  final ApiClient api;

  @override
  Widget build(BuildContext context) => ScenarioPage(
    api: api,
    title: 'Quy tắc tự động',
    path: '/bank-automation-rules',
    fields: const [
      FieldSpec('name', 'Tên quy tắc'),
      FieldSpec('condition', 'Điều kiện'),
      FieldSpec('action', 'Hành động'),
    ],
  );
}

class AssistantPage extends StatelessWidget {
  const AssistantPage({super.key, required this.api});

  final ApiClient api;

  @override
  Widget build(BuildContext context) => ScenarioPage(
    api: api,
    title: 'Trợ lý AI',
    path: '/assistant/commands',
    fields: const [
      FieldSpec('command', 'Yêu cầu'),
      FieldSpec('plan', 'Kế hoạch (tuỳ chọn)'),
    ],
  );
}

class ScenarioPage extends StatelessWidget {
  const ScenarioPage({
    super.key,
    required this.api,
    required this.title,
    required this.path,
    required this.fields,
  });

  final ApiClient api;
  final String title;
  final String path;
  final List<FieldSpec> fields;

  @override
  Widget build(BuildContext context) =>
      ResourcePage(api: api, title: title, path: path, fields: fields);
}
