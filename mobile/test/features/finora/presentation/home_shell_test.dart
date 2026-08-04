import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mobile/core/network/api_client.dart';
import 'package:mobile/core/theme/theme_controller.dart';
import 'package:mobile/features/finora/presentation/finora_pages.dart';

class _FakeApiClient extends ApiClient {
  @override
  Future<dynamic> request(
    String method,
    String path, [
    Map<String, dynamic>? body,
  ]) async {
    if (path == '/accounts') {
      return [
        {'id': 'account-1', 'name': 'Tiền mặt'},
        {'id': 'account-2', 'name': 'Ngân hàng'},
      ];
    }
    if (path.startsWith('/transactions')) return {'items': []};
    if (path == '/net-worth') {
      return {
        'netWorth': '0',
        'cash': '0',
        'liabilities': '0',
        'baseCurrency': 'VND',
      };
    }
    if (path == '/portfolios') return [];
    return [];
  }
}

void main() {
  testWidgets('phone shell exposes five destinations and global quick action', (
    tester,
  ) async {
    final view = tester.view;
    final originalPhysicalSize = view.physicalSize;
    final originalDevicePixelRatio = view.devicePixelRatio;
    view.physicalSize = const Size(390, 844);
    view.devicePixelRatio = 1;
    addTearDown(() {
      view.physicalSize = originalPhysicalSize;
      view.devicePixelRatio = originalDevicePixelRatio;
    });

    await tester.pumpWidget(
      MaterialApp(
        home: HomePage(
          api: _FakeApiClient(),
          loginBuilder: (_) => const SizedBox.shrink(),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('Trang chủ'), findsAtLeastNWidgets(1));
    expect(find.text('Tài khoản'), findsAtLeastNWidgets(1));
    expect(find.text('Giao dịch'), findsAtLeastNWidgets(1));
    expect(find.text('Cá nhân'), findsAtLeastNWidgets(1));

    await tester.tap(find.bySemanticsLabel('Thao tác nhanh'));
    await tester.pumpAndSettle();

    expect(find.text('Thao tác nhanh'), findsOneWidget);
    expect(find.text('Tạo giao dịch'), findsAtLeastNWidgets(2));
    expect(find.text('Cho vay mới'), findsAtLeastNWidgets(1));
    expect(find.text('Thu lãi / gốc'), findsOneWidget);
    expect(find.text('Chuyển tiền'), findsOneWidget);
    expect(find.text('Thêm tài sản'), findsOneWidget);

    await tester.tap(find.text('Cho vay mới'));
    await tester.pump();
    await tester.pump();

    expect(find.text('Cho vay mới'), findsAtLeastNWidgets(1));
    expect(find.text('Khách hàng'), findsOneWidget);
    expect(find.text('Tìm hoặc thêm người vay'), findsOneWidget);
    expect(find.text('Tạo khoản vay'), findsOneWidget);
  });

  testWidgets('phone shell remains usable at 320px width', (tester) async {
    final view = tester.view;
    final originalPhysicalSize = view.physicalSize;
    final originalDevicePixelRatio = view.devicePixelRatio;
    view.physicalSize = const Size(320, 640);
    view.devicePixelRatio = 1;
    addTearDown(() {
      view.physicalSize = originalPhysicalSize;
      view.devicePixelRatio = originalDevicePixelRatio;
    });

    await tester.pumpWidget(
      MaterialApp(
        home: HomePage(
          api: ApiClient(),
          loginBuilder: (_) => const SizedBox.shrink(),
        ),
      ),
    );
    await tester.pump();

    expect(find.text('Trang chủ'), findsAtLeastNWidgets(1));
    expect(find.text('Tài khoản'), findsAtLeastNWidgets(1));
    expect(find.bySemanticsLabel('Thao tác nhanh'), findsOneWidget);
    expect(find.text('Giao dịch'), findsAtLeastNWidgets(1));
    expect(find.text('Cá nhân'), findsAtLeastNWidgets(1));
  });

  testWidgets(
    '3:2 tablet uses a navigation rail and keeps all pages reachable',
    (tester) async {
      final view = tester.view;
      final originalPhysicalSize = view.physicalSize;
      final originalDevicePixelRatio = view.devicePixelRatio;
      // Xiaomi Pad 8 Pro: 3:2 display. This is the portrait logical viewport.
      view.physicalSize = const Size(854, 1280);
      view.devicePixelRatio = 1;
      addTearDown(() {
        view.physicalSize = originalPhysicalSize;
        view.devicePixelRatio = originalDevicePixelRatio;
      });

      await tester.pumpWidget(
        MaterialApp(
          home: HomePage(
            api: _FakeApiClient(),
            loginBuilder: (_) => const SizedBox.shrink(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.byType(NavigationRail), findsOneWidget);
      expect(find.bySemanticsLabel('Thao tác nhanh'), findsOneWidget);

      await tester.tap(find.byIcon(Icons.grid_view_rounded));
      await tester.pumpAndSettle();

      expect(find.text('Tất cả tiện ích'), findsOneWidget);
      expect(find.text('Bất động sản'), findsOneWidget);
    },
  );

  testWidgets('iPad landscape shows the full catalogue beside content', (
    tester,
  ) async {
    final view = tester.view;
    final originalPhysicalSize = view.physicalSize;
    final originalDevicePixelRatio = view.devicePixelRatio;
    view.physicalSize = const Size(1024, 768);
    view.devicePixelRatio = 1;
    addTearDown(() {
      view.physicalSize = originalPhysicalSize;
      view.devicePixelRatio = originalDevicePixelRatio;
    });

    await tester.pumpWidget(
      MaterialApp(
        home: HomePage(
          api: _FakeApiClient(),
          loginBuilder: (_) => const SizedBox.shrink(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('DANH MỤC TIỆN ÍCH'), findsOneWidget);
    expect(find.byType(NavigationRail), findsNothing);
    expect(find.text('Tài sản và nghĩa vụ'), findsOneWidget);
  });

  testWidgets('quick action opens the internal transfer form', (tester) async {
    final view = tester.view;
    final originalPhysicalSize = view.physicalSize;
    final originalDevicePixelRatio = view.devicePixelRatio;
    view.physicalSize = const Size(390, 844);
    view.devicePixelRatio = 1;
    addTearDown(() {
      view.physicalSize = originalPhysicalSize;
      view.devicePixelRatio = originalDevicePixelRatio;
    });

    await tester.pumpWidget(
      MaterialApp(
        home: HomePage(
          api: _FakeApiClient(),
          loginBuilder: (_) => const SizedBox.shrink(),
        ),
      ),
    );
    await tester.pump();
    await tester.tap(find.bySemanticsLabel('Thao tác nhanh'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Chuyển tiền'));
    await tester.pumpAndSettle();

    expect(find.text('Chuyển tiền'), findsAtLeastNWidgets(1));
    expect(find.text('Từ tài khoản'), findsOneWidget);
    expect(find.text('Đến tài khoản'), findsOneWidget);
    expect(find.text('Xác nhận chuyển tiền'), findsOneWidget);
  });

  testWidgets('quick action opens the asset creation form', (tester) async {
    final view = tester.view;
    final originalPhysicalSize = view.physicalSize;
    final originalDevicePixelRatio = view.devicePixelRatio;
    view.physicalSize = const Size(390, 844);
    view.devicePixelRatio = 1;
    addTearDown(() {
      view.physicalSize = originalPhysicalSize;
      view.devicePixelRatio = originalDevicePixelRatio;
    });

    await tester.pumpWidget(
      MaterialApp(
        home: HomePage(
          api: _FakeApiClient(),
          loginBuilder: (_) => const SizedBox.shrink(),
        ),
      ),
    );
    await tester.pump();
    await tester.tap(find.bySemanticsLabel('Thao tác nhanh'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Thêm tài sản'));
    await tester.pumpAndSettle();

    expect(find.text('Tên tài sản'), findsOneWidget);
    expect(find.text('Loại tài sản'), findsOneWidget);
    expect(find.text('Tạo tài sản'), findsOneWidget);
  });

  testWidgets('profile destination uses the Finora settings surface', (
    tester,
  ) async {
    final view = tester.view;
    final originalPhysicalSize = view.physicalSize;
    final originalDevicePixelRatio = view.devicePixelRatio;
    view.physicalSize = const Size(390, 844);
    view.devicePixelRatio = 1;
    addTearDown(() {
      view.physicalSize = originalPhysicalSize;
      view.devicePixelRatio = originalDevicePixelRatio;
    });

    await tester.pumpWidget(
      MaterialApp(
        home: HomePage(
          api: _FakeApiClient(),
          loginBuilder: (_) => const SizedBox.shrink(),
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Cá nhân').last);
    await tester.pumpAndSettle();

    expect(find.text('Hồ sơ Finora'), findsOneWidget);
    expect(find.text('Giao diện'), findsOneWidget);
    expect(find.text('Hiển thị số tiền'), findsOneWidget);
    expect(find.text('Liên kết ngân hàng'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Đăng xuất'),
      180,
      scrollable: find.byType(Scrollable).last,
    );
    expect(find.text('Đăng xuất'), findsOneWidget);
  });

  testWidgets(
    'dashboard keeps the executive overview without promotion cards',
    (tester) async {
      final view = tester.view;
      final originalPhysicalSize = view.physicalSize;
      final originalDevicePixelRatio = view.devicePixelRatio;
      view.physicalSize = const Size(390, 844);
      view.devicePixelRatio = 1;
      addTearDown(() {
        view.physicalSize = originalPhysicalSize;
        view.devicePixelRatio = originalDevicePixelRatio;
      });

      await tester.pumpWidget(
        MaterialApp(
          home: HomePage(
            api: _FakeApiClient(),
            loginBuilder: (_) => const SizedBox.shrink(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Tài sản và nghĩa vụ'), findsOneWidget);
      expect(find.text('Giao dịch gần đây'), findsOneWidget);
      expect(find.text('OFFER VIP'), findsNothing);
      expect(find.text('Gợi ý cho bạn'), findsNothing);
    },
  );

  testWidgets('light-only shell exposes the quick-search action', (
    tester,
  ) async {
    final view = tester.view;
    final originalPhysicalSize = view.physicalSize;
    final originalDevicePixelRatio = view.devicePixelRatio;
    view.physicalSize = const Size(390, 844);
    view.devicePixelRatio = 1;
    addTearDown(() {
      view.physicalSize = originalPhysicalSize;
      view.devicePixelRatio = originalDevicePixelRatio;
    });

    final controller = FinoraThemeController();

    await tester.pumpWidget(
      ListenableBuilder(
        listenable: controller,
        builder: (context, child) => FinoraThemeScope(
          controller: controller,
          child: MaterialApp(
            theme: ThemeData.light(),
            darkTheme: ThemeData.dark(),
            themeMode: controller.mode,
            home: HomePage(
              api: _FakeApiClient(),
              loginBuilder: (_) => const SizedBox.shrink(),
            ),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    final searchButton = find.bySemanticsLabel('Tìm kiếm');
    expect(searchButton, findsOneWidget);
    await tester.tap(searchButton);
    await tester.pumpAndSettle();

    expect(find.text('Tìm nhanh'), findsOneWidget);
    expect(find.text('Ghi giao dịch mới'), findsOneWidget);

    await tester.enterText(find.byType(TextField).last, 'ngân sách');
    await tester.pump();

    expect(find.text('Ngân sách'), findsAtLeastNWidgets(1));
    expect(find.text('Ghi giao dịch mới'), findsNothing);
    expect(controller.mode, ThemeMode.light);
  });
}
