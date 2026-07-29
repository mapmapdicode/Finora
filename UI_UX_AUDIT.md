# Finora Wealth OS — UI/UX Audit

Audit date: 2026-07-29. Scope: `mobile/` Flutter client, `frontend/` Angular client, and the existing presentation routes only. This is a UI/UX audit; no API, database, calculation, or authentication contract is changed.

## Architecture inventory

| Area | Current implementation | Audit finding |
|---|---|---|
| Mobile | Flutter Material 3, one `MaterialApp`, API client + feature repositories/view models | The primary app is operational, but `finora_pages.dart` (7,768 lines) still mixes app shell, screens, forms, sheets, styles, and reusable widgets; extraction remains incremental. |
| Mobile loans | Separate feature with remote service/repository/view model | Loan navigation and collection UI are repository/API-driven; unavailable collection history is an explicit empty state until its endpoint exists. |
| Web | Angular 21 standalone components, Tailwind utilities plus global CSS design classes | Desktop shell and 17 protected routes are present. Component templates repeat visual primitives instead of consuming a small shared component set. |
| Backend | Go API | Out of scope for this refactor; preserve all requests and business rules. |

## Screen, route, and interaction inventory

### Flutter mobile

| Surface | Entry/navigation | Forms, dialogs, and sheets |
|---|---|---|
| Login / registration | App root → `LoginPage` | Email, password, name; custom text fields |
| Home shell | `HomePage`, side navigation / phone bottom bar | Settings sheet; phone FAB opens transaction form |
| Dashboard | Home index 0 → `DashboardPage` | Fetches portfolios, accounts, transactions |
| Accounts | Index 1 → `ResourcePage` | Account create/edit sheet |
| Transactions | Index 2 → `TransactionsPage` | `_TransactionFormSheet` |
| Loans | Index 3 → `LoanPage` → borrower detail → loan detail | Collection record UI; date-picker sheet |
| Assets / properties / portfolios | Index 4, 5, 8 | Generic resource forms |
| Budget / forecast | Index 6, 7 | Budget interaction; scenario page |
| Bank / SePay | Index 9 | Hosted web view, bank-feed edit sheet |
| Automation / AI / audit | Index 10–12 | Scenario/read-only pages |
| Profile | Index 13 | Settings and logout |

### Angular web

Protected routes: dashboard, accounts, transactions, loans, assets, properties, forecasts, budgets, portfolios, SePay, automation, assistant, audit logs, forbidden. Public routes: login, register, and not-found. The shell supplies desktop sidebar, mobile navigation, workspace selector, language, amount mode, theme control, and toast stack.

### Current screen coverage matrix

| Screen / route | Applied UX focus | Remaining intentional gap |
|---|---|---|
| Login / register | Vietnamese labels, autocomplete, submit semantics, assertive auth errors | Detailed end-to-end auth flow remains outside smoke coverage |
| Not-found / forbidden | Clear recovery and permission messaging with shell-safe navigation | Destination availability still depends on router/API state |
| Dashboard | Real net-worth hierarchy, portfolio selector, historical preview, loading/error/empty states | Snapshot history depends on available portfolio data |
| Accounts | Labeled create form, type labels, read-only affordances, retry/empty states | Shared typed form primitive remains future work |
| Transactions | Transfer/create flows, localized type/status labels, compact filter ARIA names | Route interaction coverage remains smoke-level |
| Loans | API-backed list/detail/collection flow, no fabricated history, normalized errors | Collection history waits for its endpoint |
| Assets / properties | Labeled valuation forms, localized types, explicit empty/loading states | Incremental extraction from mobile presentation monolith |
| Budgets / forecast | Clear setup copy, save/run loading states, localized statuses and identifiers | Scenario/result API depth is unchanged |
| Portfolios | Portfolio selection, snapshot/bản chụp history, retry and load-more states | Long-term allocation visualization remains data-dependent |
| SePay | Connection/feed/reclassification states, filter labels, optional/required copy | Bank feed still depends on configured provider data |
| Automation | Explicit label associations, localized actions/types, preview/loading feedback | Regex remains intentionally technical |
| Assistant | Localized request/approval copy, accessible optional approval field, status feedback | Approval workflow depends on backend commands |
| Audit / read-only | Localized activity labels, permission messaging, retry/empty states | Audit data remains API-dependent |
| Profile / settings | Light semantic settings surface, amount-display and support states | Notification/support feeds remain explicit empty states |

## Reusable components already present

| Client | Existing reusable units | Gaps |
|---|---|---|
| Flutter | `PageFrame`, `FinoraSurface`, `FinoraListTile`, `FinoraSheet`, `SectionTitle`, `ErrorBox`, `EmptyState`, login-local controls | All live in a single presentation file; no shared semantic money, button, input, status, skeleton, or app-shell primitives. |
| Angular | `app-icon`, global `.w-*` / `.m3-*` CSS classes, toast service | CSS is reusable but names mix two design languages; there are no typed card, money, list row, empty-state, or modal components. |

## Findings and refactor backlog

| Screen / component | Current problem | Severity | Proposed solution | Reusable component affected | Priority |
|---|---|---|---|---|---|
| Flutter app shell | Phone bottom navigation contains 7 items and scrolls horizontally; primary action only opens a transaction form. | Critical | Adopt five-item shell: Home, Accounts, global Quick Action, Transactions, Profile. Put loans/assets/planning under contextual navigation. | `FinoraScaffold`, `FinoraQuickActionSheet` | P0 |
| Flutter loans | The earlier sample-history risk is resolved; collection history now remains empty until a real endpoint exists. | Resolved | Keep the repository/API-only rendering and preserve the explicit unavailable-data state. | `FinoraLoanDetail`, `FinoraMoneyInput` | Completed |
| Flutter presentation architecture | One 7k-line file owns nearly every page and visual primitive. | High | Extract shell, shared components, page-local widgets, and feature pages incrementally. | All mobile core widgets | P0 |
| Both clients | Colors, radii, shadows, and spacing are repeatedly hard-coded (849 style-token occurrences found in UI source). | High | Consume semantic tokens; migrate during each module phase instead of mass replacement. | Token/theme layers | P0 |
| Flutter typography | No shared type scale; financial values, labels, and headers vary by screen. | High | Use display/title/body/label scale plus tabular financial text. | `FinoraMoney`, typography tokens | P0 |
| Flutter forms | Custom fields vary and several form patterns rely on label-like placeholders or bespoke validation states. | High | Adopt labeled text/money/date field primitives with helper/error/disabled states. | `FinoraTextField`, `FinoraMoneyInput`, `FinoraDatePicker` | P1 |
| Flutter dashboard | Dashboard mixes large decorative treatment with information rather than a concise executive view. | High | Establish hero → cash flow → alerts → upcoming money → recent transactions. | `FinoraNetWorthHero`, transaction tiles | P1 |
| Angular + Flutter brand | The web client is branded WealthOS while Flutter uses Finora; purple and accent palettes differ. | High | Use Finora — Wealth OS naming and shared semantic color meanings across both clients. | App shell, theme tokens | P1 |
| Flutter destructive controls | Account deletion is passed directly through a list-tile callback; confirmation policy is not encoded in a shared pattern. | High | Route destructive actions through a confirmation dialog primitive. | `FinoraConfirmDialog` | P1 |
| Angular mobile shell | Mobile web nav lacks a central quick action and uses different destinations from the requested product IA. | Medium | Align with the five-item navigation and shared quick-action sheet/modal. | Web shell | P2 |
| Empty/loading/error states | Some mobile resources expose `EmptyState` and `ErrorBox`, but there is no standard skeleton/offline/retry coverage. | Medium | Add semantic empty, error/retry, and skeleton primitives; apply module by module. | `FinoraEmptyState`, `FinoraSkeleton` | P1 |
| Accessibility | Some compact resource-level controls still need a central target-size primitive; web shell icon/avatar controls now meet 44×44px. | Medium | Extend the 44×44px rule through shared buttons and remaining feature controls. | Buttons, icon buttons, navigation | P1 |
| Responsive financial data | Wide bottom navigation and long currency figures risk overflow. | Medium | Limit mobile navigation; render money as tabular, ellipsized/scaled appropriately. | `FinoraMoney`, shell | P1 |
| Decorative maple background | Maple art appears behind broad app areas and competes with content contrast. | Medium | Limit to low-opacity hero/header decoration and protect readable surface areas. | `FinoraScaffold`, headers | P2 |
| Loans information architecture | Contacts, loans, detail, collection and history are navigable but not expressed as a calm progressive flow. | Medium | Make borrower → loan → record collection → timeline the dominant hierarchy. | Loan cards/timeline | P1 |

## Ten most serious issues

1. The Flutter phone shell has too many scrolling navigation destinations and no global quick-action model.
2. [Resolved] Loan detail no longer contains sample collection values/history; unavailable history is explicit.
3. A single massive Flutter presentation file prevents controlled reuse and safe UI iteration.
4. Design values are dispersed rather than semantic, so screens visually drift.
5. Financial typography is not standardized or explicitly tabular across screens.
6. The dashboard does not consistently prioritize net worth, cash flow, alerts, and expected incoming money.
7. Forms lack one accessible, labeled field contract with consistent money/date/error behavior.
8. Destructive-action confirmation is not centralized.
9. Mobile and web have inconsistent Finora/WealthOS naming and palette behavior.
10. Loading, empty, offline, and retry experience is incomplete and inconsistent.

## Phase 1 delivered foundation

Flutter now has semantic color, spacing, radius, elevation, and typography tokens; theme-level typography/input/button defaults; and initial `FinoraCard`, `FinoraMoney`, and `FinoraEmptyState` primitives. Existing screen behavior remains unchanged. The next migrations will consume these primitives rather than add further local styles.

## Follow-up applied to mobile

- The phone shell now follows the required five-destination structure and the
  global Quick Action is accessible as a 56px, labeled central control.
- The loan module no longer contains visual sample borrowers, balances, or
  payment history. It is driven by its repository/ViewModel and retains a
  transparent empty state for history until an endpoint is available.
- Dashboard no longer mixes promotional/gamified cards into financial
  decision-making. It prioritizes live net-worth data, asset/liability
  indicators, quick actions and recent transactions, with explicit retry and
  empty states when the API cannot provide data.
- The mobile presentation no longer pre-fills credentials or renders a named
  person's identifiers, limits, avatar, or third-party bank branding as if
  they were live Finora data. Unavailable profile fields are explicitly marked
  as not synchronized.

## Follow-up applied to Angular web

- App shell: Finora branding, responsive five-item mobile navigation, central
  quick action, drawer backdrop, safe-area content clearance, and keyboard
  accessible menu controls are in place.
- Dashboard: the page is driven by portfolio/net-worth/snapshot data, with
  explicit loading, empty, error, historical-preview, allocation, recent
  transaction, and data-quality states; no promotional sample metrics remain
  in the active template.
- Resource screens: Accounts, Transactions, Portfolios, Assets, Properties,
  Forecast, Budget, Automation, Assistant, Audit, Loans, and SePay now expose
  clearer create/edit flows, read-only affordances, saving/loading feedback,
  retry actions, and contextual empty states.
- SePay: feed loading, request failure, retry, and genuinely empty results are
  separate states so a slow request cannot appear as an empty inbox.
- Copy and semantics: remaining user-facing English copy was localized across
  resource screens; primary form controls now use explicit label/input IDs;
  drawer and theme controls expose accessible names and keyboard behavior.
- Data trust: raw backend status values remain available to logic but are
  localized at presentation time; technical IDs and JSON/regex values remain
  visibly technical where they are needed for troubleshooting.

## Verification evidence

- Angular: `npm run lint`, `npm run build`, and the headless Karma smoke suite
pass. The current Angular test suite contains five focused smoke specs; detailed
  route-level visual verification was performed with mocked browser flows for
  Dashboard, Accounts, Transactions, and mobile shell states.
- Flutter: `flutter analyze` and the full widget test suite pass, covering the
  five-item shell, quick actions, live loan grouping, customer selection,
  profile surface, dashboard data hierarchy, and narrow-phone layouts.
- Repository hygiene: `git diff --check` passes after each applied UI pass.

## Copy and enum-label follow-up

The second Angular follow-up pass removed remaining user-visible English/error
copy and backend enum leakage from transaction paging errors, budget saves,
portfolio selection fallback, SePay confidence/reclassification affordances,
asset type badges, and automation action/type columns. These surfaces now stay
understandable for Vietnamese users even when API values are machine-oriented.

The mobile follow-up applies the same trust rule to linked-bank status badges and
readonly activity statuses, with explicit Vietnamese mappings and a visible
fallback for unknown API values. The account form now also labels portfolio IDs
as “Mã danh mục”.

The core table follow-up applies the same presentation boundary to account types,
assistant command statuses, transaction types/statuses, and loan direction/status
badges. API enums remain unchanged for behavior and persistence.

SePay connection cards now localize both connection and synchronization status;
unrecognized values remain visible as an explicit “Chưa xác định” state rather
than appearing as unexplained raw backend text.

Audit and SePay compact filter selects now expose stable IDs and Vietnamese
accessible names, preserving context when the visible text is visually adjacent
but not programmatically associated.

The mobile bank-feed edit form now describes its optional category reference as
“Mã danh mục (không bắt buộc)”, making the technical field understandable without
mixing English shorthand into the primary Vietnamese workflow.

The assistant approval control now uses “Mã phê duyệt (nếu có)” with an explicit
accessible label, so the approval flow no longer exposes the raw `approvalId`
field name to users.

Mobile activity/read-only surfaces now translate common backend action verbs
(`create`, `update`, `delete`, `approve`, authentication, and resource names)
into Vietnamese labels; unknown values remain visible as a fallback for audit
traceability.

The authenticated shell now names its workspace selector, language selector, and
amount display toggle for assistive technology, and the workspace empty option is
localized to Vietnamese.

Account and loan async status messages now use polite live regions, while SePay
feed failures use an alert role so loading outcomes are announced without
requiring visual polling.

The forecast scenario table now labels its technical identifier as “Mã kịch bản”
to keep the Vietnamese information hierarchy consistent while retaining the ID
value for troubleshooting.

Mobile drawer triggers now reference a stable navigation landmark, expose their
expanded state, and switch between “Mở”/“Đóng” labels so assistive technology can
follow the same state as the visual drawer.

The auth render audit removed the remaining English login-footer copy, made
login/register actions explicit submit buttons, and marks authentication errors
as assertive alerts for immediate feedback.

Browser verification at 390×844 confirms the current login screen renders the
Vietnamese copy, compact header, language selector, and form spacing without
horizontal overflow.

The Angular smoke suite now has five passing specs, including explicit
assertions for icon aliases, audited routes, and toast severity roles.

## Route inventory completion audit

The current Angular route inventory is covered by the audit: public auth routes
(`login`, `register`, `not-found`), protected shell/resources (`dashboard`,
`accounts`, `transactions`, `loans`, `assets`, `properties`, `forecast`,
`budgets`, `portfolios`, `sepay`, `automation`, `assistant`, `audit-logs`), and
the protected `forbidden` state all have either applied UX changes or an
explicit intentional-gap note. No route is omitted from the review scope.

The smoke suite now asserts this route inventory directly; together with the
icon alias, toast-role, and mobile Escape assertions, the Angular suite contains five passing
specs.

Mobile ErrorBox and SnackBar surfaces now normalize technical exception prefixes
and network failures into user-facing Vietnamese copy; the original API/error
flow remains unchanged for diagnostics.

The mobile suite now directly tests both normalization branches; `flutter analyze`
is clean and all 17 Flutter tests pass.

Mobile ErrorBox instances now expose a live semantic region, ensuring asynchronous
load failures are announced to assistive technology instead of requiring visual
focus.

The loan test suite now directly covers normalized API-prefix and network-error
branches, protecting the user-facing copy from regression.

Shell notification dismissal and the optional assistant approval field now expose
Vietnamese, context-accurate labels for screen readers.

Automation-rule form controls now have stable IDs and explicit label associations;
the enabled toggle is also keyboard-targetable through its visible label.

Compact transaction filters and the dashboard portfolio selector now expose
explicit Vietnamese accessible names even when their visual labels are condensed.

Visible backend identifiers are now presented as “Mã …” (connection, request,
portfolio, and account) instead of exposing the technical “ID” abbreviation.

SePay reclassification copy now distinguishes the optional category from the
required manual reason, matching the reactive-form validation rules.

Mobile support and SePay copy now removes mixed English UI terms (“Map”,
“Knowledge base”, “workspace”, and “user”) while preserving product names.

The remaining mobile navigation and multi-member SePay guidance now use the same
Vietnamese “nhật ký hoạt động” and “thành viên không gian làm việc” vocabulary.

Loan history’s intentional empty state now describes the product limitation as a
system capability, without exposing the implementation term “API”.

Loan creation/customer-loading SnackBars and customer-loading inline errors now
normalize transport prefixes and network failures into the same friendly copy as
the rest of the mobile surfaces.

The loan list empty state now applies the same normalization when rendering a
view-model error, so technical prefixes do not leak through that branch either.

Web toast messages now expose an atomic live-region role (`alert` for errors,
`status` for success/info), so async feedback is announced without stealing
focus.

Shared read-only and network fallback messages in the web shell/interceptors now
use Vietnamese product copy instead of English implementation defaults.

HTTP fallback statuses now map to Vietnamese, actionable messages while retaining
the HTTP code for diagnostic context; generic browser “Request failed” text is no
longer exposed as the primary user message.

Assistant and SePay workflow statuses now use Vietnamese terms for requests,
review, and reconciliation instead of mixed “command/review/match” copy.

The Vietnamese translation dictionary now localizes the shared workspace label;
the default language no longer falls back to the English “Workspace” value.

Snapshot action/history labels now use “bản chụp số liệu” in Vietnamese, keeping
the technical snapshot concept understandable without changing API terminology.

Mobile profile and transaction surfaces now use Vietnamese comparison wording and
sentence-case titles consistently (“so với”, “Lịch sử giao dịch”, “Tạo giao dịch”).

Transaction summary cards now also use sentence case (“Tổng thu”, “Tổng chi”) to
match surrounding financial labels.

Web shell icon and avatar controls now reserve at least 44×44px hit areas, aligning
the implemented CSS with the accessibility target-size requirement.

Sidebar navigation links and toast dismissal controls now meet the same 44×44px
minimum, closing the remaining shell-level target-size gaps.

The mobile page-heading override no longer reduces action buttons to 36px; it now
preserves the 44px minimum on narrow screens as well.

The shared `.w-btn-icon` primitive is now 44×44px by default, preventing future
feature migrations from reintroducing undersized icon targets.

Mobile-only overrides now lift compact `h-8`/`h-9` buttons and `w-select-sm`
filters to 44px without changing desktop density.

The mobile selector explicitly overrides the legacy `select.w-select-sm`
`!important` height, so the 44px guarantee is effective at runtime rather than
only nominally declared.

The mobile navigation backdrop is now a semantic button with an explicit
Vietnamese accessible name, and the shell closes the open drawer on Escape.
This removes the previous `div[role=button]` keyboard gap while preserving the
same visual overlay and touch behavior.

The Angular smoke suite now covers the drawer Escape rule at mobile and desktop
widths, keeping the new keyboard behavior regression-tested (5 specs total).

Shared web controls now restore a high-contrast `:focus-visible` ring after
their visual reset, covering primary/secondary buttons, icon controls, sidebar,
mobile navigation, and the drawer backdrop.

The shell now persists the user's dark-mode choice in local storage and restores
the theme before authenticated content renders, matching the existing amount
display preference behavior.

The shared stylesheet now honors `prefers-reduced-motion`, shortening transitions
and animations across the drawer, cards, buttons, and feedback surfaces without
removing their state changes.

The mobile personal-information sheet now combines keyboard-aware insets with a
scrollable content region, preventing the three-field form from overflowing on
short screens or when the keyboard is open.

The personal limit-change sheet no longer claims a Smart OTP request succeeded
while the integration is unavailable. It now explains the limitation, disables
the unavailable action, and offers an explicit close action.

The local-only personal-information editor now says the update applies to the
current session, and generic delete failures use the shared safe formatter while
preserving the specific linked-transaction explanation.

The remaining mobile resource and form loaders (transactions, budgets, bank
feed, accounts, and generic valued resources) now also normalize inline and
SnackBar failures through the same formatter, completing the raw-exception copy
pass across those screens.

The mobile appearance sheet now truthfully exposes only the supported light
appearance and explains that dark mode is not yet available, matching the mobile
redesign specification instead of presenting a no-op dark-mode choice.

Remaining quick-action and SePay mapping labels now use “Tạo giao dịch” and
“Đã gán…” wording, removing the last visible abbreviations/mixed-English label
found during the spec copy pass.

The custom animated CTA and personal-information action buttons now use the
specification's 48px minimum control height; decorative icon wells remain
smaller without reducing their parent hit areas.

The custom animated CTA now also exposes an explicit button semantic, label, and
enabled state, closing the screen-reader gap introduced by its GestureDetector
implementation.

Profile's limit-change/edit links and the five-star satisfaction control now
expose button semantics; each star includes its numeric label and selected state
instead of relying on visual icon/color alone.

The mobile top bar's language, notification, and settings icon controls now also
expose explicit button labels, including unread-notification state.

Profile settings now consistently describe the supported light appearance, and
the rating-history action exposes a button semantic for non-visual navigation.

The satisfaction star group now marks only the current numeric rating as
selected; filled lower stars remain visual state rather than being announced as
multiple simultaneous selections.

Registration now renders a dedicated confirmation-password field and sends its
value to the existing auth flow. Password visibility icons use the dark
text-secondary token on the light field background, restoring contrast; the
view-model suite covers the missing-confirmation guard.

After successful registration, mobile now returns users to the sign-in form with
their email preserved and password fields cleared. The success message explains
the existing email-verification requirement before they sign in, rather than
leaving them in the registration state.

The success panel now opens a dedicated email-verification screen with a
six-digit, one-time-code field, resend action, and return-to-sign-in control.
The server is currently running without SMTP configuration, so development codes
are written to the backend log rather than delivered to an inbox.

Asset creation, transfer loading/submission, and SePay connection/mapping
surfaces now pass failures through the shared `presentableError` formatter.
Network failures and technical exception prefixes therefore no longer leak into
mobile SnackBars or inline connection errors.

Readonly activity/resource pages now use the same formatter for their inline
error state, so retryable audit/support failures are consistent with the rest
of the mobile application.

## Remaining intentional gaps

These are tracked rather than hidden behind fabricated UI: Angular route-level
automated coverage is still a smoke suite; mobile presentation extraction from
the legacy monolith is incremental; loan collection history and notification
feed remain explicit empty states until their APIs exist; and a shared typed
Angular component library is still a future refactor rather than a mass rewrite.
