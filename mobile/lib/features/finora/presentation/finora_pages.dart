import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/theme/finora_colors.dart';
import 'package:mobile/features/auth/presentation/view_models/login_view_model.dart';
import 'package:webview_flutter/webview_flutter.dart';

part 'screens/scenario_pages.dart';
part 'screens/resource_support_pages.dart';
part 'models/presentation_models.dart';

abstract final class _I18n {
  static const Map<String, Map<String, String>> _strings = {
    'VN': {
      'subtitle': 'Quản lý Chi tiêu & Đầu tư Cá nhân',
      'bannerTitle': 'TÀI SẢN & DÒNG TIỀN',
      'bannerSub': 'Quản lý thu chi • Tăng trưởng danh mục đầu tư',
      'greeting': 'Chúc bạn một ngày tốt lành',
      'newAccount': 'TẠO TÀI KHOẢN MỚI',
      'loginBtn': 'Đăng nhập',
      'registerBtn': 'Tạo tài khoản',
      'emailLabel': 'Email đăng nhập',
      'passLabel': 'Mật khẩu',
      'nameLabel': 'Họ và tên',
      'switchRegister': 'Chưa có tài khoản? Tạo tài khoản',
      'switchLogin': 'Đã có tài khoản? Đăng nhập',
      'netWorth': 'Tài sản ròng',
      'aiAssistant': 'Trợ lý AI',
      'security': 'Bảo mật 100%',
      'logExpense': 'Ghi thu chi',
      'portfolio': 'Danh mục',
      'budget': 'Ngân sách',
      'forecast': 'Dự báo',
      'notifications': 'Thông báo & Cảnh báo',
      'markAllRead': 'Đánh dấu đã đọc',
      'langTitle': 'Chọn Ngôn Ngữ',
    },
    'EN': {
      'subtitle': 'Personal Finance & Wealth Management',
      'bannerTitle': 'ASSETS & CASH FLOW',
      'bannerSub': 'Expense tracking • Investment portfolio growth',
      'greeting': 'Have a wonderful day',
      'newAccount': 'CREATE NEW ACCOUNT',
      'loginBtn': 'Sign In',
      'registerBtn': 'Create Account',
      'emailLabel': 'Email Address',
      'passLabel': 'Password',
      'nameLabel': 'Full Name',
      'switchRegister': "Don't have an account? Sign Up",
      'switchLogin': 'Already have an account? Sign In',
      'netWorth': 'Net Worth',
      'aiAssistant': 'AI Assistant',
      'security': '100% Secure',
      'logExpense': 'Log Expense',
      'portfolio': 'Portfolio',
      'budget': 'Budget',
      'forecast': 'Forecast',
      'notifications': 'Notifications & Alerts',
      'markAllRead': 'Mark all as read',
      'langTitle': 'Select Language',
    },
    'JP': {
      'subtitle': '個人財務および投資管理',
      'bannerTitle': '資産とキャッシュフロー',
      'bannerSub': '収支管理 • ポートフォリオ成長',
      'greeting': '良い一日をお過ごしください',
      'newAccount': '新規アカウント作成',
      'loginBtn': 'ログイン',
      'registerBtn': 'アカウント作成',
      'emailLabel': 'メールアドレス',
      'passLabel': 'パスワード',
      'nameLabel': '氏名',
      'switchRegister': 'アカウントをお持ちでない方',
      'switchLogin': '既にアカウントをお持ちの方',
      'netWorth': '純資産',
      'aiAssistant': 'AIアシスタント',
      'security': '100% セキュリティ',
      'logExpense': '収支記録',
      'portfolio': 'ポートフォリオ',
      'budget': '予算',
      'forecast': '予測',
      'notifications': '通知とアラート',
      'markAllRead': 'すべて既読にする',
      'langTitle': '言語を選択',
    },
    'KR': {
      'subtitle': '개인 자산 및 투자 관리',
      'bannerTitle': '자산 및 현금 흐름',
      'bannerSub': '지출 추적 • 포트폴리오 성과',
      'greeting': '좋은 하루 보내세요',
      'newAccount': '새 계정 만들기',
      'loginBtn': '로그인',
      'registerBtn': '계정 만들기',
      'emailLabel': '이메일 주소',
      'passLabel': '비밀번호',
      'nameLabel': '이름',
      'switchRegister': '계정이 없으신가요? 가입하기',
      'switchLogin': '이미 계정이 있으신가요? 로그인',
      'netWorth': '순자산',
      'aiAssistant': 'AI 어시스턴트',
      'security': '100% 보안',
      'logExpense': '지출 기록',
      'portfolio': '포트폴리오',
      'budget': '예산',
      'forecast': '예측',
      'notifications': '알림 및 경고',
      'markAllRead': '모두 읽음으로 표시',
      'langTitle': '언어 선택',
    },
  };

  static String t(String lang, String key) {
    return _strings[lang]?[key] ?? _strings['VN']![key] ?? key;
  }

  static String getFlag(String code) {
    switch (code) {
      case 'EN':
        return '🇺🇸';
      case 'JP':
        return '🇯🇵';
      case 'KR':
        return '🇰🇷';
      default:
        return '🇻🇳';
    }
  }
}

class LoginPage extends StatefulWidget {
  const LoginPage({
    super.key,
    required this.viewModel,
    required this.homeBuilder,
  });
  final LoginViewModel viewModel;
  final WidgetBuilder homeBuilder;
  @override
  State<LoginPage> createState() => _LoginPageState();
}

String appAmountDisplayMode = 'full';

class _LoginPageState extends State<LoginPage> with TickerProviderStateMixin {
  final email = TextEditingController(text: 'thanhoangz');
  final password = TextEditingController(text: 'HoangThanZ6^');
  final name = TextEditingController(text: 'Than Hoang Z');
  bool registering = false;
  bool obscurePassword = true;
  String currentLang = 'VN';
  bool hasUnreadNotifications = true;

  late final AnimationController _entranceController;
  late final AnimationController _pulseController;

  late final Animation<double> _headerFade;
  late final Animation<Offset> _headerSlide;
  late final Animation<double> _formFade;
  late final Animation<Offset> _formSlide;
  late final Animation<double> _bottomNavFade;
  late final Animation<Offset> _bottomNavSlide;
  late final Animation<double> _bgScale;

  @override
  void initState() {
    super.initState();
    widget.viewModel.addListener(_onViewModelChanged);

    _entranceController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1100),
    );

    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 10),
    )..repeat(reverse: true);

    _bgScale = Tween<double>(begin: 1.0, end: 1.05).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );

    _headerFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.0, 0.5, curve: Curves.easeOut),
    );
    _headerSlide = Tween<Offset>(begin: const Offset(0, -0.2), end: Offset.zero)
        .animate(
          CurvedAnimation(
            parent: _entranceController,
            curve: const Interval(0.0, 0.6, curve: Curves.easeOutCubic),
          ),
        );

    _formFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.2, 0.7, curve: Curves.easeOut),
    );
    _formSlide = Tween<Offset>(begin: const Offset(0, 0.15), end: Offset.zero)
        .animate(
          CurvedAnimation(
            parent: _entranceController,
            curve: const Interval(0.2, 0.8, curve: Curves.easeOutCubic),
          ),
        );

    _bottomNavFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.45, 0.9, curve: Curves.easeOut),
    );
    _bottomNavSlide =
        Tween<Offset>(begin: const Offset(0, 0.25), end: Offset.zero).animate(
          CurvedAnimation(
            parent: _entranceController,
            curve: const Interval(0.45, 0.95, curve: Curves.easeOutCubic),
          ),
        );

    _entranceController.forward();
  }

  void _onViewModelChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  Future<void> submit() async {
    final authenticated = await widget.viewModel.authenticate(
      registering: registering,
      email: email.text,
      password: password.text,
      name: name.text,
    );
    if (authenticated && mounted) {
      Navigator.of(
        context,
      ).pushReplacement(MaterialPageRoute(builder: widget.homeBuilder));
    }
  }

  void _showAppSettingsSheet() {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xff200733),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Container(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(
                        Icons.settings_rounded,
                        color: Color(0xfffbbf24),
                        size: 22,
                      ),
                      const SizedBox(width: 10),
                      const Text(
                        'Cấu Hình Chế Độ Số Tiền',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const Spacer(),
                      IconButton(
                        onPressed: () => Navigator.pop(context),
                        icon: const Icon(
                          Icons.close_rounded,
                          color: Colors.white70,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Chọn cách hiển thị số tiền trên toàn bộ ứng dụng:',
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.7),
                      fontSize: 13,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildDisplayModeOption(
                    title: 'Viết tắt / Rút gọn (100)',
                    subtitle: 'Ví dụ: 100.000 VND hiển thị thành 100',
                    value: 'compact',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 10),
                  _buildDisplayModeOption(
                    title: 'Đầy đủ (100.000 VND)',
                    subtitle: 'Hiển thị nguyên văn giá trị gốc đầy đủ',
                    value: 'full',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 16),
                ],
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildDisplayModeOption({
    required String title,
    required String subtitle,
    required String value,
    required StateSetter setModalState,
  }) {
    final isSelected = appAmountDisplayMode == value;
    return InkWell(
      onTap: () {
        setState(() {
          appAmountDisplayMode = value;
        });
        setModalState(() {});
      },
      borderRadius: BorderRadius.circular(16),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 250),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isSelected
              ? const Color(0x33fbbf24)
              : Colors.white.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected
                ? const Color(0xfffbbf24)
                : Colors.white.withValues(alpha: 0.12),
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      color: isSelected
                          ? const Color(0xfffbbf24)
                          : Colors.white,
                      fontSize: 15,
                      fontWeight: isSelected
                          ? FontWeight.bold
                          : FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.6),
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            if (isSelected)
              const Icon(
                Icons.check_circle_rounded,
                color: Color(0xfffbbf24),
                size: 22,
              ),
          ],
        ),
      ),
    );
  }

  void _showLanguageSelector() {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xff200733),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      builder: (context) {
        final langs = [
          {'code': 'VN', 'flag': '🇻🇳', 'name': 'Tiếng Việt'},
          {'code': 'EN', 'flag': '🇺🇸', 'name': 'English'},
          {'code': 'JP', 'flag': '🇯🇵', 'name': '日本語'},
          {'code': 'KR', 'flag': '🇰🇷', 'name': '한국어'},
        ];
        return Container(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Icon(
                    Icons.translate_rounded,
                    color: Color(0xfffbbf24),
                    size: 22,
                  ),
                  const SizedBox(width: 10),
                  Text(
                    _I18n.t(currentLang, 'langTitle'),
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const Spacer(),
                  IconButton(
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(
                      Icons.close_rounded,
                      color: Colors.white70,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              ...langs.map((l) {
                final isSelected = l['code'] == currentLang;
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: InkWell(
                    onTap: () {
                      setState(() => currentLang = l['code']!);
                      Navigator.pop(context);
                    },
                    borderRadius: BorderRadius.circular(16),
                    child: AnimatedContainer(
                      duration: const Duration(milliseconds: 250),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                      decoration: BoxDecoration(
                        color: isSelected
                            ? const Color(0x33fbbf24)
                            : Colors.white.withValues(alpha: 0.08),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: isSelected
                              ? const Color(0xfffbbf24)
                              : Colors.white.withValues(alpha: 0.12),
                          width: isSelected ? 1.5 : 1,
                        ),
                      ),
                      child: Row(
                        children: [
                          Text(
                            l['flag']!,
                            style: const TextStyle(fontSize: 22),
                          ),
                          const SizedBox(width: 14),
                          Text(
                            l['name']!,
                            style: TextStyle(
                              color: isSelected
                                  ? const Color(0xfffbbf24)
                                  : Colors.white,
                              fontSize: 15,
                              fontWeight: isSelected
                                  ? FontWeight.bold
                                  : FontWeight.w600,
                            ),
                          ),
                          const Spacer(),
                          if (isSelected)
                            const Icon(
                              Icons.check_circle_rounded,
                              color: Color(0xfffbbf24),
                              size: 20,
                            ),
                        ],
                      ),
                    ),
                  ),
                );
              }),
              const SizedBox(height: 12),
            ],
          ),
        );
      },
    );
  }

  void _showNotificationSheet() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: const Color(0xff200733),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            final notifications = [
              {
                'icon': Icons.security_rounded,
                'color': const Color(0xff38bdf8),
                'title': 'Bảo mật tài khoản',
                'desc': 'Phát hiện đăng nhập mới từ macOS (Mac Mini).',
                'time': '10 phút trước',
                'unread': true,
              },
              {
                'icon': Icons.trending_up_rounded,
                'color': const Color(0xff4ade80),
                'title': 'Danh mục đầu tư',
                'desc': 'Tổng tài sản ròng tăng +2.4% trong tuần này.',
                'time': '2 giờ trước',
                'unread': true,
              },
              {
                'icon': Icons.pie_chart_rounded,
                'color': const Color(0xfffacc15),
                'title': 'Cảnh báo Ngân sách',
                'desc': 'Bạn đã sử dụng 75% ngân sách ăn uống tháng 7.',
                'time': 'Hôm qua',
                'unread': false,
              },
              {
                'icon': Icons.smart_toy_rounded,
                'color': const Color(0xffc084fc),
                'title': 'Trợ lý AI Finora',
                'desc': 'Đã hoàn tất phân tích báo cáo tài chính định kỳ.',
                'time': '2 ngày trước',
                'unread': false,
              },
            ];

            return Container(
              height: MediaQuery.of(context).size.height * 0.65,
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(
                        Icons.notifications_active_rounded,
                        color: Color(0xfffbbf24),
                        size: 22,
                      ),
                      const SizedBox(width: 10),
                      Text(
                        _I18n.t(currentLang, 'notifications'),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const Spacer(),
                      TextButton(
                        onPressed: () {
                          setState(() => hasUnreadNotifications = false);
                          setModalState(() {});
                        },
                        child: Text(
                          _I18n.t(currentLang, 'markAllRead'),
                          style: const TextStyle(
                            color: Color(0xfffbbf24),
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                      IconButton(
                        onPressed: () => Navigator.pop(context),
                        icon: const Icon(
                          Icons.close_rounded,
                          color: Colors.white70,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Expanded(
                    child: ListView.separated(
                      itemCount: notifications.length,
                      separatorBuilder: (_, _) => const SizedBox(height: 10),
                      itemBuilder: (context, i) {
                        final item = notifications[i];
                        final isUnread =
                            hasUnreadNotifications && (item['unread'] as bool);
                        return Container(
                          padding: const EdgeInsets.all(14),
                          decoration: BoxDecoration(
                            color: isUnread
                                ? Colors.white.withValues(alpha: 0.12)
                                : Colors.white.withValues(alpha: 0.06),
                            borderRadius: BorderRadius.circular(18),
                            border: Border.all(
                              color: isUnread
                                  ? const Color(
                                      0xfffbbf24,
                                    ).withValues(alpha: 0.3)
                                  : Colors.white.withValues(alpha: 0.1),
                            ),
                          ),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Container(
                                padding: const EdgeInsets.all(10),
                                decoration: BoxDecoration(
                                  color: (item['color'] as Color).withValues(
                                    alpha: 0.15,
                                  ),
                                  shape: BoxShape.circle,
                                ),
                                child: Icon(
                                  item['icon'] as IconData,
                                  color: item['color'] as Color,
                                  size: 20,
                                ),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        Expanded(
                                          child: Text(
                                            item['title'] as String,
                                            style: const TextStyle(
                                              color: Colors.white,
                                              fontSize: 14,
                                              fontWeight: FontWeight.bold,
                                            ),
                                          ),
                                        ),
                                        if (isUnread)
                                          Container(
                                            width: 8,
                                            height: 8,
                                            decoration: const BoxDecoration(
                                              color: Color(0xffef4444),
                                              shape: BoxShape.circle,
                                            ),
                                          ),
                                      ],
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      item['desc'] as String,
                                      style: TextStyle(
                                        color: Colors.white.withValues(
                                          alpha: 0.85,
                                        ),
                                        fontSize: 12,
                                      ),
                                    ),
                                    const SizedBox(height: 6),
                                    Text(
                                      item['time'] as String,
                                      style: TextStyle(
                                        color: Colors.white.withValues(
                                          alpha: 0.5,
                                        ),
                                        fontSize: 10,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        );
                      },
                    ),
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  @override
  void dispose() {
    _entranceController.dispose();
    _pulseController.dispose();
    widget.viewModel.removeListener(_onViewModelChanged);
    email.dispose();
    password.dispose();
    name.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    body: Stack(
      children: [
        Positioned.fill(
          child: AnimatedBuilder(
            animation: _bgScale,
            builder: (context, child) {
              return Transform.scale(
                scale: _bgScale.value,
                child: Image.asset(
                  'assets/images/login_bg.png',
                  fit: BoxFit.cover,
                  alignment: Alignment.center,
                ),
              );
            },
          ),
        ),
        Positioned.fill(
          child: Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [
                  Color(0x220f172a),
                  Color(0x331a052e),
                  Color(0x55000000),
                ],
              ),
            ),
          ),
        ),
        SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              final wide = constraints.maxWidth >= 760;
              final form = _LoginForm(
                registering: registering,
                busy: widget.viewModel.isBusy,
                error: widget.viewModel.error,
                email: email,
                password: password,
                name: name,
                obscurePassword: obscurePassword,
                lang: currentLang,
                onTogglePassword: () =>
                    setState(() => obscurePassword = !obscurePassword),
                onSubmit: submit,
                onSwitch: () {
                  setState(() => registering = !registering);
                  widget.viewModel.clearError();
                },
              );

              if (wide) {
                return Center(
                  child: SingleChildScrollView(
                    child: ConstrainedBox(
                      constraints: const BoxConstraints(maxWidth: 400),
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 12,
                        ),
                        child: Column(
                          children: [
                            FadeTransition(
                              opacity: _headerFade,
                              child: SlideTransition(
                                position: _headerSlide,
                                child: _LoginHeader(
                                  lang: currentLang,
                                  hasUnread: hasUnreadNotifications,
                                  onSelectLang: _showLanguageSelector,
                                  onOpenNotifications: _showNotificationSheet,
                                  onOpenSettings: _showAppSettingsSheet,
                                ),
                              ),
                            ),
                            const SizedBox(height: 12),
                            FadeTransition(
                              opacity: _formFade,
                              child: SlideTransition(
                                position: _formSlide,
                                child: form,
                              ),
                            ),
                            const SizedBox(height: 12),
                            FadeTransition(
                              opacity: _bottomNavFade,
                              child: SlideTransition(
                                position: _bottomNavSlide,
                                child: _LoginBottomNav(lang: currentLang),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                );
              }

              return Column(
                children: [
                  FadeTransition(
                    opacity: _headerFade,
                    child: SlideTransition(
                      position: _headerSlide,
                      child: _LoginHeader(
                        lang: currentLang,
                        hasUnread: hasUnreadNotifications,
                        onSelectLang: _showLanguageSelector,
                        onOpenNotifications: _showNotificationSheet,
                        onOpenSettings: _showAppSettingsSheet,
                      ),
                    ),
                  ),
                  Expanded(
                    child: LayoutBuilder(
                      builder: (context, constraints) {
                        return SingleChildScrollView(
                          physics: const BouncingScrollPhysics(),
                          padding: const EdgeInsets.symmetric(horizontal: 14),
                          child: ConstrainedBox(
                            constraints: BoxConstraints(
                              minHeight: constraints.maxHeight,
                            ),
                            child: Center(
                              child: Padding(
                                padding: const EdgeInsets.symmetric(
                                  vertical: 8,
                                ),
                                child: FadeTransition(
                                  opacity: _formFade,
                                  child: SlideTransition(
                                    position: _formSlide,
                                    child: form,
                                  ),
                                ),
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                  FadeTransition(
                    opacity: _bottomNavFade,
                    child: SlideTransition(
                      position: _bottomNavSlide,
                      child: _LoginBottomNav(lang: currentLang),
                    ),
                  ),
                ],
              );
            },
          ),
        ),
      ],
    ),
  );
}

class _LoginHeader extends StatelessWidget {
  const _LoginHeader({
    required this.lang,
    required this.hasUnread,
    required this.onSelectLang,
    required this.onOpenNotifications,
    this.onOpenSettings,
  });

  final String lang;
  final bool hasUnread;
  final VoidCallback onSelectLang;
  final VoidCallback onOpenNotifications;
  final VoidCallback? onOpenSettings;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 7),
        decoration: BoxDecoration(
          color: Colors.black.withValues(alpha: 0.25),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: Colors.white.withValues(alpha: 0.3)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.15),
              blurRadius: 14,
            ),
          ],
        ),
        child: Row(
          children: [
            const _BrandMark(size: 28),
            const SizedBox(width: 8),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Row(
                  children: [
                    const Text(
                      'Finora',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                        letterSpacing: -0.5,
                      ),
                    ),
                    const SizedBox(width: 5),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 5,
                        vertical: 1.5,
                      ),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [Color(0xfffbbf24), Color(0xffd97706)],
                        ),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: const Text(
                        'WEALTH OS',
                        style: TextStyle(
                          color: Color(0xff1c1917),
                          fontSize: 8,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 0.6,
                        ),
                      ),
                    ),
                  ],
                ),
                Text(
                  _I18n.t(lang, 'subtitle'),
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.9),
                    fontSize: 9,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
            const Spacer(),
            // Interactive Language Selector
            InkWell(
              onTap: onSelectLang,
              borderRadius: BorderRadius.circular(16),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(
                    color: Colors.white.withValues(alpha: 0.35),
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      _I18n.getFlag(lang),
                      style: const TextStyle(fontSize: 11.5),
                    ),
                    const SizedBox(width: 3),
                    Text(
                      lang,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(width: 2),
                    const Icon(
                      Icons.keyboard_arrow_down_rounded,
                      color: Colors.white,
                      size: 14,
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(width: 6),
            // Interactive Notification Bell
            InkWell(
              onTap: onOpenNotifications,
              borderRadius: BorderRadius.circular(99),
              child: Stack(
                children: [
                  Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.2),
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: Colors.white.withValues(alpha: 0.35),
                      ),
                    ),
                    child: const Icon(
                      Icons.notifications_none_rounded,
                      color: Colors.white,
                      size: 15,
                    ),
                  ),
                  if (hasUnread)
                    Positioned(
                      right: 2,
                      top: 2,
                      child: Container(
                        width: 7,
                        height: 7,
                        decoration: const BoxDecoration(
                          color: Color(0xffef4444),
                          shape: BoxShape.circle,
                        ),
                      ),
                    ),
                ],
              ),
            ),
            if (onOpenSettings != null) ...[
              const SizedBox(width: 6),
              InkWell(
                onTap: onOpenSettings,
                borderRadius: BorderRadius.circular(99),
                child: Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.2),
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.35),
                    ),
                  ),
                  child: const Icon(
                    Icons.settings_rounded,
                    color: Colors.white,
                    size: 15,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _LoginForm extends StatelessWidget {
  const _LoginForm({
    required this.registering,
    required this.busy,
    required this.error,
    required this.email,
    required this.password,
    required this.name,
    required this.obscurePassword,
    required this.lang,
    required this.onTogglePassword,
    required this.onSubmit,
    required this.onSwitch,
  });

  final bool registering, busy, obscurePassword;
  final String? error;
  final TextEditingController email, password, name;
  final String lang;
  final VoidCallback onTogglePassword, onSubmit, onSwitch;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.black.withValues(alpha: 0.32),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.32),
          width: 1.2,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.12),
            blurRadius: 24,
            spreadRadius: 0,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            _I18n.t(lang, 'greeting'),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              color: Colors.white.withValues(alpha: 0.9),
                              fontSize: 11.5,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                        const SizedBox(width: 3),
                        const Text('👋', style: TextStyle(fontSize: 11.5)),
                      ],
                    ),
                    const SizedBox(height: 1),
                    ShaderMask(
                      shaderCallback: (bounds) => const LinearGradient(
                        colors: [Colors.white, Color(0xfffbbf24)],
                      ).createShader(bounds),
                      child: Text(
                        registering
                            ? _I18n.t(lang, 'newAccount')
                            : (email.text.isNotEmpty
                                  ? email.text.split('@').first.toUpperCase()
                                  : 'THAN HOANG Z'),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 17,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 0.4,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.18),
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: Colors.white.withValues(alpha: 0.3),
                  ),
                ),
                child: IconButton(
                  constraints: const BoxConstraints(
                    minWidth: 34,
                    minHeight: 34,
                  ),
                  padding: EdgeInsets.zero,
                  onPressed: onSwitch,
                  tooltip: registering ? 'Quay lại đăng nhập' : 'Đổi tài khoản',
                  icon: Icon(
                    registering
                        ? Icons.login_rounded
                        : Icons.account_circle_outlined,
                    color: const Color(0xfffbbf24),
                    size: 18,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),

          AnimatedSize(
            duration: const Duration(milliseconds: 350),
            curve: Curves.easeInOutCubic,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (registering) ...[
                  _CustomGlassTextField(
                    controller: name,
                    labelText: _I18n.t(lang, 'nameLabel'),
                    icon: Icons.person_outline_rounded,
                    textCapitalization: TextCapitalization.words,
                  ),
                  const SizedBox(height: 8),
                ],
                _CustomGlassTextField(
                  controller: email,
                  labelText: _I18n.t(lang, 'emailLabel'),
                  icon: Icons.alternate_email_rounded,
                  keyboardType: TextInputType.emailAddress,
                ),
                const SizedBox(height: 8),
                _CustomGlassTextField(
                  controller: password,
                  labelText: _I18n.t(lang, 'passLabel'),
                  icon: Icons.lock_outline_rounded,
                  obscureText: obscurePassword,
                  suffixIcon: IconButton(
                    onPressed: onTogglePassword,
                    icon: Icon(
                      obscurePassword
                          ? Icons.visibility_outlined
                          : Icons.visibility_off_outlined,
                      color: Colors.white.withValues(alpha: 0.85),
                      size: 18,
                    ),
                  ),
                ),
              ],
            ),
          ),

          if (error != null)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 11,
                  vertical: 7,
                ),
                decoration: BoxDecoration(
                  color: const Color(0x66ef4444),
                  borderRadius: BorderRadius.circular(11),
                  border: Border.all(color: const Color(0xfffca5a5)),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.error_outline_rounded,
                      color: Color(0xfffca5a5),
                      size: 16,
                    ),
                    const SizedBox(width: 6),
                    Expanded(
                      child: Text(
                        error!,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

          const SizedBox(height: 12),

          _AnimatedGoldButton(
            busy: busy,
            label: registering
                ? _I18n.t(lang, 'registerBtn')
                : _I18n.t(lang, 'loginBtn'),
            onTap: onSubmit,
          ),

          const SizedBox(height: 2),
          Align(
            alignment: Alignment.center,
            child: TextButton(
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 4,
                ),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
              onPressed: onSwitch,
              child: Text(
                registering
                    ? _I18n.t(lang, 'switchLogin')
                    : _I18n.t(lang, 'switchRegister'),
                style: const TextStyle(
                  color: Color(0xfffbbf24),
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ),

          const Divider(color: Colors.white24, height: 14),

          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _CardQuickAction(
                icon: Icons.account_balance_wallet_outlined,
                label: _I18n.t(lang, 'netWorth'),
              ),
              _CardQuickAction(
                icon: Icons.smart_toy_outlined,
                label: _I18n.t(lang, 'aiAssistant'),
              ),
              _CardQuickAction(
                icon: Icons.lock_person_outlined,
                label: _I18n.t(lang, 'security'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _CustomGlassTextField extends StatelessWidget {
  const _CustomGlassTextField({
    required this.controller,
    required this.labelText,
    required this.icon,
    this.obscureText = false,
    this.keyboardType,
    this.textCapitalization = TextCapitalization.none,
    this.suffixIcon,
    this.onChanged,
  });

  final TextEditingController controller;
  final String labelText;
  final IconData icon;
  final bool obscureText;
  final TextInputType? keyboardType;
  final TextCapitalization textCapitalization;
  final Widget? suffixIcon;
  final ValueChanged<String>? onChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 2, bottom: 4),
          child: Text(
            labelText,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.9),
              fontSize: 11.5,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.2,
            ),
          ),
        ),
        TextField(
          controller: controller,
          obscureText: obscureText,
          keyboardType: keyboardType,
          textCapitalization: textCapitalization,
          onChanged: onChanged,
          style: const TextStyle(
            color: Color(0xff0f172a),
            fontSize: 13,
            fontWeight: FontWeight.w700,
          ),
          decoration: InputDecoration(
            hintText: labelText,
            hintStyle: const TextStyle(
              color: Color(0xff94a3b8),
              fontSize: 12,
              fontWeight: FontWeight.w500,
            ),
            filled: true,
            fillColor: const Color(0xfff8fafc),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 11,
            ),
            prefixIcon: Icon(icon, color: const Color(0xffd97706), size: 18),
            suffixIcon: suffixIcon,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: Color(0xffcbd5e1),
                width: 1.2,
              ),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: Color(0xffcbd5e1),
                width: 1.2,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(
                color: Color(0xfffbbf24),
                width: 1.8,
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _CardQuickAction extends StatelessWidget {
  const _CardQuickAction({required this.icon, required this.label});
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(7),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.18),
            shape: BoxShape.circle,
            border: Border.all(color: Colors.white.withValues(alpha: 0.3)),
          ),
          child: Icon(icon, color: const Color(0xfffbbf24), size: 16),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 10,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }
}

class _AnimatedGoldButton extends StatefulWidget {
  const _AnimatedGoldButton({
    required this.busy,
    required this.label,
    required this.onTap,
  });

  final bool busy;
  final String label;
  final VoidCallback? onTap;

  @override
  State<_AnimatedGoldButton> createState() => _AnimatedGoldButtonState();
}

class _AnimatedGoldButtonState extends State<_AnimatedGoldButton> {
  bool _pressed = false;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTapDown: (_) => setState(() => _pressed = true),
      onTapUp: (_) => setState(() => _pressed = false),
      onTapCancel: () => setState(() => _pressed = false),
      child: AnimatedScale(
        scale: _pressed ? 0.96 : 1.0,
        duration: const Duration(milliseconds: 120),
        curve: Curves.easeOutCubic,
        child: Container(
          height: 44,
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [Color(0xfffbbf24), Color(0xffd97706)],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(22),
            boxShadow: [
              BoxShadow(
                color: const Color(
                  0xffd97706,
                ).withValues(alpha: _pressed ? 0.25 : 0.45),
                blurRadius: _pressed ? 6 : 14,
                offset: _pressed ? const Offset(0, 2) : const Offset(0, 4),
              ),
            ],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              borderRadius: BorderRadius.circular(22),
              onTap: widget.busy ? null : widget.onTap,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  if (widget.busy)
                    const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(
                        color: Color(0xff1c1917),
                        strokeWidth: 2.2,
                      ),
                    )
                  else ...[
                    Container(
                      padding: const EdgeInsets.all(3),
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.15),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.arrow_forward_rounded,
                        color: Color(0xff1c1917),
                        size: 16,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      widget.label,
                      style: const TextStyle(
                        color: Color(0xff1c1917),
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                        letterSpacing: 0.3,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _LoginBottomNav extends StatelessWidget {
  const _LoginBottomNav({required this.lang});
  final String lang;

  @override
  Widget build(BuildContext context) {
    final items = [
      _BottomNavItem(
        icon: Icons.receipt_long_outlined,
        label: _I18n.t(lang, 'logExpense'),
      ),
      _BottomNavItem(
        icon: Icons.show_chart_rounded,
        label: _I18n.t(lang, 'portfolio'),
      ),
      _BottomNavItem(
        icon: Icons.pie_chart_outline_rounded,
        label: _I18n.t(lang, 'budget'),
      ),
      _BottomNavItem(
        icon: Icons.auto_graph_rounded,
        label: _I18n.t(lang, 'forecast'),
      ),
      _BottomNavItem(
        icon: Icons.psychology_outlined,
        label: _I18n.t(lang, 'aiAssistant'),
      ),
    ];

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 8),
        decoration: BoxDecoration(
          color: Colors.black.withValues(alpha: 0.28),
          borderRadius: BorderRadius.circular(28),
          border: Border.all(color: Colors.white.withValues(alpha: 0.3)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.15),
              blurRadius: 16,
            ),
          ],
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceAround,
          children: items,
        ),
      ),
    );
  }
}

class _BottomNavItem extends StatelessWidget {
  const _BottomNavItem({required this.icon, required this.label});
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.12),
            shape: BoxShape.circle,
            border: Border.all(color: Colors.white.withValues(alpha: 0.18)),
          ),
          child: Icon(icon, color: Colors.white, size: 20),
        ),
        const SizedBox(height: 6),
        Text(
          label,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 11,
            fontWeight: FontWeight.w600,
          ),
        ),
      ],
    );
  }
}

class _BrandMark extends StatelessWidget {
  const _BrandMark({this.size = 38});
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(size * 0.28),
        gradient: const LinearGradient(
          colors: [Color(0xfffceabb), Color(0xffdfac40), Color(0xff996515)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0xffdfac40).withValues(alpha: 0.35),
            blurRadius: 10,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Center(
        child: Transform.rotate(
          angle: 0.785398,
          child: Container(
            width: size * 0.48,
            height: size * 0.48,
            decoration: BoxDecoration(
              border: Border.all(color: const Color(0xff2a0845), width: 2.2),
              borderRadius: BorderRadius.circular(size * 0.1),
            ),
            child: Center(
              child: Container(
                width: size * 0.22,
                height: size * 0.22,
                decoration: BoxDecoration(
                  color: const Color(0xff2a0845),
                  borderRadius: BorderRadius.circular(size * 0.04),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _BankGridItem {
  const _BankGridItem({
    required this.title,
    required this.icon,
    required this.accent,
    required this.index,
    this.badge,
  });
  final String title;
  final IconData icon;
  final Color accent;
  final int index;
  final String? badge;
}

Widget _buildBankIconTile({
  required _BankGridItem item,
  required bool isSelected,
  required VoidCallback onTap,
}) {
  return InkWell(
    onTap: onTap,
    borderRadius: BorderRadius.circular(16),
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Stack(
          clipBehavior: Clip.none,
          children: [
            AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(16),
                color: isSelected
                    ? const Color(0xfffbbf24)
                    : item.accent.withValues(alpha: 0.12),
                border: Border.all(
                  color: isSelected
                      ? const Color(0xffd97706)
                      : item.accent.withValues(alpha: 0.25),
                  width: isSelected ? 2 : 1.2,
                ),
                boxShadow: isSelected
                    ? [
                        BoxShadow(
                          color: const Color(0xfffbbf24).withValues(alpha: 0.4),
                          blurRadius: 10,
                          spreadRadius: 1,
                        ),
                      ]
                    : [],
              ),
              child: Icon(
                item.icon,
                color: isSelected ? const Color(0xff1c1917) : item.accent,
                size: 24,
              ),
            ),
            if (item.badge != null)
              Positioned(
                top: -4,
                right: -6,
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 5,
                    vertical: 1.5,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xffef4444),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.white, width: 1.5),
                  ),
                  child: Text(
                    item.badge!,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 8.5,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 6),
        SizedBox(
          width: 68,
          child: Text(
            item.title,
            textAlign: TextAlign.center,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: isSelected
                  ? const Color(0xffd97706)
                  : const Color(0xff334155),
              fontSize: 11.5,
              fontWeight: isSelected ? FontWeight.w800 : FontWeight.w700,
              height: 1.15,
            ),
          ),
        ),
      ],
    ),
  );
}

Widget _buildBankSectionCard({
  required String sectionTitle,
  required List<_BankGridItem> items,
  required int selectedIndex,
  required ValueChanged<int> onSelect,
}) {
  return Container(
    margin: const EdgeInsets.only(bottom: 16),
    padding: const EdgeInsets.fromLTRB(10, 14, 10, 16),
    clipBehavior: Clip.antiAlias,
    decoration: BoxDecoration(
      color: Colors.white.withValues(alpha: 0.94),
      borderRadius: BorderRadius.circular(26),
      border: Border.all(color: Colors.white, width: 1.5),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.05),
          blurRadius: 24,
          spreadRadius: 0,
        ),
      ],
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 8, bottom: 12),
          child: Text(
            sectionTitle,
            style: const TextStyle(
              color: Color(0xff1e293b),
              fontWeight: FontWeight.w900,
              fontSize: 14.5,
              letterSpacing: 0.2,
            ),
          ),
        ),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: items.map((item) {
            return Expanded(
              child: _buildBankIconTile(
                item: item,
                isSelected: selectedIndex == item.index,
                onTap: () => onSelect(item.index),
              ),
            );
          }).toList(),
        ),
      ],
    ),
  );
}

class HomePage extends StatefulWidget {
  const HomePage({super.key, required this.api, required this.loginBuilder});
  final ApiClient api;
  final WidgetBuilder loginBuilder;
  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int index = 0;
  int refreshCounter = 0;
  final pages = const [
    NavItem('Tổng quan', Icons.dashboard_rounded),
    NavItem('Tài khoản', Icons.account_balance_rounded),
    NavItem('Giao dịch', Icons.receipt_long_rounded),
    NavItem('Khoản vay', Icons.request_quote_rounded),
    NavItem('Tài sản', Icons.inventory_2_rounded),
    NavItem('Bất động sản', Icons.home_work_rounded),
    NavItem('Ngân sách', Icons.pie_chart_rounded),
    NavItem('Dự báo', Icons.auto_graph_rounded),
    NavItem('Danh mục', Icons.category_rounded),
    NavItem('Ngân hàng', Icons.account_balance_wallet_rounded),
    NavItem('Tự động hóa', Icons.bolt_rounded),
    NavItem('Trợ lý AI', Icons.smart_toy_rounded),
    NavItem('Nhật ký audit', Icons.history_rounded),
    NavItem('Cá nhân', Icons.person_rounded),
  ];
  @override
  Widget build(BuildContext context) {
    final compact = MediaQuery.of(context).size.width < 700;
    return Scaffold(
      extendBodyBehindAppBar: true,
      backgroundColor: Colors.transparent,
      appBar: AppBar(
        backgroundColor: Colors.black.withValues(alpha: 0.25),
        elevation: 0,
        title: Row(
          children: [
            const _BrandMark(size: 32),
            const SizedBox(width: 8),
            const Text(
              'finora',
              style: TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.w900,
                fontSize: 22,
                letterSpacing: -1,
              ),
            ),
            const SizedBox(width: 6),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                gradient: FinoraColors.goldGradient,
                borderRadius: BorderRadius.circular(4),
              ),
              child: const Text(
                'WEALTH OS',
                style: TextStyle(
                  color: Color(0xff1c1917),
                  fontSize: 9,
                  fontWeight: FontWeight.w900,
                  letterSpacing: 0.8,
                ),
              ),
            ),
          ],
        ),
        actions: const [],
      ),
      drawer: null,
      body: Stack(
        children: [
          Positioned.fill(
            child: Image.asset(
              'assets/images/app_bg_maple_light.png',
              fit: BoxFit.cover,
              alignment: Alignment.center,
            ),
          ),
          Positioned.fill(
            child: Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  colors: [
                    Color(0x220f172a),
                    Color(0x331a052e),
                    Color(0x55000000),
                  ],
                ),
              ),
            ),
          ),
          SafeArea(
            child: Row(
              children: [
                if (!compact)
                  SizedBox(
                    width: 255,
                    child: Material(
                      color: Colors.black.withValues(alpha: 0.35),
                      child: _nav(),
                    ),
                  ),
                Expanded(child: _body()),
              ],
            ),
          ),
        ],
      ),
      floatingActionButton: compact
          ? Container(
              height: 56,
              width: 56,
              decoration: const BoxDecoration(
                shape: BoxShape.circle,
                gradient: FinoraColors.goldGradient,
                boxShadow: [
                  BoxShadow(
                    color: Color(0x66fbbf24),
                    blurRadius: 16,
                    spreadRadius: 2,
                    offset: Offset(0, 4),
                  ),
                ],
              ),
              child: FloatingActionButton(
                elevation: 0,
                backgroundColor: Colors.transparent,
                onPressed: () {
                  showModalBottomSheet(
                    isScrollControlled: true,
                    backgroundColor: Colors.transparent,
                    context: context,
                    builder: (_) => _TransactionFormSheet(
                      api: widget.api,
                      onSuccess: () => setState(() => refreshCounter++),
                    ),
                  );
                },
                child: const Icon(
                  Icons.add_rounded,
                  color: Color(0xff1c1917),
                  size: 28,
                ),
              ),
            )
          : null,
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: compact
          ? BottomAppBar(
              color: const Color(0xee120320),
              shape: const CircularNotchedRectangle(),
              notchMargin: 6,
              child: SizedBox(
                height: 58,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    _buildBottomNavItem(
                      icon: Icons.dashboard_rounded,
                      label: 'Trang chủ',
                      isSelected: index == 0,
                      onTap: () => setState(() => index = 0),
                    ),
                    _buildBottomNavItem(
                      icon: Icons.account_balance_wallet_rounded,
                      label: 'Tài khoản',
                      isSelected: index == 1,
                      onTap: () => setState(() => index = 1),
                    ),
                    const SizedBox(width: 48),
                    _buildBottomNavItem(
                      icon: Icons.receipt_long_rounded,
                      label: 'Giao dịch',
                      isSelected: index == 2,
                      onTap: () => setState(() => index = 2),
                    ),
                    _buildBottomNavItem(
                      icon: Icons.person_rounded,
                      label: 'Cá nhân',
                      isSelected: index == 13,
                      onTap: () => setState(() => index = 13),
                    ),
                  ],
                ),
              ),
            )
          : null,
    );
  }

  void _showSettingsModal() {
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xff200733),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Container(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(
                        Icons.tune_rounded,
                        color: Color(0xfffbbf24),
                        size: 22,
                      ),
                      const SizedBox(width: 10),
                      const Text(
                        'Cấu Hình Chế Độ Số Tiền',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const Spacer(),
                      IconButton(
                        onPressed: () => Navigator.pop(context),
                        icon: const Icon(
                          Icons.close_rounded,
                          color: Colors.white70,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Chọn cách hiển thị số tiền trên toàn bộ ứng dụng:',
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.7),
                      fontSize: 13,
                    ),
                  ),
                  const SizedBox(height: 16),
                  _buildDisplayModeTile(
                    title: 'Viết tắt / Rút gọn (100)',
                    subtitle: 'Ví dụ: 100.000 VND hiển thị thành 100',
                    value: 'compact',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 10),
                  _buildDisplayModeTile(
                    title: 'Đầy đủ (100.000 VND)',
                    subtitle: 'Hiển thị nguyên văn giá trị gốc đầy đủ',
                    value: 'full',
                    setModalState: setModalState,
                  ),
                  const SizedBox(height: 16),
                ],
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildDisplayModeTile({
    required String title,
    required String subtitle,
    required String value,
    required StateSetter setModalState,
  }) {
    final isSelected = appAmountDisplayMode == value;
    return InkWell(
      onTap: () {
        setState(() {
          appAmountDisplayMode = value;
        });
        setModalState(() {});
        try {
          widget.api.request('PUT', '/user/settings', {
            'amountDisplayMode': value,
          });
        } catch (_) {}
      },
      borderRadius: BorderRadius.circular(16),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 250),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: isSelected
              ? const Color(0x33fbbf24)
              : Colors.white.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected
                ? const Color(0xfffbbf24)
                : Colors.white.withValues(alpha: 0.12),
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      color: isSelected
                          ? const Color(0xfffbbf24)
                          : Colors.white,
                      fontSize: 15,
                      fontWeight: isSelected
                          ? FontWeight.bold
                          : FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: TextStyle(
                      color: Colors.white.withValues(alpha: 0.6),
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            if (isSelected)
              const Icon(
                Icons.check_circle_rounded,
                color: Color(0xfffbbf24),
                size: 22,
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildBottomNavItem({
    required IconData icon,
    required String label,
    required bool isSelected,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            color: isSelected ? const Color(0xfffbbf24) : Colors.white60,
            size: 22,
          ),
          const SizedBox(height: 3),
          Text(
            label,
            style: TextStyle(
              color: isSelected ? const Color(0xfffbbf24) : Colors.white60,
              fontSize: 10.5,
              fontWeight: isSelected ? FontWeight.w800 : FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  Widget _nav() => ListView(
    padding: const EdgeInsets.fromLTRB(14, 80, 14, 20),
    children: [
      const Padding(
        padding: EdgeInsets.fromLTRB(6, 4, 6, 16),
        child: Row(
          children: [
            Icon(
              Icons.account_circle_rounded,
              color: Color(0xfffbbf24),
              size: 18,
            ),
            SizedBox(width: 8),
            Text(
              'DANH MỤC TIỆN ÍCH',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w900,
                letterSpacing: 1.2,
                color: Color(0xfffbbf24),
              ),
            ),
          ],
        ),
      ),
      _buildBankSectionCard(
        sectionTitle: '👑 Quản lý Tài sản & Dòng tiền',
        items: const [
          _BankGridItem(
            title: 'Tài khoản',
            icon: Icons.account_balance_wallet_rounded,
            accent: Color(0xff38bdf8),
            index: 1,
          ),
          _BankGridItem(
            title: 'Giao dịch',
            icon: Icons.receipt_long_rounded,
            accent: Color(0xff4ade80),
            index: 2,
          ),
          _BankGridItem(
            title: 'Hạn mức',
            icon: Icons.pie_chart_outline_rounded,
            accent: Color(0xfff43f5e),
            index: 6,
          ),
          _BankGridItem(
            title: 'Danh mục',
            icon: Icons.cases_rounded,
            accent: Color(0xffa855f7),
            index: 8,
          ),
        ],
        selectedIndex: index,
        onSelect: (i) {
          setState(() => index = i);
          if (MediaQuery.of(context).size.width < 700) {
            Navigator.pop(context);
          }
        },
      ),
      _buildBankSectionCard(
        sectionTitle: '🔥 Đầu tư & Bất động sản',
        items: const [
          _BankGridItem(
            title: 'Bất động sản',
            icon: Icons.home_work_rounded,
            accent: Color(0xfff43f5e),
            index: 5,
          ),
          _BankGridItem(
            title: 'Tài sản quý',
            icon: Icons.diamond_rounded,
            accent: Color(0xfffbbf24),
            index: 4,
          ),
          _BankGridItem(
            title: 'Khoản vay',
            icon: Icons.request_quote_rounded,
            accent: Color(0xffef4444),
            index: 3,
          ),
          _BankGridItem(
            title: 'Ngân hàng',
            icon: Icons.account_balance_rounded,
            accent: Color(0xfffbbf24),
            badge: 'Mới',
            index: 9,
          ),
        ],
        selectedIndex: index,
        onSelect: (i) {
          setState(() => index = i);
          if (MediaQuery.of(context).size.width < 700) {
            Navigator.pop(context);
          }
        },
      ),
      _buildBankSectionCard(
        sectionTitle: '⚡ Tiện ích & Automation',
        items: const [
          _BankGridItem(
            title: 'Tự động hóa',
            icon: Icons.auto_fix_high_rounded,
            accent: Color(0xff38bdf8),
            index: 10,
          ),
          _BankGridItem(
            title: 'Dự báo',
            icon: Icons.trending_up_rounded,
            accent: Color(0xff4ade80),
            index: 7,
          ),
          _BankGridItem(
            title: 'Trợ lý AI',
            icon: Icons.smart_toy_rounded,
            accent: Color(0xffa855f7),
            index: 11,
          ),
          _BankGridItem(
            title: 'Nhật ký',
            icon: Icons.history_toggle_off_rounded,
            accent: Color(0xffe2e8f0),
            index: 12,
          ),
        ],
        selectedIndex: index,
        onSelect: (i) {
          setState(() => index = i);
          if (MediaQuery.of(context).size.width < 700) {
            Navigator.pop(context);
          }
        },
      ),
    ],
  );

  Widget _body() {
    switch (index) {
      case 0:
        return DashboardPage(
          key: ValueKey('dash_$refreshCounter'),
          api: widget.api,
          onNavigate: (i) => setState(() => index = i),
        );
      case 1:
        return ResourcePage(
          key: ValueKey('acc_$refreshCounter'),
          api: widget.api,
          title: 'Tài khoản',
          path: '/accounts',
          fields: const [
            FieldSpec('name', 'Tên tài khoản'),
            FieldSpec('type', 'Loại', initial: 'cash'),
            FieldSpec('currency', 'Tiền tệ', initial: 'VND'),
            FieldSpec('portfolioId', 'Portfolio ID'),
          ],
        );
      case 2:
        return TransactionsPage(
          key: ValueKey('tx_$refreshCounter'),
          api: widget.api,
        );
      case 3:
        return ResourcePage(
          api: widget.api,
          title: 'Khoản vay',
          path: '/loans',
          fields: const [
            FieldSpec(
              'direction',
              'Loại (receivable/payable)',
              initial: 'receivable',
            ),
            FieldSpec('counterparty', 'Đối tác'),
            FieldSpec('principalInitial', 'Gốc ban đầu'),
            FieldSpec('annualRate', 'Lãi suất năm', initial: '0'),
            FieldSpec('status', 'Trạng thái', initial: 'active'),
          ],
        );
      case 4:
        return ValuedResourcePage(
          api: widget.api,
          title: 'Tài sản',
          path: '/assets',
          fields: const [
            FieldSpec('name', 'Tên tài sản'),
            FieldSpec('type', 'Loại', initial: 'other'),
            FieldSpec('currency', 'Tiền tệ', initial: 'VND'),
          ],
        );
      case 5:
        return ValuedResourcePage(
          api: widget.api,
          title: 'Bất động sản',
          path: '/properties',
          fields: const [
            FieldSpec('name', 'Tên tài sản'),
            FieldSpec('currency', 'Tiền tệ', initial: 'VND'),
            FieldSpec('address', 'Địa chỉ'),
          ],
        );
      case 6:
        return BudgetPage(api: widget.api);
      case 7:
        return ForecastPage(api: widget.api);
      case 8:
        return ResourcePage(
          api: widget.api,
          title: 'Danh mục',
          path: '/portfolios',
          fields: const [
            FieldSpec('name', 'Tên danh mục'),
            FieldSpec('baseCurrency', 'Tiền tệ cơ sở', initial: 'VND'),
          ],
        );
      case 9:
        return BankPage(api: widget.api);
      case 10:
        return AutomationPage(api: widget.api);
      case 11:
        return AssistantPage(api: widget.api);
      case 13:
        return ProfilePage(
          api: widget.api,
          onOpenSettings: _showSettingsModal,
          onLogout: _logout,
        );
      default:
        return ReadonlyPage(
          api: widget.api,
          title: 'Nhật ký audit',
          path: '/audit-logs',
        );
    }
  }

  void _logout() {
    widget.api.token = null;
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(builder: widget.loginBuilder),
      (_) => false,
    );
  }
}

class PageFrame extends StatelessWidget {
  const PageFrame({
    super.key,
    required this.title,
    required this.child,
    this.action,
  });
  final String title;
  final Widget child;
  final Widget? action;

  @override
  Widget build(BuildContext context) => SafeArea(
    child: Padding(
      padding: const EdgeInsets.fromLTRB(18, 70, 18, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'FINORA / QUẢN LÝ TÀI SẢN',
                      style: TextStyle(
                        color: Color(0xfff7d070),
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                        letterSpacing: 1.2,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      title,
                      style: const TextStyle(
                        fontSize: 26,
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                        letterSpacing: -0.5,
                      ),
                    ),
                  ],
                ),
              ),
              if (action case final Widget action) action,
            ],
          ),
          const SizedBox(height: 16),
          Expanded(child: child),
        ],
      ),
    ),
  );
}

class DashboardPage extends StatefulWidget {
  const DashboardPage({super.key, required this.api, this.onNavigate});
  final ApiClient api;
  final ValueChanged<int>? onNavigate;

  @override
  State<DashboardPage> createState() => _DashboardPageState();
}

class _DashboardPageState extends State<DashboardPage> {
  Map? netWorth;
  List accounts = [], transactions = [], portfolios = [];
  String? error;
  bool loading = true;

  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    setState(() {
      loading = true;
      error = null;
    });
    try {
      portfolios = await widget.api.request('GET', '/portfolios') as List;
      accounts = await widget.api.request('GET', '/accounts') as List;
      final tx = await widget.api.request('GET', '/transactions?limit=5');
      transactions = (tx as Map)['items'] as List? ?? [];
      if (portfolios.isNotEmpty) {
        netWorth =
            await widget.api.request(
                  'GET',
                  '/portfolios/${portfolios.first['id']}/net-worth',
                )
                as Map;
      }
    } catch (e) {
      error = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Widget _buildQuickActionsRow() {
    return Padding(
      padding: const EdgeInsets.only(top: 14, bottom: 6),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: [
          _buildCircularQuickAction(
            icon: Icons.add_circle_outline_rounded,
            label: 'Tạo GD',
            color: const Color(0xfffbbf24),
            onTap: () {
              showModalBottomSheet(
                isScrollControlled: true,
                backgroundColor: Colors.transparent,
                context: context,
                builder: (_) => _TransactionFormSheet(
                  api: widget.api,
                  onSuccess: () => load(),
                ),
              );
            },
          ),
          _buildCircularQuickAction(
            icon: Icons.account_balance_wallet_rounded,
            label: 'Tài khoản',
            color: const Color(0xff38bdf8),
            onTap: () => widget.onNavigate?.call(1),
          ),
          _buildCircularQuickAction(
            icon: Icons.history_toggle_off_rounded,
            label: 'Nhật ký',
            color: const Color(0xffc084fc),
            onTap: () => widget.onNavigate?.call(12),
          ),
          _buildCircularQuickAction(
            icon: Icons.auto_fix_high_rounded,
            label: 'Tự động',
            color: const Color(0xff4ade80),
            onTap: () => widget.onNavigate?.call(10),
          ),
        ],
      ),
    );
  }

  Widget _buildCircularQuickAction({
    required IconData icon,
    required String label,
    required Color color,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 52,
            height: 52,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: Colors.white.withValues(alpha: 0.94),
              border: Border.all(color: Colors.white, width: 1.5),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.06),
                  blurRadius: 14,
                ),
              ],
            ),
            child: Icon(icon, color: color, size: 24),
          ),
          const SizedBox(height: 6),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 11.5,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSuggestionsSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Padding(
          padding: EdgeInsets.only(left: 4, bottom: 10),
          child: Text(
            'Gợi ý cho bạn',
            style: TextStyle(
              color: Colors.white,
              fontWeight: FontWeight.w900,
              fontSize: 16,
            ),
          ),
        ),
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: [
              _buildSuggestionCard(
                icon: Icons.military_tech_rounded,
                title: 'Thử thách',
                subtitle: 'Tích điểm đổi quà',
                accent: const Color(0xff00dbe7),
              ),
              const SizedBox(width: 12),
              _buildSuggestionCard(
                icon: Icons.star_rate_rounded,
                title: 'Số đẹp',
                subtitle: 'Khẳng định vị thế',
                accent: const Color(0xffeb6a1b),
              ),
              const SizedBox(width: 12),
              _buildSuggestionCard(
                icon: Icons.trending_up_rounded,
                title: 'Sinh lời kép',
                subtitle: 'Tối ưu dòng vốn',
                accent: const Color(0xffe2b5ff),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildSuggestionCard({
    required IconData icon,
    required String title,
    required String subtitle,
    required Color accent,
  }) {
    return Container(
      width: 155,
      padding: const EdgeInsets.all(14),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.94),
        borderRadius: BorderRadius.circular(22),
        border: Border(
          left: BorderSide(color: accent, width: 4),
          top: const BorderSide(color: Colors.white, width: 1.2),
          right: const BorderSide(color: Colors.white, width: 1.2),
          bottom: const BorderSide(color: Colors.white, width: 1.2),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 20,
            spreadRadius: 0,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: accent, size: 22),
          const SizedBox(height: 8),
          Text(
            title,
            style: const TextStyle(
              color: Color(0xff0f172a),
              fontWeight: FontWeight.w900,
              fontSize: 14,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            subtitle,
            style: const TextStyle(
              color: Color(0xff64748b),
              fontWeight: FontWeight.w600,
              fontSize: 11,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPromotionBanner() {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 16),
      height: 110,
      width: double.infinity,
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(26),
        gradient: const LinearGradient(
          colors: [Color(0xff4a0e72), Color(0xff8b5cf6), Color(0xffeb6a1b)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        boxShadow: [
          BoxShadow(
            color: const Color(0x338b5cf6),
            blurRadius: 24,
            spreadRadius: 0,
          ),
        ],
      ),
      child: Stack(
        children: [
          Positioned(
            right: -20,
            bottom: -20,
            child: Icon(
              Icons.account_circle_rounded,
              size: 130,
              color: Colors.white.withValues(alpha: 0.12),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(18),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xfffbbf24),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Text(
                    'OFFER VIP',
                    style: TextStyle(
                      color: Color(0xff1c1917),
                      fontSize: 9.5,
                      fontWeight: FontWeight.w900,
                      letterSpacing: 0.8,
                    ),
                  ),
                ),
                const SizedBox(height: 6),
                const Text(
                  'Đặc quyền thượng lưu cùng Finora Gold',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w900,
                    fontSize: 15.5,
                    letterSpacing: -0.2,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  'Tối ưu dòng tiền & Tích điểm thưởng VIP',
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.85),
                    fontSize: 11,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) => PageFrame(
    title: 'Tổng quan',
    action: Container(
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.12),
        shape: BoxShape.circle,
      ),
      child: IconButton(
        onPressed: load,
        icon: const Icon(Icons.refresh_rounded, color: Color(0xfffbbf24)),
      ),
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfffbbf24)),
          )
        : RefreshIndicator(
            color: const Color(0xfffbbf24),
            backgroundColor: const Color(0xff1c1917),
            onRefresh: load,
            child: ListView(
              children: [
                if (error != null) ErrorBox(error!),
                _BalanceHero(
                  value: _formatMoney(netWorth?['netWorth']),
                  currency: netWorth?['baseCurrency']?.toString() ?? 'VND',
                  accountCount: accounts.length,
                ),
                _buildQuickActionsRow(),
                const SizedBox(height: 16),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      Metric(
                        'Tiền mặt khả dụng',
                        _formatMoney(netWorth?['cash']),
                        Icons.payments_rounded,
                        accent: const Color(0xff38bdf8),
                      ),
                      const SizedBox(width: 12),
                      Metric(
                        'Nợ phải trả',
                        _formatMoney(netWorth?['liabilities']),
                        Icons.credit_card_off_rounded,
                        accent: const Color(0xfffb7185),
                      ),
                      const SizedBox(width: 12),
                      Metric(
                        'Tài khoản theo dõi',
                        accounts.length.toString(),
                        Icons.account_balance_rounded,
                        accent: const Color(0xffc084fc),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 20),
                _buildSuggestionsSection(),
                _buildPromotionBanner(),
                _buildBankSectionCard(
                  sectionTitle: '👑 Lối tắt ưu tiên',
                  items: const [
                    _BankGridItem(
                      title: 'Tài khoản',
                      icon: Icons.account_balance_wallet_rounded,
                      accent: Color(0xff38bdf8),
                      index: 1,
                    ),
                    _BankGridItem(
                      title: 'Giao dịch',
                      icon: Icons.receipt_long_rounded,
                      accent: Color(0xff4ade80),
                      index: 2,
                    ),
                    _BankGridItem(
                      title: 'Hạn mức',
                      icon: Icons.pie_chart_outline_rounded,
                      accent: Color(0xfff43f5e),
                      index: 6,
                    ),
                    _BankGridItem(
                      title: 'Danh mục',
                      icon: Icons.cases_rounded,
                      accent: Color(0xffa855f7),
                      index: 8,
                    ),
                  ],
                  selectedIndex: -1,
                  onSelect: (i) => widget.onNavigate?.call(i),
                ),
                _buildBankSectionCard(
                  sectionTitle: '🔥 Đầu tư & Bất động sản',
                  items: const [
                    _BankGridItem(
                      title: 'Bất động sản',
                      icon: Icons.home_work_rounded,
                      accent: Color(0xfff43f5e),
                      index: 5,
                    ),
                    _BankGridItem(
                      title: 'Tài sản quý',
                      icon: Icons.diamond_rounded,
                      accent: Color(0xfffbbf24),
                      index: 4,
                    ),
                    _BankGridItem(
                      title: 'Khoản vay',
                      icon: Icons.request_quote_rounded,
                      accent: Color(0xffef4444),
                      index: 3,
                    ),
                    _BankGridItem(
                      title: 'Ngân hàng',
                      icon: Icons.account_balance_rounded,
                      accent: Color(0xfffbbf24),
                      badge: 'Mới',
                      index: 9,
                    ),
                  ],
                  selectedIndex: -1,
                  onSelect: (i) => widget.onNavigate?.call(i),
                ),
                _buildBankSectionCard(
                  sectionTitle: '⚡ Tiện ích & Automation',
                  items: const [
                    _BankGridItem(
                      title: 'Tự động hóa',
                      icon: Icons.auto_fix_high_rounded,
                      accent: Color(0xff38bdf8),
                      index: 10,
                    ),
                    _BankGridItem(
                      title: 'Dự báo',
                      icon: Icons.trending_up_rounded,
                      accent: Color(0xff4ade80),
                      index: 7,
                    ),
                    _BankGridItem(
                      title: 'Trợ lý AI',
                      icon: Icons.smart_toy_rounded,
                      accent: Color(0xffa855f7),
                      index: 11,
                    ),
                    _BankGridItem(
                      title: 'Nhật ký',
                      icon: Icons.history_toggle_off_rounded,
                      accent: Color(0xffe2e8f0),
                      index: 12,
                    ),
                  ],
                  selectedIndex: -1,
                  onSelect: (i) => widget.onNavigate?.call(i),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    const Text(
                      'Giao dịch gần đây',
                      style: TextStyle(
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                        fontSize: 18,
                      ),
                    ),
                    const Spacer(),
                    TextButton(
                      onPressed: () {},
                      child: const Text(
                        'Xem tất cả',
                        style: TextStyle(
                          color: Color(0xfff7d070),
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                ...transactions.map(
                  (x) => FinoraListTile(
                    icon: x['type'] == 'income'
                        ? Icons.south_west_rounded
                        : Icons.north_east_rounded,
                    iconColor: x['type'] == 'income'
                        ? const Color(0xff4ade80)
                        : const Color(0xfffb7185),
                    title: (x['name']?.toString().trim().isNotEmpty == true)
                        ? x['name'].toString().trim()
                        : (x['note']?.toString().isNotEmpty == true
                              ? x['note'].toString()
                              : (x['type'] == 'income'
                                    ? 'Thu nhập'
                                    : 'Chi tiêu')),
                    subtitle: x['occurredAt']?.toString() ?? '',
                    amount:
                        "${x['type'] == 'income' ? '+' : '-'}${_formatMoney(x['amount'])} ${x['currency'] ?? 'VND'}",
                  ),
                ),
                if (transactions.isEmpty)
                  const EmptyState('Chưa có giao dịch gần đây'),
              ],
            ),
          ),
  );

  String _formatMoney(dynamic raw) {
    if (raw == null) return '0';
    final d = double.tryParse(raw.toString());
    if (d == null) return raw.toString();
    final value = appAmountDisplayMode == 'compact' ? (d / 1000.0) : d;
    final parts = value.toStringAsFixed(2).split('.');
    final intPart = parts[0].replaceAllRegExp(
      RegExp(r'\B(?=(\d{3})+(?!\d))'),
      ',',
    );
    return parts[1] == '00' ? intPart : '$intPart.${parts[1]}';
  }
}

extension on String {
  String replaceAllRegExp(RegExp regex, String replacement) {
    return replaceAllMapped(regex, (match) => replacement);
  }
}

class _BalanceHero extends StatefulWidget {
  const _BalanceHero({
    required this.value,
    required this.currency,
    required this.accountCount,
  });
  final String value, currency;
  final int accountCount;

  @override
  State<_BalanceHero> createState() => _BalanceHeroState();
}

class _BalanceHeroState extends State<_BalanceHero> {
  bool hideBalance = false;

  @override
  Widget build(BuildContext context) => Container(
    width: double.infinity,
    padding: const EdgeInsets.all(22),
    clipBehavior: Clip.antiAlias,
    decoration: BoxDecoration(
      borderRadius: BorderRadius.circular(28),
      gradient: const LinearGradient(
        colors: [
          Color(0xff2a0845),
          Color(0xff5b2382),
          Color(0xff7c45a4),
          Color(0xffeb6a1b),
        ],
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      ),
      border: Border.all(
        color: Colors.white.withValues(alpha: 0.35),
        width: 1.5,
      ),
      boxShadow: [
        BoxShadow(
          color: const Color(0x335b2382),
          blurRadius: 32,
          spreadRadius: 0,
        ),
      ],
    ),
    child: Stack(
      children: [
        Positioned(
          right: -25,
          bottom: -25,
          child: Icon(
            Icons.account_circle_rounded,
            size: 150,
            color: Colors.white.withValues(alpha: 0.12),
          ),
        ),
        Positioned(
          right: -10,
          top: -20,
          child: Container(
            width: 130,
            height: 130,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: const Color(0xfffbbf24).withValues(alpha: 0.15),
            ),
          ),
        ),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(6),
                  decoration: BoxDecoration(
                    color: const Color(0xfffbbf24).withValues(alpha: 0.22),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.account_balance_wallet_rounded,
                    color: Color(0xfffbbf24),
                    size: 18,
                  ),
                ),
                const SizedBox(width: 8),
                const Text(
                  'TỔNG TÀI SẢN RÒNG',
                  style: TextStyle(
                    color: Color(0xfffbbf24),
                    fontWeight: FontWeight.w900,
                    fontSize: 11.5,
                    letterSpacing: 1.2,
                  ),
                ),
                const Spacer(),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 3,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xfffbbf24),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Text(
                    'WEALTH OS VIP',
                    style: TextStyle(
                      color: Color(0xff1c1917),
                      fontSize: 9,
                      fontWeight: FontWeight.w900,
                      letterSpacing: 0.8,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              hideBalance ? '••••••••••••' : widget.value,
              style: const TextStyle(
                fontSize: 30,
                fontWeight: FontWeight.w900,
                color: Colors.white,
                letterSpacing: -0.5,
                shadows: [
                  Shadow(
                    color: Color(0x66000000),
                    blurRadius: 10,
                    offset: Offset(0, 4),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 2),
            Text(
              widget.currency,
              style: const TextStyle(
                color: Color(0xfffef08a),
                fontWeight: FontWeight.w900,
                fontSize: 13.5,
                letterSpacing: 0.5,
              ),
            ),
            const SizedBox(height: 18),
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 7,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.16),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.3),
                    ),
                  ),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.check_circle_rounded,
                        color: Color(0xff4ade80),
                        size: 14,
                      ),
                      const SizedBox(width: 6),
                      Text(
                        '${widget.accountCount} tài khoản đang theo dõi',
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 11.5,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ),
                ),
                const Spacer(),
                Container(
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.16),
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.3),
                    ),
                  ),
                  child: IconButton(
                    onPressed: () => setState(() => hideBalance = !hideBalance),
                    icon: Icon(
                      hideBalance
                          ? Icons.visibility_off_outlined
                          : Icons.visibility_outlined,
                      color: const Color(0xfffbbf24),
                      size: 20,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ],
    ),
  );
}

class Metric extends StatelessWidget {
  const Metric(
    this.label,
    this.value,
    this.icon, {
    super.key,
    required this.accent,
  });
  final String label, value;
  final IconData icon;
  final Color accent;

  @override
  Widget build(BuildContext context) => Container(
    width: 190,
    padding: const EdgeInsets.all(16),
    clipBehavior: Clip.antiAlias,
    decoration: BoxDecoration(
      color: Colors.white.withValues(alpha: 0.94),
      borderRadius: BorderRadius.circular(24),
      border: Border.all(color: Colors.white, width: 1.2),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.04),
          blurRadius: 20,
          spreadRadius: 0,
        ),
      ],
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: accent.withValues(alpha: 0.14),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(icon, color: accent, size: 20),
        ),
        const SizedBox(height: 14),
        Text(
          value,
          style: const TextStyle(
            fontWeight: FontWeight.w900,
            color: Color(0xff0f172a),
            fontSize: 17,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: const TextStyle(
            color: Color(0xff64748b),
            fontWeight: FontWeight.w600,
            fontSize: 11.5,
          ),
        ),
      ],
    ),
  );
}

class ResourcePage extends StatefulWidget {
  const ResourcePage({
    super.key,
    required this.api,
    required this.title,
    required this.path,
    required this.fields,
  });
  final ApiClient api;
  final String title, path;
  final List<FieldSpec> fields;
  @override
  State<ResourcePage> createState() => _ResourcePageState();
}

class _ResourcePageState extends State<ResourcePage> {
  late Map<String, TextEditingController> ctl;
  List items = [];
  bool loading = true;
  String? error;
  @override
  void initState() {
    super.initState();
    ctl = {
      for (final f in widget.fields)
        f.key: TextEditingController(text: f.initial),
    };
    load();
  }

  @override
  void dispose() {
    for (final x in ctl.values) {
      x.dispose();
    }
    super.dispose();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      final data = await widget.api.request('GET', widget.path);
      items = data is List ? data : ((data as Map)['items'] as List? ?? []);
      error = null;
    } catch (e) {
      error = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> create() async {
    try {
      await widget.api.request('POST', widget.path, {
        for (final x in ctl.entries)
          if (x.value.text.trim().isNotEmpty) x.key: x.value.text.trim(),
      });
      for (final f in widget.fields) {
        ctl[f.key]!.text = f.initial;
      }
      if (mounted) Navigator.pop(context);
      await load();
    } catch (e) {
      if (mounted) showError(context, e.toString());
    }
  }

  void openForm() {
    if (widget.path == '/accounts' || widget.title == 'Tài khoản') {
      showModalBottomSheet(
        isScrollControlled: true,
        backgroundColor: Colors.transparent,
        context: context,
        builder: (_) => _AccountFormSheet(
          api: widget.api,
          onSuccess: () {
            Navigator.pop(context);
            load();
          },
        ),
      );
      return;
    }

    showModalBottomSheet(
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      context: context,
      builder: (_) => FinoraSheet(
        title: 'Thêm ${widget.title}',
        subtitle: 'Thông tin sẽ được lưu vào user hiện tại.',
        child: Column(
          children: [
            ...widget.fields.map(
              (f) => Padding(
                padding: const EdgeInsets.only(bottom: 12),
                child: _CustomGlassTextField(
                  controller: ctl[f.key]!,
                  labelText: f.label,
                  icon: Icons.edit_note_rounded,
                ),
              ),
            ),
            const SizedBox(height: 8),
            _AnimatedGoldButton(
              busy: false,
              label: 'Lưu thông tin',
              onTap: create,
            ),
          ],
        ),
      ),
    );
  }

  Future<void> deleteItem(String id, String itemName) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xee120320),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(24),
          side: BorderSide(color: Colors.white.withValues(alpha: 0.3)),
        ),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: const Color(0xffef4444).withValues(alpha: 0.15),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.delete_outline_rounded,
                color: Color(0xffef4444),
                size: 22,
              ),
            ),
            const SizedBox(width: 10),
            Text(
              'Xóa ${widget.title}',
              style: const TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
        content: Text(
          'Bạn có chắc chắn muốn xóa "$itemName"?\nChỉ những tài khoản không dính dáng bất kỳ khoản nào mới có thể xóa.',
          style: TextStyle(
            color: Colors.white.withValues(alpha: 0.85),
            fontSize: 13,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Hủy', style: TextStyle(color: Colors.white70)),
          ),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: const Color(0xffef4444),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(14),
              ),
            ),
            onPressed: () => Navigator.pop(context, true),
            child: const Text(
              'Xóa ngay',
              style: TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    try {
      await widget.api.request('DELETE', '${widget.path}/$id');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            backgroundColor: const Color(0xff15803d),
            content: Text('Đã xóa "$itemName" thành công!'),
          ),
        );
      }
      await load();
    } catch (e) {
      if (mounted) {
        final rawErr = e.toString();
        final msg = rawErr.contains('ACCOUNT_HAS_TRANSACTIONS')
            ? 'Không thể xóa "$itemName" vì đang có giao dịch hoặc dòng tiền liên kết.'
            : 'Không thể xóa "$itemName": $rawErr';
        showError(context, msg);
      }
    }
  }

  @override
  Widget build(BuildContext context) => PageFrame(
    title: widget.title,
    action: Row(
      children: [
        IconButton(
          onPressed: load,
          icon: const Icon(Icons.refresh_rounded, color: Color(0xfffbbf24)),
        ),
        const SizedBox(width: 4),
        FilledButton.icon(
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xfffbbf24),
            foregroundColor: const Color(0xff1c1917),
          ),
          onPressed: openForm,
          icon: const Icon(Icons.add_rounded),
          label: const Text(
            'Thêm',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
        ),
      ],
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfffbbf24)),
          )
        : ListView(
            children: [
              const _ScreenIntro(
                'Quản lý thông tin tập trung, rõ ràng và an toàn.',
              ),
              if (error != null) ErrorBox(error!),
              if (items.isEmpty) const EmptyState('Chưa có dữ liệu'),
              ...items.map((x) {
                final id = x['id']?.toString() ?? '';
                final name =
                    x['name']?.toString() ??
                    x['counterparty']?.toString() ??
                    x['id']?.toString() ??
                    '-';
                final tile = FinoraListTile(
                  icon: _iconForTitle(widget.title),
                  title: name,
                  subtitle: _details(x),
                  badge:
                      x['status']?.toString() ??
                      x['currency']?.toString() ??
                      '',
                );

                if (id.isEmpty) return tile;

                return Dismissible(
                  key: ValueKey('res_${widget.title}_$id'),
                  direction: DismissDirection.endToStart,
                  confirmDismiss: (_) async {
                    await deleteItem(id, name);
                    return false;
                  },
                  background: Container(
                    margin: const EdgeInsets.only(bottom: 10),
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    alignment: Alignment.centerRight,
                    decoration: BoxDecoration(
                      color: const Color(0xddef4444),
                      borderRadius: BorderRadius.circular(24),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0x66ef4444),
                          blurRadius: 16,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: const Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        Icon(
                          Icons.delete_forever_rounded,
                          color: Colors.white,
                          size: 24,
                        ),
                        SizedBox(width: 6),
                        Text(
                          'Xóa',
                          style: TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.w900,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                  child: tile,
                );
              }),
            ],
          ),
  );
  String _details(dynamic x) {
    if (x is! Map) return '';

    final parts = <String>[];

    if (x.containsKey('type') && x['type'] != null) {
      final t = x['type'].toString();
      final typeLabel = switch (t) {
        'cash' => 'Tiền mặt',
        'bank' => 'Ngân hàng',
        'gold' => 'Vàng kim',
        'real_estate' => 'Bất động sản',
        'income' => 'Thu nhập',
        'expense' => 'Chi tiêu',
        _ => t,
      };
      parts.add('Loại: $typeLabel');
    }

    if (x.containsKey('balance') &&
        x['balance'] != null &&
        x['balance'].toString().isNotEmpty) {
      final b = x['balance'].toString();
      final curr = x['currency']?.toString() ?? 'VND';
      parts.add('Số dư: $b $curr');
    }

    if (x.containsKey('principal') && x['principal'] != null) {
      parts.add('Gốc: ${x['principal']} ${x['currency'] ?? 'VND'}');
    }
    if (x.containsKey('rate') && x['rate'] != null) {
      parts.add('Lãi suất: ${x['rate']}%');
    }

    final rawDate =
        x['createdAt']?.toString() ??
        x['updatedAt']?.toString() ??
        x['occurredAt']?.toString();
    final formattedDate = _formatDate(rawDate);
    if (formattedDate.isNotEmpty) {
      parts.add('Ngày: $formattedDate');
    }

    if (parts.isNotEmpty) {
      return parts.join(' • ');
    }

    const noisyKeys = {
      'id',
      'userId',
      'portfolioId',
      'createdAt',
      'updatedAt',
      'occurredAt',
      'name',
      'type',
      'currency',
      'balance',
    };

    final fallbackEntries = x.entries
        .where(
          (e) =>
              !noisyKeys.contains(e.key) &&
              e.value != null &&
              e.value.toString().isNotEmpty,
        )
        .take(2)
        .map((e) => '${e.key}: ${e.value}')
        .join(' • ');

    return fallbackEntries;
  }
}

String _formatDate(String? raw) {
  if (raw == null || raw.trim().isEmpty) return '';
  try {
    final dt = DateTime.parse(raw).toLocal();
    final day = dt.day.toString().padLeft(2, '0');
    final month = dt.month.toString().padLeft(2, '0');
    final year = dt.year;
    final hour = dt.hour.toString().padLeft(2, '0');
    final minute = dt.minute.toString().padLeft(2, '0');
    return '$day/$month/$year $hour:$minute';
  } catch (_) {
    return raw;
  }
}

class PersonalInfoPage extends StatefulWidget {
  const PersonalInfoPage({super.key, required this.api});
  final ApiClient api;

  @override
  State<PersonalInfoPage> createState() => _PersonalInfoPageState();
}

class _PersonalInfoPageState extends State<PersonalInfoPage> {
  String _email = 'hoangxuan.ks6@gmail.com';
  String _phone = '*** *** *399';
  String _address = '4C26, CAU TRE, NGO QUYEN, HAI PHONG';

  void _showEditBasicInfoModal() {
    final phoneCtrl = TextEditingController(text: _phone);
    final emailCtrl = TextEditingController(text: _email);
    final addressCtrl = TextEditingController(text: _address);
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: EdgeInsets.only(
          left: 20,
          right: 20,
          top: 20,
          bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: const Color(0xffcbd5e1),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              'Chỉnh sửa thông tin cơ bản',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: Color(0xff1e293b),
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: phoneCtrl,
              decoration: const InputDecoration(
                labelText: 'Số điện thoại',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: emailCtrl,
              decoration: const InputDecoration(
                labelText: 'Email liên hệ',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: addressCtrl,
              decoration: const InputDecoration(
                labelText: 'Địa chỉ liên hệ',
                border: OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              height: 46,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xff6b21a8),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                onPressed: () {
                  setState(() {
                    _phone = phoneCtrl.text.trim();
                    _email = emailCtrl.text.trim();
                    _address = addressCtrl.text.trim();
                  });
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text(
                        'Đã cập nhật thông tin cá nhân thành công!',
                      ),
                    ),
                  );
                },
                child: const Text(
                  'Lưu thay đổi',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showLimitChangeModal() {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                width: 36,
                height: 4,
                decoration: BoxDecoration(
                  color: const Color(0xffcbd5e1),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              'Thay đổi hạn mức giao dịch',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: Color(0xff1e293b),
              ),
            ),
            const SizedBox(height: 12),
            const Text(
              'Hạn mức tối đa hiện tại: 300,000,000 VND/ngày\nXác thực bằng Smart OTP để tăng hoặc điều chỉnh hạn mức.',
              style: TextStyle(fontSize: 13, color: Color(0xff64748b)),
            ),
            const SizedBox(height: 20),
            SizedBox(
              width: double.infinity,
              height: 46,
              child: ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xffd97706),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                onPressed: () {
                  Navigator.pop(ctx);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(
                      content: Text(
                        'Yêu cầu điều chỉnh hạn mức đã được tiếp nhận và xử lý!',
                      ),
                    ),
                  );
                },
                child: const Text(
                  'Xác nhận qua Smart OTP',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    const double dailyLimit = 300000000;
    const double usedLimit = 936207;
    const double progress = (usedLimit / dailyLimit);

    return Scaffold(
      backgroundColor: const Color(0xfff8fafc),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0.5,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_rounded, color: Color(0xff1e293b)),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          'Thông tin cá nhân',
          style: TextStyle(
            color: Color(0xff1e293b),
            fontSize: 17,
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        physics: const BouncingScrollPhysics(),
        children: [
          // Section 1: Dữ liệu sinh trắc học
          _buildSectionHeader('Dữ liệu sinh trắc học'),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 10,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Khuôn mặt',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: Color(0xff1e293b),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 10,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xffdcfce7),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Text(
                    'Đã cập nhật',
                    style: TextStyle(
                      color: Color(0xff15803d),
                      fontSize: 11.5,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Section 2: Hạn mức tài khoản
          _buildSectionHeader('Hạn mức tài khoản'),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 12,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Hạn mức hiện tại của Bạn là:',
                      style: TextStyle(
                        fontSize: 13.5,
                        fontWeight: FontWeight.w600,
                        color: Color(0xff334155),
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 3.5,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xffdcfce7),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Text(
                        'Tối đa',
                        style: TextStyle(
                          color: Color(0xff15803d),
                          fontSize: 11.5,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                const Text(
                  '300,000,000 VND/ngày và\n300,000,000 VND/lần',
                  style: TextStyle(
                    fontSize: 14.5,
                    fontWeight: FontWeight.w900,
                    color: Color(0xff6b21a8),
                    height: 1.3,
                  ),
                ),
                const SizedBox(height: 14),
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: progress,
                    minHeight: 5,
                    backgroundColor: const Color(0xfff1f5f9),
                    valueColor: const AlwaysStoppedAnimation<Color>(
                      Color(0xff8b5cf6),
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                const Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Đã giao dịch trong ngày',
                          style: TextStyle(
                            fontSize: 11,
                            color: Color(0xff64748b),
                          ),
                        ),
                        SizedBox(height: 2),
                        Text(
                          '936,207 VND',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: Color(0xff0f172a),
                          ),
                        ),
                      ],
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          'Hạn mức còn lại',
                          style: TextStyle(
                            fontSize: 11,
                            color: Color(0xff64748b),
                          ),
                        ),
                        SizedBox(height: 2),
                        Text(
                          '299,063,793 VND',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.bold,
                            color: Color(0xff0f172a),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                const SizedBox(height: 14),
                GestureDetector(
                  onTap: _showLimitChangeModal,
                  child: const Text(
                    'Thay đổi hạn mức giao dịch',
                    style: TextStyle(
                      fontSize: 12.5,
                      fontWeight: FontWeight.bold,
                      color: Color(0xffd97706),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Section 3: Thông tin cơ bản
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildSectionHeader('Thông tin cơ bản'),
              GestureDetector(
                onTap: _showEditBasicInfoModal,
                child: const Row(
                  children: [
                    Icon(
                      Icons.edit_outlined,
                      size: 14,
                      color: Color(0xffd97706),
                    ),
                    SizedBox(width: 4),
                    Text(
                      'Chỉnh sửa',
                      style: TextStyle(
                        fontSize: 12.5,
                        fontWeight: FontWeight.bold,
                        color: Color(0xffd97706),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 12,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                _buildInfoRow('Tên đăng nhập', '0857869399'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Mã KH (CIF)', '03769945', showInfoIcon: true),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Số điện thoại', _phone),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Email', _email),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Địa chỉ liên hệ', _address, isMultiLine: true),
              ],
            ),
          ),
          const SizedBox(height: 20),

          // Section 4: Thông tin định danh
          _buildSectionHeader('Thông tin định danh'),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(18),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.04),
                  blurRadius: 12,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                _buildInfoRow('Số CMND/CCCD/Hộ chiếu', '031097006012'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Họ và tên', 'NGUYEN XUAN HOANG'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Ngày sinh', '12/08/1997'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Giới tính', 'Nam'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Quốc tịch', 'Việt Nam'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow('Ngày cấp', '15/05/2021'),
                const Divider(height: 1, indent: 0, endIndent: 0),
                _buildInfoRow(
                  'Nơi cấp',
                  'Cục Cảnh sát QLHC về TTXH',
                  isMultiLine: true,
                ),
              ],
            ),
          ),
          const SizedBox(height: 30),
        ],
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Text(
      title,
      style: const TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.bold,
        color: Color(0xff1e293b),
      ),
    );
  }

  Widget _buildInfoRow(
    String label,
    String value, {
    bool showInfoIcon = false,
    bool isMultiLine = false,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 11),
      child: Row(
        crossAxisAlignment: isMultiLine
            ? CrossAxisAlignment.start
            : CrossAxisAlignment.center,
        children: [
          Expanded(
            flex: 4,
            child: Row(
              children: [
                Flexible(
                  child: Text(
                    label,
                    style: const TextStyle(
                      fontSize: 12.5,
                      color: Color(0xff64748b),
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
                if (showInfoIcon) ...[
                  const SizedBox(width: 4),
                  const Icon(
                    Icons.info_outline_rounded,
                    size: 14,
                    color: Color(0xffd97706),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            flex: 6,
            child: Text(
              value,
              textAlign: TextAlign.right,
              style: const TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.bold,
                color: Color(0xff0f172a),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class ProfilePage extends StatefulWidget {
  const ProfilePage({
    super.key,
    required this.api,
    required this.onOpenSettings,
    required this.onLogout,
  });

  final ApiClient api;
  final VoidCallback onOpenSettings;
  final VoidCallback onLogout;

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  int _satisfactionRating = 0;
  String _sepaySubtitle = 'Kết nối ngân hàng qua SePay';

  @override
  void initState() {
    super.initState();
    _loadSePaySummary();
  }

  Future<void> _loadSePaySummary() async {
    try {
      final data = await widget.api.request('GET', '/me/sepay') as Map;
      final accounts = (data['bankAccounts'] as List?) ?? const [];
      if (!mounted) return;
      if (accounts.isEmpty) {
        setState(() => _sepaySubtitle = 'Kết nối ngân hàng qua SePay');
        return;
      }
      final profile = data['profile'] as Map?;
      final synced = profile?['lastSyncedAt']?.toString();
      setState(
        () => _sepaySubtitle =
            '${accounts.length} tài khoản đã liên kết · ${_relativeSync(synced)}',
      );
    } catch (_) {
      // This menu remains usable while the profile endpoint is temporarily unavailable.
    }
  }

  String _relativeSync(String? raw) {
    final value = raw == null ? null : DateTime.tryParse(raw);
    if (value == null || value.year == 1970) return 'Chưa đồng bộ';
    final minutes = DateTime.now().difference(value.toLocal()).inMinutes;
    if (minutes < 1) return 'Đồng bộ vừa xong';
    if (minutes < 60) return 'Đồng bộ $minutes phút trước';
    return 'Đồng bộ ${minutes ~/ 60} giờ trước';
  }

  Future<void> _openSePay() async {
    await Navigator.of(context).push(
      MaterialPageRoute(builder: (_) => SePayConnectionPage(api: widget.api)),
    );
    await _loadSePaySummary();
  }

  void _openPersonalInfo() {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => PersonalInfoPage(api: widget.api),
      ),
    );
  }

  void _showSupportModal(String title, String detail) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text(
          title,
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
        ),
        content: Text(
          detail,
          style: const TextStyle(fontSize: 13.5, color: Color(0xff334155)),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text(
              'Đóng',
              style: TextStyle(
                color: Color(0xff6b21a8),
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return PageFrame(
      title: 'Cá nhân',
      child: ListView(
        physics: const BouncingScrollPhysics(),
        children: [
          // Header Card (Avatar + Name + CIF + Badge) -> Click opens Personal Info Screen
          InkWell(
            onTap: _openPersonalInfo,
            borderRadius: BorderRadius.circular(20),
            child: Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.06),
                    blurRadius: 16,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: Column(
                children: [
                  Stack(
                    alignment: Alignment.center,
                    children: [
                      Container(
                        width: 76,
                        height: 76,
                        decoration: const BoxDecoration(
                          shape: BoxShape.circle,
                          gradient: LinearGradient(
                            colors: [
                              Color(0xffd97706),
                              Color(0xfffbbf24),
                              Color(0xfff59e0b),
                            ],
                          ),
                        ),
                      ),
                      Container(
                        width: 70,
                        height: 70,
                        decoration: const BoxDecoration(
                          shape: BoxShape.circle,
                          color: Colors.white,
                        ),
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(35),
                          child: Image.network(
                            'https://api.dicebear.com/7.x/bottts/png?seed=Hoang',
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stackTrace) =>
                                const Icon(
                                  Icons.person_rounded,
                                  size: 40,
                                  color: Color(0xff6b21a8),
                                ),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'NGUYEN XUAN HOANG',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w900,
                          color: Color(0xff2e1065),
                          letterSpacing: 0.3,
                        ),
                      ),
                      SizedBox(width: 4),
                      Icon(
                        Icons.chevron_right_rounded,
                        color: Color(0xffd97706),
                        size: 22,
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'Mã KH (CIF): 03769945',
                    style: TextStyle(
                      fontSize: 11.5,
                      color: Color(0xff64748b),
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 3.5,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xffdcfce7),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: const Color(0xff86efac)),
                    ),
                    child: const Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          Icons.check_circle_rounded,
                          color: Color(0xff16a34a),
                          size: 13,
                        ),
                        SizedBox(width: 4),
                        Text(
                          'Đã xác thực',
                          style: TextStyle(
                            color: Color(0xff15803d),
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 14),

          // Main Support & Contact Services Menu (matching Image 1)
          Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 14,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                _buildMenuItemTile(
                  icon: Icons.chat_rounded,
                  title: 'Chat ngay trên App TPBank',
                  iconColor: const Color(0xffd97706),
                  onTap: () => _showSupportModal(
                    'Chat ngay trên App',
                    'Kết nối trợ lý hỗ trợ trực tiếp 24/7.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.forum_rounded,
                  title: 'Chat qua Messenger',
                  iconColor: const Color(0xff2563eb),
                  onTap: () => _showSupportModal(
                    'Messenger',
                    'Đang chuyển hướng tới Messenger hỗ trợ...',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.mark_chat_read_rounded,
                  title: 'Chat/Gọi qua Zalo',
                  iconColor: const Color(0xff0284c7),
                  onTap: () => _showSupportModal(
                    'Zalo Official',
                    'Đang mở trang Zalo CSKH chính thức...',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.email_rounded,
                  title: 'Gửi Email',
                  iconColor: const Color(0xff06b6d4),
                  onTap: () => _showSupportModal(
                    'Gửi Email',
                    'Vui lòng gửi ý kiến về hotro@tpbank.com.vn',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.phone_rounded,
                  title: 'Gọi Hotline',
                  iconColor: const Color(0xff16a34a),
                  onTap: () => _showSupportModal(
                    'Gọi Hotline',
                    'Đang gọi 1900 58 58 85...',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.warning_amber_rounded,
                  title: 'Yêu cầu trợ giúp/Báo lỗi',
                  iconColor: const Color(0xffef4444),
                  onTap: () => _showSupportModal(
                    'Báo lỗi & Trợ giúp',
                    'Gửi yêu cầu hỗ trợ sự cố giao dịch.',
                  ),
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.tune_rounded,
                  title: 'Cấu hình hiển thị số tiền',
                  subtitle: 'Rút gọn (100) vs Đầy đủ (100.000 VND)',
                  iconColor: const Color(0xff7c3aed),
                  onTap: widget.onOpenSettings,
                ),
                const Divider(height: 1, indent: 52, endIndent: 16),
                _buildMenuItemTile(
                  icon: Icons.account_balance,
                  title: 'Liên kết ngân hàng qua SePay',
                  subtitle: _sepaySubtitle,
                  iconColor: const Color(0xff2563eb),
                  onTap: _openSePay,
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),

          // Satisfaction Rating Card (matching Image 1)
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.05),
                  blurRadius: 14,
                  offset: const Offset(0, 3),
                ),
              ],
            ),
            child: Column(
              children: [
                const Text(
                  'Mức độ hài lòng về App TPBank?',
                  style: TextStyle(
                    fontSize: 13.5,
                    fontWeight: FontWeight.w700,
                    color: Color(0xff334155),
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: List.generate(5, (index) {
                    final starIndex = index + 1;
                    final isFilled = starIndex <= _satisfactionRating;
                    return GestureDetector(
                      onTap: () {
                        setState(() {
                          _satisfactionRating = starIndex;
                        });
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(
                            content: Text(
                              'Cảm ơn bạn đã đánh giá $_satisfactionRating sao!',
                            ),
                            duration: const Duration(seconds: 2),
                          ),
                        );
                      },
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 6),
                        child: Icon(
                          isFilled
                              ? Icons.star_rounded
                              : Icons.star_outline_rounded,
                          size: 34,
                          color: isFilled
                              ? const Color(0xfff59e0b)
                              : const Color(0xffcbd5e1),
                        ),
                      ),
                    );
                  }),
                ),
                const SizedBox(height: 12),
                const Divider(height: 1),
                const SizedBox(height: 10),
                GestureDetector(
                  onTap: () => _showSupportModal(
                    'Lịch sử đánh giá',
                    'Bạn chưa có lịch sử đánh giá nào trước đó.',
                  ),
                  child: const Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        'Lịch sử đánh giá',
                        style: TextStyle(
                          fontSize: 12.5,
                          fontWeight: FontWeight.bold,
                          color: Color(0xffd97706),
                        ),
                      ),
                      SizedBox(width: 4),
                      Icon(
                        Icons.chevron_right_rounded,
                        size: 18,
                        color: Color(0xffd97706),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 12),

          // App Version (matching Image 1)
          const Center(
            child: Text(
              'Phiên bản: 10.12.52',
              style: TextStyle(
                fontSize: 12,
                color: Color(0xff94a3b8),
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          const SizedBox(height: 14),

          // Logout Button Card (matching Image 1)
          InkWell(
            onTap: widget.onLogout,
            borderRadius: BorderRadius.circular(18),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(18),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.05),
                    blurRadius: 14,
                    offset: const Offset(0, 3),
                  ),
                ],
              ),
              child: const Row(
                children: [
                  Icon(
                    Icons.output_rounded,
                    color: Color(0xffef4444),
                    size: 20,
                  ),
                  SizedBox(width: 12),
                  Text(
                    'Đăng xuất',
                    style: TextStyle(
                      fontSize: 13.5,
                      fontWeight: FontWeight.bold,
                      color: Color(0xffef4444),
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 24),
        ],
      ),
    );
  }

  Widget _buildMenuItemTile({
    required IconData icon,
    required String title,
    String? subtitle,
    required Color iconColor,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 13),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(7),
              decoration: BoxDecoration(
                color: iconColor.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(icon, color: iconColor, size: 18),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: const TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: Color(0xff1e293b),
                    ),
                  ),
                  if (subtitle != null) ...[
                    const SizedBox(height: 2),
                    Text(
                      subtitle,
                      style: const TextStyle(
                        fontSize: 10.5,
                        color: Color(0xff64748b),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const Icon(
              Icons.chevron_right_rounded,
              color: Color(0xffd97706),
              size: 20,
            ),
          ],
        ),
      ),
    );
  }
}

class TransactionsPage extends StatefulWidget {
  const TransactionsPage({super.key, required this.api});
  final ApiClient api;
  @override
  State<TransactionsPage> createState() => _TransactionsPageState();
}

class _TransactionsPageState extends State<TransactionsPage> {
  List items = [];
  bool loading = true;
  String? error;
  String selectedFilter = 'all'; // 'all', 'income', 'expense', 'transfer'
  final searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    load();
  }

  @override
  void dispose() {
    searchController.dispose();
    super.dispose();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      final x =
          await widget.api.request('GET', '/transactions?limit=100') as Map;
      items = x['items'] as List? ?? [];
      error = null;
    } catch (e) {
      error = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  void form() => showModalBottomSheet(
    isScrollControlled: true,
    backgroundColor: Colors.transparent,
    context: context,
    builder: (_) =>
        _TransactionFormSheet(api: widget.api, onSuccess: () => load()),
  );

  double get _totalIncome {
    double sum = 0;
    for (final item in items) {
      if (item['type'] == 'income') {
        sum += double.tryParse(item['amount']?.toString() ?? '0') ?? 0;
      }
    }
    return sum;
  }

  double get _totalExpense {
    double sum = 0;
    for (final item in items) {
      if (item['type'] == 'expense') {
        sum += double.tryParse(item['amount']?.toString() ?? '0') ?? 0;
      }
    }
    return sum;
  }

  double get _netCashFlow => _totalIncome - _totalExpense;

  List get _filteredItems {
    final query = searchController.text.trim().toLowerCase();
    return items.where((x) {
      final type = x['type']?.toString() ?? '';
      if (selectedFilter == 'income' && type != 'income') return false;
      if (selectedFilter == 'expense' && type != 'expense') return false;
      if (selectedFilter == 'transfer' && type != 'transfer') return false;

      if (query.isNotEmpty) {
        final note = x['note']?.toString().toLowerCase() ?? '';
        final amountStr = x['amount']?.toString() ?? '';
        return note.contains(query) || amountStr.contains(query);
      }
      return true;
    }).toList();
  }

  Map<String, List> get _groupedItems {
    final Map<String, List> map = {};
    for (final item in _filteredItems) {
      final rawDate =
          item['occurredAt']?.toString() ?? item['createdAt']?.toString();
      final dateKey = _formatGroupDate(rawDate);
      map.putIfAbsent(dateKey, () => []).add(item);
    }
    return map;
  }

  String _formatGroupDate(String? raw) {
    if (raw == null || raw.isEmpty) return 'Khác';
    try {
      final dt = DateTime.parse(raw).toLocal();
      final now = DateTime.now();
      if (dt.year == now.year && dt.month == now.month && dt.day == now.day) {
        return 'Hôm nay - ${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}';
      }
      final yesterday = now.subtract(const Duration(days: 1));
      if (dt.year == yesterday.year &&
          dt.month == yesterday.month &&
          dt.day == yesterday.day) {
        return 'Hôm qua - ${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}';
      }
      return '${dt.day.toString().padLeft(2, '0')}/${dt.month.toString().padLeft(2, '0')}/${dt.year}';
    } catch (_) {
      return 'Khác';
    }
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Lịch sử Giao dịch',
    action: Row(
      children: [
        IconButton(
          onPressed: load,
          icon: const Icon(Icons.refresh_rounded, color: Color(0xfffbbf24)),
        ),
        const SizedBox(width: 4),
        FilledButton.icon(
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xfffbbf24),
            foregroundColor: const Color(0xff1c1917),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
            ),
          ),
          onPressed: form,
          icon: const Icon(Icons.add_rounded, size: 18),
          label: const Text(
            'Thêm mới',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
        ),
      ],
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfffbbf24)),
          )
        : ListView(
            physics: const BouncingScrollPhysics(),
            children: [
              // KPI Analytics Cards Header
              Row(
                children: [
                  Expanded(
                    child: _buildMetricCard(
                      title: 'Tổng Thu',
                      amount: _totalIncome,
                      isIncome: true,
                      icon: Icons.south_west_rounded,
                      accentColor: const Color(0xff4ade80),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _buildMetricCard(
                      title: 'Tổng Chi',
                      amount: _totalExpense,
                      isIncome: false,
                      icon: Icons.north_east_rounded,
                      accentColor: const Color(0xfffb7185),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              _buildNetFlowCard(),
              const SizedBox(height: 16),

              // Search Bar & Filter Pills
              _buildSearchAndFilterSection(),
              const SizedBox(height: 16),

              if (error != null) ErrorBox(error!),

              // Grouped List Items
              if (_groupedItems.isEmpty)
                const EmptyState('Không tìm thấy giao dịch nào')
              else
                ..._groupedItems.entries.map((entry) {
                  final dateLabel = entry.key;
                  final list = entry.value;
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 4,
                          vertical: 8,
                        ),
                        child: Row(
                          children: [
                            Icon(
                              Icons.calendar_today_rounded,
                              size: 14,
                              color: Colors.white.withValues(alpha: 0.6),
                            ),
                            const SizedBox(width: 6),
                            Text(
                              dateLabel,
                              style: TextStyle(
                                color: Colors.white.withValues(alpha: 0.9),
                                fontSize: 13,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const Spacer(),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.white.withValues(alpha: 0.1),
                                borderRadius: BorderRadius.circular(10),
                              ),
                              child: Text(
                                '${list.length} giao dịch',
                                style: const TextStyle(
                                  color: Colors.white70,
                                  fontSize: 10.5,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      ...list.map((x) => _buildTransactionCard(x)),
                      const SizedBox(height: 8),
                    ],
                  );
                }),
            ],
          ),
  );

  Widget _buildMetricCard({
    required String title,
    required double amount,
    required bool isIncome,
    required IconData icon,
    required Color accentColor,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: accentColor.withValues(alpha: 0.25),
          width: 1,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.06),
            blurRadius: 14,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(5),
                decoration: BoxDecoration(
                  color: accentColor.withValues(alpha: 0.12),
                  shape: BoxShape.circle,
                ),
                child: Icon(icon, color: accentColor, size: 12),
              ),
              const SizedBox(width: 6),
              Text(
                title,
                style: const TextStyle(
                  color: Color(0xff64748b),
                  fontSize: 10.5,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          FittedBox(
            fit: BoxFit.scaleDown,
            child: Text(
              '${isIncome ? '+' : '-'}${formatCurrency(amount)}',
              style: TextStyle(
                color: accentColor,
                fontSize: 14.5,
                fontWeight: FontWeight.w900,
                letterSpacing: -0.2,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNetFlowCard() {
    final net = _netCashFlow;
    final isPositive = net >= 0;
    final accentColor = isPositive
        ? const Color(0xff10b981)
        : const Color(0xfff43f5e);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: accentColor.withValues(alpha: 0.25),
          width: 1,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.06),
            blurRadius: 14,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: accentColor.withValues(alpha: 0.12),
              shape: BoxShape.circle,
            ),
            child: Icon(
              isPositive
                  ? Icons.trending_up_rounded
                  : Icons.trending_down_rounded,
              color: accentColor,
              size: 16,
            ),
          ),
          const SizedBox(width: 10),
          const Text(
            'Dòng tiền Ròng:',
            style: TextStyle(
              color: Color(0xff1e293b),
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const Spacer(),
          Text(
            '${isPositive ? '+' : ''}${formatCurrency(net)}',
            style: TextStyle(
              color: accentColor,
              fontSize: 14,
              fontWeight: FontWeight.w900,
              letterSpacing: -0.2,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSearchAndFilterSection() {
    return Column(
      children: [
        // Search Input Bar
        Container(
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.white),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withValues(alpha: 0.05),
                blurRadius: 14,
                offset: const Offset(0, 3),
              ),
            ],
          ),
          child: TextField(
            controller: searchController,
            onChanged: (_) => setState(() {}),
            style: const TextStyle(
              color: Color(0xff0f172a),
              fontSize: 12.5,
              fontWeight: FontWeight.w600,
            ),
            decoration: InputDecoration(
              hintText: 'Tìm kiếm theo ghi chú, số tiền...',
              hintStyle: const TextStyle(
                color: Color(0xff94a3b8),
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
              prefixIcon: const Icon(
                Icons.search_rounded,
                color: Color(0xffd97706),
                size: 18,
              ),
              suffixIcon: searchController.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(
                        Icons.clear_rounded,
                        color: Color(0xff64748b),
                        size: 16,
                      ),
                      onPressed: () {
                        searchController.clear();
                        setState(() {});
                      },
                    )
                  : null,
              border: InputBorder.none,
              contentPadding: const EdgeInsets.symmetric(vertical: 10),
            ),
          ),
        ),
        const SizedBox(height: 8),

        // Filter Pills
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          child: Row(
            children: [
              _buildFilterPill('all', 'Tất cả', Icons.apps_rounded),
              const SizedBox(width: 6),
              _buildFilterPill('income', 'Thu nhập', Icons.south_west_rounded),
              const SizedBox(width: 6),
              _buildFilterPill('expense', 'Chi tiêu', Icons.north_east_rounded),
              const SizedBox(width: 6),
              _buildFilterPill(
                'transfer',
                'Chuyển tiền',
                Icons.swap_horiz_rounded,
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildFilterPill(String filterKey, String label, IconData icon) {
    final isSelected = selectedFilter == filterKey;
    return InkWell(
      onTap: () => setState(() => selectedFilter = filterKey),
      borderRadius: BorderRadius.circular(14),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 6.5),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xfffbbf24) : Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(
            color: isSelected
                ? const Color(0xffd97706)
                : Colors.white.withValues(alpha: 0.8),
            width: 1,
          ),
          boxShadow: [
            BoxShadow(
              color: isSelected
                  ? const Color(0x22d97706)
                  : Colors.black.withValues(alpha: 0.04),
              blurRadius: isSelected ? 10 : 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              icon,
              size: 13,
              color: isSelected
                  ? const Color(0xff1c1917)
                  : const Color(0xff64748b),
            ),
            const SizedBox(width: 5),
            Text(
              label,
              style: TextStyle(
                color: isSelected
                    ? const Color(0xff1c1917)
                    : const Color(0xff334155),
                fontSize: 11.5,
                fontWeight: isSelected ? FontWeight.w900 : FontWeight.w600,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTransactionCard(Map item) {
    final isIncome = item['type'] == 'income';
    final isTransfer = item['type'] == 'transfer';
    final name = item['name']?.toString().trim();
    final note = item['note']?.toString().trim();
    final title = (name != null && name.isNotEmpty)
        ? name
        : ((note != null && note.isNotEmpty)
              ? note
              : (isIncome
                    ? 'Thu nhập'
                    : (isTransfer ? 'Chuyển tiền' : 'Chi tiêu')));
    final amountVal = double.tryParse(item['amount']?.toString() ?? '0') ?? 0.0;
    final icon = isIncome
        ? Icons.south_west_rounded
        : (isTransfer ? Icons.swap_horiz_rounded : Icons.north_east_rounded);
    final color = isIncome
        ? const Color(0xff4ade80)
        : (isTransfer ? const Color(0xff38bdf8) : const Color(0xfffb7185));

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: FinoraListTile(
        icon: icon,
        iconColor: color,
        title: title,
        subtitle: _formatDate(
          item['occurredAt']?.toString() ?? item['createdAt']?.toString(),
        ),
        amount:
            "${isIncome ? '+' : (isTransfer ? '' : '-')}${formatCurrency(amountVal)}",
      ),
    );
  }
}

class _TransactionFormSheet extends StatefulWidget {
  const _TransactionFormSheet({required this.api, required this.onSuccess});
  final ApiClient api;
  final VoidCallback onSuccess;

  @override
  State<_TransactionFormSheet> createState() => _TransactionFormSheetState();
}

class _TransactionFormSheetState extends State<_TransactionFormSheet> {
  String type = 'expense';
  String? selectedAccountId;
  List<Map<String, dynamic>> accounts = [];
  bool loadingAccounts = true;

  final nameController = TextEditingController();
  final amountController = TextEditingController();
  final noteController = TextEditingController();
  DateTime selectedDate = DateTime.now();
  bool submitting = false;

  @override
  void initState() {
    super.initState();
    _loadAccounts();
  }

  Future<void> _loadAccounts() async {
    try {
      final res = await widget.api.request('GET', '/accounts');
      final list = res is List
          ? res
          : (res is Map ? res['items'] as List? ?? [] : []);
      setState(() {
        accounts = List<Map<String, dynamic>>.from(
          list.map((e) => Map<String, dynamic>.from(e as Map)),
        );
        if (accounts.isNotEmpty) {
          selectedAccountId = accounts.first['id']?.toString();
        }
        loadingAccounts = false;
      });
    } catch (_) {
      if (mounted) setState(() => loadingAccounts = false);
    }
  }

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: selectedDate,
      firstDate: DateTime(2020),
      lastDate: DateTime(2030),
      builder: (context, child) {
        return Theme(
          data: ThemeData.light().copyWith(
            colorScheme: const ColorScheme.light(
              primary: Color(0xffd97706),
              onPrimary: Colors.white,
              surface: Colors.white,
              onSurface: Color(0xff0f172a),
            ),
          ),
          child: child!,
        );
      },
    );
    if (picked != null) {
      setState(() => selectedDate = picked);
    }
  }

  Future<void> _submit() async {
    final rawName = nameController.text.trim();
    final rawAmount = amountController.text.trim();
    if (rawAmount.isEmpty) {
      showError(context, 'Vui lòng nhập số tiền.');
      return;
    }

    final parsedVal = parseSmartAmount(rawAmount);
    if (parsedVal <= 0) {
      showError(context, 'Số tiền không hợp lệ.');
      return;
    }

    if (selectedAccountId == null || selectedAccountId!.isEmpty) {
      showError(context, 'Vui lòng chọn hoặc tạo tài khoản trước.');
      return;
    }

    final txnName = rawName.isNotEmpty
        ? rawName
        : (type == 'income' ? 'Thu nhập' : 'Chi tiêu');

    setState(() => submitting = true);
    try {
      await widget.api.request('POST', '/transactions', {
        'accountId': selectedAccountId,
        'name': txnName,
        'type': type,
        'amount': parsedVal.toStringAsFixed(0),
        'currency': 'VND',
        'note': noteController.text.trim(),
        'occurredAt': selectedDate.toUtc().toIso8601String(),
        'status': 'posted',
      });
      if (mounted) {
        if (Navigator.of(context).canPop()) {
          Navigator.of(context).pop();
        }
        widget.onSuccess();
      }
    } catch (e) {
      if (mounted) showError(context, e.toString());
    } finally {
      if (mounted) setState(() => submitting = false);
    }
  }

  @override
  void dispose() {
    nameController.dispose();
    amountController.dispose();
    noteController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final parsedAmount = parseSmartAmount(amountController.text);
    final isToday =
        DateTime.now().year == selectedDate.year &&
        DateTime.now().month == selectedDate.month &&
        DateTime.now().day == selectedDate.day;
    final dateLabel = isToday
        ? 'Hôm nay (${selectedDate.day.toString().padLeft(2, '0')}/${selectedDate.month.toString().padLeft(2, '0')}/${selectedDate.year})'
        : '${selectedDate.day.toString().padLeft(2, '0')}/${selectedDate.month.toString().padLeft(2, '0')}/${selectedDate.year}';

    return FinoraSheet(
      title: 'Tạo Giao dịch',
      subtitle:
          'Ghi nhận thu chi nhanh. Số tiền được tự động nội suy thông minh.',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: InkWell(
                  onTap: () => setState(() => type = 'expense'),
                  borderRadius: BorderRadius.circular(16),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: type == 'expense'
                          ? const Color(0xffffe4e6)
                          : const Color(0xfff1f5f9),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: type == 'expense'
                            ? const Color(0xfff43f5e)
                            : const Color(0xffcbd5e1),
                        width: type == 'expense' ? 2 : 1,
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.north_east_rounded,
                          color: type == 'expense'
                              ? const Color(0xffbe123c)
                              : const Color(0xff64748b),
                          size: 18,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          'Chi tiêu',
                          style: TextStyle(
                            color: type == 'expense'
                                ? const Color(0xffbe123c)
                                : const Color(0xff475569),
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: InkWell(
                  onTap: () => setState(() => type = 'income'),
                  borderRadius: BorderRadius.circular(16),
                  child: AnimatedContainer(
                    duration: const Duration(milliseconds: 200),
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    decoration: BoxDecoration(
                      color: type == 'income'
                          ? const Color(0xffdcfce7)
                          : const Color(0xfff1f5f9),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: type == 'income'
                            ? const Color(0xff22c55e)
                            : const Color(0xffcbd5e1),
                        width: type == 'income' ? 2 : 1,
                      ),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.south_west_rounded,
                          color: type == 'income'
                              ? const Color(0xff15803d)
                              : const Color(0xff64748b),
                          size: 18,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          'Thu nhập',
                          style: TextStyle(
                            color: type == 'income'
                                ? const Color(0xff15803d)
                                : const Color(0xff475569),
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          if (loadingAccounts)
            const Center(
              child: CircularProgressIndicator(color: Color(0xffd97706)),
            )
          else if (accounts.isEmpty)
            const Text(
              'Chưa có tài khoản nào. Vui lòng tạo tài khoản trước.',
              style: TextStyle(
                color: Color(0xffef4444),
                fontSize: 13,
                fontWeight: FontWeight.bold,
              ),
            )
          else
            DropdownButtonFormField<String>(
              initialValue: selectedAccountId,
              dropdownColor: Colors.white,
              style: const TextStyle(
                color: Color(0xff0f172a),
                fontWeight: FontWeight.w800,
                fontSize: 14,
              ),
              items: accounts.map((acc) {
                return DropdownMenuItem<String>(
                  value: acc['id']?.toString(),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.account_balance_wallet_rounded,
                        color: Color(0xffd97706),
                        size: 18,
                      ),
                      const SizedBox(width: 8),
                      Text(acc['name']?.toString() ?? 'Tài khoản'),
                    ],
                  ),
                );
              }).toList(),
              onChanged: (v) => setState(() => selectedAccountId = v),
              decoration: InputDecoration(
                labelText: 'Chọn tài khoản',
                labelStyle: const TextStyle(
                  color: Color(0xff64748b),
                  fontWeight: FontWeight.w600,
                ),
                filled: true,
                fillColor: const Color(0xfff8fafc),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(
                    color: Color(0xffcbd5e1),
                    width: 1.2,
                  ),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(
                    color: Color(0xffcbd5e1),
                    width: 1.2,
                  ),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(
                    color: Color(0xffd97706),
                    width: 1.8,
                  ),
                ),
              ),
            ),
          const SizedBox(height: 14),
          _CustomGlassTextField(
            controller: nameController,
            labelText: 'Tên giao dịch (VD: Lương tháng, Cà phê...)',
            icon: Icons.title_rounded,
          ),
          // Quick Amount Suggestion Chips
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                _buildQuickAmountChip('100k', '100'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('200k', '200'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('500k', '500'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('1M', '1tr'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('2M', '2tr'),
                const SizedBox(width: 6),
                _buildQuickAmountChip('5M', '5tr'),
              ],
            ),
          ),
          const SizedBox(height: 10),
          _CustomGlassTextField(
            controller: amountController,
            keyboardType: TextInputType.text,
            labelText: 'Số tiền (Nhập 100 = 100k, 1000 = 1tr, 1.5tr)',
            icon: Icons.attach_money_rounded,
            onChanged: (_) => setState(() {}),
          ),
          if (amountController.text.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 6, left: 12),
              child: Row(
                children: [
                  const Icon(
                    Icons.auto_awesome_rounded,
                    color: Color(0xffd97706),
                    size: 14,
                  ),
                  const SizedBox(width: 4),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 3,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xfffef3c7),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      '= ${formatCurrency(parsedAmount)}',
                      style: const TextStyle(
                        color: Color(0xffb45309),
                        fontWeight: FontWeight.w900,
                        fontSize: 13,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          const SizedBox(height: 14),
          InkWell(
            onTap: _pickDate,
            borderRadius: BorderRadius.circular(16),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
              decoration: BoxDecoration(
                color: const Color(0xfff8fafc),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: const Color(0xffcbd5e1), width: 1.2),
              ),
              child: Row(
                children: [
                  const Icon(
                    Icons.calendar_today_rounded,
                    color: Color(0xffd97706),
                    size: 20,
                  ),
                  const SizedBox(width: 12),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Ngày thực hiện',
                        style: TextStyle(
                          color: Color(0xff64748b),
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        dateLabel,
                        style: const TextStyle(
                          color: Color(0xff0f172a),
                          fontWeight: FontWeight.w800,
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  const Spacer(),
                  const Icon(
                    Icons.edit_calendar_rounded,
                    color: Color(0xff94a3b8),
                    size: 18,
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 14),
          _CustomGlassTextField(
            controller: noteController,
            labelText: 'Ghi chú (Tùy chọn)',
            icon: Icons.note_alt_rounded,
          ),
          const SizedBox(height: 18),
          _AnimatedGoldButton(
            busy: submitting,
            label: 'Lưu giao dịch',
            onTap: _submit,
          ),
        ],
      ),
    );
  }

  Widget _buildQuickAmountChip(String label, String valueToSet) {
    return InkWell(
      onTap: () {
        setState(() {
          amountController.text = valueToSet;
        });
      },
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
        decoration: BoxDecoration(
          color: const Color(0xfffef3c7),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xfffde68a)),
        ),
        child: Text(
          '+$label',
          style: const TextStyle(
            color: Color(0xffb45309),
            fontSize: 11.5,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
  }
}

double parseSmartAmount(String rawInput, {String currency = 'VND'}) {
  final clean = rawInput
      .trim()
      .toLowerCase()
      .replaceAll(',', '.')
      .replaceAll(' ', '');
  if (clean.isEmpty) return 0.0;

  final isVnd = currency.toUpperCase() == 'VND';

  if (RegExp(r'^[0-9.]+(k)$').hasMatch(clean)) {
    final numStr = clean.replaceAll('k', '');
    final val = double.tryParse(numStr) ?? 0.0;
    return val * 1000;
  }

  if (RegExp(r'^[0-9.]+(tr|m|trieu|triệu)$').hasMatch(clean)) {
    final numStr = clean.replaceAll(RegExp(r'(tr|m|trieu|triệu)'), '');
    final val = double.tryParse(numStr) ?? 0.0;
    return val * 1000000;
  }

  if (RegExp(r'^[0-9.]+(ty|b|tỷ)$').hasMatch(clean)) {
    final numStr = clean.replaceAll(RegExp(r'(ty|b|tỷ)'), '');
    final val = double.tryParse(numStr) ?? 0.0;
    return val * 1000000000;
  }

  final val = double.tryParse(clean);
  if (val != null) {
    if (isVnd &&
        val > 0 &&
        (appAmountDisplayMode == 'compact' || val < 10000)) {
      return val * 1000;
    }
    return val;
  }

  return 0.0;
}

String formatCurrency(double amount, {String currency = 'VND'}) {
  final intVal = amount.round();
  final formattedStr = intVal.toString().replaceAllMapped(
    RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'),
    (Match m) => '${m[1]}.',
  );
  return '$formattedStr $currency';
}

class BudgetPage extends StatefulWidget {
  const BudgetPage({super.key, required this.api});
  final ApiClient api;
  @override
  State<BudgetPage> createState() => _BudgetPageState();
}

class _BudgetPageState extends State<BudgetPage> {
  final period = TextEditingController(text: '2026-07'),
      category = TextEditingController(),
      limit = TextEditingController();
  dynamic data;
  String? err;
  bool loading = false;
  Future<void> load() async {
    setState(() => loading = true);
    try {
      data = await widget.api.request('GET', '/budgets/${period.text}');
      err = null;
    } catch (e) {
      err = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> save() async {
    try {
      await widget.api.request('PUT', '/budgets/${period.text}', {
        'categoryId': category.text,
        'limit': limit.text,
      });
      load();
    } catch (e) {
      if (mounted) showError(context, e.toString());
    }
  }

  @override
  void dispose() {
    period.dispose();
    category.dispose();
    limit.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Ngân sách',
    child: ListView(
      children: [
        const _ScreenIntro(
          'Đặt giới hạn thông minh để kế hoạch tài chính luôn đúng hướng.',
        ),
        FinoraSurface(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Tra cứu kỳ ngân sách',
                style: TextStyle(
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  fontSize: 16,
                ),
              ),
              const SizedBox(height: 14),
              _CustomGlassTextField(
                controller: period,
                labelText: 'Kỳ ngân sách (YYYY-MM)',
                icon: Icons.calendar_month_rounded,
              ),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                style: OutlinedButton.styleFrom(
                  side: const BorderSide(color: Color(0xfff7d070)),
                  foregroundColor: const Color(0xfff7d070),
                ),
                onPressed: loading ? null : load,
                icon: const Icon(Icons.search_rounded),
                label: const Text('Xem ngân sách'),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        FinoraSurface(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Thiết lập hạn mức',
                style: TextStyle(
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  fontSize: 16,
                ),
              ),
              const SizedBox(height: 14),
              _CustomGlassTextField(
                controller: category,
                labelText: 'Category ID',
                icon: Icons.sell_outlined,
              ),
              const SizedBox(height: 12),
              _CustomGlassTextField(
                controller: limit,
                keyboardType: TextInputType.number,
                labelText: 'Hạn mức',
                icon: Icons.account_balance_wallet_outlined,
              ),
              const SizedBox(height: 14),
              FilledButton.icon(
                style: FilledButton.styleFrom(
                  backgroundColor: const Color(0xfff7d070),
                  foregroundColor: const Color(0xff200733),
                ),
                onPressed: save,
                icon: const Icon(Icons.check_rounded),
                label: const Text(
                  'Lưu hạn mức',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
        ),
        if (err != null) ErrorBox(err!),
        if (data != null)
          Padding(
            padding: const EdgeInsets.only(top: 16),
            child: FinoraSurface(
              child: Text(
                JsonEncoder.withIndent('  ').convert(data),
                style: const TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 12,
                  color: Colors.white,
                ),
              ),
            ),
          ),
      ],
    ),
  );
}

class SePayConnectionPage extends StatefulWidget {
  const SePayConnectionPage({super.key, required this.api});
  final ApiClient api;

  @override
  State<SePayConnectionPage> createState() => _SePayConnectionPageState();
}

class _SePayConnectionPageState extends State<SePayConnectionPage> {
  List<dynamic> _accounts = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
    });
    try {
      final data = await widget.api.request('GET', '/me/sepay') as Map;
      _accounts = (data['bankAccounts'] as List?) ?? const [];
      _error = null;
    } catch (error) {
      _error = error.toString();
    }
    if (mounted) {
      setState(() {
        _loading = false;
      });
    }
  }

  Future<void> _link() async {
    try {
      final data =
          await widget.api.request('POST', '/me/sepay/link-session') as Map;
      final url = data['hosted_link_url']?.toString() ?? '';
      if (url.isEmpty) {
        throw Exception('SePay không trả về đường dẫn liên kết.');
      }
      if (!mounted) {
        return;
      }
      final completion = await Navigator.of(context).push<Map<String, dynamic>>(
        MaterialPageRoute(builder: (_) => _HostedLinkPage(url: url)),
      );
      final accountNumber =
          completion?['account_number']?.toString() ??
          completion?['accountNumber']?.toString() ??
          '';
      if (accountNumber.isNotEmpty) {
        await widget.api.request('POST', '/me/sepay/bank-accounts/sync', {
          'accountNumber': accountNumber,
        });
      }
      await _load();
    } catch (error) {
      if (mounted) showError(context, error.toString());
    }
  }

  Future<void> _mapAccount(Map account) async {
    try {
      final accounts = await widget.api.request('GET', '/accounts') as List;
      if (!mounted) return;
      final selected = await showModalBottomSheet<Map>(
        context: context,
        builder: (context) => SafeArea(
          child: ListView(
            shrinkWrap: true,
            children: [
              const ListTile(
                title: Text(
                  'Chọn tài khoản Finora',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
              const Padding(
                padding: EdgeInsets.fromLTRB(16, 0, 16, 12),
                child: Text(
                  'Nếu user có nhiều thành viên, giao dịch ngân hàng sẽ được chia sẻ theo quyền của user đó.',
                  style: TextStyle(color: Color(0xff64748b), fontSize: 12),
                ),
              ),
              ...accounts.map(
                (item) => ListTile(
                  leading: const Icon(Icons.account_balance_wallet_outlined),
                  title: Text(item['name']?.toString() ?? 'Tài khoản'),
                  subtitle: Text(item['currency']?.toString() ?? 'VND'),
                  onTap: () => Navigator.pop(
                    context,
                    Map<String, dynamic>.from(item as Map),
                  ),
                ),
              ),
            ],
          ),
        ),
      );
      if (selected == null) return;
      await widget.api.request(
        'POST',
        '/me/sepay/bank-accounts/${account['id']}/map',
        {'accountId': selected['id']},
      );
      await _load();
    } catch (error) {
      if (mounted) showError(context, error.toString());
    }
  }

  Future<void> _unlink(Map account) async {
    final accepted = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Ngắt liên kết?'),
        content: const Text(
          'Giao dịch đã ghi nhận vẫn được giữ lại. Bạn có thể liên kết lại sau.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Hủy'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Ngắt liên kết'),
          ),
        ],
      ),
    );
    if (accepted != true) return;
    try {
      await widget.api.request(
        'POST',
        '/me/sepay/bank-accounts/${account['id']}/unlink',
      );
      await _load();
    } catch (error) {
      if (mounted) showError(context, error.toString());
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Liên kết SePay')),
    body: _loading
        ? const Center(child: CircularProgressIndicator())
        : RefreshIndicator(
            onRefresh: _load,
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: const Color(0xffeff6ff),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: const Row(
                    children: [
                      Icon(Icons.shield_outlined, color: Color(0xff2563eb)),
                      SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Finora chỉ có quyền đọc giao dịch và số dư. Chúng tôi không lưu mật khẩu hoặc OTP của bạn.',
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                FilledButton.icon(
                  onPressed: _link,
                  icon: const Icon(Icons.add_link),
                  label: const Text('Liên kết tài khoản ngân hàng'),
                ),
                if (_error != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: ErrorBox(_error!),
                  ),
                const SizedBox(height: 20),
                const Text(
                  'TÀI KHOẢN ĐÃ LIÊN KẾT',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: Color(0xff475569),
                  ),
                ),
                const SizedBox(height: 8),
                if (_accounts.isEmpty)
                  const EmptyState(
                    'Chưa có tài khoản ngân hàng nào được liên kết.',
                  ),
                ..._accounts.map((raw) {
                  final account = Map<String, dynamic>.from(raw as Map);
                  final mapping = account['mapping'];
                  final canIn = account['supportsIn'] == true;
                  final canOut = account['supportsOut'] == true;
                  return Card(
                    child: Padding(
                      padding: const EdgeInsets.all(14),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(
                                Icons.account_balance,
                                color: Color(0xff2563eb),
                              ),
                              const SizedBox(width: 10),
                              Expanded(
                                child: Text(
                                  account['bankName']?.toString().isNotEmpty ==
                                          true
                                      ? account['bankName'].toString()
                                      : account['bankCode']?.toString() ??
                                            'Ngân hàng',
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                              Text(account['status']?.toString() ?? ''),
                            ],
                          ),
                          const SizedBox(height: 8),
                          Text(
                            account['accountNumberMasked']?.toString() ??
                                '••••',
                          ),
                          const SizedBox(height: 10),
                          Wrap(
                            spacing: 8,
                            children: [
                              Chip(
                                label: Text(
                                  canIn ? 'Tiền vào' : 'Không hỗ trợ tiền vào',
                                ),
                              ),
                              Chip(
                                label: Text(
                                  canOut ? 'Tiền ra' : 'Không hỗ trợ tiền ra',
                                ),
                              ),
                            ],
                          ),
                          const Divider(),
                          if (mapping == null)
                            Align(
                              alignment: Alignment.centerRight,
                              child: TextButton.icon(
                                onPressed: () => _mapAccount(account),
                                icon: const Icon(Icons.link),
                                label: const Text('Map tài khoản Finora'),
                              ),
                            )
                          else
                            Row(
                              children: [
                                Expanded(
                                  child: Text(
                                    'Đã map vào tài khoản Finora',
                                    style: TextStyle(
                                      color: Colors.green.shade700,
                                    ),
                                  ),
                                ),
                                TextButton(
                                  onPressed: () => _unlink(account),
                                  child: const Text('Ngắt liên kết'),
                                ),
                              ],
                            ),
                        ],
                      ),
                    ),
                  );
                }),
              ],
            ),
          ),
  );
}

class _HostedLinkPage extends StatefulWidget {
  const _HostedLinkPage({required this.url});
  final String url;
  @override
  State<_HostedLinkPage> createState() => _HostedLinkPageState();
}

class _HostedLinkPageState extends State<_HostedLinkPage> {
  late final WebViewController _controller;
  @override
  void initState() {
    super.initState();
    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..addJavaScriptChannel(
        'BankHubEvents',
        onMessageReceived: (message) {
          try {
            final event = jsonDecode(message.message) as Map;
            final name = event['event']?.toString() ?? '';
            final metadata = Map<String, dynamic>.from(
              (event['metadata'] as Map?) ?? const {},
            );
            if (name == 'FINISHED_BANK_ACCOUNT_LINK') {
              Navigator.of(context).pop(metadata);
            } else if (name == 'BANKHUB_CLOSE_LINK' ||
                name == 'FINISHED_BANK_ACCOUNT_UNLINK') {
              Navigator.of(context).pop();
            } else if (name == 'BANKHUB_TOKEN_EXPIRED' ||
                name == 'BANKHUB_SESSION_EXPIRED') {
              showError(context, 'Phiên liên kết SePay đã hết hạn.');
              Navigator.of(context).pop();
            }
          } catch (_) {
            // Ignore malformed cross-window messages; they are not provider data.
          }
        },
      )
      ..setNavigationDelegate(
        NavigationDelegate(
          onPageFinished: (_) => _controller.runJavaScript('''
            (function () {
              if (window.__finoraBankHubListener) return;
              window.__finoraBankHubListener = true;
              window.addEventListener('message', function (event) {
                var host = '';
                try { host = new URL(event.origin).hostname; } catch (_) { return; }
                if (host !== 'bankhub.sepay.vn' && !host.endsWith('.sepay.vn')) return;
                var data = event.data;
                if (typeof data === 'string') { try { data = JSON.parse(data); } catch (_) { return; } }
                if (data && data.event) BankHubEvents.postMessage(JSON.stringify(data));
              });
            })();
          '''),
        ),
      )
      ..loadRequest(Uri.parse(widget.url));
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Kết nối ngân hàng')),
    body: WebViewWidget(controller: _controller),
  );
}

class BankPage extends StatefulWidget {
  const BankPage({super.key, required this.api});
  final ApiClient api;
  @override
  State<BankPage> createState() => _BankPageState();
}

class _BankPageState extends State<BankPage> {
  List feed = [];
  String? err;
  bool loading = true;
  String state = 'needs_review';
  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      final data =
          await widget.api.request('GET', '/me/bank-feed?state=$state') as Map;
      feed = (data['items'] as List?) ?? const [];
      err = null;
    } catch (e) {
      err = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> _action(
    String id,
    String action, [
    Map<String, dynamic>? payload,
  ]) async {
    try {
      await widget.api.request('POST', '/bank-feed/$id/$action', payload);
      await load();
    } catch (e) {
      if (mounted) showError(context, e.toString());
    }
  }

  Future<void> _edit(Map item) async {
    final result = await showModalBottomSheet<Map<String, dynamic>>(
      context: context,
      isScrollControlled: true,
      builder: (_) => _BankFeedEditSheet(api: widget.api, item: item),
    );
    if (result != null) await _action(item['id'].toString(), 'correct', result);
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Cần kiểm tra',
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfff7d070)),
          )
        : ListView(
            children: [
              const _ScreenIntro(
                'Rà soát giao dịch trước khi Finora ghi nhận vào sổ.',
              ),
              if (err != null) ErrorBox(err!),
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                child: Row(
                  children: [
                    for (final tab in const [
                      ('needs_review', 'Cần kiểm tra'),
                      ('ai_tagged', 'AI đã gắn'),
                      ('confirmed', 'Đã xác nhận'),
                      ('ignored', 'Bỏ qua'),
                    ])
                      Padding(
                        padding: const EdgeInsets.only(right: 8),
                        child: ChoiceChip(
                          label: Text(tab.$2),
                          selected: state == tab.$1,
                          onSelected: (_) {
                            setState(() => state = tab.$1);
                            load();
                          },
                        ),
                      ),
                  ],
                ),
              ),
              const SizedBox(height: 14),
              ...feed.map((raw) {
                final item = Map<String, dynamic>.from(raw as Map);
                final inbound = item['direction'] == 'in';
                final suggestions = (item['suggestions'] as List?) ?? const [];
                final suggestion = suggestions.isEmpty
                    ? null
                    : Map<String, dynamic>.from(suggestions.first as Map);
                return Card(
                  child: Padding(
                    padding: const EdgeInsets.all(14),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(
                              inbound ? Icons.south_west : Icons.north_east,
                              color: inbound ? Colors.green : Colors.red,
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                inbound ? 'Thu' : 'Chi',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  color: inbound ? Colors.green : Colors.red,
                                ),
                              ),
                            ),
                            Text(
                              '${item['amount'] ?? ''} ${item['currency'] ?? 'VND'}',
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(_redactBankContent(item['description'])),
                        Text(
                          'Ngân hàng · ${item['occurredAt']?.toString().replaceFirst('T', ' ').split('.').first ?? ''}',
                          style: const TextStyle(
                            color: Color(0xff64748b),
                            fontSize: 12,
                          ),
                        ),
                        if (suggestion != null)
                          Padding(
                            padding: const EdgeInsets.only(top: 10),
                            child: Text(
                              '${suggestion['suggestedName'] ?? 'Gợi ý AI'} · ${suggestion['reason'] ?? ''} · ${_confidenceLabel(suggestion['confidence'])}',
                              style: const TextStyle(
                                color: Color(0xff2563eb),
                                fontSize: 12,
                              ),
                            ),
                          ),
                        if (suggestion == null)
                          const Padding(
                            padding: EdgeInsets.only(top: 10),
                            child: Text(
                              'Chưa phân loại',
                              style: TextStyle(
                                color: Color(0xff64748b),
                                fontSize: 12,
                              ),
                            ),
                          ),
                        if (state == 'needs_review' || state == 'ai_tagged')
                          Row(
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              TextButton(
                                onPressed: () =>
                                    _action(item['id'].toString(), 'ignore'),
                                child: const Text('Bỏ qua'),
                              ),
                              TextButton(
                                onPressed: () => _edit(item),
                                child: const Text('Sửa'),
                              ),
                              FilledButton(
                                onPressed: () =>
                                    _action(item['id'].toString(), 'confirm'),
                                child: const Text('Xác nhận'),
                              ),
                            ],
                          ),
                      ],
                    ),
                  ),
                );
              }),
              if (feed.isEmpty)
                const EmptyState('Không có giao dịch trong mục này.'),
            ],
          ),
  );
}

String _redactBankContent(dynamic value) {
  final text = value?.toString().trim() ?? '';
  if (text.isEmpty) return 'Giao dịch ngân hàng';
  final masked = text.replaceAll(RegExp(r'\b\d{6,}\b'), '••••');
  final runes = masked.runes.toList();
  return runes.length <= 90
      ? masked
      : '${String.fromCharCodes(runes.take(90))}…';
}

String _confidenceLabel(dynamic value) {
  final confidence = value is num
      ? value.toDouble()
      : double.tryParse(value?.toString() ?? '') ?? 0;
  if (confidence >= 90) return 'Khớp cao ${confidence.round()}%';
  if (confidence > 0) return 'Cần xác nhận ${confidence.round()}%';
  return 'Chưa phân loại';
}

class _BankFeedEditSheet extends StatefulWidget {
  const _BankFeedEditSheet({required this.api, required this.item});
  final ApiClient api;
  final Map item;
  @override
  State<_BankFeedEditSheet> createState() => _BankFeedEditSheetState();
}

class _BankFeedEditSheetState extends State<_BankFeedEditSheet> {
  late final TextEditingController _name = TextEditingController(
    text: widget.item['description']?.toString() ?? '',
  );
  final _category = TextEditingController();
  final _note = TextEditingController();
  bool _remember = false;
  List<dynamic> _accounts = const [];
  String? _accountId;
  late String _type = widget.item['direction'] == 'in' ? 'income' : 'expense';
  @override
  void initState() {
    super.initState();
    final mapping = widget.item['mapping'];
    _accountId = mapping is Map ? mapping['accountId']?.toString() : null;
    _loadAccounts();
  }

  Future<void> _loadAccounts() async {
    try {
      final data = await widget.api.request('GET', '/accounts');
      if (mounted) setState(() => _accounts = (data as List?) ?? const []);
    } catch (_) {
      // The mapped account remains the server-side fallback if this optional
      // picker cannot be loaded (for example, offline while editing).
    }
  }

  @override
  void dispose() {
    _name.dispose();
    _category.dispose();
    _note.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => FinoraSheet(
    title: 'Sửa giao dịch',
    subtitle: 'Dữ liệu gốc từ ngân hàng không thể thay đổi.',
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        DropdownButtonFormField<String>(
          initialValue: _type,
          items: const [
            DropdownMenuItem(value: 'income', child: Text('Thu')),
            DropdownMenuItem(value: 'expense', child: Text('Chi')),
          ],
          onChanged: (value) => setState(() => _type = value!),
          decoration: const InputDecoration(labelText: 'Loại'),
        ),
        TextField(
          controller: _name,
          decoration: const InputDecoration(labelText: 'Tên giao dịch'),
        ),
        TextField(
          controller: _category,
          decoration: const InputDecoration(
            labelText: 'Danh mục (ID, tùy chọn)',
          ),
        ),
        DropdownButtonFormField<String>(
          initialValue:
              _accounts.any((item) => item['id']?.toString() == _accountId)
              ? _accountId
              : null,
          items: _accounts
              .map(
                (raw) => DropdownMenuItem<String>(
                  value: raw['id']?.toString(),
                  child: Text(raw['name']?.toString() ?? 'Tài khoản Finora'),
                ),
              )
              .toList(),
          onChanged: (value) => setState(() => _accountId = value),
          decoration: const InputDecoration(labelText: 'Tài khoản Finora'),
        ),
        TextField(
          controller: _note,
          decoration: const InputDecoration(labelText: 'Ghi chú'),
        ),
        SwitchListTile(
          contentPadding: EdgeInsets.zero,
          value: _remember,
          onChanged: (value) => setState(() => _remember = value),
          title: const Text('Ghi nhớ lựa chọn này cho các lần sau'),
        ),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: const Color(0xfff8fafc),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Text(
            'Dữ liệu gốc từ ngân hàng\n${widget.item['description'] ?? ''}\n${widget.item['amount'] ?? ''} ${widget.item['currency'] ?? ''}',
            style: const TextStyle(fontSize: 12, color: Color(0xff475569)),
          ),
        ),
        const SizedBox(height: 16),
        FilledButton(
          onPressed: () => Navigator.pop(context, {
            'name': _name.text.trim(),
            'categoryId': _category.text.trim(),
            'accountId': _accountId ?? '',
            'note': _note.text.trim(),
            'type': _type,
            'rememberChoice': _remember,
          }),
          child: const Text('Lưu và xác nhận'),
        ),
      ],
    ),
  );
}

class FinoraSheet extends StatelessWidget {
  const FinoraSheet({
    super.key,
    required this.title,
    required this.subtitle,
    required this.child,
  });
  final String title, subtitle;
  final Widget child;
  @override
  Widget build(BuildContext context) => Padding(
    padding: EdgeInsets.fromLTRB(
      16,
      0,
      16,
      16 + MediaQuery.of(context).viewInsets.bottom,
    ),
    child: Container(
      padding: const EdgeInsets.fromLTRB(20, 16, 20, 22),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: Colors.white, width: 1.5),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.22),
            blurRadius: 36,
            offset: const Offset(0, -10),
          ),
        ],
      ),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                width: 44,
                height: 5,
                decoration: BoxDecoration(
                  color: const Color(0xffcbd5e1),
                  borderRadius: BorderRadius.circular(9),
                ),
              ),
            ),
            const SizedBox(height: 18),
            Text(
              title,
              style: const TextStyle(
                color: Color(0xff0f172a),
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 5),
            Text(
              subtitle,
              style: const TextStyle(
                color: Color(0xff64748b),
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 20),
            child,
          ],
        ),
      ),
    ),
  );
}

class _AccountFormSheet extends StatefulWidget {
  const _AccountFormSheet({required this.api, required this.onSuccess});
  final ApiClient api;
  final VoidCallback onSuccess;

  @override
  State<_AccountFormSheet> createState() => _AccountFormSheetState();
}

class _AccountFormSheetState extends State<_AccountFormSheet> {
  String selectedType = 'cash';
  final nameController = TextEditingController(text: 'Ví Tiền mặt');
  final balanceController = TextEditingController();
  bool submitting = false;

  final accountTypes = const [
    {
      'key': 'cash',
      'title': 'Tiền mặt',
      'icon': Icons.payments_rounded,
      'color': Color(0xff4ade80),
      'defaultName': 'Ví Tiền mặt',
    },
    {
      'key': 'bank',
      'title': 'Ngân hàng',
      'icon': Icons.account_balance_rounded,
      'color': Color(0xff38bdf8),
      'defaultName': 'Tài khoản Ngân hàng',
    },
    {
      'key': 'gold',
      'title': 'Vàng',
      'icon': Icons.monetization_on_rounded,
      'color': Color(0xfffbbf24),
      'defaultName': 'Vàng SJC 9999',
    },
    {
      'key': 'real_estate',
      'title': 'Đất / BĐS',
      'icon': Icons.home_work_rounded,
      'color': Color(0xfff43f5e),
      'defaultName': 'Bất động sản / Đất',
    },
  ];

  Color _accountTypeColor(String key) => switch (key) {
    'cash' => const Color(0xff15803d),
    'bank' => const Color(0xff0369a1),
    'gold' => const Color(0xffa16207),
    _ => const Color(0xffbe123c),
  };

  void _onSelectType(Map<String, dynamic> item) {
    setState(() {
      final oldDefault =
          accountTypes.firstWhere(
                (x) => x['key'] == selectedType,
                orElse: () => accountTypes.first,
              )['defaultName']
              as String;

      selectedType = item['key'] as String;

      if (nameController.text.trim().isEmpty ||
          nameController.text == oldDefault) {
        nameController.text = item['defaultName'] as String;
      }
    });
  }

  Future<void> _submit() async {
    setState(() => submitting = true);
    try {
      final selectedItem = accountTypes.firstWhere(
        (x) => x['key'] == selectedType,
        orElse: () => accountTypes.first,
      );
      final finalName = nameController.text.trim().isNotEmpty
          ? nameController.text.trim()
          : selectedItem['defaultName'] as String;

      final payload = <String, dynamic>{
        'name': finalName,
        'type': selectedType,
        'currency': 'VND',
      };

      if (balanceController.text.trim().isNotEmpty) {
        payload['balance'] = balanceController.text.trim();
      }

      await widget.api.request('POST', '/accounts', payload);
      if (mounted) widget.onSuccess();
    } catch (e) {
      if (mounted) showError(context, e.toString());
    } finally {
      if (mounted) setState(() => submitting = false);
    }
  }

  @override
  void dispose() {
    nameController.dispose();
    balanceController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FinoraSheet(
      title: 'Thêm Tài khoản',
      subtitle:
          'Chọn loại tài khoản & số dư. Thông tin khác sẽ được tự động nội suy.',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'LOẠI TÀI KHOẢN / TÀI SẢN',
            style: TextStyle(
              color: Color(0xff92400e),
              fontSize: 11,
              fontWeight: FontWeight.w800,
              letterSpacing: 1.1,
            ),
          ),
          const SizedBox(height: 12),
          GridView.count(
            crossAxisCount: 2,
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            mainAxisSpacing: 10,
            crossAxisSpacing: 10,
            childAspectRatio: 2.3,
            children: accountTypes.map((item) {
              final isSelected = selectedType == item['key'];
              final color = item['color'] as Color;
              final foreground = _accountTypeColor(item['key'] as String);
              return InkWell(
                onTap: () => _onSelectType(item),
                borderRadius: BorderRadius.circular(18),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: isSelected
                        ? color.withValues(alpha: 0.2)
                        : const Color(0xfff1f5f9),
                    borderRadius: BorderRadius.circular(18),
                    border: Border.all(
                      color: isSelected ? foreground : const Color(0xff94a3b8),
                      width: isSelected ? 2.2 : 1.4,
                    ),
                    boxShadow: isSelected
                        ? [
                            BoxShadow(
                              color: foreground.withValues(alpha: 0.18),
                              blurRadius: 10,
                            ),
                          ]
                        : null,
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: color.withValues(alpha: 0.18),
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          item['icon'] as IconData,
                          color: foreground,
                          size: 18,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          item['title'] as String,
                          style: TextStyle(
                            color: const Color(0xff0f172a),
                            fontWeight: isSelected
                                ? FontWeight.bold
                                : FontWeight.w600,
                            fontSize: 13,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      if (isSelected)
                        Icon(
                          Icons.check_circle_rounded,
                          color: foreground,
                          size: 16,
                        ),
                    ],
                  ),
                ),
              );
            }).toList(),
          ),
          const SizedBox(height: 16),
          _CustomGlassTextField(
            controller: balanceController,
            keyboardType: TextInputType.number,
            labelText: 'Số dư / Giá trị ban đầu (VND)',
            icon: Icons.payments_outlined,
          ),
          const SizedBox(height: 12),
          _CustomGlassTextField(
            controller: nameController,
            labelText: 'Tên tài khoản (Tự động gợi ý)',
            icon: Icons.edit_note_rounded,
          ),
          const SizedBox(height: 18),
          _AnimatedGoldButton(
            busy: submitting,
            label: 'Lưu tài khoản',
            onTap: _submit,
          ),
        ],
      ),
    );
  }
}

class FinoraSurface extends StatelessWidget {
  const FinoraSurface({super.key, required this.child});
  final Widget child;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
    clipBehavior: Clip.antiAlias,
    decoration: BoxDecoration(
      color: Colors.white,
      borderRadius: BorderRadius.circular(16),
      border: Border.all(color: Colors.white, width: 1),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.05),
          blurRadius: 12,
          offset: const Offset(0, 3),
        ),
      ],
    ),
    child: child,
  );
}

class _ScreenIntro extends StatelessWidget {
  const _ScreenIntro(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: Text(
      text,
      style: TextStyle(
        color: Colors.white.withValues(alpha: 0.85),
        height: 1.35,
        fontSize: 12.5,
        fontWeight: FontWeight.w600,
      ),
    ),
  );
}

class SectionTitle extends StatelessWidget {
  const SectionTitle(this.text, {super.key, required this.icon});
  final String text;
  final IconData icon;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 8),
    child: Row(
      children: [
        Container(
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.2),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: const Color(0xfffbbf24), size: 16),
        ),
        const SizedBox(width: 8),
        Text(
          text,
          style: const TextStyle(
            fontWeight: FontWeight.w800,
            color: Colors.white,
            fontSize: 14.5,
          ),
        ),
      ],
    ),
  );
}

class FinoraListTile extends StatelessWidget {
  const FinoraListTile({
    super.key,
    required this.icon,
    required this.title,
    required this.subtitle,
    this.amount,
    this.badge,
    this.iconColor,
    this.onDelete,
  });
  final IconData icon;
  final String title, subtitle;
  final String? amount, badge;
  final Color? iconColor;
  final VoidCallback? onDelete;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(bottom: 6),
    child: FinoraSurface(
      child: Row(
        children: [
          Container(
            width: 36,
            height: 36,
            decoration: BoxDecoration(
              color: (iconColor ?? const Color(0xffd97706)).withValues(
                alpha: 0.12,
              ),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(
              icon,
              color: iconColor ?? const Color(0xffd97706),
              size: 18,
            ),
          ),
          const SizedBox(width: 11),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    color: Color(0xff0f172a),
                    fontSize: 13,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  subtitle,
                  style: const TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w500,
                    color: Color(0xff64748b),
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
          if (amount != null)
            Text(
              amount!,
              style: TextStyle(
                fontWeight: FontWeight.w900,
                color: iconColor ?? const Color(0xffd97706),
                fontSize: 13,
              ),
            ),
          if (badge != null)
            Container(
              margin: const EdgeInsets.only(left: 8),
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
              decoration: BoxDecoration(
                color: (iconColor ?? const Color(0xffd97706)).withValues(
                  alpha: 0.14,
                ),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                badge!,
                style: TextStyle(
                  color: iconColor ?? const Color(0xffd97706),
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          if (onDelete != null)
            Padding(
              padding: const EdgeInsets.only(left: 6),
              child: IconButton(
                onPressed: onDelete,
                tooltip: 'Xóa tài khoản',
                icon: const Icon(
                  Icons.delete_outline_rounded,
                  color: Color(0xfff87171),
                  size: 20,
                ),
              ),
            ),
        ],
      ),
    ),
  );
}

IconData _iconForTitle(String title) => switch (title) {
  'Tài khoản' => Icons.account_balance_rounded,
  'Khoản vay' => Icons.request_quote_rounded,
  'Tài sản' => Icons.inventory_2_rounded,
  'Bất động sản' => Icons.home_work_rounded,
  'Danh mục' => Icons.category_rounded,
  'Dự báo' => Icons.auto_graph_rounded,
  'Trợ lý AI' => Icons.smart_toy_rounded,
  _ => Icons.bolt_rounded,
};

class ErrorBox extends StatelessWidget {
  const ErrorBox(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext c) => Padding(
    padding: const EdgeInsets.only(bottom: 12),
    child: Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0x44ef4444),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0x88ef4444)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.info_outline_rounded, color: Color(0xfffca5a5)),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              text,
              style: const TextStyle(
                color: Colors.white,
                height: 1.35,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    ),
  );
}

class EmptyState extends StatelessWidget {
  const EmptyState(this.text, {super.key});
  final String text;
  @override
  Widget build(BuildContext c) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 46, horizontal: 24),
    child: Center(
      child: Column(
        children: [
          Container(
            width: 62,
            height: 62,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: Colors.white.withValues(alpha: 0.1),
            ),
            child: const Icon(
              Icons.inbox_rounded,
              color: Color(0xfff7d070),
              size: 29,
            ),
          ),
          const SizedBox(height: 14),
          Text(
            text,
            textAlign: TextAlign.center,
            style: const TextStyle(
              fontWeight: FontWeight.w900,
              color: Colors.white,
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 5),
          Text(
            'Dữ liệu mới sẽ xuất hiện tại đây.',
            textAlign: TextAlign.center,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.7),
              fontSize: 13,
            ),
          ),
        ],
      ),
    ),
  );
}

void showError(BuildContext c, String e) =>
    ScaffoldMessenger.of(c).showSnackBar(
      SnackBar(
        behavior: SnackBarBehavior.floating,
        backgroundColor: const Color(0xff200733),
        content: Text(e, style: const TextStyle(color: Colors.white)),
      ),
    );
