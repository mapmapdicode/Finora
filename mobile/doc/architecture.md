# Kiến trúc Flutter của Finora

## Quy ước áp dụng

Ứng dụng dùng **feature-first MVVM + Clean Architecture thực dụng**. Quy ước
này áp dụng trực tiếp từ bài “Clean Architecture in Flutter 2026”: View,
ViewModel, Repository, Service tách trách nhiệm; domain/use-case chỉ được thêm
khi có logic phức tạp, tái sử dụng, hoặc phối hợp nhiều repository.

```
lib/
  app/                 # Composition root và điều phối dependency
  core/                # Hạ tầng dùng chung, không thuộc feature nào
    config/
    network/
    theme/
  features/
    <feature>/
      presentation/    # View (widget), ViewModel và UI state
      domain/           # Entity, repository contract, use-case (khi cần)
      data/             # DTO/model, service và repository implementation
```

## Luật phụ thuộc

```text
View → ViewModel → Repository contract → Repository implementation → Service → ApiClient
```

- **View** chỉ render, xử lý layout/animation và gửi event.
- **ViewModel** giữ `loading/error` và các command người dùng gọi.
- **Repository** là nơi chuyển DTO thành entity, quản lý session/cache/retry khi
  cần; ViewModel không biết REST endpoint.
- **Service** là wrapper không state cho một data source; không chứa UI state
  hay nghiệp vụ.

`FinoraApp` là composition root. Hiện tại dự án dùng manual DI qua
`AppDependencies` để không thêm dependency state-management trước khi cần;
toàn bộ dependency vẫn được inject qua constructor, nên có thể thay bằng
Riverpod sau này mà không đổi domain/data hoặc UI contract.

`ApiClient` là HTTP boundary duy nhất. Tránh gọi HTTP, truy cập `Platform`,
hoặc nhúng quy tắc nghiệp vụ trực tiếp trong widget.

## Feature mẫu đã áp dụng: Auth

```text
features/auth/
  domain/entities/auth_credentials.dart, auth_session.dart
  domain/repositories/auth_repository.dart
  data/services/auth_remote_service.dart
  data/repositories/auth_repository_impl.dart
  presentation/view_models/login_view_model.dart
```

Khi người dùng bấm đăng nhập, `LoginPage` chỉ gọi
`LoginViewModel.authenticate()`. ViewModel gọi interface `AuthRepository`; bản
triển khai gọi `AuthRemoteService`; service gọi `ApiClient`. Repository kiểm
tra response, tạo `AuthSession` và cập nhật token/user cho networking.
Điều này làm test ViewModel không cần Flutter widget hay backend thật.

## Feature đã áp dụng: Loans

Nghiệp vụ khoản cho vay được tách theo đúng tuyến phụ thuộc trên:

```text
features/loans/
  domain/entities/loan.dart
  domain/repositories/loan_repository.dart
  data/services/loan_remote_service.dart
  data/repositories/loan_repository_impl.dart
  presentation/view_models/loan_view_model.dart
  presentation/screens/loan_page.dart
```

`LoanPage` chỉ hiển thị tổng gốc đang chạy, lãi/ngày, lãi phát sinh, lịch thu
và các thao tác giải ngân/thu hồi. `LoanViewModel` điều phối tải dữ liệu từ
`/loans`, `/loans/summary`, `/loans/schedule` và ghi nhận thu lãi/gốc qua
repository. Quy tắc tính lãi, kỳ cuối tháng và bút toán nằm ở backend; mobile
không tự tính hoặc tự xác nhận lãi đã nhận.

## Lộ trình tách feature hiện hữu

`features/finora/presentation/finora_pages.dart` vẫn giữ các widget giao diện
cũ để tránh thay đổi UX trong đợt refactor. Logic auth trong widget này đã được
chuyển sang `features/auth/`; bước tiếp theo là di chuyển riêng widget login
vào `features/auth/presentation/screens/`, rồi xử lý theo thứ tự `dashboard`,
`transactions` và các resource page. Mỗi màn hình mới có một ViewModel tương
ứng; không tạo use-case chỉ để bọc một lần gọi repository.

## Chất lượng

- Chạy `dart format lib test` và `flutter analyze` trước khi tạo PR.
- Đặt unit test cạnh feature trong `test/features/<feature>/`.
- Đặt widget test theo view và integration test trong `integration_test/`.
- Dùng `package:` import khi đi từ `test/` vào `lib/`; trong `lib/`, giữ import
  ổn định, có thứ tự và không import `src` từ package khác.

## Nguồn tham khảo

- [Flutter app architecture guide](https://docs.flutter.dev/app-architecture/guide)
- [Clean Architecture in Flutter 2026](https://dev.to/techwithsam/clean-architecture-in-flutter-2026-practical-implementation-guide-1dfb)
- [Dart package layout conventions](https://dart.dev/tools/pub/package-layout)
- [Effective Dart style](https://dart.dev/effective-dart/style)
