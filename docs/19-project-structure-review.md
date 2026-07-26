# Rà soát cấu trúc dự án Finora / WealthOS

> Ngày rà soát: 25-07-2026. Phạm vi: cấu trúc thư mục, ranh giới dependency,
> khả năng mở rộng và kiểm thử. Đây là đánh giá kiến trúc từ mã nguồn hiện có,
> không phải kiểm thử bảo mật hay đánh giá nghiệp vụ tài chính.

## Kết luận ngắn

Đây là một **multi-application repository** tốt cho giai đoạn prototype đến
MVP: backend Go là modular monolith, frontend Angular tổ chức theo feature và
mobile Flutter đã có baseline Clean Architecture. Mô hình phù hợp để phát triển
một sản phẩm tài chính chung API trên web và mobile.

Rủi ro chính không nằm ở số lượng thư mục mà ở ba điểm:

1. Backend đang dồn phần lớn nghiệp vụ vào hai tệp rất lớn (`WealthService` và
   `WealthHandler`), nên ranh giới domain thực tế chưa rõ.
2. Frontend và mobile chưa dùng một API contract được sinh/kiểm tra tự động;
   các model TypeScript, DTO Go và JSON Flutter có thể lệch nhau.
3. Flutter đã tách logic auth, nhưng phần UI còn lại vẫn nằm trong một tệp
   presentation 3.284 dòng. Nó là điểm nóng cần refactor theo từng feature.

Không nên viết lại toàn bộ. Hướng an toàn là tách theo **vertical slice**:
hoàn thiện một feature từ API đến hai client, có contract và test, rồi lặp lại.

## 1. Bản đồ cấu trúc hiện tại

```text
Finora/
├── backend/                   Go 1.25, Gin, PostgreSQL
│   ├── cmd/server/            binary entry point
│   └── internal/
│       ├── app/               composition root / startup
│       ├── auth/              token helpers
│       ├── config/            environment config
│       ├── db/                migrations
│       ├── domain/            entities và type nghiệp vụ
│       ├── http/              router, middleware, handler, DTO
│       ├── integration/       SePay và Hermes adapters
│       ├── job/               worker nền
│       ├── service/           business/application logic
│       └── storage/           Store interface + memory/PostgreSQL store
├── frontend/                  Angular 21 standalone app
│   └── src/app/
│       ├── core/              auth, API, guard, interceptor, toast
│       ├── features/          route/page theo capability
│       └── shared/            TypeScript API models
├── mobile/                    Flutter app iOS/Android
│   └── lib/
│       ├── app/               composition root và manual DI
│       ├── core/              config, network, theme
│       └── features/
│           ├── auth/          Clean Architecture mẫu
│           └── finora/        UI legacy còn gom chung
├── docs/                      product, API, domain, DB và delivery plan
└── docker-compose*.yml        local/prod deployment manifests
```

### Luồng runtime hiện tại

```text
Angular browser ─┐
                 ├─ HTTP /api/v1 ─→ Gin router/middleware
Flutter mobile ──┘                      ↓
                                     handler
                                       ↓
                                WealthService
                                  ↓         ↓
                              Store      integrations/jobs
                               ↓
                    In-memory store hoặc PostgreSQL
```

Đây là modular monolith, không phải microservices. Với quy mô hiện tại, đó là
lựa chọn đúng: một deployable unit, transaction và quan sát vận hành đơn giản.

## 2. Đánh giá cấp repository

| Hạng mục | Hiện trạng | Nhận định |
| --- | --- | --- |
| Tách deployable | `backend/`, `frontend/`, `mobile/` độc lập | Tốt; mỗi app có manifest và README riêng. |
| Tài liệu domain | Có vision, business rule, API, database, roadmap | Tốt; tài liệu là lợi thế lớn nếu được giữ khớp mã nguồn. |
| Local runtime | Docker Compose có PostgreSQL + backend + frontend | Tốt cho onboarding. Mobile chạy native nên không nằm trong Compose là hợp lý. |
| Workspace/tooling chung | Chưa có task runner, CI workflow, contract generation hay root quality command | Thiếu; đây là điểm cần đầu tư sớm. |
| Version-control hygiene | `mobile/` hiện đang là thư mục untracked trong worktree tại thời điểm rà soát | Cần xác nhận trước khi merge: mọi mã mobile phải được add/commit có chủ đích. |

### Khuyến nghị ở root

1. Thêm `Makefile` hoặc `justfile` với `check`, `test`, `dev`, `compose-up`.
   `check` nên gọi `go test ./...`, Angular lint/test/build và Flutter
   analyze/test.
2. Thêm CI (GitHub Actions hoặc hệ tương đương) chạy các lệnh trên khi PR.
3. Chọn một API contract làm source of truth: OpenAPI trong `docs/api/` là lựa
   chọn thực tế nhất. Sinh TypeScript client/models và Dart DTO từ contract,
   hoặc ít nhất validate response bằng contract test.
4. Chuẩn hóa tên sản phẩm. Root/backend/frontend còn dùng **WealthOS** trong
   khi mobile dùng **Finora**; xác định Finora là brand còn WealthOS là tên nội
   bộ hay đổi đồng nhất trước khi public.

## 3. Backend Go

### Điểm tốt

- Có entry point tối giản ở `cmd/server`; `internal/app` lo compose config,
  storage, migration, service, HTTP server và worker. Đây là dependency
  direction dễ hiểu.
- Có `internal/` nên module nội bộ không vô tình thành public API Go.
- Router có version `/api/v1`, health/readiness endpoints, middleware cho CORS,
  request ID, error envelope, auth, workspace membership, rate limit và
  idempotency. Với domain tài chính, đây là nền tảng đáng giá.
- Domain model tách khỏi handler và storage; Store có in-memory implementation
  cho test/bootstrap và PostgreSQL implementation cho production.
- Có 7 test Go trải trên service, storage, middleware, handler và auth.

### Điểm cần xử lý

| Mức | Quan sát | Hệ quả | Hướng xử lý |
| --- | --- | --- | --- |
| P0 trước production | `User.Password` và `Authenticate()` đang so sánh password trực tiếp; README cũng mô tả in-memory bootstrap | Không đạt mức bảo vệ credential production | Dùng Argon2id/bcrypt, chỉ lưu hash, migration/rehash strategy, rate-limit và audit login rõ ràng. |
| P0 trước production | Config có fallback `local-dev-secret` và CORS `*` | Dễ deploy sai cấu hình | Khi `APP_ENV=production`, fail fast nếu JWT secret, DB URL hoặc CORS allow-list thiếu/không an toàn. |
| P1 | `internal/service/wealth_service.go` 2.240 dòng; `internal/http/handler/wealth_handler.go` 2.665 dòng | Mọi capability phụ thuộc cùng một vùng mã; tăng conflict và khó test | Tách theo bounded context: `ledger`, `portfolio`, `loans`, `budget`, `banking`, `assistant`. |
| P1 | `storage.Store` 111 dòng bao toàn bộ domain | Mọi storage implementation phải biết mọi feature; coupling tăng theo feature | Chia port/interface nhỏ theo use case (`AccountRepository`, `TransactionRepository`, ...), inject đúng phần cần dùng. |
| P1 | Snapshot net-worth có state package-level trong service | Khó scale ngang, khó kiểm soát lifecycle/cache và test isolation | Đưa cache/snapshot thành dependency riêng, hoặc persist snapshot vào DB/job. |
| P2 | Handler chứa DTO binding, auth flow, orchestration và một phần policy | HTTP layer khó đọc/đổi transport | Mỗi feature có `transport/http`, request/response DTO và application service riêng. |

### Cấu trúc backend đích đề xuất

Không cần ép "clean architecture" thuần. Nên tách dần theo feature, vẫn giữ một
binary và một DB:

```text
backend/internal/
  platform/                 # config, db, http middleware, observability
  modules/
    identity/
      application/          # register/login commands
      domain/                # user, credential policy
      infrastructure/        # postgres repository, password hasher
      transport/http/        # handler + DTO
    ledger/
    portfolio/
    lending/
    budgeting/
    banking/
    assistant/
```

Quy tắc: `transport` chỉ parse/serialize; `application` điều phối use case;
`domain` chứa invariant; `infrastructure` chứa SQL/API provider. Module chỉ gọi
module khác qua application-facing interface/event, không gọi repository của
nhau tùy tiện.

## 4. Frontend Angular

### Điểm tốt

- `core/` đã chứa concern global đúng chỗ: authentication, API, interceptor,
  route guard và toast.
- `features/` được chia theo capability người dùng, không theo loại file toàn
  cục. Đây là cấu trúc Angular dễ mở rộng.
- Standalone components và routing tập trung trong `app.routes.ts` phù hợp với
  Angular hiện đại.
- Model API có type thay vì để `any` lan rộng.

### Điểm cần xử lý

| Mức | Quan sát | Hướng xử lý |
| --- | --- | --- |
| P1 | `shared/models.ts` 313 dòng là tập trung mọi API model | Chuyển model/DTO vào feature hoặc generated `api-client`; chỉ giữ primitive UI chung ở `shared/`. |
| P1 | `ApiService` 354 dòng | Chia theo feature client: `accounts-api`, `transactions-api`, `banking-api`; mỗi feature tự sở hữu data-access layer. |
| P1 | Component feature có thể vừa quản lý form, request, mapping và presentation | Tạo `features/<name>/data-access/`, `ui/`, `pages/`; container page giữ orchestration, UI component chỉ nhận input/output. |
| P2 | Chỉ thấy 1 frontend spec | Tăng unit test cho service/interceptor/guard và component critical; E2E cho login, workspace selection, create transaction. |

### Cấu trúc frontend đích đề xuất

```text
src/app/
  core/                      # singleton application concerns
  shared/                    # reusable UI/primitives, không chứa feature API
  features/
    transactions/
      data-access/           # API client, facade/store, DTO mapper
      pages/                  # routed container
      ui/                     # presentational components
      models/                 # chỉ khi không generated từ OpenAPI
```

Không cần thay toàn bộ state management ngay. RxJS/service hiện có có thể giữ;
chỉ cần giới hạn API call và transformation ở `data-access`, không rải trong
template/page.

## 5. Mobile Flutter

### Điểm tốt

- `app/` là composition root và `core/` tách config/network/theme khỏi UI.
- Auth đã có một vertical slice đúng hướng: `domain` (entity + contract),
  `data` (remote service + implementation), `presentation` (ViewModel).
- `LoginViewModel` được test bằng fake repository, chứng minh UI logic không
  cần backend thật.
- `flutter analyze` và `flutter test` đã chạy sạch ở lần rà soát trước đó.

### Điểm cần xử lý

| Mức | Quan sát | Hướng xử lý |
| --- | --- | --- |
| P1 | `features/finora/presentation/finora_pages.dart` 3.284 dòng, bao login, dashboard, navigation, resource page và widget dùng chung | Tách theo feature, trước hết di chuyển `LoginPage` cùng widget con vào `features/auth/presentation/screens/`; sau đó dashboard và transactions. |
| P1 | Home/resource pages còn nhận `ApiClient` trực tiếp | Lặp lại mẫu auth: mỗi feature có repository/service và ViewModel. View không gọi `ApiClient`. |
| P1 | `ApiClient` có token/workspace mutable | Tách `SessionStore`/`AuthSession` khỏi HTTP client. Điều này giúp logout, refresh token và test độc lập rõ ràng hơn. |
| P2 | Manual DI là hợp lý hiện tại nhưng composition sẽ lớn khi feature tăng | Chuyển sang Riverpod/provider khi có nhiều shared state, caching hoặc lifecycle phức tạp; không cần làm trước khi cần. |
| P2 | 2 test Flutter hiện hữu | Thêm test ViewModel cho account/transaction, widget test form states, và `integration_test` cho login → create transaction. |

### Cấu trúc mobile đích đề xuất

```text
lib/
  app/                       # router, theme binding, DI
  core/                      # config, network, session, design system
  shared/                    # widget/model thật sự dùng xuyên feature
  features/
    auth/
      presentation/screens/ widgets/ view_models/
      domain/entities/ repositories/
      data/services/ repositories/
    transactions/
    dashboard/
    accounts/
```

Use case chỉ thêm khi một command kết hợp nhiều repository hoặc có quy tắc tái
sử dụng. Không tạo use case chỉ để bọc một repository call.

## 6. API contract và cross-client consistency

Đây là ranh giới quan trọng nhất vì cùng backend phục vụ Angular và Flutter.

Hiện tại có ba representation của cùng dữ liệu:

1. Go `domain` và HTTP DTO.
2. TypeScript model ở `frontend/src/app/shared/models.ts`.
3. Dart entity/Map JSON ở mobile.

Nếu thay đổi field, enum, pagination hoặc error envelope mà chỉ sửa một nơi,
client còn lại có thể compile nhưng lỗi runtime. Đề xuất:

1. Viết OpenAPI 3.1 cho `/api/v1`, bao gồm auth, error envelope, headers
   `X-Workspace-ID` và `Idempotency-Key`.
2. CI validate endpoint/response với contract.
3. Generate TypeScript và Dart API DTO/client; mapper DTO → UI/domain model vẫn
   sống trong từng feature.
4. Version endpoint có breaking change bằng `/api/v2` hoặc compatibility
   policy rõ ràng, không thay đổi field âm thầm.

## 7. Lộ trình ưu tiên

### Sprint 1 — nền production và quality gate

- Bỏ plaintext password; production config fail-fast; giới hạn CORS.
- Commit rõ `mobile/`, thêm root task runner và CI.
- Chuẩn hóa Finora/WealthOS naming và environment matrix.

### Sprint 2 — contract và transaction slice

- Công bố OpenAPI cho auth/workspace/transactions.
- Tách backend `identity` và `ledger` trước.
- Chia Angular/Flutter transaction thành data-access/repository + presentation.
- Thêm contract test và E2E create/list transaction.

### Sprint 3 — banking và observability

- Tách banking/integration module, outbox/event processing nếu webhook cần
  at-least-once delivery.
- Structured logging với request/correlation ID, metrics, tracing và audit
  retention policy.
- Persist net-worth snapshot/caching theo workspace thay cho state package-level.

## 8. Definition of Done cho feature mới

Một feature chỉ nên được coi là hoàn tất khi có:

- API contract, auth/workspace/idempotency/error behavior được nêu rõ.
- Backend handler mỏng, application logic test được, repository port nhỏ.
- Angular data-access tách khỏi UI và test tình huống lỗi.
- Flutter service/repository/ViewModel theo feature, không gọi HTTP trong View.
- Unit test cho business/UI state quan trọng; integration/E2E cho happy path.
- Logs/audit phù hợp nếu feature thay đổi dữ liệu tài chính.

## Tài liệu liên quan

- [Product vision](01-product-vision.md)
- [Domain model](03-domain-model.md)
- [API design](05-api-design.md)
- [Database design](04-database-design.md)
- [Mobile architecture](../mobile/doc/architecture.md)
