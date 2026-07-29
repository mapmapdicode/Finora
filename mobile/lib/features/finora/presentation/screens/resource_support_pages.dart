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
    setState(() {
      loading = true;
      err = null;
    });
    try {
      final response = await widget.api.request('GET', widget.path);
      data = response is List
          ? response
          : ((response as Map)['items'] as List? ?? []);
    } catch (error) {
      err = presentableError(error);
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  @override
  Widget build(BuildContext context) => PageFrame(
    title: widget.title,
    action: IconButton(
      tooltip: 'Tải lại ${widget.title.toLowerCase()}',
      onPressed: load,
      icon: const Icon(Icons.refresh_rounded, color: FinoraColors.primary),
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: FinoraColors.primary),
          )
        : (err != null && data.isEmpty)
        ? FinoraEmptyState(
            title: 'Chưa thể tải ${widget.title.toLowerCase()}',
            message: 'Kiểm tra kết nối rồi thử lại.',
            icon: Icons.cloud_off_rounded,
            action: FilledButton.icon(
              onPressed: load,
              icon: const Icon(Icons.refresh_rounded),
              label: const Text('Thử lại'),
            ),
          )
        : ListView(
            padding: const EdgeInsets.only(bottom: FinoraSpace.xxl),
            children: [
              const _ScreenIntro(
                'Dấu vết hoạt động được lưu để đảm bảo minh bạch.',
              ),
              if (err != null) ErrorBox(err!),
              ...data.map((item) {
                if (item is Map) {
                  final action = _activityActionLabel(
                    item['action']?.toString() ??
                        item['entityType']?.toString(),
                  );
                  final date = _formatDate(item['createdAt']?.toString());
                  final status = _activityStatusLabel(item['status']);
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
                    style: FinoraTypography.bodySmall.copyWith(
                      color: FinoraColors.textPrimary,
                    ),
                  ),
                );
              }),
              if (data.isEmpty)
                FinoraEmptyState(
                  title: 'Chưa có dữ liệu',
                  message: 'Các hoạt động mới sẽ xuất hiện tại đây.',
                  icon: Icons.history_toggle_off_rounded,
                ),
            ],
          ),
  );
}

String _activityStatusLabel(dynamic value) {
  return switch (value?.toString()) {
    'success' || 'completed' || 'done' => 'Hoàn tất',
    'pending' || 'running' => 'Đang xử lý',
    'failed' || 'error' => 'Thất bại',
    null || '' => 'Đã ghi nhận',
    _ => 'Trạng thái: ${value.toString()}',
  };
}

String _activityActionLabel(String? value) {
  const labels = {
    'create': 'Tạo mới',
    'created': 'Tạo mới',
    'update': 'Cập nhật',
    'updated': 'Cập nhật',
    'delete': 'Xoá',
    'deleted': 'Xoá',
    'approve': 'Phê duyệt',
    'approved': 'Phê duyệt',
    'login': 'Đăng nhập',
    'logout': 'Đăng xuất',
    'account': 'Tài khoản',
    'transaction': 'Giao dịch',
    'loan': 'Khoản vay',
    'asset': 'Tài sản',
    'property': 'Bất động sản',
  };
  final raw = value?.trim() ?? '';
  if (raw.isEmpty) return 'Hoạt động hệ thống';
  final normalized = raw.toLowerCase().replaceAll(RegExp(r'[_-]+'), ' ');
  for (final entry in labels.entries) {
    if (normalized == entry.key || normalized.contains(entry.key)) {
      return entry.value;
    }
  }
  return raw;
}
