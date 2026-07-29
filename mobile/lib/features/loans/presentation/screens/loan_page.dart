import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/theme/finora_colors.dart';
import 'package:mobile/core/theme/finora_tokens.dart';
import 'package:mobile/core/theme/finora_typography.dart';
import 'package:mobile/core/utils/vietnamese_money_input.dart';
import 'package:mobile/core/widgets/finora_core_widgets.dart';
import 'package:mobile/features/loans/domain/entities/loan.dart';
import 'package:mobile/features/loans/presentation/view_models/loan_view_model.dart';

class LoanPage extends StatefulWidget {
  const LoanPage({
    super.key,
    required this.viewModel,
    this.api,
    this.autoOpenCreate = false,
  });
  final LoanViewModel viewModel;
  final ApiClient? api;
  final bool autoOpenCreate;

  @override
  State<LoanPage> createState() => _LoanPageState();
}

String loanErrorMessage(Object error) {
  final raw = error.toString().trim();
  if (raw.isEmpty) return 'Đã xảy ra lỗi. Vui lòng thử lại.';
  if (raw.contains('SocketException') || raw.contains('Failed host lookup')) {
    return 'Không thể kết nối máy chủ. Kiểm tra mạng rồi thử lại.';
  }
  return raw.replaceFirst(RegExp(r'^(Exception|ApiException):\s*'), '').trim();
}

class _LoanPageState extends State<LoanPage> {
  @override
  void initState() {
    super.initState();
    widget.viewModel.addListener(_onChanged);
    widget.viewModel.load();
    if (widget.autoOpenCreate) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _openCreateSheet());
    }
  }

  void _onChanged() {
    if (mounted) setState(() {});
  }

  Future<void> _openCreateSheet() async {
    final created = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) =>
          _LoanCreateSheet(viewModel: widget.viewModel, api: widget.api),
    );
    if (created == true && mounted) await widget.viewModel.load();
  }

  Future<void> _openInterestConverter() => showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    backgroundColor: Colors.transparent,
    builder: (_) => const _InterestRateConverterSheet(),
  );

  @override
  void dispose() {
    widget.viewModel
      ..removeListener(_onChanged)
      ..dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final model = widget.viewModel;
    return Scaffold(
      backgroundColor: FinoraColors.background,
      appBar: AppBar(
        backgroundColor: FinoraColors.background,
        foregroundColor: FinoraColors.textPrimary,
        title: const Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Khoản vay', style: FinoraTypography.h3),
            Text(
              'Đầu mối và các khoản phải thu',
              style: FinoraTypography.caption,
            ),
          ],
        ),
        actions: [
          IconButton(
            tooltip: 'Quy đổi lãi suất',
            onPressed: _openInterestConverter,
            icon: const Icon(Icons.calculate_rounded),
          ),
          IconButton(
            tooltip: 'Tạo khoản vay',
            onPressed: model.loading ? null : _openCreateSheet,
            icon: const Icon(Icons.add_rounded),
          ),
          IconButton(
            tooltip: 'Tải lại khoản vay',
            onPressed: model.loading ? null : model.load,
            icon: const Icon(Icons.refresh_rounded),
          ),
        ],
      ),
      body: model.loading
          ? const Center(child: CircularProgressIndicator())
          : model.error != null
          ? FinoraEmptyState(
              icon: Icons.cloud_off_rounded,
              title: 'Chưa tải được khoản vay',
              message: loanErrorMessage(model.error!),
              action: FilledButton.icon(
                onPressed: model.load,
                icon: const Icon(Icons.refresh_rounded),
                label: const Text('Thử lại'),
              ),
            )
          : _LoanList(viewModel: model),
    );
  }
}

class _LoanList extends StatelessWidget {
  const _LoanList({required this.viewModel});
  final LoanViewModel viewModel;

  @override
  Widget build(BuildContext context) {
    final groups = _BorrowerGroup.fromLoans(
      viewModel.loans,
      viewModel.accruals,
    );
    return ListView(
      padding: const EdgeInsets.fromLTRB(
        FinoraSpace.xl,
        FinoraSpace.xs,
        FinoraSpace.xl,
        FinoraSpace.xxl,
      ),
      children: [
        _LoanSummaryCard(summary: viewModel.summary),
        const SizedBox(height: FinoraSpace.md),
        _InterestConverterEntry(
          onTap: () => showModalBottomSheet<void>(
            context: context,
            isScrollControlled: true,
            backgroundColor: Colors.transparent,
            builder: (_) => const _InterestRateConverterSheet(),
          ),
        ),
        const SizedBox(height: FinoraSpace.xl),
        Row(
          children: [
            const Text('Đầu mối', style: FinoraTypography.h3),
            const Spacer(),
            Text('${groups.length} đầu mối', style: FinoraTypography.bodySmall),
          ],
        ),
        const SizedBox(height: FinoraSpace.sm),
        if (groups.isEmpty)
          const FinoraEmptyState(
            icon: Icons.request_quote_outlined,
            title: 'Chưa có khoản vay',
            message: 'Khoản vay mới sẽ xuất hiện ở đây sau khi được tạo.',
          )
        else
          ...groups.map(
            (group) => Padding(
              padding: const EdgeInsets.only(bottom: FinoraSpace.sm),
              child: _BorrowerCard(
                group: group,
                onTap: () => Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (_) =>
                        _BorrowerDetailPage(group: group, viewModel: viewModel),
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }
}

class _LoanSummaryCard extends StatelessWidget {
  const _LoanSummaryCard({required this.summary});
  final LoanSummary summary;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.all(FinoraSpace.xl),
    decoration: const BoxDecoration(
      gradient: LinearGradient(
        colors: [FinoraColors.purple, FinoraColors.primaryDeep],
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      ),
      borderRadius: FinoraRadius.xl,
      boxShadow: FinoraElevation.floating,
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'TỔNG GỐC ĐANG VẬN HÀNH',
          style: TextStyle(
            color: Colors.white70,
            fontWeight: FontWeight.w700,
            letterSpacing: .5,
          ),
        ),
        const SizedBox(height: FinoraSpace.xs),
        FinoraMoney(
          _amount(summary.activePrincipal),
          color: Colors.white,
          style: FinoraTypography.display,
        ),
        const SizedBox(height: FinoraSpace.lg),
        Row(
          children: [
            _summaryMetric('Lãi/ngày', summary.dailyInterest),
            const SizedBox(width: FinoraSpace.xl),
            _summaryMetric('Lãi cộng dồn', summary.accruedInterest),
          ],
        ),
      ],
    ),
  );

  Widget _summaryMetric(String label, String value) => Expanded(
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: FinoraSpace.xxs),
        Text(
          _money(_amount(value)),
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    ),
  );
}

class _InterestConverterEntry extends StatelessWidget {
  const _InterestConverterEntry({required this.onTap});
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => FinoraCard(
    onTap: onTap,
    padding: const EdgeInsets.all(FinoraSpace.md),
    child: Row(
      children: [
        Container(
          width: 44,
          height: 44,
          decoration: const BoxDecoration(
            color: FinoraColors.primarySoft,
            borderRadius: FinoraRadius.md,
          ),
          child: const Icon(
            Icons.calculate_rounded,
            color: FinoraColors.primary,
          ),
        ),
        const SizedBox(width: FinoraSpace.sm),
        const Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Quy đổi lãi suất', style: FinoraTypography.title),
              SizedBox(height: FinoraSpace.xxs),
              Text(
                'Từ “nghìn/đầu triệu” sang % và ngược lại',
                style: FinoraTypography.caption,
              ),
            ],
          ),
        ),
        const Icon(Icons.arrow_forward_rounded, color: FinoraColors.primary),
      ],
    ),
  );
}

class _BorrowerCard extends StatelessWidget {
  const _BorrowerCard({required this.group, required this.onTap});
  final _BorrowerGroup group;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => FinoraCard(
    onTap: onTap,
    child: Row(
      children: [
        CircleAvatar(
          radius: 22,
          backgroundColor: FinoraColors.primarySoft,
          foregroundColor: FinoraColors.primaryDeep,
          child: Text(_initials(group.name), style: FinoraTypography.title),
        ),
        const SizedBox(width: FinoraSpace.sm),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(group.name, style: FinoraTypography.title),
              const SizedBox(height: FinoraSpace.xxs),
              Text(
                '${group.loans.length} khoản vay',
                style: FinoraTypography.caption.copyWith(
                  color: FinoraColors.textSecondary,
                ),
              ),
              const SizedBox(height: FinoraSpace.xs),
              Text(
                'Tổng gốc ${_money(group.principal)}',
                style: FinoraTypography.bodySmall,
              ),
            ],
          ),
        ),
        SizedBox(
          width: 96,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                'Lãi cộng dồn',
                style: FinoraTypography.caption.copyWith(
                  color: FinoraColors.textSecondary,
                ),
              ),
              const SizedBox(height: FinoraSpace.xxs),
              FittedBox(
                fit: BoxFit.scaleDown,
                alignment: Alignment.centerRight,
                child: Text(
                  _money(group.accrued),
                  style: FinoraTypography.title.copyWith(
                    color: FinoraColors.primaryDeep,
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    ),
  );
}

class _BorrowerDetailPage extends StatefulWidget {
  const _BorrowerDetailPage({required this.group, required this.viewModel});
  final _BorrowerGroup group;
  final LoanViewModel viewModel;

  @override
  State<_BorrowerDetailPage> createState() => _BorrowerDetailPageState();
}

class _BorrowerDetailPageState extends State<_BorrowerDetailPage> {
  late final Future<_BorrowerLoanMetrics> _metrics = _loadMetrics();

  Future<_BorrowerLoanMetrics> _loadMetrics() async {
    final histories = await Future.wait(
      widget.group.loans.map((loan) => widget.viewModel.payments(loan.id)),
    );
    return _BorrowerLoanMetrics.fromLoans(widget.group.loans, histories);
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    backgroundColor: FinoraColors.background,
    appBar: AppBar(
      backgroundColor: FinoraColors.background,
      foregroundColor: FinoraColors.textPrimary,
      title: Text(widget.group.name),
    ),
    body: ListView(
      padding: const EdgeInsets.all(FinoraSpace.md),
      children: [
        FutureBuilder<_BorrowerLoanMetrics>(
          future: _metrics,
          builder: (context, snapshot) {
            final metrics =
                snapshot.data ??
                _BorrowerLoanMetrics.fromLoans(widget.group.loans, const []);
            return _BorrowerOverview(
              metrics: metrics,
              loading: snapshot.connectionState == ConnectionState.waiting,
            );
          },
        ),
        const SizedBox(height: FinoraSpace.xl),
        const Text('Khoản vay', style: FinoraTypography.h3),
        const SizedBox(height: FinoraSpace.sm),
        ...widget.group.loans.map(
          (loan) => Padding(
            padding: const EdgeInsets.only(bottom: FinoraSpace.sm),
            child: _LoanCard(
              loan: loan,
              accrual: widget.viewModel.accruals[loan.id],
              onTap: () => Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (_) =>
                      _LoanDetailPage(loan: loan, viewModel: widget.viewModel),
                ),
              ),
            ),
          ),
        ),
      ],
    ),
  );
}

class _BorrowerOverview extends StatelessWidget {
  const _BorrowerOverview({required this.metrics, required this.loading});
  final _BorrowerLoanMetrics metrics;
  final bool loading;

  @override
  Widget build(BuildContext context) => Column(
    children: [
      Container(
        width: double.infinity,
        padding: const EdgeInsets.all(FinoraSpace.lg),
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            colors: [FinoraColors.purple, FinoraColors.primaryDeep],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: FinoraRadius.xl,
          boxShadow: FinoraElevation.floating,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'TỔNG QUAN KHOẢN VAY',
              style: TextStyle(
                color: Color(0xdfffffff),
                fontSize: 11,
                fontWeight: FontWeight.w800,
                letterSpacing: .6,
              ),
            ),
            const SizedBox(height: FinoraSpace.md),
            Row(
              children: [
                Expanded(
                  child: _overviewMetric(
                    'Đã cho vay',
                    _money(metrics.lentPrincipal),
                  ),
                ),
                const SizedBox(width: FinoraSpace.md),
                Expanded(
                  child: _overviewMetric(
                    'Đã thu hồi gốc',
                    _money(metrics.recoveredPrincipal),
                  ),
                ),
              ],
            ),
            const SizedBox(height: FinoraSpace.md),
            Row(
              children: [
                Expanded(
                  child: _overviewMetric(
                    'Lãi đã nhận',
                    loading
                        ? 'Đang tổng hợp'
                        : _money(metrics.receivedInterest),
                  ),
                ),
                const SizedBox(width: FinoraSpace.md),
                Expanded(
                  child: _overviewMetric(
                    'Dư nợ còn lại',
                    _money(metrics.outstandingPrincipal),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
      const SizedBox(height: FinoraSpace.md),
      FinoraCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(
                  Icons.task_alt_rounded,
                  color: FinoraColors.success,
                  size: 20,
                ),
                const SizedBox(width: FinoraSpace.xs),
                const Expanded(
                  child: Text(
                    'Tỷ lệ quyết toán',
                    style: FinoraTypography.title,
                  ),
                ),
                Text(
                  '${(metrics.settlementRate * 100).round()}%',
                  style: FinoraTypography.title.copyWith(
                    color: FinoraColors.success,
                  ),
                ),
              ],
            ),
            const SizedBox(height: FinoraSpace.sm),
            ClipRRect(
              borderRadius: FinoraRadius.full,
              child: LinearProgressIndicator(
                value: metrics.settlementRate,
                minHeight: 8,
                backgroundColor: FinoraColors.primarySoft,
                valueColor: const AlwaysStoppedAnimation(FinoraColors.success),
              ),
            ),
            const SizedBox(height: FinoraSpace.xs),
            Text(
              '${metrics.settledLoans}/${metrics.loanCount} khoản đã quyết toán',
              style: FinoraTypography.caption.copyWith(
                color: FinoraColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
      const SizedBox(height: FinoraSpace.md),
      FinoraCard(
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: const BoxDecoration(
                color: FinoraColors.primarySoft,
                borderRadius: FinoraRadius.sm,
              ),
              child: const Icon(
                Icons.auto_graph_rounded,
                color: FinoraColors.primary,
              ),
            ),
            const SizedBox(width: FinoraSpace.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Dự báo lãi 30 ngày',
                    style: FinoraTypography.title,
                  ),
                  const SizedBox(height: FinoraSpace.xxs),
                  Text(
                    metrics.outstandingPrincipal > 0
                        ? 'Ước tính theo dư nợ và lãi suất hiện tại'
                        : 'Không còn dư nợ để dự báo lãi',
                    style: FinoraTypography.caption.copyWith(
                      color: FinoraColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: FinoraSpace.sm),
            FittedBox(
              fit: BoxFit.scaleDown,
              alignment: Alignment.centerRight,
              child: Text(
                _money(metrics.forecastInterest30Days),
                style: FinoraTypography.title.copyWith(
                  color: FinoraColors.primaryDeep,
                ),
              ),
            ),
          ],
        ),
      ),
    ],
  );

  Widget _overviewMetric(String label, String value) => Container(
    padding: const EdgeInsets.all(FinoraSpace.sm),
    decoration: const BoxDecoration(
      color: Color(0x20ffffff),
      borderRadius: FinoraRadius.sm,
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            color: Color(0xdfffffff),
            fontSize: 11,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: FinoraSpace.xxs),
        Text(
          value,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 14,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    ),
  );
}

class _BorrowerLoanMetrics {
  const _BorrowerLoanMetrics({
    required this.lentPrincipal,
    required this.recoveredPrincipal,
    required this.receivedInterest,
    required this.outstandingPrincipal,
    required this.forecastInterest30Days,
    required this.settledLoans,
    required this.loanCount,
  });

  final double lentPrincipal;
  final double recoveredPrincipal;
  final double receivedInterest;
  final double outstandingPrincipal;
  final double forecastInterest30Days;
  final int settledLoans;
  final int loanCount;

  double get settlementRate => loanCount == 0 ? 0 : settledLoans / loanCount;

  factory _BorrowerLoanMetrics.fromLoans(
    List<Loan> loans,
    List<List<LoanPaymentRecord>> histories,
  ) {
    var lentPrincipal = 0.0;
    var recoveredPrincipal = 0.0;
    var receivedInterest = 0.0;
    var outstandingPrincipal = 0.0;
    var forecastInterest30Days = 0.0;
    var settledLoans = 0;

    for (var index = 0; index < loans.length; index++) {
      final loan = loans[index];
      final payments = index < histories.length ? histories[index] : const [];
      final receivedPrincipal = payments.fold<double>(
        0,
        (total, payment) => total + _amount(payment.principal),
      );
      final initial = _amount(loan.principalInitial);
      final balance = _amount(loan.principalBalance);
      lentPrincipal += initial > 0 ? initial : balance + receivedPrincipal;
      recoveredPrincipal += receivedPrincipal;
      receivedInterest += payments.fold<double>(
        0,
        (total, payment) => total + _amount(payment.interest),
      );
      outstandingPrincipal += balance;
      if (_loanStatusPresentation(loan).label == 'Đã quyết toán') {
        settledLoans++;
      } else {
        forecastInterest30Days += _dailyInterest(loan) * 30;
      }
    }

    return _BorrowerLoanMetrics(
      lentPrincipal: lentPrincipal,
      recoveredPrincipal: recoveredPrincipal,
      receivedInterest: receivedInterest,
      outstandingPrincipal: outstandingPrincipal,
      forecastInterest30Days: forecastInterest30Days,
      settledLoans: settledLoans,
      loanCount: loans.length,
    );
  }
}

class _LoanCard extends StatelessWidget {
  const _LoanCard({
    required this.loan,
    required this.accrual,
    required this.onTap,
  });
  final Loan loan;
  final LoanAccrual? accrual;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => FinoraCard(
    onTap: onTap,
    child: Row(
      children: [
        const Icon(Icons.request_quote_outlined, color: FinoraColors.primary),
        const SizedBox(width: FinoraSpace.sm),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                _money(_amount(loan.principalBalance)),
                style: FinoraTypography.title,
              ),
              const SizedBox(height: FinoraSpace.xxs),
              Row(
                children: [
                  Text(
                    '${loan.days} ngày',
                    style: FinoraTypography.caption.copyWith(
                      color: FinoraColors.textSecondary,
                    ),
                  ),
                  const SizedBox(width: FinoraSpace.xs),
                  _LoanStatusBadge(loan: loan),
                ],
              ),
            ],
          ),
        ),
        SizedBox(
          width: 88,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                'Lãi cộng dồn',
                style: FinoraTypography.caption.copyWith(
                  color: FinoraColors.textSecondary,
                ),
              ),
              FittedBox(
                fit: BoxFit.scaleDown,
                alignment: Alignment.centerRight,
                child: Text(
                  _money(
                    _amount(accrual?.accruedInterest ?? loan.accruedInterest),
                  ),
                  style: FinoraTypography.bodySmall,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: FinoraSpace.xs),
        const Icon(
          Icons.chevron_right_rounded,
          color: FinoraColors.textSecondary,
        ),
      ],
    ),
  );
}

class _LoanStatusBadge extends StatelessWidget {
  const _LoanStatusBadge({required this.loan});
  final Loan loan;

  @override
  Widget build(BuildContext context) {
    final presentation = _loanStatusPresentation(loan);
    return Semantics(
      label: 'Trạng thái: ${presentation.label}',
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 4),
        decoration: BoxDecoration(
          color: presentation.background,
          borderRadius: FinoraRadius.full,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(presentation.icon, size: 12, color: presentation.foreground),
            const SizedBox(width: 4),
            Text(
              presentation.label,
              style: FinoraTypography.caption.copyWith(
                color: presentation.foreground,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _LoanStatusPresentation {
  const _LoanStatusPresentation({
    required this.label,
    required this.foreground,
    required this.background,
    required this.icon,
  });

  final String label;
  final Color foreground;
  final Color background;
  final IconData icon;
}

class _LoanDetailPage extends StatefulWidget {
  const _LoanDetailPage({required this.loan, required this.viewModel});
  final Loan loan;
  final LoanViewModel viewModel;

  @override
  State<_LoanDetailPage> createState() => _LoanDetailPageState();
}

class _LoanDetailPageState extends State<_LoanDetailPage> {
  List<LoanPaymentRecord> _payments = const [];
  bool _loadingPayments = true;
  Object? _paymentsError;
  int _paymentLoadVersion = 0;
  DateTime? _lastInterestPaidAt;
  bool _deleting = false;

  @override
  void initState() {
    super.initState();
    _loadPayments();
  }

  Future<void> _openCollection() async {
    final payment = await showModalBottomSheet<LoanPaymentRecord>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) =>
          _LoanCollectionSheet(loan: widget.loan, viewModel: widget.viewModel),
    );
    if (!mounted || payment == null) return;

    // The POST response is authoritative. Rendering it directly avoids a
    // stale or eventually-consistent history GET hiding a successful receipt.
    setState(() {
      _paymentLoadVersion++;
      _payments = _mergePayment(payment);
      _paymentsError = null;
      _loadingPayments = false;
      if (_amount(payment.interest) > 0) {
        _lastInterestPaidAt = _paymentDate(payment);
      }
    });
  }

  Future<void> _loadPayments() async {
    final loadVersion = ++_paymentLoadVersion;
    setState(() {
      _loadingPayments = true;
      _paymentsError = null;
    });
    try {
      final payments = await widget.viewModel.payments(widget.loan.id);
      if (!mounted || loadVersion != _paymentLoadVersion) return;
      setState(() => _payments = payments);
    } catch (error) {
      if (!mounted || loadVersion != _paymentLoadVersion) return;
      setState(() => _paymentsError = error);
    } finally {
      if (mounted && loadVersion == _paymentLoadVersion) {
        setState(() => _loadingPayments = false);
      }
    }
  }

  List<LoanPaymentRecord> _mergePayment(LoanPaymentRecord payment) {
    final merged = [
      payment,
      ..._payments.where((item) => item.id != payment.id),
    ];
    merged.sort((a, b) => _paymentDate(b).compareTo(_paymentDate(a)));
    return merged;
  }

  DateTime _paymentDate(LoanPaymentRecord payment) =>
      DateTime.tryParse(payment.occurredAt) ??
      DateTime.fromMillisecondsSinceEpoch(0);

  Future<void> _deleteLoan() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Xóa khoản vay?'),
        content: const Text(
          'Khoản vay chưa có lịch sử thu sẽ bị xóa vĩnh viễn. '
          'Các khoản đã có lịch sử thu được giữ lại để bảo toàn đối soát.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Hủy'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: FinoraColors.danger),
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Xóa khoản vay'),
          ),
        ],
      ),
    );
    if (confirmed != true || _deleting) return;
    setState(() => _deleting = true);
    try {
      await widget.viewModel.delete(widget.loan.id);
      if (mounted) Navigator.pop(context);
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(loanErrorMessage(error))));
      }
    } finally {
      if (mounted) setState(() => _deleting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final loan = widget.loan;
    final accrual = widget.viewModel.accruals[loan.id];
    final accrued = _amount(accrual?.accruedInterest ?? loan.accruedInterest);
    final lastInterestPaidAt =
        _lastInterestPaidAt ??
        DateTime.tryParse(accrual?.lastInterestPaidAt ?? '');
    final principal = _amount(loan.principalBalance);
    final daily = _dailyInterest(loan);
    return Scaffold(
      backgroundColor: FinoraColors.background,
      appBar: AppBar(
        backgroundColor: FinoraColors.background,
        foregroundColor: FinoraColors.textPrimary,
        title: const Text('Chi tiết khoản vay'),
        actions: [
          IconButton(
            tooltip: 'Xóa khoản vay',
            onPressed: _deleting ? null : _deleteLoan,
            icon: _deleting
                ? const SizedBox.square(
                    dimension: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.delete_outline_rounded),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(FinoraSpace.md),
        children: [
          FinoraCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(loan.borrower, style: FinoraTypography.h3),
                    ),
                    const SizedBox(width: FinoraSpace.sm),
                    _LoanStatusBadge(loan: loan),
                  ],
                ),
                const SizedBox(height: FinoraSpace.lg),
                _detailRow('Gốc', _money(principal)),
                _detailRow('Lãi/ngày', _money(daily)),
                _detailRow(
                  'Số ngày đang vay',
                  '${accrual?.days ?? loan.days} ngày',
                ),
                _detailRow('Lãi cộng dồn', _money(accrued)),
                if (lastInterestPaidAt != null)
                  _detailRow(
                    'Lần nhận lãi gần nhất',
                    _date(lastInterestPaidAt.toLocal()),
                  ),
                const Divider(height: FinoraSpace.xl),
                _detailRow(
                  'Tổng phải thu',
                  _money(principal + accrued),
                  emphasize: true,
                ),
              ],
            ),
          ),
          const SizedBox(height: FinoraSpace.lg),
          FilledButton.icon(
            onPressed: _openCollection,
            icon: const Icon(Icons.payments_rounded),
            label: const Text('Ghi nhận thu'),
          ),
          const SizedBox(height: FinoraSpace.xl),
          const Text('Lịch sử thu', style: FinoraTypography.h3),
          const SizedBox(height: FinoraSpace.sm),
          if (_loadingPayments)
            const Center(child: CircularProgressIndicator())
          else if (_paymentsError != null)
            FinoraEmptyState(
              icon: Icons.cloud_off_rounded,
              title: 'Chưa tải được lịch sử thu',
              message: loanErrorMessage(_paymentsError!),
              action: TextButton.icon(
                onPressed: _loadPayments,
                icon: const Icon(Icons.refresh_rounded),
                label: const Text('Thử lại'),
              ),
            )
          else if (_payments.isEmpty)
            const FinoraEmptyState(
              icon: Icons.timeline_rounded,
              title: 'Chưa có khoản thu',
              message: 'Các lần thu lãi hoặc gốc sẽ xuất hiện tại đây.',
            )
          else
            Column(
              children: _payments.map((payment) {
                final principalAmount = _amount(payment.principal);
                final interestAmount = _amount(payment.interest);
                final feeAmount = _amount(payment.fee);
                final total = principalAmount + interestAmount + feeAmount;
                return Padding(
                  padding: const EdgeInsets.only(bottom: FinoraSpace.sm),
                  child: FinoraCard(
                    child: Row(
                      children: [
                        const CircleAvatar(
                          backgroundColor: FinoraColors.primarySoft,
                          foregroundColor: FinoraColors.primary,
                          child: Icon(Icons.payments_rounded),
                        ),
                        const SizedBox(width: FinoraSpace.sm),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                _date(
                                  DateTime.tryParse(
                                        payment.occurredAt,
                                      )?.toLocal() ??
                                      DateTime.now(),
                                ),
                                style: FinoraTypography.label,
                              ),
                              const SizedBox(height: 3),
                              Text(
                                _paymentSummary(
                                  payment,
                                  principalAmount,
                                  interestAmount,
                                ),
                                style: FinoraTypography.caption,
                              ),
                            ],
                          ),
                        ),
                        Text(_money(total), style: FinoraTypography.label),
                      ],
                    ),
                  ),
                );
              }).toList(),
            ),
        ],
      ),
    );
  }

  Widget _detailRow(String label, String value, {bool emphasize = false}) =>
      Padding(
        padding: const EdgeInsets.only(bottom: FinoraSpace.sm),
        child: Row(
          children: [
            Expanded(
              child: Text(
                label,
                style: FinoraTypography.bodySmall.copyWith(
                  color: FinoraColors.textSecondary,
                ),
              ),
            ),
            Flexible(
              child: Text(
                value,
                textAlign: TextAlign.end,
                style: emphasize
                    ? FinoraTypography.title
                    : FinoraTypography.body,
              ),
            ),
          ],
        ),
      );
}

class _LoanCreateSheet extends StatefulWidget {
  const _LoanCreateSheet({required this.viewModel, this.api});
  final LoanViewModel viewModel;
  final ApiClient? api;

  @override
  State<_LoanCreateSheet> createState() => _LoanCreateSheetState();
}

class _LoanCreateSheetState extends State<_LoanCreateSheet> {
  final principal = TextEditingController();
  final dailyRate = TextEditingController();
  _LoanCustomer? customer;
  bool submitting = false;

  Future<void> _pickCustomer() async {
    if (widget.api == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Không thể tải danh sách khách hàng.')),
      );
      return;
    }
    final selected = await showModalBottomSheet<_LoanCustomer>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) => _CustomerPickerSheet(api: widget.api!),
    );
    if (selected != null && mounted) {
      setState(() => customer = selected);
    }
  }

  Future<void> _submit() async {
    final principalAmount = _parseLoanAmount(principal.text);
    final dailyRatePerMillion = _parseDailyRatePerMillion(dailyRate.text);
    if (customer == null || principalAmount <= 0 || dailyRatePerMillion <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Chọn khách hàng, số gốc và lãi mỗi triệu/ngày hợp lệ.',
          ),
        ),
      );
      return;
    }
    setState(() => submitting = true);
    try {
      final now = DateTime.now();
      await widget.viewModel.create({
        'customerId': customer!.id,
        'counterparty': customer!.name,
        'direction': 'receivable',
        'principalInitial': principalAmount.toStringAsFixed(0),
        'annualRate': _annualRateFromDaily(
          dailyRatePerMillion,
        ).toStringAsFixed(3),
        'dailyRatePerMillion': dailyRatePerMillion.toStringAsFixed(2),
        'dayCountBasis': 'actual_365',
        'startAt': now.toUtc().toIso8601String(),
        'dueAt': now.add(const Duration(days: 30)).toUtc().toIso8601String(),
        'interestCompounding': false,
      });
      if (mounted) {
        Navigator.pop(context, true);
      }
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(loanErrorMessage(error))));
      }
    } finally {
      if (mounted) {
        setState(() => submitting = false);
      }
    }
  }

  @override
  void dispose() {
    principal.dispose();
    dailyRate.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Container(
      padding: EdgeInsets.fromLTRB(
        FinoraSpace.lg,
        FinoraSpace.sm,
        FinoraSpace.lg,
        FinoraSpace.xl + MediaQuery.of(context).viewInsets.bottom,
      ),
      decoration: const BoxDecoration(
        color: FinoraColors.surfaceElevated,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: const BoxDecoration(
                  color: FinoraColors.borderStrong,
                  borderRadius: FinoraRadius.full,
                ),
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            const Text('Cho vay mới', style: FinoraTypography.h3),
            const SizedBox(height: FinoraSpace.xs),
            Text(
              'Nhập số gốc và mức lãi dân gian. Finora tự quy đổi phần còn lại.',
              style: FinoraTypography.bodySmall.copyWith(
                color: FinoraColors.textSecondary,
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            _LoanCustomerField(customer: customer, onTap: _pickCustomer),
            const SizedBox(height: FinoraSpace.sm),
            TextField(
              controller: principal,
              keyboardType: TextInputType.text,
              onChanged: (_) => setState(() {}),
              decoration: const InputDecoration(
                labelText: 'Số gốc',
                hintText: 'Ví dụ: 30tr, 30t hoặc 30m',
              ),
            ),
            const SizedBox(height: FinoraSpace.sm),
            Container(
              padding: const EdgeInsets.all(FinoraSpace.md),
              decoration: BoxDecoration(
                color: FinoraColors.primarySoft,
                borderRadius: FinoraRadius.md,
                border: Border.all(
                  color: FinoraColors.primary.withValues(alpha: .35),
                ),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Row(
                    children: [
                      Icon(Icons.bolt_rounded, color: FinoraColors.primary),
                      SizedBox(width: 8),
                      Text(
                        'LÃI MỖI TRIỆU / NGÀY',
                        style: FinoraTypography.label,
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: dailyRate,
                    autofocus: true,
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
                    onChanged: (_) => setState(() {}),
                    decoration: const InputDecoration(
                      labelText: 'Mức lãi',
                      hintText: 'Ví dụ: 3',
                      suffixText: 'nghìn đồng',
                      helperText: 'Nhập 3 = 3.000đ cho mỗi 1 triệu đồng/ngày',
                    ),
                  ),
                  const SizedBox(height: 10),
                  _LoanRatePreview(
                    principal: _parseLoanAmount(principal.text),
                    dailyRatePerMillion: _parseDailyRatePerMillion(
                      dailyRate.text,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            FilledButton.icon(
              onPressed: submitting ? null : _submit,
              icon: submitting
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : const Icon(Icons.check_rounded),
              label: const Text('Tạo khoản vay'),
            ),
          ],
        ),
      ),
    ),
  );
}

class _LoanRatePreview extends StatelessWidget {
  const _LoanRatePreview({
    required this.principal,
    required this.dailyRatePerMillion,
  });

  final double principal;
  final double dailyRatePerMillion;

  @override
  Widget build(BuildContext context) {
    if (dailyRatePerMillion <= 0) {
      return const Text(
        'Nhập mức lãi, ví dụ 3 hoặc 3000.',
        style: FinoraTypography.caption,
      );
    }
    final dailyInterest = principal / 1000000 * dailyRatePerMillion;
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: .7),
        borderRadius: FinoraRadius.sm,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Đã hiểu: ${_money(dailyRatePerMillion)} / 1 triệu / ngày',
            style: FinoraTypography.label.copyWith(
              color: FinoraColors.primaryDeep,
            ),
          ),
          const SizedBox(height: 3),
          Text(
            principal > 0
                ? 'Khoản này sinh lãi khoảng ${_money(dailyInterest)} mỗi ngày · tương đương ${_formatNumber(_annualRateFromDaily(dailyRatePerMillion), decimals: 1)}%/năm.'
                : 'Tương đương ${_formatNumber(_annualRateFromDaily(dailyRatePerMillion), decimals: 1)}%/năm.',
            style: FinoraTypography.caption,
          ),
        ],
      ),
    );
  }
}

class _LoanCustomer {
  const _LoanCustomer({required this.id, required this.name, this.phone = ''});
  final String id;
  final String name;
  final String phone;

  factory _LoanCustomer.fromJson(Map<dynamic, dynamic> json) => _LoanCustomer(
    id: json['id']?.toString() ?? '',
    name: json['name']?.toString() ?? '',
    phone: json['phone']?.toString() ?? '',
  );
}

class _LoanCustomerField extends StatelessWidget {
  const _LoanCustomerField({required this.customer, required this.onTap});
  final _LoanCustomer? customer;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => Semantics(
    button: true,
    label: customer == null ? 'Chọn khách hàng' : 'Đổi khách hàng',
    child: Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: FinoraRadius.md,
        child: Container(
          constraints: const BoxConstraints(minHeight: 56),
          padding: const EdgeInsets.symmetric(
            horizontal: FinoraSpace.md,
            vertical: FinoraSpace.sm,
          ),
          decoration: BoxDecoration(
            color: customer == null
                ? const Color(0xfffaf9ff)
                : FinoraColors.primarySoft,
            borderRadius: FinoraRadius.md,
            border: Border.all(
              color: customer == null
                  ? FinoraColors.border
                  : FinoraColors.purple,
            ),
          ),
          child: Row(
            children: [
              Container(
                width: 32,
                height: 32,
                decoration: const BoxDecoration(
                  color: Colors.white,
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  customer == null
                      ? Icons.person_search_rounded
                      : Icons.person_rounded,
                  size: 18,
                  color: FinoraColors.primary,
                ),
              ),
              const SizedBox(width: FinoraSpace.sm),
              Expanded(
                child: customer == null
                    ? const Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text('Khách hàng', style: FinoraTypography.label),
                          SizedBox(height: 2),
                          Text(
                            'Tìm hoặc thêm người vay',
                            style: FinoraTypography.bodySmall,
                          ),
                        ],
                      )
                    : Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(customer!.name, style: FinoraTypography.title),
                          if (customer!.phone.isNotEmpty) ...[
                            const SizedBox(height: 2),
                            Text(
                              customer!.phone,
                              style: FinoraTypography.caption,
                            ),
                          ],
                        ],
                      ),
              ),
              Text(
                customer == null ? 'Chọn' : 'Đổi',
                style: FinoraTypography.label.copyWith(
                  color: FinoraColors.primary,
                ),
              ),
              const SizedBox(width: 2),
              const Icon(
                Icons.chevron_right_rounded,
                color: FinoraColors.primary,
              ),
            ],
          ),
        ),
      ),
    ),
  );
}

class _CustomerPickerSheet extends StatefulWidget {
  const _CustomerPickerSheet({required this.api});
  final ApiClient api;

  @override
  State<_CustomerPickerSheet> createState() => _CustomerPickerSheetState();
}

class _CustomerPickerSheetState extends State<_CustomerPickerSheet> {
  final _search = TextEditingController();
  List<_LoanCustomer> _customers = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load([String query = '']) async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final encoded = Uri.encodeQueryComponent(query);
      final data = await widget.api.request(
        'GET',
        '/customers?q=$encoded&limit=50',
      );
      final rows = data is List
          ? data
          : ((data as Map?)?['items'] as List? ?? []);
      if (!mounted || query != _search.text.trim()) return;
      setState(
        () => _customers = rows
            .whereType<Map>()
            .map(_LoanCustomer.fromJson)
            .where((item) => item.id.isNotEmpty && item.name.isNotEmpty)
            .toList(),
      );
    } catch (error) {
      if (mounted) setState(() => _error = loanErrorMessage(error));
    } finally {
      if (mounted && query == _search.text.trim()) {
        setState(() => _loading = false);
      }
    }
  }

  Future<void> _createCustomer() async {
    final created = await showModalBottomSheet<_LoanCustomer>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (_) =>
          _CustomerCreateSheet(api: widget.api, initialName: _search.text),
    );
    if (created != null && mounted) Navigator.pop(context, created);
  }

  @override
  void dispose() {
    _search.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Container(
      height: MediaQuery.of(context).size.height * .78,
      padding: EdgeInsets.fromLTRB(
        FinoraSpace.xl,
        FinoraSpace.sm,
        FinoraSpace.xl,
        FinoraSpace.md + MediaQuery.of(context).viewInsets.bottom,
      ),
      decoration: const BoxDecoration(
        color: FinoraColors.surface,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: Column(
        children: [
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: const BoxDecoration(
                color: FinoraColors.borderStrong,
                borderRadius: FinoraRadius.full,
              ),
            ),
          ),
          const SizedBox(height: FinoraSpace.lg),
          const Align(
            alignment: Alignment.centerLeft,
            child: Text('Chọn khách hàng', style: FinoraTypography.h3),
          ),
          const SizedBox(height: FinoraSpace.xs),
          TextField(
            controller: _search,
            autofocus: true,
            textCapitalization: TextCapitalization.words,
            onChanged: (value) => _load(value.trim()),
            decoration: const InputDecoration(
              hintText: 'Tìm tên hoặc số điện thoại',
              prefixIcon: Icon(Icons.search_rounded),
            ),
          ),
          const SizedBox(height: FinoraSpace.md),
          Expanded(child: _buildResults()),
          const SizedBox(height: FinoraSpace.sm),
          SizedBox(
            width: double.infinity,
            child: OutlinedButton.icon(
              onPressed: _createCustomer,
              icon: const Icon(Icons.person_add_alt_1_rounded),
              label: Text(
                _search.text.trim().isEmpty
                    ? 'Thêm khách hàng mới'
                    : 'Tạo “${_search.text.trim()}”',
              ),
            ),
          ),
        ],
      ),
    ),
  );

  Widget _buildResults() {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      return FinoraEmptyState(
        icon: Icons.cloud_off_rounded,
        title: 'Chưa tải được khách hàng',
        message: 'Kiểm tra kết nối rồi thử lại.',
        action: FilledButton(onPressed: _load, child: const Text('Thử lại')),
      );
    }
    if (_customers.isEmpty) {
      return FinoraEmptyState(
        icon: Icons.people_outline_rounded,
        title: _search.text.trim().isEmpty
            ? 'Chưa có khách hàng'
            : 'Không tìm thấy khách hàng',
        message: 'Bạn có thể thêm người vay mới ngay bây giờ.',
      );
    }
    return ListView.separated(
      itemCount: _customers.length,
      separatorBuilder: (_, _) => const SizedBox(height: FinoraSpace.xs),
      itemBuilder: (context, index) {
        final item = _customers[index];
        return FinoraCard(
          onTap: () => Navigator.pop(context, item),
          padding: const EdgeInsets.symmetric(
            horizontal: FinoraSpace.md,
            vertical: FinoraSpace.sm,
          ),
          child: Row(
            children: [
              CircleAvatar(
                radius: 20,
                backgroundColor: FinoraColors.primarySoft,
                foregroundColor: FinoraColors.primary,
                child: Text(
                  _initials(item.name),
                  style: FinoraTypography.bodySmall,
                ),
              ),
              const SizedBox(width: FinoraSpace.sm),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(item.name, style: FinoraTypography.title),
                    Text(
                      item.phone.isEmpty ? 'Chưa có số điện thoại' : item.phone,
                      style: FinoraTypography.caption.copyWith(
                        color: FinoraColors.textSecondary,
                      ),
                    ),
                  ],
                ),
              ),
              const Icon(
                Icons.chevron_right_rounded,
                color: FinoraColors.textSecondary,
              ),
            ],
          ),
        );
      },
    );
  }
}

class _CustomerCreateSheet extends StatefulWidget {
  const _CustomerCreateSheet({required this.api, required this.initialName});
  final ApiClient api;
  final String initialName;

  @override
  State<_CustomerCreateSheet> createState() => _CustomerCreateSheetState();
}

class _CustomerCreateSheetState extends State<_CustomerCreateSheet> {
  late final _name = TextEditingController(text: widget.initialName.trim());
  final _phone = TextEditingController();
  bool _saving = false;

  Future<void> _save() async {
    final name = _name.text.trim();
    if (name.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Nhập tên khách hàng.')));
      return;
    }
    setState(() => _saving = true);
    try {
      final data = await widget.api.request('POST', '/customers', {
        'name': name,
        'phone': _phone.text.trim(),
      });
      if (mounted) {
        Navigator.pop(context, _LoanCustomer.fromJson(data as Map));
      }
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(loanErrorMessage(error))));
      }
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  void dispose() {
    _name.dispose();
    _phone.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Container(
      padding: EdgeInsets.fromLTRB(
        FinoraSpace.xl,
        FinoraSpace.sm,
        FinoraSpace.xl,
        FinoraSpace.xl + MediaQuery.of(context).viewInsets.bottom,
      ),
      decoration: const BoxDecoration(
        color: FinoraColors.surface,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: const BoxDecoration(
                  color: FinoraColors.borderStrong,
                  borderRadius: FinoraRadius.full,
                ),
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            const Text('Thêm khách hàng', style: FinoraTypography.h3),
            const SizedBox(height: FinoraSpace.xs),
            Text(
              'Số điện thoại được lưu để hỗ trợ gọi và nhắc thu sau này.',
              style: FinoraTypography.bodySmall.copyWith(
                color: FinoraColors.textSecondary,
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            TextField(
              controller: _name,
              autofocus: true,
              textCapitalization: TextCapitalization.words,
              decoration: const InputDecoration(labelText: 'Tên khách hàng'),
            ),
            const SizedBox(height: FinoraSpace.sm),
            TextField(
              controller: _phone,
              keyboardType: TextInputType.phone,
              decoration: const InputDecoration(
                labelText: 'Số điện thoại',
                hintText: 'Ví dụ: 090 123 4567',
                helperText: 'Không bắt buộc',
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            FilledButton.icon(
              onPressed: _saving ? null : _save,
              icon: _saving
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(
                        color: Colors.white,
                        strokeWidth: 2,
                      ),
                    )
                  : const Icon(Icons.check_rounded),
              label: const Text('Lưu và chọn khách hàng'),
            ),
          ],
        ),
      ),
    ),
  );
}

class _LoanCollectionSheet extends StatefulWidget {
  const _LoanCollectionSheet({required this.loan, required this.viewModel});
  final Loan loan;
  final LoanViewModel viewModel;

  @override
  State<_LoanCollectionSheet> createState() => _LoanCollectionSheetState();
}

class _LoanCollectionSheetState extends State<_LoanCollectionSheet> {
  bool _interestMode = true;
  int _days = 1;
  DateTime _receivedAt = DateTime.now();
  late final TextEditingController _amountController;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _amountController = TextEditingController(
      text: _suggested.round().toString(),
    );
  }

  double get _suggested => _interestMode
      ? _dailyInterest(widget.loan) * _days
      : _amount(widget.loan.principalBalance);

  void _updateSuggestion() =>
      setState(() => _amountController.text = _suggested.round().toString());

  Future<void> _pickDate() async {
    final date = await showDatePicker(
      context: context,
      initialDate: _receivedAt,
      firstDate: DateTime(2020),
      lastDate: DateTime.now().add(const Duration(days: 1)),
    );
    if (date != null) {
      setState(() => _receivedAt = date);
    }
  }

  Future<void> _save() async {
    final amount = parseVietnameseMoneyInput(_amountController.text);
    if (amount <= 0) return;
    setState(() => _saving = true);
    try {
      final payment = await widget.viewModel.receive(widget.loan.id, {
        'principalAmount': _interestMode ? '0' : amount.toStringAsFixed(0),
        'interestAmount': _interestMode ? amount.toStringAsFixed(0) : '0',
        'interestDays': _interestMode ? _days : 0,
        'feeAmount': '0',
        'waivedAmount': '0',
        'occurredAt': _receivedAt.toUtc().toIso8601String(),
      });
      if (mounted) {
        Navigator.pop(context, payment);
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Không thể lưu khoản thu. Vui lòng thử lại.'),
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _saving = false);
      }
    }
  }

  @override
  void dispose() {
    _amountController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => SafeArea(
    top: false,
    child: Container(
      padding: EdgeInsets.fromLTRB(
        FinoraSpace.xl,
        FinoraSpace.sm,
        FinoraSpace.xl,
        MediaQuery.of(context).viewInsets.bottom + FinoraSpace.xl,
      ),
      decoration: const BoxDecoration(
        color: FinoraColors.surface,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Center(
              child: Container(
                width: 40,
                height: 4,
                decoration: const BoxDecoration(
                  color: FinoraColors.borderStrong,
                  borderRadius: FinoraRadius.full,
                ),
              ),
            ),
            const SizedBox(height: FinoraSpace.lg),
            const Text('Ghi nhận thu', style: FinoraTypography.h3),
            const SizedBox(height: FinoraSpace.md),
            SegmentedButton<bool>(
              segments: const [
                ButtonSegment(value: true, label: Text('Thu lãi')),
                ButtonSegment(value: false, label: Text('Thu gốc')),
              ],
              selected: {_interestMode},
              onSelectionChanged: (value) {
                _interestMode = value.first;
                _updateSuggestion();
              },
            ),
            const SizedBox(height: FinoraSpace.lg),
            if (_interestMode) ...[
              Text('Số ngày nhận lãi', style: FinoraTypography.label),
              const SizedBox(height: FinoraSpace.xs),
              Wrap(
                spacing: FinoraSpace.xs,
                children: [1, 3, 7, 30]
                    .map(
                      (days) => ChoiceChip(
                        label: Text('$days ngày'),
                        selected: _days == days,
                        onSelected: (_) {
                          _days = days;
                          _updateSuggestion();
                        },
                      ),
                    )
                    .toList(),
              ),
              const SizedBox(height: FinoraSpace.sm),
              Row(
                children: [
                  IconButton(
                    onPressed: _days > 1
                        ? () {
                            _days--;
                            _updateSuggestion();
                          }
                        : null,
                    icon: const Icon(Icons.remove_circle_outline),
                  ),
                  Text('$_days ngày', style: FinoraTypography.title),
                  IconButton(
                    onPressed: () {
                      _days++;
                      _updateSuggestion();
                    },
                    icon: const Icon(Icons.add_circle_outline),
                  ),
                ],
              ),
              Text(
                'Tự tính: ${_money(_amount(widget.loan.dailyRatePerMillion))} / 1 triệu / ngày × ${_formatNumber(_amount(widget.loan.principalBalance) / 1000000, decimals: 2)} triệu × $_days ngày = ${_money(_suggested)}',
                style: FinoraTypography.bodySmall.copyWith(
                  color: FinoraColors.textSecondary,
                ),
              ),
              const SizedBox(height: FinoraSpace.lg),
            ],
            TextField(
              controller: _amountController,
              keyboardType: TextInputType.number,
              inputFormatters: [FilteringTextInputFormatter.digitsOnly],
              decoration: const InputDecoration(
                labelText: 'Số tiền thực nhận',
                helperText: 'Chỉ nhập chữ số. Số tiền tự tính theo số ngày.',
              ),
            ),
            const SizedBox(height: FinoraSpace.md),
            OutlinedButton.icon(
              onPressed: _pickDate,
              icon: const Icon(Icons.calendar_today_outlined),
              label: Text('Ngày nhận: ${_date(_receivedAt)}'),
            ),
            const SizedBox(height: FinoraSpace.lg),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: _saving ? null : _save,
                child: _saving
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Lưu khoản thu'),
              ),
            ),
          ],
        ),
      ),
    ),
  );
}

enum _RateInputKind { perMillion, dailyPercent, monthlyPercent, annualPercent }

class _RateConversion {
  const _RateConversion(this.dailyPercent);
  final double dailyPercent;

  double get perMillionPerDay => dailyPercent * 10000;
  double get monthlyPercent => dailyPercent * 30;
  double get annualPercent => dailyPercent * 365;
  double interestFor(double principal, int days) =>
      principal * dailyPercent / 100 * days;
}

class _InterestRateConverterSheet extends StatefulWidget {
  const _InterestRateConverterSheet();

  @override
  State<_InterestRateConverterSheet> createState() =>
      _InterestRateConverterSheetState();
}

class _InterestRateConverterSheetState
    extends State<_InterestRateConverterSheet> {
  final _value = TextEditingController(text: '3000');
  _RateInputKind _kind = _RateInputKind.perMillion;

  _RateConversion get _conversion {
    final value = _parseInterestRate(_value.text);
    final dailyPercent = switch (_kind) {
      _RateInputKind.perMillion => value / 10000,
      _RateInputKind.dailyPercent => value,
      _RateInputKind.monthlyPercent => value / 30,
      _RateInputKind.annualPercent => value / 365,
    };
    return _RateConversion(dailyPercent.clamp(0, double.infinity));
  }

  String get _label => switch (_kind) {
    _RateInputKind.perMillion => 'Đồng lãi / 1 triệu / ngày',
    _RateInputKind.dailyPercent => 'Lãi suất theo ngày (%)',
    _RateInputKind.monthlyPercent => 'Lãi suất theo tháng (%)',
    _RateInputKind.annualPercent => 'Lãi suất theo năm (%)',
  };

  String get _helper => _kind == _RateInputKind.perMillion
      ? 'Ví dụ: 3.000 tương đương “3 nghìn/đầu triệu/ngày”.'
      : 'Lãi đơn, quy đổi theo 30 ngày/tháng và 365 ngày/năm.';

  void _setQuickRate(double perMillion) {
    setState(() {
      _kind = _RateInputKind.perMillion;
      _value.text = _formatNumber(perMillion);
    });
  }

  @override
  void dispose() {
    _value.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final conversion = _conversion;
    return SafeArea(
      top: false,
      child: Container(
        constraints: BoxConstraints(
          maxHeight: MediaQuery.of(context).size.height * .91,
        ),
        padding: EdgeInsets.fromLTRB(
          FinoraSpace.xl,
          FinoraSpace.sm,
          FinoraSpace.xl,
          FinoraSpace.xl + MediaQuery.of(context).viewInsets.bottom,
        ),
        decoration: const BoxDecoration(
          color: FinoraColors.surface,
          borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        ),
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: const BoxDecoration(
                    color: FinoraColors.borderStrong,
                    borderRadius: FinoraRadius.full,
                  ),
                ),
              ),
              const SizedBox(height: FinoraSpace.lg),
              Row(
                children: [
                  Container(
                    width: 42,
                    height: 42,
                    decoration: const BoxDecoration(
                      color: FinoraColors.primarySoft,
                      borderRadius: FinoraRadius.md,
                    ),
                    child: const Icon(
                      Icons.calculate_rounded,
                      color: FinoraColors.primary,
                    ),
                  ),
                  const SizedBox(width: FinoraSpace.sm),
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Quy đổi lãi suất', style: FinoraTypography.h3),
                        SizedBox(height: 2),
                        Text(
                          'Lãi đơn · 30 ngày/tháng · 365 ngày/năm',
                          style: FinoraTypography.caption,
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    tooltip: 'Đóng',
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close_rounded),
                  ),
                ],
              ),
              const SizedBox(height: FinoraSpace.lg),
              Text(
                'CÁCH NHẬP LÃI',
                style: FinoraTypography.label.copyWith(
                  color: FinoraColors.textSecondary,
                ),
              ),
              const SizedBox(height: FinoraSpace.xs),
              Wrap(
                spacing: FinoraSpace.xs,
                runSpacing: FinoraSpace.xs,
                children: [
                  _rateKindChip(_RateInputKind.perMillion, 'Đồng / triệu'),
                  _rateKindChip(_RateInputKind.dailyPercent, '% ngày'),
                  _rateKindChip(_RateInputKind.monthlyPercent, '% tháng'),
                  _rateKindChip(_RateInputKind.annualPercent, '% năm'),
                ],
              ),
              const SizedBox(height: FinoraSpace.md),
              TextField(
                controller: _value,
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                onChanged: (_) => setState(() {}),
                decoration: InputDecoration(
                  labelText: _label,
                  helperText: _helper,
                ),
              ),
              const SizedBox(height: FinoraSpace.sm),
              Wrap(
                spacing: FinoraSpace.xs,
                children: [2000, 2500, 3000, 3500, 4000]
                    .map(
                      (rate) => ActionChip(
                        label: Text('${_formatNumber(rate / 1000)} nghìn'),
                        onPressed: () => _setQuickRate(rate.toDouble()),
                      ),
                    )
                    .toList(),
              ),
              const SizedBox(height: FinoraSpace.xl),
              _RateResultCard(conversion: conversion),
              const SizedBox(height: FinoraSpace.lg),
              _RateFormula(conversion: conversion),
              const SizedBox(height: FinoraSpace.lg),
              _RateExamples(conversion: conversion),
            ],
          ),
        ),
      ),
    );
  }

  Widget _rateKindChip(_RateInputKind kind, String label) => ChoiceChip(
    label: Text(label),
    selected: _kind == kind,
    onSelected: (_) => setState(() => _kind = kind),
  );
}

class _RateResultCard extends StatelessWidget {
  const _RateResultCard({required this.conversion});
  final _RateConversion conversion;

  @override
  Widget build(BuildContext context) => Container(
    width: double.infinity,
    padding: const EdgeInsets.all(FinoraSpace.md),
    decoration: BoxDecoration(
      color: FinoraColors.primarySoft,
      borderRadius: FinoraRadius.lg,
      border: Border.all(color: FinoraColors.border),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('MỨC LÃI TƯƠNG ĐƯƠNG', style: FinoraTypography.label),
        const SizedBox(height: FinoraSpace.xs),
        Text(
          '${_formatNumber(conversion.perMillionPerDay, decimals: 2)} đồng / đầu triệu / ngày',
          style: FinoraTypography.title.copyWith(color: FinoraColors.primary),
        ),
        const SizedBox(height: FinoraSpace.md),
        Row(
          children: [
            _rateMetric('% / ngày', conversion.dailyPercent),
            _rateMetric('% / tháng', conversion.monthlyPercent),
            _rateMetric('% / năm', conversion.annualPercent),
          ],
        ),
      ],
    ),
  );

  Widget _rateMetric(String label, double value) => Expanded(
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(_formatNumber(value, decimals: 3), style: FinoraTypography.title),
        const SizedBox(height: 2),
        Text(label, style: FinoraTypography.caption),
      ],
    ),
  );
}

class _RateFormula extends StatelessWidget {
  const _RateFormula({required this.conversion});
  final _RateConversion conversion;

  @override
  Widget build(BuildContext context) => FinoraCard(
    padding: const EdgeInsets.all(FinoraSpace.md),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Công thức quy đổi', style: FinoraTypography.title),
        const SizedBox(height: FinoraSpace.sm),
        Text(
          'Lãi %/ngày = ${_formatNumber(conversion.perMillionPerDay, decimals: 2)} ÷ 1.000.000 × 100 = ${_formatNumber(conversion.dailyPercent, decimals: 3)}%',
          style: FinoraTypography.bodySmall,
        ),
        const SizedBox(height: FinoraSpace.xs),
        Text(
          '%/tháng = %/ngày × 30 · %/năm = %/ngày × 365',
          style: FinoraTypography.caption.copyWith(
            color: FinoraColors.textSecondary,
          ),
        ),
      ],
    ),
  );
}

class _RateExamples extends StatelessWidget {
  const _RateExamples({required this.conversion});
  final _RateConversion conversion;

  @override
  Widget build(BuildContext context) => FinoraCard(
    padding: const EdgeInsets.all(FinoraSpace.md),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('Ví dụ tiền lãi phải trả', style: FinoraTypography.title),
        const SizedBox(height: 2),
        Text(
          'Lãi đơn, chưa gồm phí và gốc.',
          style: FinoraTypography.caption.copyWith(
            color: FinoraColors.textSecondary,
          ),
        ),
        const SizedBox(height: FinoraSpace.sm),
        const Divider(height: 1),
        ...[50000000.0, 100000000.0, 200000000.0].map(
          (principal) => Padding(
            padding: const EdgeInsets.symmetric(vertical: FinoraSpace.sm),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Vay ${_money(principal)}',
                  style: FinoraTypography.bodySmall,
                ),
                const SizedBox(height: 4),
                Text(
                  'Ngày ${_money(conversion.interestFor(principal, 1))}  ·  Tháng ${_money(conversion.interestFor(principal, 30))}  ·  Năm ${_money(conversion.interestFor(principal, 365))}',
                  style: FinoraTypography.caption.copyWith(
                    color: FinoraColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    ),
  );
}

class _BorrowerGroup {
  const _BorrowerGroup(this.name, this.loans, this.principal, this.accrued);
  final String name;
  final List<Loan> loans;
  final double principal;
  final double accrued;

  static List<_BorrowerGroup> fromLoans(
    List<Loan> loans,
    Map<String, LoanAccrual> accruals,
  ) {
    final buckets = <String, List<Loan>>{};
    for (final loan in loans) {
      buckets.putIfAbsent(loan.borrower, () => []).add(loan);
    }
    return buckets.entries.map((entry) {
      final principal = entry.value.fold<double>(
        0,
        (sum, loan) => sum + _amount(loan.principalBalance),
      );
      final accrued = entry.value.fold<double>(
        0,
        (sum, loan) =>
            sum +
            _amount(accruals[loan.id]?.accruedInterest ?? loan.accruedInterest),
      );
      return _BorrowerGroup(entry.key, entry.value, principal, accrued);
    }).toList()..sort((a, b) => b.principal.compareTo(a.principal));
  }
}

double _amount(String value) =>
    double.tryParse(value.replaceAll(RegExp(r'[^0-9.-]'), '')) ?? 0;

double _parseLoanAmount(String raw) {
  return parseVietnameseMoneyInput(raw);
}

double _parseDailyRatePerMillion(String raw) {
  final value = raw.trim().toLowerCase().replaceAll(' ', '');
  if (value.isEmpty) return 0;
  final match = RegExp(r'^([0-9.,]+)(k|nghìn|nghin)?$').firstMatch(value);
  if (match == null) return 0;
  final parsed =
      double.tryParse(
        _normalizeLoanNumber(match.group(1) ?? '', hasSuffix: false),
      ) ??
      0;
  if (parsed <= 0) return 0;
  if (match.group(2) != null) return parsed * 1000;
  // In this field, a short value is a convenient shorthand in thousands.
  // Thus 3 means 3,000 VND per one million VND per day.
  return parsed < 1000 ? parsed * 1000 : parsed;
}

double _annualRateFromDaily(double dailyRatePerMillion) =>
    dailyRatePerMillion * 365 * 100 / 1000000;

String _normalizeLoanNumber(String raw, {required bool hasSuffix}) {
  final dots = '.'.allMatches(raw).length;
  final commas = ','.allMatches(raw).length;
  if (dots > 0 && commas > 0) {
    final decimalAt = raw.lastIndexOf(RegExp(r'[.,]'));
    final integer = raw.substring(0, decimalAt).replaceAll(RegExp(r'[.,]'), '');
    return '$integer.${raw.substring(decimalAt + 1)}';
  }
  final separator = dots > 0 ? '.' : (commas > 0 ? ',' : '');
  if (separator.isEmpty) return raw;
  final count = dots + commas;
  final lastGroup = raw.substring(raw.lastIndexOf(separator) + 1);
  if (count > 1 || (!hasSuffix && lastGroup.length == 3)) {
    return raw.replaceAll(separator, '');
  }
  return raw.replaceAll(separator, '.');
}

double _parseInterestRate(String raw) {
  final normalized = raw
      .trim()
      .toLowerCase()
      .replaceAll('đ', '')
      .replaceAll('vnd', '')
      .replaceAll('nghìn', 'k')
      .replaceAll('nghin', 'k')
      .replaceAll(' ', '')
      .replaceAll(',', '.');
  final match = RegExp(r'^([0-9.]+)(k)?$').firstMatch(normalized);
  if (match == null) return 0;
  final value = double.tryParse(match.group(1) ?? '') ?? 0;
  return match.group(2) == 'k' ? value * 1000 : value;
}

String _formatNumber(num value, {int decimals = 0}) {
  final rounded = value.toStringAsFixed(decimals);
  return rounded.replaceFirst(RegExp(r'\.?0+$'), '');
}

double _dailyInterest(Loan loan) =>
    _amount(loan.principalBalance) *
    _amount(loan.dailyRatePerMillion) /
    1000000;

String _paymentSummary(
  LoanPaymentRecord payment,
  double principal,
  double interest,
) {
  if (interest > 0 && principal <= 0) {
    final interestDays = payment.interestDays ?? 0;
    final days = interestDays > 0 ? ' · $interestDays ngày' : '';
    return 'Thu lãi$days · ${_money(interest)}';
  }
  if (principal > 0 && interest <= 0) return 'Thu gốc · ${_money(principal)}';
  return 'Gốc ${_money(principal)} · Lãi ${_money(interest)}';
}

String _money(num value) {
  final digits = value.abs().round().toString();
  final output = StringBuffer();
  for (var i = 0; i < digits.length; i++) {
    if (i > 0 && (digits.length - i) % 3 == 0) output.write(',');
    output.write(digits[i]);
  }
  return '${value < 0 ? '-' : ''}$output VND';
}

String _date(DateTime value) =>
    '${value.day.toString().padLeft(2, '0')}/${value.month.toString().padLeft(2, '0')}/${value.year}';

String _initials(String name) => name
    .trim()
    .split(RegExp(r'\s+'))
    .where((part) => part.isNotEmpty)
    .take(2)
    .map((part) => part[0].toUpperCase())
    .join();

_LoanStatusPresentation _loanStatusPresentation(Loan loan) {
  final status = loan.status.trim().toLowerCase();
  // A zero principal is settled even if an older API response has not yet
  // reflected the closed status.
  if (status == 'closed' || _amount(loan.principalBalance) <= 0) {
    return const _LoanStatusPresentation(
      label: 'Đã quyết toán',
      foreground: FinoraColors.primaryDeep,
      background: FinoraColors.primarySoft,
      icon: Icons.task_alt_rounded,
    );
  }
  if (status == 'active') {
    return const _LoanStatusPresentation(
      label: 'Hoạt động',
      foreground: Color(0xff18864b),
      background: Color(0xffe9f8ef),
      icon: Icons.radio_button_checked_rounded,
    );
  }
  if (status == 'overdue') {
    return const _LoanStatusPresentation(
      label: 'Quá hạn',
      foreground: Color(0xffb45309),
      background: Color(0xfffff4df),
      icon: Icons.schedule_rounded,
    );
  }
  return _LoanStatusPresentation(
    label: _statusLabel(status),
    foreground: FinoraColors.textSecondary,
    background: const Color(0xfff1f1f3),
    icon: Icons.info_outline_rounded,
  );
}

String _statusLabel(String status) => switch (status) {
  'draft' => 'Nháp',
  'restructured' => 'Cơ cấu lại',
  'written_off' => 'Đã xóa nợ',
  'cancelled' => 'Đã hủy',
  _ => status.isEmpty ? 'Chưa xác định' : status,
};
