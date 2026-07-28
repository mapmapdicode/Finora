part of '../finora_pages.dart';

class ValuedResourcePage extends StatelessWidget {
  const ValuedResourcePage({
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

class ReadonlyPage extends StatefulWidget {
  const ReadonlyPage({
    super.key,
    required this.api,
    required this.title,
    required this.path,
  });

  final ApiClient api;
  final String title;
  final String path;

  @override
  State<ReadonlyPage> createState() => _ReadonlyPageState();
}

class _ReadonlyPageState extends State<ReadonlyPage> {
  List data = [];
  String? err;
  bool loading = true;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      final response = await widget.api.request('GET', widget.path);
      data = response is List
          ? response
          : ((response as Map)['items'] as List? ?? []);
    } catch (error) {
      err = error.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  @override
  Widget build(BuildContext context) => PageFrame(
    title: widget.title,
    action: IconButton(
      onPressed: load,
      icon: const Icon(Icons.refresh_rounded, color: Color(0xfffbbf24)),
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfffbbf24)),
          )
        : ListView(
            children: [
              const _ScreenIntro(
                'Dấu vết hoạt động được lưu để đảm bảo minh bạch.',
              ),
              if (err != null) ErrorBox(err!),
              ...data.map((item) {
                if (item is Map) {
                  final action =
                      item['action']?.toString() ??
                      item['entityType']?.toString() ??
                      'Hoạt động hệ thống';
                  final date = _formatDate(item['createdAt']?.toString());
                  final status = item['status']?.toString() ?? 'Thành công';
                  return FinoraListTile(
                    icon: Icons.history_toggle_off_rounded,
                    title: action,
                    subtitle: date.isNotEmpty
                        ? 'Thời gian: $date'
                        : 'Đã ghi nhận nhật ký',
                    badge: status,
                  );
                }
                return FinoraSurface(
                  child: Text(
                    item.toString(),
                    style: const TextStyle(color: Colors.white, fontSize: 13),
                  ),
                );
              }),
              if (data.isEmpty) const EmptyState('Chưa có dữ liệu'),
            ],
          ),
  );
}
