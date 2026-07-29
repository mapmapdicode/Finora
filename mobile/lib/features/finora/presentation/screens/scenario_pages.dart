part of '../finora_pages.dart';

/// Lightweight resource-backed screens for forecasts, automation and AI.
///
/// This stays in the same Dart library as the shared resource page while the
/// presentation folder is progressively split into screen-focused files.
class ForecastPage extends StatefulWidget {
  const ForecastPage({super.key, required this.api});

  final ApiClient api;

  @override
  State<ForecastPage> createState() => _ForecastPageState();
}

class _ForecastPageState extends State<ForecastPage> {
  Map<String, dynamic>? _netWorth;
  List<Map<String, dynamic>> _transactions = const [];
  String? _error;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final results = await Future.wait([
        widget.api.request('GET', '/net-worth'),
        widget.api.request('GET', '/transactions?limit=200'),
      ]);
      final transactionResponse = results[1];
      final rows = transactionResponse is Map
          ? transactionResponse['items'] as List? ?? const []
          : const [];
      if (!mounted) return;
      setState(() {
        _netWorth = Map<String, dynamic>.from(results[0] as Map);
        _transactions = rows
            .whereType<Map>()
            .map((item) => Map<String, dynamic>.from(item))
            .toList();
      });
    } catch (error) {
      if (mounted) setState(() => _error = presentableError(error));
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  double _amount(dynamic value) =>
      double.tryParse(value?.toString().replaceAll(',', '') ?? '') ?? 0;

  double _impact(Map<String, dynamic> item) {
    final amount = _amount(item['amount']);
    switch (item['type']?.toString()) {
      case 'income':
      case 'loan_payment':
        return amount;
      case 'expense':
      case 'investment':
      case 'loan_disbursement':
        return -amount;
      default:
        return 0;
    }
  }

  DateTime? _occurredAt(Map<String, dynamic> item) {
    final value = item['occurredAt']?.toString();
    return value == null ? null : DateTime.tryParse(value)?.toLocal();
  }

  bool _sameDay(DateTime left, DateTime right) =>
      left.year == right.year &&
      left.month == right.month &&
      left.day == right.day;

  String _money(double amount) {
    final parts = amount.toStringAsFixed(0).split('.');
    return parts.first.replaceAllRegExp(RegExp(r'\B(?=(\d{3})+(?!\d))'), ',');
  }

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final monthEnd = DateTime(now.year, now.month + 1, 0);
    final current = _amount(_netWorth?['netWorth']);
    final planned = _transactions.where((item) {
      final when = _occurredAt(item);
      return when != null && when.isAfter(now) && !when.isAfter(monthEnd);
    }).toList();
    final projected =
        current + planned.fold<double>(0, (sum, item) => sum + _impact(item));

    return PageFrame(
      title: 'Dự báo tháng ${now.month}',
      action: IconButton(
        tooltip: 'Tải lại dự báo',
        onPressed: _loading ? null : _load,
        icon: const Icon(Icons.refresh_rounded, color: FinoraColors.accentGold),
      ),
      child: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
          ? FinoraEmptyState(
              title: 'Chưa thể lập dự báo',
              message: _error!,
              icon: Icons.cloud_off_rounded,
              action: FilledButton.icon(
                onPressed: _load,
                icon: const Icon(Icons.refresh_rounded),
                label: const Text('Thử lại'),
              ),
            )
          : RefreshIndicator(
              onRefresh: _load,
              child: ListView(
                padding: const EdgeInsets.only(bottom: FinoraSpace.xxl),
                children: [
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [FinoraColors.primary, FinoraColors.purple],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      borderRadius: BorderRadius.circular(24),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Row(
                          children: [
                            Icon(Icons.auto_graph_rounded, color: Colors.white),
                            SizedBox(width: 8),
                            Text(
                              'TÀI SẢN RÒNG DỰ KIẾN CUỐI THÁNG',
                              style: TextStyle(
                                color: Colors.white70,
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                letterSpacing: .6,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        Text(
                          '${_money(projected)} ${_netWorth?['baseCurrency'] ?? 'VND'}',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 28,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          planned.isEmpty
                              ? 'Chưa có giao dịch nào được lên lịch cho phần còn lại của tháng.'
                              : 'Bao gồm ${planned.length} khoản dự kiến còn lại trong tháng.',
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 12,
                            height: 1.35,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: FinoraSpace.lg),
                  Row(
                    children: [
                      Expanded(
                        child: Metric(
                          'Hiện tại',
                          _money(current),
                          Icons.account_balance_wallet_rounded,
                          accent: FinoraColors.info,
                        ),
                      ),
                      const SizedBox(width: FinoraSpace.sm),
                      Expanded(
                        child: Metric(
                          'Biến động dự kiến',
                          '${planned.fold<double>(0, (sum, item) => sum + _impact(item)) >= 0 ? '+' : ''}${_money(planned.fold<double>(0, (sum, item) => sum + _impact(item)))}',
                          Icons.trending_up_rounded,
                          accent: FinoraColors.success,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: FinoraSpace.xl),
                  const Text('Dòng tiền từng ngày', style: FinoraTypography.h3),
                  const SizedBox(height: 4),
                  const Text(
                    'Các khoản đã ghi nhận hoặc được lên lịch từ hôm nay đến hết tháng.',
                    style: FinoraTypography.caption,
                  ),
                  const SizedBox(height: FinoraSpace.md),
                  ...List.generate(monthEnd.day - now.day + 1, (index) {
                    final date = DateTime(now.year, now.month, now.day + index);
                    final entries = _transactions.where((item) {
                      final when = _occurredAt(item);
                      return when != null && _sameDay(when, date);
                    }).toList();
                    final dailyImpact = entries
                        .where((item) {
                          final when = _occurredAt(item);
                          return when != null && when.isAfter(now);
                        })
                        .fold<double>(0, (sum, item) => sum + _impact(item));
                    return _ForecastDayCard(
                      date: date,
                      isToday: index == 0,
                      entries: entries,
                      impact: dailyImpact,
                      money: _money,
                    );
                  }),
                ],
              ),
            ),
    );
  }
}

class _ForecastDayCard extends StatelessWidget {
  const _ForecastDayCard({
    required this.date,
    required this.isToday,
    required this.entries,
    required this.impact,
    required this.money,
  });

  final DateTime date;
  final bool isToday;
  final List<Map<String, dynamic>> entries;
  final double impact;
  final String Function(double) money;

  @override
  Widget build(BuildContext context) => Container(
    margin: const EdgeInsets.only(bottom: 10),
    padding: const EdgeInsets.all(14),
    decoration: BoxDecoration(
      color: isToday ? FinoraColors.primarySoft : Colors.white,
      borderRadius: BorderRadius.circular(16),
      border: Border.all(
        color: isToday
            ? FinoraColors.primary.withValues(alpha: .45)
            : FinoraColors.border,
      ),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
              decoration: BoxDecoration(
                color: isToday ? FinoraColors.primary : const Color(0xfff1f5f9),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                isToday ? 'Hôm nay' : '${date.day}/${date.month}',
                style: TextStyle(
                  color: isToday ? Colors.white : FinoraColors.textPrimary,
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            const Spacer(),
            if (impact != 0)
              Text(
                '${impact > 0 ? '+' : ''}${money(impact)}',
                style: TextStyle(
                  color: impact > 0
                      ? FinoraColors.success
                      : FinoraColors.danger,
                  fontWeight: FontWeight.w800,
                ),
              ),
          ],
        ),
        const SizedBox(height: 9),
        if (entries.isEmpty)
          const Text(
            'Chưa có khoản thu hoặc chi dự kiến.',
            style: FinoraTypography.caption,
          )
        else
          ...entries.map((item) {
            final value =
                double.tryParse(item['amount']?.toString() ?? '') ?? 0;
            final isIncome = item['type']?.toString() == 'income';
            return Padding(
              padding: const EdgeInsets.only(bottom: 5),
              child: Row(
                children: [
                  Icon(
                    isIncome
                        ? Icons.south_west_rounded
                        : Icons.north_east_rounded,
                    size: 16,
                    color: isIncome
                        ? FinoraColors.success
                        : FinoraColors.danger,
                  ),
                  const SizedBox(width: 7),
                  Expanded(
                    child: Text(
                      item['name']?.toString().isNotEmpty == true
                          ? item['name'].toString()
                          : item['note']?.toString().isNotEmpty == true
                          ? item['note'].toString()
                          : 'Giao dịch',
                      style: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                  Text(
                    '${isIncome ? '+' : '-'}${money(value)}',
                    style: TextStyle(
                      color: isIncome
                          ? FinoraColors.success
                          : FinoraColors.danger,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
            );
          }),
      ],
    ),
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

class AssistantPage extends StatefulWidget {
  const AssistantPage({super.key, required this.api});

  final ApiClient api;

  @override
  State<AssistantPage> createState() => _AssistantPageState();
}

class _AssistantPageState extends State<AssistantPage> {
  final _composer = TextEditingController();
  final _scrollController = ScrollController();
  final List<_AiMessage> _messages = [];
  bool _loading = false;
  bool _historyLoaded = false;

  static const _suggestions = [
    'Tóm tắt tình hình tài sản của tôi',
    'Gợi ý tối ưu ngân sách tháng này',
    'Phân tích các khoản chi gần đây',
  ];

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    try {
      final raw = await widget.api.request('GET', '/assistant/commands');
      final rows = raw is List ? raw : ((raw as Map?)?['items'] as List? ?? []);
      if (!mounted) return;
      setState(() {
        _messages.addAll(
          rows.whereType<Map>().take(12).expand((row) {
            final prompt = row['command']?.toString();
            final response =
                row['response']?.toString() ??
                row['answer']?.toString() ??
                row['plan']?.toString();
            return [
              if (prompt != null && prompt.isNotEmpty) _AiMessage(prompt, true),
              if (response != null && response.isNotEmpty)
                _AiMessage(response, false),
            ];
          }),
        );
        _historyLoaded = true;
      });
    } catch (_) {
      if (mounted) setState(() => _historyLoaded = true);
    }
  }

  Future<void> _send([String? prompt]) async {
    final text = (prompt ?? _composer.text).trim();
    if (text.isEmpty || _loading) return;
    _composer.clear();
    setState(() {
      _messages.add(_AiMessage(text, true));
      _loading = true;
    });
    _scrollToEnd();
    try {
      final raw = await widget.api.request('POST', '/assistant/commands', {
        'command': text,
      });
      final reply = _responseText(raw);
      if (mounted) setState(() => _messages.add(_AiMessage(reply, false)));
    } catch (error) {
      if (mounted) {
        setState(
          () => _messages.add(
            _AiMessage(
              'Mình chưa thể hoàn tất yêu cầu này. '
              'Vui lòng thử lại khi kết nối ổn định hơn.',
              false,
              isError: true,
            ),
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
      _scrollToEnd();
    }
  }

  String _responseText(dynamic raw) {
    if (raw is String && raw.trim().isNotEmpty) return raw;
    if (raw is Map) {
      for (final key in ['response', 'answer', 'content', 'message', 'plan']) {
        final value = raw[key]?.toString();
        if (value != null && value.isNotEmpty) return value;
      }
    }
    return 'Mình đã ghi nhận yêu cầu và đang chuẩn bị phân tích cho bạn.';
  }

  void _scrollToEnd() => WidgetsBinding.instance.addPostFrameCallback((_) {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 260),
        curve: Curves.easeOutCubic,
      );
    }
  });

  @override
  void dispose() {
    _composer.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: FinoraColors.background,
    appBar: AppBar(
      title: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Trợ lý Finora', style: FinoraTypography.h3),
          Text('Không gian kiến thức cá nhân', style: FinoraTypography.caption),
        ],
      ),
      actions: [
        IconButton(
          tooltip: 'Cuộc trò chuyện mới',
          onPressed: _loading ? null : () => setState(_messages.clear),
          icon: const Icon(Icons.edit_note_rounded),
        ),
      ],
    ),
    body: Stack(
      children: [
        Positioned.fill(
          child: IgnorePointer(
            child: Opacity(
              opacity: .045,
              child: Image.asset(
                'assets/images/app_bg_maple_light.png',
                fit: BoxFit.cover,
              ),
            ),
          ),
        ),
        Column(
          children: [
            _AiModeBar(historyLoaded: _historyLoaded),
            Expanded(
              child: _messages.isEmpty && _historyLoaded
                  ? _AiWelcome(onSuggestion: _send)
                  : ListView.builder(
                      controller: _scrollController,
                      padding: const EdgeInsets.fromLTRB(24, 8, 24, 16),
                      itemCount: _messages.length + (_loading ? 1 : 0),
                      itemBuilder: (context, index) {
                        if (index == _messages.length) {
                          return const _AiThinking();
                        }
                        return _AiBubble(message: _messages[index]);
                      },
                    ),
            ),
            _AiComposer(controller: _composer, busy: _loading, onSend: _send),
          ],
        ),
      ],
    ),
  );
}

class _AiMessage {
  const _AiMessage(this.text, this.isUser, {this.isError = false});
  final String text;
  final bool isUser;
  final bool isError;
}

class _AiModeBar extends StatelessWidget {
  const _AiModeBar({required this.historyLoaded});
  final bool historyLoaded;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.fromLTRB(24, 8, 24, 12),
    child: Row(
      children: [
        const _AiPill(
          icon: Icons.auto_awesome_rounded,
          label: 'Finora AI',
          active: true,
        ),
        const SizedBox(width: 8),
        const _AiPill(icon: Icons.menu_book_rounded, label: 'Kho kiến thức'),
        const Spacer(),
        if (!historyLoaded)
          const SizedBox(
            width: 16,
            height: 16,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
      ],
    ),
  );
}

class _AiPill extends StatelessWidget {
  const _AiPill({required this.icon, required this.label, this.active = false});
  final IconData icon;
  final String label;
  final bool active;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
    decoration: BoxDecoration(
      color: active ? FinoraColors.primarySoft : Colors.white,
      borderRadius: FinoraRadius.full,
      border: Border.all(
        color: active ? FinoraColors.purple : FinoraColors.border,
      ),
    ),
    child: Row(
      children: [
        Icon(
          icon,
          size: 15,
          color: active ? FinoraColors.primary : FinoraColors.textSecondary,
        ),
        const SizedBox(width: 6),
        Text(
          label,
          style: FinoraTypography.caption.copyWith(
            color: active ? FinoraColors.primary : FinoraColors.textSecondary,
          ),
        ),
      ],
    ),
  );
}

class _AiWelcome extends StatelessWidget {
  const _AiWelcome({required this.onSuggestion});
  final ValueChanged<String> onSuggestion;
  @override
  Widget build(BuildContext context) => Center(
    child: SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          Container(
            width: 72,
            height: 72,
            decoration: const BoxDecoration(
              color: FinoraColors.primarySoft,
              shape: BoxShape.circle,
            ),
            child: const Icon(
              Icons.auto_awesome_rounded,
              color: FinoraColors.primary,
              size: 32,
            ),
          ),
          const SizedBox(height: 20),
          const Text(
            'Hôm nay mình có thể giúp gì?',
            style: FinoraTypography.h2,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 8),
          Text(
            'Hỏi về dòng tiền, kế hoạch tài chính, hoặc dữ liệu Finora của bạn.',
            style: FinoraTypography.bodySmall.copyWith(
              color: FinoraColors.textSecondary,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 24),
          ..._AssistantPageState._suggestions.map(
            (item) => Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: FinoraCard(
                onTap: () => onSuggestion(item),
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 14,
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.north_east_rounded,
                      size: 18,
                      color: FinoraColors.primary,
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(item, style: FinoraTypography.bodySmall),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    ),
  );
}

class _AiBubble extends StatelessWidget {
  const _AiBubble({required this.message});
  final _AiMessage message;
  @override
  Widget build(BuildContext context) => Padding(
    padding: EdgeInsets.only(
      top: 8,
      left: message.isUser ? 48 : 0,
      right: message.isUser ? 0 : 20,
    ),
    child: Column(
      crossAxisAlignment: message.isUser
          ? CrossAxisAlignment.end
          : CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: message.isUser ? FinoraColors.primary : Colors.white,
            borderRadius: BorderRadius.only(
              topLeft: const Radius.circular(16),
              topRight: const Radius.circular(16),
              bottomLeft: Radius.circular(message.isUser ? 16 : 4),
              bottomRight: Radius.circular(message.isUser ? 4 : 16),
            ),
            border: message.isUser
                ? null
                : Border.all(
                    color: message.isError
                        ? FinoraColors.danger
                        : FinoraColors.border,
                  ),
            boxShadow: message.isUser ? const [] : FinoraElevation.card,
          ),
          child: Text(
            message.text,
            style: FinoraTypography.bodySmall.copyWith(
              color: message.isUser ? Colors.white : FinoraColors.textPrimary,
              height: 1.55,
            ),
          ),
        ),
        if (!message.isUser && !message.isError)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                IconButton(
                  tooltip: 'Sao chép',
                  visualDensity: VisualDensity.compact,
                  onPressed: () async {
                    await Clipboard.setData(ClipboardData(text: message.text));
                    if (context.mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Đã sao chép phản hồi.')),
                      );
                    }
                  },
                  icon: const Icon(
                    Icons.copy_rounded,
                    size: 16,
                    color: FinoraColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
      ],
    ),
  );
}

class _AiThinking extends StatelessWidget {
  const _AiThinking();
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: 16),
    child: Row(
      children: [
        const SizedBox(
          width: 18,
          height: 18,
          child: CircularProgressIndicator(strokeWidth: 2),
        ),
        const SizedBox(width: 10),
        Text(
          'Finora đang suy nghĩ…',
          style: FinoraTypography.caption.copyWith(
            color: FinoraColors.textSecondary,
          ),
        ),
      ],
    ),
  );
}

class _AiComposer extends StatelessWidget {
  const _AiComposer({
    required this.controller,
    required this.busy,
    required this.onSend,
  });
  final TextEditingController controller;
  final bool busy;
  final VoidCallback onSend;
  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Container(
      padding: const EdgeInsets.fromLTRB(24, 12, 24, 16),
      decoration: const BoxDecoration(
        color: Color(0xeefafafc),
        border: Border(top: BorderSide(color: FinoraColors.border)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Expanded(
            child: TextField(
              controller: controller,
              enabled: !busy,
              minLines: 1,
              maxLines: 4,
              textInputAction: TextInputAction.send,
              onSubmitted: (_) => onSend(),
              decoration: const InputDecoration(
                hintText: 'Nhắn cho Finora…',
                prefixIcon: Icon(Icons.add_rounded),
              ),
            ),
          ),
          const SizedBox(width: 8),
          FilledButton(
            onPressed: busy ? null : onSend,
            style: FilledButton.styleFrom(
              minimumSize: const Size(48, 48),
              padding: EdgeInsets.zero,
              shape: const CircleBorder(),
            ),
            child: busy
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(
                      color: Colors.white,
                      strokeWidth: 2,
                    ),
                  )
                : const Icon(Icons.arrow_upward_rounded),
          ),
        ],
      ),
    ),
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
