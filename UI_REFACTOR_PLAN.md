# Finora UI Refactor Plan

This plan preserves API contracts, backend models, calculation logic, authentication, and real-data behavior. Any visual state requiring unavailable data must present a TODO-backed empty state rather than fabricated values.

The companion audit includes a current screen coverage matrix mapping every public,
protected, mobile, and shell surface to its applied UX focus and remaining gap.

| Phase | Scope | Exit criteria |
|---|---|---|
| 1 — Foundations | Semantic Flutter tokens/theme, typography, core card/money/empty primitives; document both clients | Build, analyze, and tests pass; no behavior regression. |
| 2 — Shell | Finora app shell, responsive five-item navigation, global Quick Action sheet, shared headers | Navigation works at small/standard/large phone widths. |
| 3 — Dashboard | Executive overview hierarchy and reusable summary/alert/transaction components | Net worth, cash flow, alerts, upcoming money, and recent transactions remain real-data driven. |
| 4 — Accounts & transactions | Account scanning, semantic transaction rows, create/edit transaction forms, confirmations | Numeric formatting, error/empty/loading states and transfers verified. |
| 5 — Loans | Contacts → loan → detail → editable collection → timeline; remove mock presentation values | Loan calculations continue to come from existing domain/API code. |
| 6 — Assets & investments | Assets, real estate, portfolios, valuations and allocation | Long monetary values and empty states tested. |
| 7 — Planning | Budgets, forecasts, financial scenarios | Existing planning rules unchanged. |
| 8 — Automation, AI & profile | Bank connection, automation rules, assistant, security/settings | Correct permission/read-only affordances. |
| 9 — Polish | Accessibility, skeletons, offline/retry, micro-animation, visual QA | 44px targets, contrast, dynamic text, safe areas, no overflow. |

## Standard execution for every phase

1. Search for an existing reusable component before adding another.
2. Migrate only the scoped screens; do not mass-rewrite unrelated presentation code.
3. Build the affected client, run lint/analyze and existing tests.
4. Validate navigation, empty/loading/error and long-money states.
5. Record backend-data gaps as explicit TODOs; do not fake feature behavior.

## Immediate component order

1. Tokens, typography, `FinoraCard`, `FinoraMoney`, `FinoraEmptyState` — completed in Phase 1.
2. Button, icon button, labeled text/money/date inputs, status badge, confirmation dialog.
3. App scaffold/header, list/tile, bottom sheet, quick action sheet.
4. Dashboard hero, transaction tile, loan/contact card and collection timeline.

## Progress update — mobile

- Phase 2 shell: completed for phone layouts. The scrolling seven-item bar is
  replaced by five stable destinations (Home, Accounts, Quick Action,
  Transactions, Profile). The central action opens a single global sheet and
  routes lending, collection, transfer, and asset work to their real modules.
- Phase 5 loan baseline: completed. The loan experience now renders only
  repository/API data, groups active loans by counterparty, and uses the flow
  borrower → loan → detail → editable collection. Collection history is an
  explicit empty state until the backend supplies a history endpoint; sample
  financial records are not shown to users.
- Phase 5 creation flow: Quick Action now opens the API-backed loan-create
  sheet directly. It validates principal/rates, accepts compact VND input
  (for example `50tr`), and reloads the borrower list after creation.
- Phase 3 dashboard baseline: completed. The mobile dashboard is now an
  executive overview (net worth, asset/liability summary, quick actions, and
  recent transactions). Non-data-backed promotion and gamified suggestion
  blocks were removed; unavailable dashboard data has retry and empty states.
- Phase 4 transfer flow: the global Quick Action now opens a real, mobile-safe
  transfer sheet backed by `POST /transfers`; it loads accounts, requires two
  distinct selections, validates the amount, and refreshes the active view on
  success. The remaining account/transaction visual migration stays in Phase 4.
- Quick Action asset flow: `Thêm tài sản` now opens an API-backed asset-create
  sheet (`POST /assets`) with name/type validation instead of merely routing to
  a list. The collection action intentionally routes to loan selection, since
  a collection must be attached to a specific loan.
- Responsive regression: the shell is exercised at both 390px and 320px in
  widget tests; the latter guards the five-destination navigation against
  narrow-phone rendering regressions.
- Cross-module state handling: accounts, transactions, audit/read-only pages,
  budgets and dashboard now distinguish unavailable API data from an empty
  result and provide a retry or contextual empty state.
- Profile: the active mobile profile route is now a compact Finora settings
  surface (profile-sync state, amount display, SePay, support, logout). The
  old bank-template layout is no longer reachable from the shell and remains
  only as migration reference pending deletion with the legacy screen split.
- Profile display settings: the amount-display bottom sheet now uses the same
  light semantic surface, typography, border, and primary-action tokens as the
  rest of the mobile app; the legacy dark/gold presentation was removed while
  preserving the existing preference update request.
- Notification entry point: the login notification sheet no longer renders
  fabricated device, portfolio, budget, or AI events. It now presents a
  light-theme empty state until a notification feed is available.
- Read-only resource surfaces: refresh controls now expose screen-reader
  tooltips and use the primary token; fallback raw values use readable light
  typography instead of hard-coded white text.
- Web copy pass: remaining user-facing English labels in Assets, Properties,
  Transactions, and Automation were translated while preserving technical
  identifiers and regex syntax.
- Mobile web shell spacing: the fixed bottom navigation now has a shared
  safe-area-aware content clearance, and the drawer backdrop aligns with the
  64px phone header so long forms are not obscured or left with a visual gap.
- Accessibility polish: mobile/web navigation and theme controls now expose
  Vietnamese labels and expanded state to assistive technology; remaining
  public empty/error actions use consistent Vietnamese copy.
- SePay feed states: bank-feed loading, request failure, retry, and genuinely
  empty results are now distinct, preventing a transient empty inbox from
  being mistaken for missing data.
- AI assistant feedback: mobile assistant responses now have a working copy
  action with confirmation feedback; the unsupported thumbs-up affordance was
  removed until a feedback endpoint exists.
- Forecast status copy: backend status values remain intact for behavior, but
  the table now presents localized labels (Hoàn tất/Đang chạy/Chờ chạy) so
  users do not need to understand internal state names.
- Accessibility verification: Angular lint exposed a keyboard-inaccessible
  mobile drawer backdrop; it now has a focus target, semantic label, and
  Enter/Escape handlers. The full Angular lint suite is green.
- Auth form semantics: login and registration labels now explicitly reference
  stable input IDs, improving screen-reader navigation and autofill behavior.
- Transaction/automation form semantics: primary account, transfer, amount,
  note, and rule controls now have explicit label-to-input associations for
  more reliable keyboard and screen-reader navigation.
- Asset/property form semantics: create and valuation controls now expose
  explicit IDs and label associations for name, type, portfolio, amount,
  currency, source, and effective time fields.
- Loan/SePay form semantics: loan creation, collection, and bank-feed
  reclassification controls now have explicit label associations for
  counterparty, principal, rates, accounts, payment amounts, and reason.
- Assistant/portfolio form semantics: command, execution-plan, portfolio-name,
  and base-currency controls now expose explicit label associations.
- Notification trust state: the mobile login shell no longer advertises unread
  notifications while the notification feed is unavailable; the empty state
  and badge now agree until real notification data is supplied.
- Verification: Angular lint, production build, and headless test runner now
  pass; the current Angular suite contains five focused smoke specs, while detailed
  interaction coverage remains in the Flutter widget tests.
- Audit handoff: `UI_UX_AUDIT.md` now records the applied Angular follow-up,
  verification evidence, and intentional API/component-library gaps instead
  of implying those gaps are complete.
- Icon registry integrity: loading and offline states now resolve to dedicated
  `progress_activity` and `cloud_off` paths instead of the generic fallback
  icon used previously.
- Icon alias audit: directional action icons and the assistant robot icon now
  have explicit registry aliases (`arrow_downward`, `arrow_forward`,
  `arrow_upward`, `smart_toy`), preventing silent fallback glyphs in navigation
  and assistant surfaces.
- Copy and data-label audit: transaction/budget errors, portfolio fallback,
  SePay confidence/reclassification copy, asset types, and automation actions
  now render Vietnamese user-facing labels instead of English or backend enum
  values.
- Mobile enum-label audit: account-link statuses and readonly activity statuses
  now use Vietnamese labels, while unknown values remain visibly marked instead
  of silently disappearing; account forms also use “Mã danh mục” consistently.
- Core table-label audit: account type, assistant command status, transaction
  type/status, and loan direction/status now render localized labels without
  changing API values or filtering behavior.
- SePay connection audit: provider connection and sync statuses now use
  localized labels with an explicit unknown-state fallback, matching the feed
  status treatment.
- Filter accessibility audit: Audit and SePay filter selects now expose stable
  IDs and Vietnamese accessible names, so icon-free compact filter bars remain
  understandable to screen readers.
- Mobile form-copy audit: bank-feed reclassification now labels the optional
  category reference as “Mã danh mục (không bắt buộc)”, keeping technical input
  context clear without exposing the English “ID” shorthand.
- Assistant approval input audit: the inline approval field now has a
  Vietnamese placeholder and accessible name instead of exposing the raw
  `approvalId` backend key.
- Mobile activity audit: common audit actions are now mapped from backend verbs
  to Vietnamese labels, with a visible raw fallback for unknown actions.
- Shell accessibility audit: workspace/language selectors and the amount-mode
  toggle now expose Vietnamese accessible names; the workspace placeholder is
  localized as well.
- Async feedback audit: account/loan status messages now announce politely and
  SePay feed failures use an alert role, making asynchronous outcomes audible to
  assistive technology.
- Forecast copy audit: the technical scenario table header now reads “Mã kịch
  bản” instead of the unexplained English “ID”.
- Mobile drawer semantics: the top and bottom menu triggers now reference the
  navigation drawer via `aria-controls`, expose dynamic open/closed labels, and
  the drawer has a stable ID and landmark label.
- Auth render audit: the login footer is fully Vietnamese; login/register
  buttons explicitly submit their forms and auth errors announce assertively.
- Browser verification: regenerated the 390×844 login render after the auth
  copy changes; the current snapshot shows Vietnamese labels and the expected
  mobile spacing/header without overflow.
- Verification hardening: Angular smoke coverage now includes an icon-registry
  alias assertion, protecting loading/offline/assistant/action icons from
  silently regressing to the fallback glyph.
- Mobile error-copy audit: shared ErrorBox/SnackBar rendering now strips
  technical exception prefixes and maps network failures to a clear Vietnamese
  retry message without changing underlying error handling.
- Error-copy test hardening: the mobile widget suite now covers technical-prefix
  stripping and network-error wording; Flutter suite is green at 17 tests.
- Error accessibility: mobile `ErrorBox` is now a live semantic region so async
  failures are announced to assistive technology as soon as they render.
- Shell copy consistency: notification dismissal and the optional assistant
  approval control now use Vietnamese, context-accurate accessible names.
- Automation form semantics: all rule inputs/selects and the enabled toggle now
  expose explicit label associations for keyboard and assistive-technology users.
- Compact filter semantics: transaction account/type filters and the dashboard
  portfolio selector now expose explicit Vietnamese accessible names.
- Identifier copy audit: visible connection/request/portfolio/account identifiers
  now use user-facing “Mã …” labels instead of the technical “ID” shorthand.
- SePay validation copy: the reclassification form now marks category as optional
  and manual reason as required, matching the actual validators.
- Mobile copy consistency: support, audit, AI, and SePay surfaces now use
  Vietnamese terms for account mapping, knowledge base, workspace, and user.
- Mobile vocabulary follow-up: navigation and multi-member SePay guidance now
  use consistent activity-log and workspace-member wording.
- Loan empty-state copy: the unavailable history message now uses product-facing
  “Hệ thống…” wording instead of exposing an API implementation detail.
- Loan error-copy consistency: loan SnackBars and customer-loading errors now use
  normalized, network-aware Vietnamese messages.
- Loan list error surface: the top-level empty state now normalizes view-model
  errors before rendering them to users.
- Toast announcement semantics: web notifications now use atomic `alert` or
  `status` roles according to severity while preserving non-blocking focus.
- Shared fallback copy: read-only workspace and network interceptor messages now
  use Vietnamese wording consistent with the default product language.
- HTTP fallback copy: status-based interceptor messages now use actionable
  Vietnamese wording and retain codes only as diagnostic context.
- Workflow copy consistency: Assistant and SePay statuses now use Vietnamese
  request/review/reconciliation terminology.
- Translation dictionary cleanup: the shared workspace label is now Vietnamese
  in the default locale, preventing English leakage in shell controls.
- Financial copy clarity: snapshot action/history labels now use user-facing
  “bản chụp số liệu” wording while preserving backend/API names.
- Mobile typography copy: profile and transaction titles now use sentence case
  and Vietnamese comparison wording instead of mixed casing/“vs”.
- Financial summary copy: transaction total cards now use sentence case for
  consistent hierarchy.
- Tap-target hardening: web shell icon/avatar controls now reserve 44×44px
  interaction areas without changing their semantic labels or actions.
- Shell target-size completion: sidebar links and toast dismissal controls now
  also reserve 44×44px interaction areas.
- Mobile heading actions: the narrow-screen override now preserves the 44px
  minimum instead of shrinking page actions to 36px.
- Shared icon primitive: `.w-btn-icon` now defaults to 44×44px for safe future
  reuse across feature screens.
- Mobile compact controls: `h-8`/`h-9` buttons and small selects receive a
  44px minimum only below the mobile breakpoint.
- Select specificity fix: the mobile `select.w-select-sm` rule now wins over the
  legacy 34px `!important` height so the target-size guarantee is effective.
- Mobile drawer semantics: the navigation backdrop is now a real button and
  Escape closes the open drawer, improving keyboard parity without changing
  navigation flow.
- Drawer regression coverage: the Angular smoke suite now verifies the Escape
  close rule at mobile and desktop widths (5 specs total).
- Focus visibility: shared web controls now expose a consistent high-contrast
  `:focus-visible` ring for keyboard users after native outlines are reset.
- Theme persistence: the shell now saves and restores the dark-mode preference
  alongside the amount-display preference, avoiding a surprising reset on reload.
- Motion accessibility: the shared stylesheet now honors `prefers-reduced-motion`
  for all transitions and animations across the product shell.
- Mobile form resilience: the personal-information bottom sheet now scrolls
  within its keyboard-aware inset so fields and the save action remain reachable
  on compact viewports.
- Capability honesty: the unconnected Smart OTP limit-change flow now presents
  a disabled “Chưa khả dụng” state instead of a fabricated success confirmation.
- Local-state honesty: profile editing now distinguishes session-only changes,
  and delete errors no longer expose raw exception prefixes.
- Resource error consistency: transaction, budget, bank-feed, account, and
  generic resource loaders now use the shared formatter for inline/retry errors.
- Appearance capability alignment: the mobile settings sheet no longer exposes
  a non-functional dark-mode radio option; it documents the supported light-only
  state until a real dark theme exists.
- Copy cleanup: the quick action and SePay mapping labels now use full
  Vietnamese wording instead of “Tạo GD”/“map”.
- Target-size alignment: custom animated and profile action buttons now meet the
  48px mobile control-height requirement from the redesign specification.
- Custom CTA semantics: the animated gesture-based CTA now exposes button role,
  label, and disabled state to assistive technology.
- Profile interaction semantics: limit/edit links and satisfaction stars now
  expose button roles, labels, and selected state for non-visual navigation.
- Top-bar semantics: mobile language, notification, and settings controls now
  expose labels and unread state to assistive technology.
- Profile settings consistency: the appearance subtitle is explicitly
  light-only and the rating-history link now has a button semantic.
- Rating semantics: only the active satisfaction score is announced as selected;
  visual fill for lower stars no longer implies multi-select state.
- Registration repair: restore the confirmation-password control, wire it into
  authentication, and use high-contrast password-visibility icons with a
  regression test for missing confirmation.
- Registration handoff: after account creation, return to sign-in with the email
  prefilled and password fields cleared; explain email verification before login.
- Verification completion: add a mobile six-digit verification screen with
  resend/back actions; document that SMTP must be configured for real delivery.
- Mobile error consistency: asset, transfer, and SePay flows now use the shared
  `presentableError` formatter for actionable Vietnamese error copy instead of
  exposing raw exception strings.
- Readonly resource errors: audit/support list pages now normalize failures
  before rendering inline errors, keeping retry states free of raw exception
  prefixes.
- Route inventory audit: reconciled the Angular route table against the UX audit;
  public auth/error routes and all protected resource routes are explicitly
  covered, with remaining gaps documented rather than silently omitted.
- Route-test hardening: Angular smoke coverage now asserts the audited public,
  protected, and wildcard route paths remain present.
- Remaining phases are intentionally not bulk-rewritten: dashboard, accounts,
  transactions, assets, planning, automation, and profile should migrate to
  the shared components one module at a time.
