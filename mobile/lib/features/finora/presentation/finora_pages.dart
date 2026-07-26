import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/theme/finora_colors.dart';
import 'package:mobile/features/auth/presentation/view_models/login_view_model.dart';

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
      'workspaceLabel': 'Tên workspace',
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
      'workspaceLabel': 'Workspace Name',
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
      'workspaceLabel': 'ワークスペース名',
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
      'workspaceLabel': '워크스페이스 이름',
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

class _LoginPageState extends State<LoginPage> with TickerProviderStateMixin {
  final email = TextEditingController(text: 'demo@wealthos.vn');
  final password = TextEditingController(text: 'demo-pass');
  final name = TextEditingController();
  final workspace = TextEditingController(text: 'My Finora');
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
    _headerSlide = Tween<Offset>(
      begin: const Offset(0, -0.2),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.0, 0.6, curve: Curves.easeOutCubic),
    ));

    _formFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.2, 0.7, curve: Curves.easeOut),
    );
    _formSlide = Tween<Offset>(
      begin: const Offset(0, 0.15),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.2, 0.8, curve: Curves.easeOutCubic),
    ));

    _bottomNavFade = CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.45, 0.9, curve: Curves.easeOut),
    );
    _bottomNavSlide = Tween<Offset>(
      begin: const Offset(0, 0.25),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _entranceController,
      curve: const Interval(0.45, 0.95, curve: Curves.easeOutCubic),
    ));

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
      workspaceName: workspace.text,
    );
    if (authenticated && mounted) {
      Navigator.of(
        context,
      ).pushReplacement(MaterialPageRoute(builder: widget.homeBuilder));
    }
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
    workspace.dispose();
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
                    workspace: workspace,
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
                          constraints: const BoxConstraints(maxWidth: 460),
                          child: Padding(
                            padding: const EdgeInsets.all(24),
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
                                      onOpenNotifications:
                                          _showNotificationSheet,
                                    ),
                                  ),
                                ),
                                const SizedBox(height: 24),
                                FadeTransition(
                                  opacity: _formFade,
                                  child: SlideTransition(
                                    position: _formSlide,
                                    child: form,
                                  ),
                                ),
                                const SizedBox(height: 24),
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
                          ),
                        ),
                      ),
                      Expanded(
                        child: Center(
                          child: SingleChildScrollView(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 20,
                              vertical: 12,
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
  });

  final String lang;
  final bool hasUnread;
  final VoidCallback onSelectLang;
  final VoidCallback onOpenNotifications;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: Colors.black.withValues(alpha: 0.25),
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: Colors.white.withValues(alpha: 0.3)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.15),
              blurRadius: 16,
            ),
          ],
        ),
        child: Row(
          children: [
            const _BrandMark(size: 34),
            const SizedBox(width: 10),
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
                        fontSize: 19,
                        fontWeight: FontWeight.w900,
                        letterSpacing: -0.5,
                      ),
                    ),
                    const SizedBox(width: 6),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 2,
                      ),
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [Color(0xfffbbf24), Color(0xffd97706)],
                        ),
                        borderRadius: BorderRadius.circular(5),
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
                Text(
                  _I18n.t(lang, 'subtitle'),
                  style: TextStyle(
                    color: Colors.white.withValues(alpha: 0.9),
                    fontSize: 10,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
            const Spacer(),
            // Interactive Language Selector
            InkWell(
              onTap: onSelectLang,
              borderRadius: BorderRadius.circular(20),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(
                    color: Colors.white.withValues(alpha: 0.35),
                  ),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      _I18n.getFlag(lang),
                      style: const TextStyle(fontSize: 13),
                    ),
                    const SizedBox(width: 4),
                    Text(
                      lang,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(width: 2),
                    const Icon(
                      Icons.keyboard_arrow_down_rounded,
                      color: Colors.white,
                      size: 16,
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(width: 8),
            // Interactive Notification Bell
            InkWell(
              onTap: onOpenNotifications,
              borderRadius: BorderRadius.circular(99),
              child: Stack(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
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
                      size: 18,
                    ),
                  ),
                  if (hasUnread)
                    Positioned(
                      right: 3,
                      top: 3,
                      child: Container(
                        width: 8,
                        height: 8,
                        decoration: const BoxDecoration(
                          color: Color(0xffef4444),
                          shape: BoxShape.circle,
                        ),
                      ),
                    ),
                ],
              ),
            ),
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
    required this.workspace,
    required this.obscurePassword,
    required this.lang,
    required this.onTogglePassword,
    required this.onSubmit,
    required this.onSwitch,
  });

  final bool registering, busy, obscurePassword;
  final String? error;
  final TextEditingController email, password, name, workspace;
  final String lang;
  final VoidCallback onTogglePassword, onSubmit, onSwitch;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.black.withValues(alpha: 0.32),
        borderRadius: BorderRadius.circular(28),
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.32),
          width: 1.2,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.3),
            blurRadius: 28,
            offset: const Offset(0, 10),
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
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                        const SizedBox(width: 4),
                        const Text('👋', style: TextStyle(fontSize: 13)),
                      ],
                    ),
                    const SizedBox(height: 2),
                    ShaderMask(
                      shaderCallback: (bounds) => const LinearGradient(
                        colors: [Colors.white, Color(0xfffbbf24)],
                      ).createShader(bounds),
                      child: Text(
                        registering
                            ? _I18n.t(lang, 'newAccount')
                            : (email.text.isNotEmpty
                                  ? email.text.split('@').first.toUpperCase()
                                  : 'DEMO WEALTH'),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 20,
                          fontWeight: FontWeight.w900,
                          letterSpacing: 0.5,
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
                  onPressed: onSwitch,
                  tooltip: registering ? 'Quay lại đăng nhập' : 'Đổi tài khoản',
                  icon: Icon(
                    registering
                        ? Icons.login_rounded
                        : Icons.account_circle_outlined,
                    color: const Color(0xfffbbf24),
                    size: 22,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

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
                  const SizedBox(height: 12),
                ],
                _CustomGlassTextField(
                  controller: email,
                  labelText: _I18n.t(lang, 'emailLabel'),
                  icon: Icons.alternate_email_rounded,
                  keyboardType: TextInputType.emailAddress,
                ),
                const SizedBox(height: 12),
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
                      size: 20,
                    ),
                  ),
                ),
                if (registering) ...[
                  const SizedBox(height: 12),
                  _CustomGlassTextField(
                    controller: workspace,
                    labelText: _I18n.t(lang, 'workspaceLabel'),
                    icon: Icons.dashboard_customize_outlined,
                  ),
                ],
              ],
            ),
          ),

          if (error != null)
            Padding(
              padding: const EdgeInsets.only(top: 12),
              child: Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  color: const Color(0x66ef4444),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: const Color(0xfffca5a5)),
                ),
                child: Row(
                  children: [
                    const Icon(
                      Icons.error_outline_rounded,
                      color: Color(0xfffca5a5),
                      size: 18,
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        error!,
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

          const SizedBox(height: 18),

          _AnimatedGoldButton(
            busy: busy,
            label: registering
                ? _I18n.t(lang, 'registerBtn')
                : _I18n.t(lang, 'loginBtn'),
            onTap: onSubmit,
          ),

          const SizedBox(height: 12),
          Align(
            alignment: Alignment.center,
            child: TextButton(
              onPressed: onSwitch,
              child: Text(
                registering
                    ? _I18n.t(lang, 'switchLogin')
                    : _I18n.t(lang, 'switchRegister'),
                style: const TextStyle(
                  color: Color(0xfffbbf24),
                  fontSize: 13,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ),

          const Divider(color: Colors.white24, height: 20),

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
  });

  final TextEditingController controller;
  final String labelText;
  final IconData icon;
  final bool obscureText;
  final TextInputType? keyboardType;
  final TextCapitalization textCapitalization;
  final Widget? suffixIcon;

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      obscureText: obscureText,
      keyboardType: keyboardType,
      textCapitalization: textCapitalization,
      style: const TextStyle(
        color: Colors.white,
        fontSize: 14,
        fontWeight: FontWeight.w600,
      ),
      decoration: InputDecoration(
        labelText: labelText,
        labelStyle: TextStyle(
          color: Colors.white.withValues(alpha: 0.95),
          fontSize: 13,
          fontWeight: FontWeight.w600,
        ),
        filled: true,
        fillColor: Colors.black.withValues(alpha: 0.22),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 16,
          vertical: 14,
        ),
        prefixIcon: Icon(icon, color: const Color(0xfffbbf24), size: 20),
        suffixIcon: suffixIcon,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: Colors.white.withValues(alpha: 0.3)),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: Colors.white.withValues(alpha: 0.3)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: const BorderSide(color: Color(0xfffbbf24), width: 1.8),
        ),
      ),
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
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.18),
            shape: BoxShape.circle,
            border: Border.all(color: Colors.white.withValues(alpha: 0.3)),
          ),
          child: Icon(icon, color: const Color(0xfffbbf24), size: 20),
        ),
        const SizedBox(height: 6),
        Text(
          label,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 11,
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
          height: 52,
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [Color(0xfffbbf24), Color(0xffd97706)],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(26),
            boxShadow: [
              BoxShadow(
                color: const Color(0xffd97706).withValues(alpha: _pressed ? 0.25 : 0.45),
                blurRadius: _pressed ? 8 : 18,
                offset: _pressed ? const Offset(0, 2) : const Offset(0, 6),
              ),
            ],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              borderRadius: BorderRadius.circular(26),
              onTap: widget.busy ? null : widget.onTap,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  if (widget.busy)
                    const SizedBox(
                      width: 22,
                      height: 22,
                      child: CircularProgressIndicator(
                        color: Color(0xff1c1917),
                        strokeWidth: 2.5,
                      ),
                    )
                  else ...[
                    Container(
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.15),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.arrow_forward_rounded,
                        color: Color(0xff1c1917),
                        size: 20,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Text(
                      widget.label,
                      style: const TextStyle(
                        color: Color(0xff1c1917),
                        fontSize: 16,
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

class HomePage extends StatefulWidget {
  const HomePage({super.key, required this.api, required this.loginBuilder});
  final ApiClient api;
  final WidgetBuilder loginBuilder;
  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int index = 0;
  final pages = const [
    NavItem('Tổng quan', Icons.dashboard_rounded),
    NavItem('Tài khoản', Icons.account_balance_rounded),
    NavItem('Giao dịch', Icons.receipt_long_rounded),
    NavItem('Khoản vay', Icons.request_quote_rounded),
    NavItem('Tài sản', Icons.inventory_2_rounded),
    NavItem('Bất động sản', Icons.home_work_rounded),
    NavItem('Ngân sách', Icons.pie_chart_rounded),
    NavItem('Dự báo', Icons.auto_graph_rounded),
    NavItem('Danh mục', Icons.workspaces_rounded),
    NavItem('Ngân hàng', Icons.account_balance_wallet_rounded),
    NavItem('Tự động hóa', Icons.bolt_rounded),
    NavItem('Trợ lý AI', Icons.smart_toy_rounded),
    NavItem('Nhật ký audit', Icons.history_rounded),
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
        actions: [
          IconButton(
            tooltip: 'Chuyển workspace',
            onPressed: _selectWorkspace,
            icon: const Icon(Icons.grid_view_rounded, color: Color(0xfffbbf24)),
          ),
          IconButton(
            tooltip: 'Đăng xuất',
            onPressed: _logout,
            icon: const Icon(Icons.logout_rounded, color: Colors.white70),
          ),
        ],
      ),
      drawer: compact
          ? Drawer(backgroundColor: const Color(0xee120320), child: _nav())
          : null,
      body: Stack(
        children: [
          Positioned.fill(
            child: Image.asset(
              'assets/images/login_bg.png',
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
      bottomNavigationBar: compact
          ? Container(
              color: Colors.black.withValues(alpha: 0.35),
              child: NavigationBarTheme(
                data: NavigationBarThemeData(
                  backgroundColor: Colors.transparent,
                  indicatorColor: const Color(0xfffbbf24),
                  labelTextStyle: WidgetStateProperty.resolveWith(
                    (states) => TextStyle(
                      color: states.contains(WidgetState.selected)
                          ? const Color(0xfffbbf24)
                          : Colors.white70,
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  iconTheme: WidgetStateProperty.resolveWith(
                    (states) => IconThemeData(
                      color: states.contains(WidgetState.selected)
                          ? const Color(0xff1c1917)
                          : Colors.white70,
                    ),
                  ),
                ),
                child: NavigationBar(
                  selectedIndex: index > 4 ? 0 : index,
                  onDestinationSelected: (i) => setState(() => index = i),
                  destinations: pages
                      .take(5)
                      .map(
                        (p) => NavigationDestination(
                          icon: Icon(p.icon),
                          label: p.title,
                        ),
                      )
                      .toList(),
                ),
              ),
            )
          : null,
    );
  }

  Widget _nav() => ListView(
    padding: const EdgeInsets.fromLTRB(12, 86, 12, 12),
    children: [
      const Padding(
        padding: EdgeInsets.fromLTRB(14, 4, 14, 14),
        child: Text(
          'KHÔNG GIAN TÀI CHÍNH',
          style: TextStyle(
            fontSize: 11,
            fontWeight: FontWeight.w800,
            letterSpacing: 1.1,
            color: Color(0xfffbbf24),
          ),
        ),
      ),
      ...pages.asMap().entries.map(
        (e) => Padding(
          padding: const EdgeInsets.only(bottom: 3),
          child: ListTile(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
            ),
            selectedColor: const Color(0xff1c1917),
            selectedTileColor: const Color(0xfffbbf24),
            selected: index == e.key,
            leading: Icon(
              e.value.icon,
              color: index == e.key ? const Color(0xff200733) : Colors.white70,
            ),
            title: Text(
              e.value.title,
              style: TextStyle(
                fontWeight: FontWeight.w700,
                color: index == e.key ? const Color(0xff200733) : Colors.white,
              ),
            ),
            onTap: () {
              setState(() => index = e.key);
              if (MediaQuery.of(context).size.width < 700) {
                Navigator.pop(context);
              }
            },
          ),
        ),
      ),
    ],
  );

  Widget _body() {
    switch (index) {
      case 0:
        return DashboardPage(api: widget.api);
      case 1:
        return ResourcePage(
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
        return TransactionsPage(api: widget.api);
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
      default:
        return ReadonlyPage(
          api: widget.api,
          title: 'Nhật ký audit',
          path: '/audit-logs',
        );
    }
  }

  Future<void> _selectWorkspace() async {
    final data = await widget.api.request('GET', '/workspaces') as List;
    if (!mounted) return;
    showModalBottomSheet(
      context: context,
      backgroundColor: const Color(0xff200733),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      builder: (_) => Container(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'CHỌN WORKSPACE',
              style: TextStyle(
                color: Color(0xfff7d070),
                fontWeight: FontWeight.bold,
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 14),
            ...data.map(
              (x) => ListTile(
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(14),
                ),
                tileColor: Colors.white.withValues(alpha: 0.08),
                title: Text(
                  x['name']?.toString() ?? '-',
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                subtitle: Text(
                  x['baseCurrency']?.toString() ?? '',
                  style: TextStyle(color: Colors.white.withValues(alpha: 0.7)),
                ),
                trailing: const Icon(
                  Icons.arrow_forward_ios_rounded,
                  color: Color(0xfff7d070),
                  size: 16,
                ),
                onTap: () {
                  widget.api.workspaceId = x['id'].toString();
                  Navigator.pop(context);
                  setState(() {});
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _logout() {
    widget.api.token = null;
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(builder: widget.loginBuilder),
      (_) => false,
    );
  }
}

class NavItem {
  const NavItem(this.title, this.icon);
  final String title;
  final IconData icon;
}

class FieldSpec {
  const FieldSpec(this.key, this.label, {this.initial = ''});
  final String key, label, initial;
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
  const DashboardPage({super.key, required this.api});
  final ApiClient api;
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
        icon: const Icon(Icons.refresh_rounded, color: Color(0xfff7d070)),
      ),
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfff7d070)),
          )
        : RefreshIndicator(
            color: const Color(0xfff7d070),
            backgroundColor: const Color(0xff200733),
            onRefresh: load,
            child: ListView(
              children: [
                if (error != null) ErrorBox(error!),
                _BalanceHero(
                  value: _formatMoney(netWorth?['netWorth']),
                  currency: netWorth?['baseCurrency']?.toString() ?? 'VND',
                  accountCount: accounts.length,
                ),
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
                const SizedBox(height: 24),
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
                    title: x['note']?.toString().isNotEmpty == true
                        ? x['note'].toString()
                        : (x['type'] == 'income' ? 'Thu nhập' : 'Chi tiêu'),
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
    final parts = d.toStringAsFixed(2).split('.');
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
    decoration: BoxDecoration(
      borderRadius: BorderRadius.circular(28),
      color: Colors.black.withValues(alpha: 0.35),
      border: Border.all(
        color: Colors.white.withValues(alpha: 0.35),
        width: 1.2,
      ),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.3),
          blurRadius: 28,
          offset: const Offset(0, 10),
        ),
      ],
    ),
    child: Stack(
      children: [
        Positioned(
          right: -18,
          top: -32,
          child: Container(
            width: 140,
            height: 140,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: const Color(0xfffbbf24).withValues(alpha: 0.12),
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
                    color: Colors.white.withValues(alpha: 0.14),
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
                    fontSize: 11,
                    letterSpacing: 1.2,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            ShaderMask(
              shaderCallback: (bounds) => const LinearGradient(
                colors: [Colors.white, Color(0xfffbbf24), Colors.white],
              ).createShader(bounds),
              child: Text(
                hideBalance ? '••••••••••••' : widget.value,
                style: const TextStyle(
                  fontSize: 28,
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  letterSpacing: -0.5,
                ),
              ),
            ),
            Text(
              widget.currency,
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.8),
                fontWeight: FontWeight.bold,
                fontSize: 13,
              ),
            ),
            const SizedBox(height: 18),
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.14),
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(
                      color: Colors.white.withValues(alpha: 0.2),
                    ),
                  ),
                  child: Text(
                    '${widget.accountCount} tài khoản đang theo dõi',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
                const Spacer(),
                IconButton(
                  onPressed: () => setState(() => hideBalance = !hideBalance),
                  icon: Icon(
                    hideBalance
                        ? Icons.visibility_off_outlined
                        : Icons.visibility_outlined,
                    color: const Color(0xfffbbf24),
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
    decoration: BoxDecoration(
      color: const Color(0x661e0734),
      borderRadius: BorderRadius.circular(22),
      border: Border.all(color: Colors.white.withValues(alpha: 0.18)),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.25),
          blurRadius: 16,
          offset: const Offset(0, 6),
        ),
      ],
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: accent.withValues(alpha: 0.18),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(icon, color: accent, size: 20),
        ),
        const SizedBox(height: 14),
        Text(
          value,
          style: const TextStyle(
            fontWeight: FontWeight.w900,
            color: Colors.white,
            fontSize: 16,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: TextStyle(
            color: Colors.white.withValues(alpha: 0.75),
            fontWeight: FontWeight.w600,
            fontSize: 11,
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
        subtitle: 'Thông tin sẽ được lưu vào workspace hiện tại.',
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
                final name = x['name']?.toString() ??
                    x['counterparty']?.toString() ??
                    x['id']?.toString() ??
                    '-';
                final tile = FinoraListTile(
                  icon: _iconForTitle(widget.title),
                  title: name,
                  subtitle: _details(x),
                  badge: x['status']?.toString() ??
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
  String _details(dynamic x) => x is Map
      ? x.entries
            .where((e) => !['id', 'workspaceId', 'name'].contains(e.key))
            .take(3)
            .map((e) => '${e.key}: ${e.value}')
            .join(' • ')
      : '';
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
  final amount = TextEditingController(),
      note = TextEditingController(),
      account = TextEditingController(),
      currency = TextEditingController(text: 'VND');
  String type = 'expense';
  @override
  void initState() {
    super.initState();
    load();
  }

  @override
  void dispose() {
    amount.dispose();
    note.dispose();
    account.dispose();
    currency.dispose();
    super.dispose();
  }

  Future<void> load() async {
    setState(() => loading = true);
    try {
      final x =
          await widget.api.request('GET', '/transactions?limit=50') as Map;
      items = x['items'] as List? ?? [];
      error = null;
    } catch (e) {
      error = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> create() async {
    try {
      await widget.api.request('POST', '/transactions', {
        'accountId': account.text,
        'type': type,
        'amount': amount.text,
        'currency': currency.text,
        'note': note.text,
        'status': 'posted',
      });
      if (mounted) Navigator.pop(context);
      load();
    } catch (e) {
      if (mounted) showError(context, e.toString());
    }
  }

  void form() => showModalBottomSheet(
    isScrollControlled: true,
    backgroundColor: Colors.transparent,
    context: context,
    builder: (_) => FinoraSheet(
      title: 'Tạo giao dịch',
      subtitle: 'Ghi nhận dòng tiền cho tài khoản của bạn.',
      child: Column(
        children: [
          DropdownButtonFormField<String>(
            initialValue: type,
            dropdownColor: const Color(0xff200733),
            style: const TextStyle(color: Colors.white),
            items: const [
              DropdownMenuItem(value: 'expense', child: Text('Chi tiêu')),
              DropdownMenuItem(value: 'income', child: Text('Thu nhập')),
            ],
            onChanged: (v) => type = v!,
            decoration: InputDecoration(
              labelText: 'Loại',
              filled: true,
              fillColor: Colors.white.withValues(alpha: 0.14),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(16),
              ),
            ),
          ),
          const SizedBox(height: 12),
          _CustomGlassTextField(
            controller: account,
            labelText: 'Account ID',
            icon: Icons.account_balance_rounded,
          ),
          const SizedBox(height: 12),
          _CustomGlassTextField(
            controller: amount,
            keyboardType: TextInputType.number,
            labelText: 'Số tiền',
            icon: Icons.attach_money_rounded,
          ),
          const SizedBox(height: 12),
          _CustomGlassTextField(
            controller: currency,
            labelText: 'Tiền tệ',
            icon: Icons.currency_exchange_rounded,
          ),
          const SizedBox(height: 12),
          _CustomGlassTextField(
            controller: note,
            labelText: 'Ghi chú',
            icon: Icons.note_rounded,
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            height: 48,
            child: FilledButton.icon(
              style: FilledButton.styleFrom(
                backgroundColor: const Color(0xfff7d070),
                foregroundColor: const Color(0xff200733),
              ),
              onPressed: create,
              icon: const Icon(Icons.check_rounded),
              label: const Text(
                'Lưu giao dịch',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
            ),
          ),
        ],
      ),
    ),
  );
  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Giao dịch',
    action: Row(
      children: [
        IconButton(
          onPressed: load,
          icon: const Icon(Icons.refresh_rounded, color: Color(0xfff7d070)),
        ),
        FilledButton.icon(
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xfff7d070),
            foregroundColor: const Color(0xff200733),
          ),
          onPressed: form,
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
            child: CircularProgressIndicator(color: Color(0xfff7d070)),
          )
        : ListView(
            children: [
              const _ScreenIntro(
                'Theo dõi mọi khoản thu chi theo thời gian thực.',
              ),
              if (error != null) ErrorBox(error!),
              ...items.map(
                (x) => FinoraListTile(
                  icon: x['type'] == 'income'
                      ? Icons.south_west_rounded
                      : Icons.north_east_rounded,
                  iconColor: x['type'] == 'income'
                      ? const Color(0xff4ade80)
                      : const Color(0xfffb7185),
                  title:
                      x['note']?.toString() ??
                      (x['type'] == 'income' ? 'Thu nhập' : 'Chi tiêu'),
                  subtitle: x['occurredAt']?.toString() ?? '',
                  amount:
                      "${x['type'] == 'income' ? '+' : '-'}${x['amount'] ?? ''} ${x['currency'] ?? 'VND'}",
                ),
              ),
              if (items.isEmpty) const EmptyState('Chưa có giao dịch'),
            ],
          ),
  );
}

class ValuedResourcePage extends StatelessWidget {
  const ValuedResourcePage({
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
  final String title, path;
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
      final x = await widget.api.request('GET', widget.path);
      data = x is List ? x : ((x as Map)['items'] as List? ?? []);
    } catch (e) {
      err = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: widget.title,
    action: IconButton(
      onPressed: load,
      icon: const Icon(Icons.refresh_rounded, color: Color(0xfff7d070)),
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfff7d070)),
          )
        : ListView(
            children: [
              const _ScreenIntro(
                'Dấu vết hoạt động được lưu để đảm bảo minh bạch.',
              ),
              if (err != null) ErrorBox(err!),
              ...data.map(
                (x) => FinoraSurface(
                  child: Text(
                    JsonEncoder.withIndent('  ').convert(x),
                    style: const TextStyle(
                      fontFamily: 'monospace',
                      fontSize: 12,
                      color: Colors.white,
                    ),
                  ),
                ),
              ),
              if (data.isEmpty) const EmptyState('Chưa có dữ liệu'),
            ],
          ),
  );
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

class ForecastPage extends StatelessWidget {
  const ForecastPage({super.key, required this.api});
  final ApiClient api;
  @override
  Widget build(BuildContext c) => ScenarioPage(
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
  Widget build(BuildContext c) => ScenarioPage(
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
  Widget build(BuildContext c) => ScenarioPage(
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
  final String title, path;
  final List<FieldSpec> fields;
  @override
  Widget build(BuildContext c) =>
      ResourcePage(api: api, title: title, path: path, fields: fields);
}

class BankPage extends StatefulWidget {
  const BankPage({super.key, required this.api});
  final ApiClient api;
  @override
  State<BankPage> createState() => _BankPageState();
}

class _BankPageState extends State<BankPage> {
  List con = [], feed = [];
  String? err;
  bool loading = true;
  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      con = await widget.api.request('GET', '/bank-connections') as List;
      feed = await widget.api.request('GET', '/bank-feed-transactions') as List;
      err = null;
    } catch (e) {
      err = e.toString();
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  Future<void> connect() async {
    try {
      final x = await widget.api.request(
        'POST',
        '/integrations/sepay/connect',
        {'provider': 'sepay', 'scope': 'read_transactions'},
      );
      if (mounted) {
        showDialog(
          context: context,
          builder: (_) => AlertDialog(
            backgroundColor: const Color(0xff200733),
            title: const Text(
              'Kết nối SePay',
              style: TextStyle(color: Colors.white),
            ),
            content: Text(
              'Mở URL này trong trình duyệt để hoàn tất:\n${x['connectUrl'] ?? ''}',
              style: const TextStyle(color: Colors.white70),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context),
                child: const Text(
                  'Đóng',
                  style: TextStyle(color: Color(0xfff7d070)),
                ),
              ),
            ],
          ),
        );
      }
      load();
    } catch (e) {
      if (mounted) showError(context, e.toString());
    }
  }

  @override
  Widget build(BuildContext c) => PageFrame(
    title: 'Ngân hàng & SePay',
    action: FilledButton.icon(
      style: FilledButton.styleFrom(
        backgroundColor: const Color(0xfff7d070),
        foregroundColor: const Color(0xff200733),
      ),
      onPressed: connect,
      icon: const Icon(Icons.add_link_rounded),
      label: const Text(
        'Kết nối',
        style: TextStyle(fontWeight: FontWeight.bold),
      ),
    ),
    child: loading
        ? const Center(
            child: CircularProgressIndicator(color: Color(0xfff7d070)),
          )
        : ListView(
            children: [
              const _ScreenIntro(
                'Liên kết nguồn tiền để theo dõi giao dịch liền mạch.',
              ),
              if (err != null) ErrorBox(err!),
              const SectionTitle(
                'Kết nối ngân hàng',
                icon: Icons.account_balance_rounded,
              ),
              ...con.map(
                (x) => FinoraListTile(
                  icon: Icons.account_balance_rounded,
                  title: x['provider']?.toString() ?? '',
                  subtitle: 'Trạng thái kết nối',
                  badge: x['status']?.toString() ?? '',
                ),
              ),
              const SizedBox(height: 18),
              const SectionTitle(
                'Giao dịch ngân hàng',
                icon: Icons.swap_horiz_rounded,
              ),
              ...feed.map(
                (x) => FinoraListTile(
                  icon: Icons.receipt_long_rounded,
                  title: x['description']?.toString() ?? '',
                  subtitle: x['postingState']?.toString() ?? '',
                  amount: x['amount']?.toString() ?? '',
                ),
              ),
              if (con.isEmpty && feed.isEmpty)
                const EmptyState('Chưa có kết nối ngân hàng'),
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
        color: Colors.black.withValues(alpha: 0.85),
        borderRadius: BorderRadius.circular(28),
        border: Border.all(
          color: Colors.white.withValues(alpha: 0.32),
          width: 1.2,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.4),
            blurRadius: 30,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                width: 42,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.white30,
                  borderRadius: BorderRadius.circular(9),
                ),
              ),
            ),
            const SizedBox(height: 18),
            Text(
              title,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 5),
            Text(
              subtitle,
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.75),
                fontSize: 13,
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

  void _onSelectType(Map<String, dynamic> item) {
    setState(() {
      final oldDefault = accountTypes.firstWhere(
        (x) => x['key'] == selectedType,
        orElse: () => accountTypes.first,
      )['defaultName'] as String;

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
              color: Color(0xfffbbf24),
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
              return InkWell(
                onTap: () => _onSelectType(item),
                borderRadius: BorderRadius.circular(18),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: isSelected
                        ? color.withValues(alpha: 0.22)
                        : Colors.white.withValues(alpha: 0.08),
                    borderRadius: BorderRadius.circular(18),
                    border: Border.all(
                      color: isSelected
                          ? color
                          : Colors.white.withValues(alpha: 0.18),
                      width: isSelected ? 1.8 : 1,
                    ),
                    boxShadow: isSelected
                        ? [
                            BoxShadow(
                              color: color.withValues(alpha: 0.25),
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
                          color: color,
                          size: 18,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          item['title'] as String,
                          style: TextStyle(
                            color: isSelected ? Colors.white : Colors.white70,
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
                          color: color,
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
    padding: const EdgeInsets.all(18),
    decoration: BoxDecoration(
      color: Colors.black.withValues(alpha: 0.32),
      borderRadius: BorderRadius.circular(24),
      border: Border.all(
        color: Colors.white.withValues(alpha: 0.28),
        width: 1.2,
      ),
      boxShadow: [
        BoxShadow(
          color: Colors.black.withValues(alpha: 0.25),
          blurRadius: 20,
          offset: const Offset(0, 8),
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
    padding: const EdgeInsets.only(bottom: 16),
    child: Text(
      text,
      style: TextStyle(
        color: Colors.white.withValues(alpha: 0.85),
        height: 1.4,
        fontWeight: FontWeight.w500,
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
    padding: const EdgeInsets.only(bottom: 10),
    child: Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.14),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Icon(icon, color: const Color(0xfffbbf24), size: 18),
        ),
        const SizedBox(width: 9),
        Text(
          text,
          style: const TextStyle(
            fontWeight: FontWeight.w900,
            color: Colors.white,
            fontSize: 16,
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
    padding: const EdgeInsets.only(bottom: 10),
    child: FinoraSurface(
      child: Row(
        children: [
          Container(
            width: 42,
            height: 42,
            decoration: BoxDecoration(
              color: (iconColor ?? const Color(0xfffbbf24)).withValues(
                alpha: 0.18,
              ),
              borderRadius: BorderRadius.circular(14),
            ),
            child: Icon(
              icon,
              color: iconColor ?? const Color(0xfffbbf24),
              size: 21,
            ),
          ),
          const SizedBox(width: 13),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontWeight: FontWeight.w900,
                    color: Colors.white,
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  subtitle,
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.white.withValues(alpha: 0.75),
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
                color: iconColor ?? const Color(0xfffbbf24),
                fontSize: 14,
              ),
            ),
          if (badge != null)
            Container(
              margin: const EdgeInsets.only(left: 8),
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
              decoration: BoxDecoration(
                color: const Color(0xfffbbf24).withValues(alpha: 0.2),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                badge!,
                style: const TextStyle(
                  color: Color(0xfffbbf24),
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
  'Danh mục' => Icons.workspaces_rounded,
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
