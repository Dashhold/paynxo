# Implementation Plan: React Native Mobile Application

## Overview

This implementation plan transforms the React Native Expo mobile application design into discrete coding tasks. The mobile app replicates the Payment Gateway Operations, Commission & Settlement System web frontend with role-based access for five user types (SuperAdmin, Admin, Company, Affiliate, Merchant). The implementation follows a layered architecture: presentation layer (screens), business logic layer (hooks), and data layer (services), connecting to the existing Go backend API.

Key implementation milestones:
1. **Foundation**: Project setup, core services (API client, token storage, cache, network monitoring)
2. **Business Logic**: Custom hooks for authentication, data fetching, collections, forms
3. **UI Components**: Reusable components (forms, lists, badges, loading states, error displays)
4. **Navigation**: Portal routing system with role-based drawer navigation
5. **Feature Screens**: Authentication, dashboard, collections management (companies, merchants, affiliates, gateways, banks, transactions, settlements, ledgers, reports, leases)
6. **Testing**: Property-based tests for pure functions and universal behaviors, unit tests for components

## Tasks

- [x] 1. Set up project structure and core dependencies
  - Initialize Expo React Native project with TypeScript template
  - Install React Navigation 6.x (stack + drawer navigation)
  - Install core dependencies: axios, @react-native-async-storage/async-storage, expo-secure-store, @react-native-community/netinfo
  - Install form libraries: react-hook-form, yup, @hookform/resolvers
  - Install UI component library: React Native Paper or Native Base
  - Install testing libraries: jest, @testing-library/react-native, fast-check
  - Create folder structure: components/, screens/, services/, hooks/, models/, utils/, navigation/, context/, config/
  - Configure app.json with Android settings (minSdkVersion: 21, permissions, app metadata)
  - Configure tsconfig.json with strict type checking
  - Configure jest.config.js for React Native testing
  - Create README.md with setup and build instructions
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 2. Implement API client service with interceptors
  - [x] 2.1 Create services/apiClient.ts with ApiClient class
    - Implement HTTP methods: get<T>, post<T>, put<T>, delete<T> with Axios
    - Configure base URL and 30-second timeout
    - Create TypeScript interfaces: ApiResponse<T>, ApiError, ApiClientConfig
    - _Requirements: 29.1, 29.2_
  
  - [x] 2.2 Implement authentication interceptors
    - Add request interceptor to attach Bearer token from TokenStore to Authorization header
    - Add response interceptor to handle 401 responses (clear token, clear cache, navigate to login)
    - Add response interceptor to handle network errors with user-friendly messages
    - Enforce HTTPS-only for production builds
    - _Requirements: 3.2, 4.1, 4.2, 4.3, 4.4, 29.3, 29.4_
  
  - [ ]* 2.3 Write unit tests for API client
    - Test HTTP methods with mocked Axios responses
    - Test request interceptor attaches token correctly
    - Test 401 response triggers session termination
    - Test network error handling

- [x] 3. Implement secure token storage service
  - Create services/tokenStore.ts with SecureTokenStore class implementing TokenStore interface
  - Implement saveToken(), getToken(), clearToken(), hasToken() using expo-secure-store
  - Store JWT with key 'auth_token'
  - Create TypeScript interface for TokenStore
  - _Requirements: 3.1, 3.3_


- [ ]* 3.1 Write unit tests for token storage
    - Test saveToken, getToken, clearToken, hasToken operations

- [x] 4. Implement cache manager service
  - Create services/cacheManager.ts with CacheManager class
  - Implement set<T>(), get<T>(), has(), isStale(), invalidate(), invalidatePattern(), clear()
  - Use AsyncStorage for cache storage with TTL validation (default: 5 minutes)
  - Document cache key structure: "{collection}:{id?}:{filter?}"
  - Create TypeScript interfaces: CacheEntry<T>, CacheConfig
  - _Requirements: 28.1, 28.2, 28.5_

- [ ]* 4.1 Write unit tests for cache manager
    - Test cache operations with TTL expiry scenarios
    - Test pattern-based invalidation

- [x] 5. Implement network monitoring service
  - Create services/networkMonitor.ts with NetworkMonitor class
  - Implement getCurrentState(), subscribe(), isOnline() using @react-native-community/netinfo
  - Create TypeScript interface for NetworkState
  - _Requirements: 28.3, 32.1_

- [ ]* 5.1 Write unit tests for network monitor
    - Test connectivity state detection with mocked NetInfo


- [x] 6. Create TypeScript type definitions for all data models
  - Create models/auth.ts: LoginRequest, LoginResponse, UserPrincipal
  - Create models/entities.ts: Company, Merchant, Affiliate, Gateway, Bank, BankAccount
  - Create models/transaction.ts: Transaction, CommissionBreakdown, Settlement, LedgerEntry
  - Create models/admin.ts: Lease, Report, ReportParameters, DashboardMetrics
  - Create models/api.ts: PaginatedRequest, PaginatedResponse<T>, ApiError
  - Define enums: TransactionStatus, SettlementStatus, LeaseStatus, EntityStatus, UserRole
  - Export all types from models/index.ts
  - _Requirements: All (supports type safety across entire app)_

- [x] 7. Implement formatting utility functions
  - Create utils/formatters.ts with pure formatting functions
  - Implement formatCurrency(amount): string - Indian Rupee format with ₹, comma separators (lakhs/crores), 2 decimals
  - Implement formatNumber(value): string - locale-appropriate number formatting with commas
  - Implement formatPercentage(value): string - adds % symbol and preserves numeric value
  - Implement formatDate(date): string - readable format (e.g., "Jan 15, 2024")
  - Implement formatRelativeDate(date): string - relative dates ("Today", "Yesterday", "2 days ago")
  - Implement formatDateTime(date): string - full timestamp ("Jan 15, 2024 at 3:45 PM")
  - _Requirements: 37.1, 37.2, 37.3, 37.4_


- [ ]* 7.1 Write property-based tests for formatting functions
    - **Property 1: Currency Formatting Consistency** - for any numeric value, formatted currency includes ₹, Indian separators, 2 decimals (fast-check, 100 iterations)
    - **Property 2: Percentage Formatting** - for any numeric value, formatted percentage includes % and preserves value (fast-check, 100 iterations)
    - **Property 7: Decimal Place Limitation** - for any amount with arbitrary decimals, formatting truncates/rounds to 2 decimal places (fast-check, 100 iterations)
    - Tag: "Feature: react-native-mobile-app, Property {N}: {description}"
    - **Validates: Requirements 37.1, 37.2, 37.3, 37.4**

- [ ]* 7.2 Write unit tests for formatting edge cases
    - Test negative numbers, zero, very large numbers, null/undefined handling

- [x] 8. Implement validation utility functions
  - Create utils/validators.ts with pure validation functions
  - Implement isValidEmail(email): boolean - RFC 5322 simplified pattern validation
  - Implement isValidPhone(phone): boolean - phone number format validation
  - Implement isNumeric(value): boolean - check if string represents valid numeric value
  - Implement isPositiveNumber(value): boolean - check if number > 0
  - Implement isRequiredFieldValid(value): boolean - check if field is not empty or whitespace-only
  - _Requirements: 43.1, 43.2, 43.3, 43.4, 43.5, 61.1, 61.2_


- [ ]* 8.1 Write property-based tests for validation functions
    - **Property 3: Email Validation Consistency** - accepts all valid email formats, rejects invalid formats (fast-check, 100 iterations)
    - **Property 4: Phone Validation Consistency** - accepts valid phone formats, rejects invalid formats (fast-check, 100 iterations)
    - **Property 5: Numeric Input Validation** - accepts numeric strings, rejects non-numeric strings (fast-check, 100 iterations)
    - **Property 6: Positive Number Validation** - accepts positive numbers, rejects zero and negatives (fast-check, 100 iterations)
    - **Property 8: Required Field Validation** - prevents submission when empty or whitespace-only (fast-check, 100 iterations)
    - Tag: "Feature: react-native-mobile-app, Property {N}: {description}"
    - **Validates: Requirements 43.1, 43.2, 43.3, 43.4, 43.5, 61.1, 61.2**

- [ ]* 8.2 Write unit tests for validation edge cases
    - Test specific valid/invalid inputs for each validator

- [ ]* 9. Write property-based tests for universal API client behaviors
  - **Property 9: 401 Response Triggers Token Clearing** - for any API endpoint returning 401, verify token is cleared, cache is cleared, navigation to login occurs (fast-check with mocked API, 100 iterations)
  - **Property 10: Bearer Token Attachment** - for any authenticated API request to any endpoint, verify Bearer token is in Authorization header (fast-check with random endpoints, 100 iterations)
  - Set up mocked API client responses for property tests
  - Generate random endpoint paths and HTTP methods for testing
  - Tag: "Feature: react-native-mobile-app, Property {N}: {description}"
  - **Validates: Requirements 3.2, 4.1, 4.2, 4.3**


- [x] 10. Implement useAuth hook for authentication state management
  - Create hooks/useAuth.ts with useAuth hook
  - Manage authentication state: isAuthenticated, isLoading, user (UserPrincipal), error
  - Implement login(userId, password) that calls POST /api/auth/login
  - Store JWT token in SecureStore on successful login
  - Extract UserPrincipal from login response and update state
  - Handle errors: 401 → "Invalid credentials", 403 → server message, network → "Cannot reach server"
  - Implement logout() that calls POST /api/auth/logout, clears token, clears cache, navigates to login
  - Implement restoreSession() that validates stored token with GET /api/me
  - Navigate to appropriate portal based on user role after successful authentication
  - Create TypeScript interfaces: AuthState, UseAuthReturn
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.10, 3.4, 3.5, 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ]* 10.1 Write unit tests for useAuth hook
    - Test login, logout, session restoration with mocked API and navigation

- [x] 11. Implement useAPI hook for data fetching with caching
  - Create hooks/useAPI.ts with useAPI<T> generic hook
  - Fetch data from API using ApiClient with caching via CacheManager
  - Manage state: data, loading, error
  - Implement refetch() for manual refresh
  - Implement mutate(newData) to update local cache
  - Support options: cacheKey, cacheTTL, skipCache, onSuccess, onError
  - Create TypeScript interfaces: UseAPIOptions, UseAPIReturn<T>
  - _Requirements: 28.1, 28.2, 30.1, 30.3_


- [ ]* 11.1 Write unit tests for useAPI hook
    - Test data fetching, caching, refetch, mutate with mocked API

- [x] 12. Implement useCollection hook for CRUD operations
  - Create hooks/useCollection.ts with useCollection<T> generic hook
  - Implement collection data fetching with pagination (pageSize: 50)
  - Implement filtering, sorting (field + order asc/desc), searching
  - Implement refresh() for pull-to-refresh
  - Implement loadMore() for infinite scroll
  - Implement create(item), update(id, item), remove(id) with cache invalidation
  - Implement setFilters(), setSort(), setSearch() for UI control
  - Track totalCount, hasMore for pagination UI
  - Create TypeScript interfaces: UseCollectionOptions, UseCollectionReturn<T>
  - _Requirements: 17.6, 17.7, 17.8, 18.6, 18.7, 18.8, 19.6, 19.7, 19.8_

- [ ]* 12.1 Write unit tests for useCollection hook
    - Test CRUD operations, pagination, filtering, sorting with mocked API

- [x] 13. Implement useForm hook wrapper for react-hook-form
  - Create hooks/useForm.ts wrapping react-hook-form with Yup validation
  - Integrate yupResolver for schema validation
  - Implement handleSubmit with loading state management
  - Expose control, errors, isSubmitting, isValid, reset(), setValue()
  - Support options: schema (Yup), defaultValues, onSubmit callback
  - Create TypeScript interfaces: UseFormOptions<T>, UseFormReturn<T>
  - _Requirements: 43.1, 43.6, 44.1, 44.2, 60.1, 60.2_


- [ ]* 13.1 Write unit tests for useForm hook
    - Test form validation, submission, error handling with Yup schemas

- [x] 14. Implement core reusable UI components
  - [x] 14.1 Create FormInput component
    - Create components/FormInput.tsx with React Hook Form integration
    - Support input types: text, email, password, number
    - Display label above input, error message below in red
    - Implement password visibility toggle for secure text entry
    - Support appropriate keyboard types (numeric, email-address, etc.)
    - Support props: multiline, disabled, placeholder
    - Implement accessibility labels (WCAG AA)
    - _Requirements: 2.9, 43.1, 43.2, 43.3, 43.6, 44.3_
  
  - [x] 14.2 Create StatusBadge component
    - Create components/StatusBadge.tsx with color-coded status pills
    - Support status types with colors: success (green), pending (yellow), failed (red), active (green), inactive (gray), suspended (gray), expired (gray), revoked (red)
    - _Requirements: 38.1, 38.2, 38.3_
  
  - [x] 14.3 Create ListItem component
    - Create components/ListItem.tsx for list row display
    - Support: title, subtitle, rightText, status badge, left icon, chevron right indicator
    - Implement TouchableOpacity for onPress handler
    - Implement accessibility labels (WCAG AA)
    - _Requirements: 12.2, 13.3, 14.2, 15.2, 16.2, 17.3, 18.3, 19.3_


- [ ]* 14.4 Write component tests for core UI components
    - Test FormInput, StatusBadge, ListItem with React Native Testing Library

- [x] 15. Implement additional reusable UI components
  - [x] 15.1 Create EmptyState component
    - Create components/EmptyState.tsx for empty list displays
    - Support: icon, title, message, optional action button with onActionPress
    - _Requirements: 40.1_
  
  - [x] 15.2 Create ErrorDisplay component
    - Create components/ErrorDisplay.tsx for error messages
    - Display error message from ApiError or generic Error
    - Support retry button with onRetry callback
    - Parse structured ApiError vs generic Error for appropriate messaging
    - _Requirements: 29.1, 29.2, 29.3, 29.5_
  
  - [x] 15.3 Create LoadingSkeleton component
    - Create components/LoadingSkeleton.tsx for loading placeholders
    - Support skeleton types: list, details, dashboard
    - Implement shimmering animation effect
    - _Requirements: 30.1, 30.3_
  
  - [x] 15.4 Create OfflineBanner component
    - Create components/OfflineBanner.tsx for offline indicator
    - Display persistent banner when network is unavailable
    - _Requirements: 28.3_


- [ ]* 15.5 Write component tests for additional UI components
    - Test EmptyState, ErrorDisplay, LoadingSkeleton, OfflineBanner with React Native Testing Library

- [x] 16. Create authentication context and provider
  - Create context/AuthContext.tsx with AuthContext and AuthProvider
  - Wrap authentication state from useAuth hook in context
  - Export useAuthContext hook for consuming authentication state
  - Provide authentication state and actions (login, logout, restoreSession) to entire app
  - Create TypeScript interface for AuthContextValue
  - _Requirements: All authentication requirements_

- [ ]* 16.1 Write unit tests for AuthProvider
    - Test AuthProvider and context consumption

- [x] 17. Set up navigation structure and root navigator
  - Create navigation/RootNavigator.tsx as main navigation entry point
  - Implement conditional navigation: AuthNavigator when not authenticated, PortalWrapper when authenticated
  - Create navigation/AuthNavigator.tsx with stack navigator for LoginScreen
  - Configure navigation theme and options (header styles, transitions)
  - Set up deep linking configuration for major sections
  - Create TypeScript types for navigation stack parameters
  - _Requirements: 2.10, 6.1, 7.1, 8.1, 9.1, 10.1_


- [ ]* 17.1 Write property-based test for role-based portal routing
    - **Property 11: Role-Based Portal Routing** - for any valid user role, verify correct portal component is rendered (fast-check, 100 iterations)
    - Tag: "Feature: react-native-mobile-app, Property 11: Role-Based Portal Routing"
    - **Validates: Requirements 2.10**

- [x] 18. Implement portal wrapper and drawer navigation
  - Create navigation/PortalWrapper.tsx that selects portal based on UserPrincipal.role
  - Implement role-to-portal mapping: SuperAdmin→SuperAdminPortal, Admin→AdminPortal, Company→CompanyPortal, Affiliate→AffiliatePortal, Merchant→MerchantPortal
  - Render appropriate drawer navigation for each portal
  - Include header with user info (name, role) and logout button
  - Display offline indicator banner when network is unavailable
  - Support deep link routing to specific screens
  - Handle unknown roles gracefully with error screen
  - _Requirements: 6.1, 7.1, 8.1, 9.1, 10.1_

- [ ]* 18.1 Write component tests for portal selection logic
    - Test portal rendering for each role

- [x] 19. Checkpoint - Ensure foundation is solid
  - Ensure all tests pass for services, hooks, components
  - Verify authentication flow works end-to-end (login, session restore, logout)
  - Verify API client interceptors work correctly
  - Ask the user if questions arise


- [x] 20. Implement SuperAdmin portal navigation
  - Create navigation/SuperAdminPortal.tsx with drawer navigator
  - Add drawer navigation items: Dashboard, Companies, Merchants, Affiliates, Gateways, Banks, Transactions, Settlements, Ledgers, Reports, Leases
  - Configure drawer UI with user profile section at top
  - Highlight currently active screen in drawer
  - Include logout option in drawer
  - Set up stack navigators for each section (list → details → form screens)
  - Configure header options for each screen
  - _Requirements: 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8, 6.9, 6.10, 6.11, 6.12_

- [ ]* 20.1 Write component tests for SuperAdmin portal navigation
    - Test drawer navigation and screen routing

- [x] 21. Implement Admin portal navigation
  - Create navigation/AdminPortal.tsx with drawer navigator
  - Add drawer navigation items: Dashboard, Companies, Merchants, Affiliates, Gateways, Banks, Transactions, Settlements, Ledgers, Reports (no Leases)
  - Reuse drawer UI components from SuperAdmin portal
  - Set up stack navigators for each section
  - _Requirements: 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8, 7.9, 7.10, 7.11, 7.12_

- [ ]* 21.1 Write component tests for Admin portal navigation
    - Verify Leases navigation is not present


- [x] 22. Implement Company portal navigation
  - Create navigation/CompanyPortal.tsx with drawer navigator
  - Add drawer navigation items: Dashboard, Merchants (company-owned only), Transactions, Settlements, Ledger
  - Configure drawer UI for company role
  - Set up stack navigators for each section
  - Verify no access to Companies, Affiliates, Gateways, Banks, or Reports
  - _Requirements: 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_

- [ ]* 22.1 Write component tests for Company portal navigation
    - Test company-scoped navigation

- [x] 23. Implement Affiliate portal navigation
  - Create navigation/AffiliatePortal.tsx with drawer navigator
  - Add drawer navigation items: Dashboard, Merchants (referred by affiliate only), Transactions, Ledger
  - Configure drawer UI for affiliate role
  - Set up stack navigators for each section
  - Verify no access to Companies, Affiliates, Gateways, Banks, Settlements, or Reports
  - _Requirements: 9.2, 9.3, 9.4, 9.5, 9.6_

- [ ]* 23.1 Write component tests for Affiliate portal navigation
    - Test affiliate-scoped navigation


- [x] 24. Implement Merchant portal navigation
  - Create navigation/MerchantPortal.tsx with drawer navigator
  - Add drawer navigation items: Dashboard, Transactions, Bank Accounts, Ledger
  - Configure drawer UI for merchant role
  - Set up stack navigators for each section
  - Verify no access to Companies, Affiliates, Merchants, Gateways, Banks, Settlements, or Reports
  - _Requirements: 10.2, 10.3, 10.4, 10.5, 10.6_

- [ ]* 24.1 Write component tests for Merchant portal navigation
    - Test merchant-scoped navigation

- [x] 25. Implement Login screen
  - Create screens/auth/LoginScreen.tsx with login form
  - Implement form with userId (text input) and password (secure text input) fields
  - Add password visibility toggle control
  - Display "Payment Gateway Ops" branding and logo
  - Call useAuth login() function on form submission
  - Display loading indicator on submit button during authentication
  - Display error messages: 401 → "Invalid credentials", 403 → server message, network → "Cannot reach server"
  - Navigate to appropriate portal on successful login based on user role
  - Support auto-fill and keyboard management (dismiss on tap outside)
  - Implement accessibility labels for all inputs and buttons (WCAG AA)
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, 2.10_


- [ ]* 25.1 Write component tests for Login screen
    - Test login flow, error handling, navigation

- [x] 26. Implement Dashboard screen
  - Create screens/dashboard/DashboardScreen.tsx
  - Fetch dashboard data from API using useAPI hook (endpoint varies by role)
  - Display key metrics in card layout with labels, values, and optional trend indicators
  - Display recent activity list with timestamps and descriptions
  - Implement pull-to-refresh functionality
  - Display LoadingSkeleton while fetching data
  - Display ErrorDisplay with retry button on fetch failure
  - Optimize layout for mobile screen sizes with scrollable content
  - Use formatCurrency and formatNumber utilities for metric values
  - Implement accessibility labels for metrics (WCAG AA)
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6_

- [ ]* 26.1 Write component tests for Dashboard screen
    - Test data fetching, loading states, error handling, pull-to-refresh

- [x] 27. Implement generic ListScreen component
  - Create screens/ListScreen.tsx as generic component with TypeScript generics <T>
  - Accept props: collectionName, title, renderItem, onItemPress, filterOptions, searchPlaceholder, canCreate, onCreatePress
  - Use useCollection hook to fetch and manage collection data
  - Implement search bar at top with debounced search (300ms delay)
  - Implement filter buttons for common criteria with active state indicators
  - Display "Clear filters" button when filters are active
  - Implement FlatList with optimized rendering (windowSize, removeClippedSubviews)
  - Implement pull-to-refresh functionality
  - Implement infinite scroll pagination (onEndReached to load more pages)
  - Display LoadingSkeleton while initial data loads
  - Display EmptyState when no items found
  - Display ErrorDisplay with retry on fetch failure
  - Support smooth 60fps scrolling with memoized list items
  - _Requirements: 12.2, 12.8, 13.3, 13.11, 14.2, 14.7, 15.2, 15.7, 16.2, 16.7, 17.3, 17.8, 17.11, 18.3, 18.8, 18.11, 19.3, 19.8, 19.11, 21.2, 21.9_


- [ ]* 27.1 Write component tests for ListScreen
    - Test search, filter, pagination, error handling

- [x] 28. Implement generic DetailsScreen component
  - Create screens/DetailsScreen.tsx as generic component with TypeScript generics <T>
  - Accept props: itemId, collectionName, renderDetails, actions, canEdit, onEditPress
  - Use useAPI hook to fetch single item details
  - Display item details using renderDetails render prop
  - Display action buttons at bottom (configurable: primary, danger variants)
  - Display edit button in header if canEdit is true
  - Display LoadingSkeleton while data loads
  - Display ErrorDisplay with retry on fetch failure
  - Support scrollable content for long details
  - Implement accessibility labels for action buttons (WCAG AA)
  - _Requirements: 12.4, 13.4, 14.4, 15.4, 16.4, 21.4_

- [ ]* 28.1 Write component tests for DetailsScreen
    - Test data fetching, actions, edit navigation, error handling

- [x] 29. Implement generic FormScreen component
  - Create screens/FormScreen.tsx as generic component with TypeScript generics <T>
  - Accept props: title, fields (array of FormField configs), schema (Yup), defaultValues, onSubmit, submitLabel
  - Use useForm hook with Yup validation
  - Dynamically render form fields based on field type (text, email, number, select, date, switch)
  - Display validation errors inline below each field in red
  - Disable submit button when validation errors exist or form is submitting
  - Display loading state on submit button ("Saving..." or "Creating...")
  - Dismiss keyboard on submit
  - Support appropriate keyboard types for each field (numeric, email-address, etc.)
  - Navigate back on successful submission
  - Display error message on submission failure
  - Implement accessibility labels for all form inputs (WCAG AA)
  - _Requirements: 12.3, 12.5, 12.6, 13.5, 13.6, 14.3, 14.5, 14.6, 15.3, 15.5, 15.6, 16.3, 16.5, 16.6, 21.3, 21.5, 21.6, 24.1, 24.2, 24.3, 24.4, 24.5, 24.6, 24.7, 25.1, 25.2, 25.3, 25.4, 25.5, 25.6, 25.7, 26.1, 26.2, 26.3, 26.4, 26.5, 26.6, 26.7_


- [ ]* 29.1 Write component tests for FormScreen
    - Test form validation, submission, error handling

- [x] 30. Checkpoint - Ensure navigation and core screens work
  - Verify all portal navigations render correctly for each role
  - Verify Login screen authenticates and navigates to correct portal
  - Verify Dashboard screen fetches and displays data
  - Verify generic ListScreen, DetailsScreen, FormScreen work as expected
  - Ask the user if questions arise

- [x] 31. Implement Companies management screens
  - [x] 31.1 Create CompanyListScreen
    - Create screens/companies/CompanyListScreen.tsx using ListScreen component
    - Fetch companies via GET /api/companies using useCollection hook
    - Display company list items with name, email, status badge
    - Implement search by name and email
    - Implement filter by status (active, inactive)
    - Add "Create Company" button for authorized roles
    - _Requirements: 12.1, 12.2, 12.3, 12.8, 12.9, 12.10_
  
  - [x] 31.2 Create CompanyDetailsScreen
    - Create screens/companies/CompanyDetailsScreen.tsx using DetailsScreen component
    - Display company information: name, email, phone, address, status, timestamps
    - Add "Edit Company" button in header
    - Add "Payment Management" navigation action
    - _Requirements: 12.4, 12.7_


  - [x] 31.3 Create CompanyFormScreen
    - Create screens/companies/CompanyFormScreen.tsx using FormScreen component
    - Include fields: name, email, phone, address, status
    - Implement Yup schema validation for required fields and email format
    - Handle POST /api/companies (create) and PUT /api/companies/{id} (update)
    - Display success message and navigate to list on successful save
    - _Requirements: 12.5, 12.6, 24.1, 24.2, 24.3, 24.4, 24.5, 24.6, 24.7_

- [ ]* 31.4 Write component tests for Companies management
    - Test list, details, create, edit flows

- [x] 32. Implement Merchants management screens
  - [x] 32.1 Create MerchantListScreen
    - Create screens/merchants/MerchantListScreen.tsx using ListScreen component
    - Fetch merchants via GET /api/merchants (server-side scoped to company if Company role)
    - Display merchant list items with name, company name, status badge
    - Implement search by name and company
    - Implement filter by status (active, inactive)
    - Add "Create Merchant" button for SuperAdmin/Admin only (hide for Company role)
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.11, 13.12, 13.13_
  
  - [x] 32.2 Create MerchantDetailsScreen
    - Create screens/merchants/MerchantDetailsScreen.tsx using DetailsScreen component
    - Display merchant information: name, company, affiliate, email, phone, commission rate, status, timestamps
    - Display associated bank accounts list with "Manage Bank Accounts" action
    - Display assigned payment gateways with "Manage Gateways" action
    - Add "Edit Merchant" button in header for authorized roles
    - _Requirements: 13.5, 13.6, 13.9, 13.10_


  - [x] 32.3 Create MerchantFormScreen
    - Create screens/merchants/MerchantFormScreen.tsx using FormScreen component
    - Include fields: name, companyId (select), affiliateId (select, optional), email, phone, commissionRate (number), status
    - Implement Yup schema validation for required fields, email format, positive commission rate
    - Handle POST /api/merchants (create) and PUT /api/merchants/{id} (update)
    - Display success message and navigate to list on successful save
    - _Requirements: 13.7, 13.8, 25.1, 25.2, 25.3, 25.4, 25.5, 25.6, 25.7_

- [ ]* 32.4 Write component tests for Merchants management
    - Test list, details, create, edit, scoped access

- [x] 33. Implement Affiliates management screens
  - [x] 33.1 Create AffiliateListScreen
    - Create screens/affiliates/AffiliateListScreen.tsx using ListScreen component
    - Fetch affiliates via GET /api/affiliates using useCollection hook
    - Display affiliate list items with name, email, commission rate, status badge
    - Implement search by name and email
    - Implement filter by status (active, inactive)
    - Add "Create Affiliate" button
    - _Requirements: 14.1, 14.2, 14.3, 14.7, 14.8, 14.9_
  
  - [x] 33.2 Create AffiliateDetailsScreen
    - Create screens/affiliates/AffiliateDetailsScreen.tsx using DetailsScreen component
    - Display affiliate information: name, email, phone, commission rate, status, timestamps
    - Add "Edit Affiliate" button in header
    - _Requirements: 14.4, 14.5_


  - [x] 33.3 Create AffiliateFormScreen
    - Create screens/affiliates/AffiliateFormScreen.tsx using FormScreen component
    - Include fields: name, email, phone, commissionRate (number), status
    - Implement Yup schema validation for required fields, email format, positive commission rate
    - Handle POST /api/affiliates (create) and PUT /api/affiliates/{id} (update)
    - Display success message and navigate to list on successful save
    - _Requirements: 14.6, 26.1, 26.2, 26.3, 26.4, 26.5, 26.6, 26.7_

- [ ]* 33.4 Write component tests for Affiliates management
    - Test list, details, create, edit

- [x] 34. Implement Gateways management screens
  - [x] 34.1 Create GatewayListScreen
    - Create screens/gateways/GatewayListScreen.tsx using ListScreen component
    - Fetch gateways via GET /api/gateways using useCollection hook
    - Display gateway list items with name, type, status badge
    - Implement search by name
    - Implement filter by status (active, inactive)
    - Add "Create Gateway" button
    - _Requirements: 15.1, 15.2, 15.3, 15.7, 15.8, 15.9_
  
  - [x] 34.2 Create GatewayDetailsScreen
    - Create screens/gateways/GatewayDetailsScreen.tsx using DetailsScreen component
    - Display gateway information: name, type, status, credentials (masked), timestamps
    - Add "Edit Gateway" button in header
    - _Requirements: 15.4, 15.5_


  - [x] 34.3 Create GatewayFormScreen
    - Create screens/gateways/GatewayFormScreen.tsx using FormScreen component
    - Include fields: name, type (select), status, credentials (secure inputs)
    - Implement Yup schema validation for required fields
    - Handle POST /api/gateways (create) and PUT /api/gateways/{id} (update)
    - Display success message and navigate to list on successful save
    - _Requirements: 15.6_

- [ ]* 34.4 Write component tests for Gateways management
    - Test list, details, create, edit

- [x] 35. Implement Banks management screens
  - [x] 35.1 Create BankListScreen
    - Create screens/banks/BankListScreen.tsx using ListScreen component
    - Fetch banks via GET /api/banks using useCollection hook
    - Display bank list items with name, code, swift code
    - Implement search by name and code
    - Add "Create Bank" button
    - _Requirements: 16.1, 16.2, 16.3, 16.7, 16.8, 16.9_
  
  - [x] 35.2 Create BankDetailsScreen
    - Create screens/banks/BankDetailsScreen.tsx using DetailsScreen component
    - Display bank information: name, code, swift code, timestamps
    - Add "Edit Bank" button in header
    - _Requirements: 16.4, 16.5_


  - [x] 35.3 Create BankFormScreen
    - Create screens/banks/BankFormScreen.tsx using FormScreen component
    - Include fields: name, code, swiftCode (optional)
    - Implement Yup schema validation for required fields
    - Handle POST /api/banks (create) and PUT /api/banks/{id} (update)
    - Display success message and navigate to list on successful save
    - _Requirements: 16.6_

- [ ]* 35.4 Write component tests for Banks management
    - Test list, details, create, edit

- [x] 36. Implement Transactions viewing screens
  - [x] 36.1 Create TransactionListScreen
    - Create screens/transactions/TransactionListScreen.tsx using ListScreen component
    - Fetch transactions via GET /api/transactions (server-side filtered by role)
    - Display transaction list items with date, merchant name, amount, status badge
    - Implement search by merchant, reference
    - Implement filter by date range, status, merchant, amount range
    - Implement sort by date, amount, status
    - Implement pagination with infinite scroll
    - Implement pull-to-refresh
    - _Requirements: 17.1, 17.2, 17.3, 17.6, 17.7, 17.8, 17.9, 17.10, 17.11_
  
  - [x] 36.2 Create TransactionDetailsScreen
    - Create screens/transactions/TransactionDetailsScreen.tsx using DetailsScreen component
    - Display transaction information: merchant, company, affiliate, gateway, amount, currency, status, commission breakdown, reference, timestamps
    - Use formatCurrency for amount display
    - _Requirements: 17.4, 17.5_


- [ ]* 36.3 Write component tests for Transactions viewing
    - Test list with filters, sorting, pagination, details display

- [x] 37. Implement Settlements management screens
  - [x] 37.1 Create SettlementListScreen
    - Create screens/settlements/SettlementListScreen.tsx using ListScreen component
    - Fetch settlements via GET /api/settlements (server-side scoped to company if Company role)
    - Display settlement list items with date, merchant name, amount, status badge
    - Implement search by merchant, reference
    - Implement filter by date range, status, merchant
    - Implement sort by date, amount
    - Implement pagination with infinite scroll
    - Implement pull-to-refresh
    - _Requirements: 18.1, 18.2, 18.3, 18.6, 18.7, 18.8, 18.9, 18.10, 18.11_
  
  - [x] 37.2 Create SettlementDetailsScreen
    - Create screens/settlements/SettlementDetailsScreen.tsx using DetailsScreen component
    - Display settlement information: merchant, company, amount, currency, status, settlement date, bank account, reference, timestamps
    - Use formatCurrency for amount display
    - _Requirements: 18.4, 18.5_

- [ ]* 37.3 Write component tests for Settlements management
    - Test list with filters, sorting, pagination, details display


- [x] 38. Implement Ledgers viewing screens
  - [x] 38.1 Create LedgerListScreen
    - Create screens/ledger/LedgerListScreen.tsx using ListScreen component
    - Fetch ledger entries from appropriate endpoint based on role (server-side filtered)
    - Display ledger entries with date, description, debit, credit, balance
    - Display current balance prominently at top
    - Implement filter by date range, entry type
    - Implement sort by date
    - Implement pagination with infinite scroll
    - Implement pull-to-refresh
    - Use formatCurrency for amount display
    - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.6, 19.7, 19.8, 19.9, 19.10, 19.11_
  
  - [x] 38.2 Create LedgerEntryDetailsScreen
    - Create screens/ledger/LedgerEntryDetailsScreen.tsx using DetailsScreen component
    - Display ledger entry information: entity, transaction/settlement reference, type, amount, balance, description, timestamps
    - Use formatCurrency for amount display
    - _Requirements: 19.5_

- [ ]* 38.3 Write component tests for Ledgers viewing
    - Test list with filters, sorting, pagination, details display


- [x] 39. Implement Reports generation screens
  - [x] 39.1 Create ReportGenerationScreen
    - Create screens/reports/ReportGenerationScreen.tsx
    - Display report generation form with report type selection
    - Include date range selection (start date, end date)
    - Include filter selections: merchant, company, affiliate, gateway (using select dropdowns)
    - Implement "Generate Report" button that fetches report data from API
    - Display loading indicator while generating report
    - _Requirements: 20.1, 20.2, 20.3, 20.4, 20.5, 20.8_
  
  - [x] 39.2 Create ReportResultsScreen
    - Create screens/reports/ReportResultsScreen.tsx
    - Display report results in mobile-optimized format (tables, charts)
    - Provide export options (PDF, CSV) with share functionality
    - Display error message with retry button on generation failure
    - _Requirements: 20.6, 20.7, 20.9_

- [ ]* 39.3 Write component tests for Reports generation
    - Test report generation flow, results display, error handling

- [x] 40. Implement Leases management screens (SuperAdmin only)
  - [x] 40.1 Create LeaseListScreen
    - Create screens/leases/LeaseListScreen.tsx using ListScreen component
    - Fetch leases via GET /api/leases using useCollection hook
    - Display lease list items with tenant name, status badge, start date, end date
    - Implement search by tenant
    - Implement filter by status (active, suspended, expired, revoked)
    - Implement sort by start date, end date
    - Add "Create Lease" button
    - Verify only accessible to SuperAdmin role
    - _Requirements: 21.1, 21.2, 21.3, 21.8, 21.9, 21.10, 21.12_


  - [x] 40.2 Create LeaseDetailsScreen
    - Create screens/leases/LeaseDetailsScreen.tsx using DetailsScreen component
    - Display lease information: tenant details, status, access level, start date, end date, timestamps
    - Allow editing lease status (active, suspended, revoked)
    - Add "Edit Lease" button in header
    - _Requirements: 21.4, 21.5, 21.6_
  
  - [x] 40.3 Create LeaseFormScreen
    - Create screens/leases/LeaseFormScreen.tsx using FormScreen component
    - Include fields: tenantId (select), status, accessLevel, startDate, endDate
    - Implement Yup schema validation for required fields, date validation (end date > start date)
    - Handle POST /api/leases (create) and PUT /api/leases/{id} (update)
    - Display success message and navigate to list on successful save
    - _Requirements: 21.3, 21.7, 21.11_

- [ ]* 40.4 Write component tests for Leases management
    - Test list, details, create, edit, SuperAdmin-only access

- [x] 41. Implement Bank Accounts management screens
  - [x] 41.1 Create BankAccountListScreen
    - Create screens/bankAccounts/BankAccountListScreen.tsx using ListScreen component
    - Display bank accounts for merchant (scoped to current user if Merchant role)
    - Display bank account list items with bank name, account number, status badge
    - Add "Add Bank Account" button
    - _Requirements: 22.1, 22.2, 22.3, 22.4, 22.9_


  - [x] 41.2 Create BankAccountDetailsScreen
    - Create screens/bankAccounts/BankAccountDetailsScreen.tsx using DetailsScreen component
    - Display bank account information: bank name, account number, account holder name, IFSC code, status, timestamps
    - Allow editing account information
    - Allow deleting bank account with confirmation
    - _Requirements: 22.5, 22.6, 22.7_
  
  - [x] 41.3 Create BankAccountFormScreen
    - Create screens/bankAccounts/BankAccountFormScreen.tsx using FormScreen component
    - Include fields: bankId (select), accountNumber, accountHolderName, ifscCode (optional), status
    - Implement Yup schema validation for required fields, account number format
    - Handle POST and PUT via updating merchant record (PUT /api/merchants/{id})
    - Display success message on save
    - _Requirements: 22.8_

- [ ]* 41.4 Write component tests for Bank Accounts management
    - Test list, details, add, edit, delete

- [x] 42. Implement Payment Gateway Assignment screens
  - [x] 42.1 Create GatewayAssignmentScreen
    - Create screens/gateways/GatewayAssignmentScreen.tsx for merchant gateway assignment
    - Display currently assigned gateways with status and credentials
    - Provide "Assign Gateway" button to add new gateway
    - Allow removing assigned gateway with confirmation
    - _Requirements: 23.1, 23.2, 23.6_


  - [x] 42.2 Create GatewayAssignmentFormScreen
    - Create screens/gateways/GatewayAssignmentFormScreen.tsx
    - Display available gateways for selection
    - Include fields for gateway-specific credentials (dynamic based on gateway type)
    - Implement Yup schema validation for credentials
    - Handle assignment via updating merchant record (PUT /api/merchants/{id})
    - Display loading indicator while saving
    - _Requirements: 23.3, 23.4, 23.5, 23.7_

- [ ]* 42.3 Write component tests for Gateway Assignment
    - Test gateway assignment, removal, credentials entry

- [x] 43. Implement Company Payment recording screen
  - Create screens/payments/CompanyPaymentScreen.tsx
  - Display payment recording form for companies
  - Include fields: company selection (dropdown), amount (number), date (date picker), description (text)
  - Implement Yup schema validation for required fields, positive amount
  - Handle POST to appropriate API endpoint for payment recording
  - Display success message on successful payment recording
  - Display error message with details on failure
  - Display loading indicator while submitting
  - _Requirements: 27.1, 27.2, 27.3, 27.4, 27.5, 27.6, 27.7_

- [ ]* 43.1 Write component tests for Company Payment recording
    - Test payment form, validation, submission, error handling


- [x] 44. Checkpoint - Ensure all feature screens work
  - Verify all collection management screens (Companies, Merchants, Affiliates, Gateways, Banks) work correctly
  - Verify Transactions, Settlements, Ledgers viewing works with proper filtering and pagination
  - Verify Reports generation works
  - Verify Leases management (SuperAdmin only)
  - Verify Bank Accounts and Gateway Assignment work
  - Verify Company Payment recording works
  - Ask the user if questions arise

- [x] 45. Implement offline data caching integration
  - Integrate CacheManager with all data fetching hooks (useAPI, useCollection)
  - Cache API responses after successful fetch
  - Display cached data with visual offline indicator when network unavailable
  - Display offline indicator banner in navigation bar when network is down
  - Prevent data modifications when offline (disable create/update/delete buttons)
  - Refresh cached data automatically when network connectivity is restored
  - Clear cached data on logout
  - _Requirements: 28.1, 28.2, 28.3, 28.4, 28.5, 28.6_

- [ ]* 45.1 Write integration tests for offline caching
    - Test cache behavior, offline mode, network restoration


- [x] 46. Implement error handling and user feedback
  - Ensure all API requests display structured error messages from ApiError
  - Display "Cannot reach the server. Please check your connection." for network errors
  - Display field-specific validation errors below input fields in red
  - Display success messages (toast/snackbar) after successful create, update, delete operations
  - Display retry buttons for transient errors
  - Ensure no technical error details (stack traces, internal codes) are displayed to users
  - Implement error logging for debugging purposes (console.log in development)
  - _Requirements: 29.1, 29.2, 29.3, 29.4, 29.5, 29.6, 29.7_

- [ ]* 46.1 Write tests for error handling
    - Test API error display, network error display, field validation errors, success messages

- [x] 47. Implement loading state indicators
  - Ensure all data fetching displays loading spinner or skeleton
  - Disable submit buttons during form submission with "Saving..." or "Creating..." text
  - Display skeleton screens for list views while data loads
  - Display pull-to-refresh indicator when refreshing
  - Display loading indicator in navigation bar for background operations
  - _Requirements: 30.1, 30.2, 30.3, 30.4, 30.5_

- [ ]* 47.1 Write tests for loading states
    - Test loading indicators, skeleton screens, button states


- [x] 48. Implement accessibility improvements (WCAG AA compliance)
  - Add accessibility labels to all interactive elements (buttons, inputs, list items)
  - Ensure proper heading hierarchy in screens
  - Ensure sufficient color contrast for text and UI elements (4.5:1 for normal text, 3:1 for large text)
  - Ensure touch targets are at least 44x44 points
  - Support screen reader navigation with proper focus management
  - Add semantic labels to form fields and validation errors
  - Test with TalkBack (Android screen reader)
  - _Requirements: 31.1, 31.2, 31.3, 31.4, 31.5_

- [ ]* 48.1 Write accessibility tests
    - Test accessibility labels, contrast ratios, touch target sizes

- [x] 49. Implement mobile-specific UI optimizations
  - Ensure responsive layout for screen sizes 4.7" to 6.7"
  - Support portrait and landscape orientations with proper layout adjustments
  - Implement smooth 60fps scrolling with optimized FlatList rendering
  - Implement keyboard management (dismiss on tap outside, avoid input overlap)
  - Implement haptic feedback for important actions (delete, submit)
  - Optimize images and assets for mobile (compression, appropriate resolutions)
  - Implement swipe gestures for common actions (swipe to delete, swipe to go back)
  - _Requirements: 1.4, 1.5, 35.1, 36.1, 36.2, 36.3, 36.4_


- [ ]* 49.1 Write tests for mobile UI optimizations
    - Test responsive layouts, orientation changes, gesture handling

- [x] 50. Configure Android build and generate APK
  - Configure app.json for Android build settings (package name, version, permissions)
  - Configure app icons and splash screen for Android
  - Set up environment-specific configuration (development, staging, production API URLs)
  - Configure code signing for release builds
  - Build APK using Expo EAS Build or expo build:android
  - Test APK installation on physical Android device (API level 21+)
  - Verify all features work correctly on installed APK
  - Document build process and distribution instructions in README
  - _Requirements: 1.2, 1.3_

- [ ] 51. Final testing and bug fixes
  - Run full test suite (unit tests, component tests, property-based tests)
  - Test all user flows end-to-end for each role
  - Test offline mode thoroughly (cache, network restoration, error handling)
  - Test authentication flows (login, session restore, logout, token expiry)
  - Test all CRUD operations for each collection
  - Test all filters, sorting, pagination, search functionality
  - Test error scenarios (network errors, API errors, validation errors)
  - Test on multiple Android devices with different screen sizes
  - Fix any bugs discovered during testing
  - Verify performance (60fps scrolling, fast app launch, smooth navigation)


- [x] 52. Final checkpoint - Ensure all tests pass and app is ready for deployment
  - Ensure all unit tests pass
  - Ensure all component tests pass
  - Ensure all property-based tests pass (minimum 100 iterations each)
  - Ensure APK builds successfully
  - Ensure app works correctly on physical Android devices
  - Ask the user if questions arise or if ready for deployment

## Notes

- Tasks marked with `*` are optional testing tasks and can be skipped for faster MVP delivery
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at major milestones
- Property-based tests validate pure formatting/validation functions and universal API behaviors
- Unit tests validate component rendering, user interactions, and business logic
- Integration tests validate end-to-end flows and API interactions
- The implementation follows a layered architecture: data layer (services) → business logic layer (hooks) → presentation layer (screens/components)
- Generic reusable components (ListScreen, DetailsScreen, FormScreen) reduce code duplication across feature screens
- Role-based portal routing ensures each user sees only authorized features
- Offline caching provides graceful degradation when network is unavailable


## Task Dependency Graph

```json
{
  "waves": [
    { "id": 0, "tasks": ["1"] },
    { "id": 1, "tasks": ["2.1", "3", "5", "6"] },
    { "id": 2, "tasks": ["2.2", "3.1", "4", "5.1", "7", "8"] },
    { "id": 3, "tasks": ["2.3", "4.1", "7.1", "7.2", "8.1", "8.2", "10", "11", "13"] },
    { "id": 4, "tasks": ["9", "10.1", "11.1", "12", "13.1", "14.1", "14.2", "14.3"] },
    { "id": 5, "tasks": ["12.1", "14.4", "15.1", "15.2", "15.3", "15.4", "16"] },
    { "id": 6, "tasks": ["15.5", "16.1", "17"] },
    { "id": 7, "tasks": ["17.1", "18"] },
    { "id": 8, "tasks": ["18.1", "20", "21", "22", "23", "24"] },
    { "id": 9, "tasks": ["20.1", "21.1", "22.1", "23.1", "24.1", "25"] },
    { "id": 10, "tasks": ["25.1", "26"] },
    { "id": 11, "tasks": ["26.1", "27"] },
    { "id": 12, "tasks": ["27.1", "28"] },
    { "id": 13, "tasks": ["28.1", "29"] },
    { "id": 14, "tasks": ["29.1", "31.1", "31.2", "32.1", "32.2", "33.1", "33.2", "34.1", "34.2", "35.1", "35.2"] },
    { "id": 15, "tasks": ["31.3", "32.3", "33.3", "34.3", "35.3", "36.1", "36.2", "37.1", "37.2", "38.1", "38.2"] },
    { "id": 16, "tasks": ["31.4", "32.4", "33.4", "34.4", "35.4", "36.3", "37.3", "38.3", "39.1", "39.2", "40.1", "40.2", "41.1", "41.2", "42.1"] },
    { "id": 17, "tasks": ["39.3", "40.3", "41.3", "42.2", "43"] },
    { "id": 18, "tasks": ["40.4", "41.4", "42.3", "43.1", "45"] },
    { "id": 19, "tasks": ["45.1", "46", "47"] },
    { "id": 20, "tasks": ["46.1", "47.1", "48"] },
    { "id": 21, "tasks": ["48.1", "49"] },
    { "id": 22, "tasks": ["49.1", "50"] },
    { "id": 23, "tasks": ["51"] }
  ]
}
```
