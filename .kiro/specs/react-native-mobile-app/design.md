# Technical Design Document

## Overview

This document presents the technical design for a React Native Expo mobile application (APK) that replicates the Payment Gateway Operations, Commission & Settlement System web frontend. The mobile app provides role-based access to five user types (SuperAdmin, Admin, Company, Affiliate, Merchant), connecting to the existing Go backend API.

### System Context

The mobile application serves as a native Android client for the existing payment gateway system. It authenticates against the Go backend API, receives JWT tokens with role information, and displays role-appropriate UI portals. The app supports offline data caching, pull-to-refresh, and mobile-optimized interactions while maintaining feature parity with the web frontend.

### Key Design Goals

1. **Role-Based Portal System**: Automatic portal selection based on server-provided user role
2. **Offline-First Architecture**: Cache API responses for offline viewing
3. **Mobile-Optimized UX**: Touch-friendly, responsive design with native mobile patterns
4. **Secure Token Management**: JWT-based authentication with secure storage
5. **Feature Parity**: Complete replication of web frontend capabilities
6. **Performance**: 60fps scrolling, optimized list rendering, efficient caching

### Technology Stack

- **Framework**: React Native with Expo SDK 50+
- **Language**: TypeScript
- **State Management**: React Context API + Custom Hooks
- **Network**: Axios for HTTP client
- **Storage**: expo-secure-store for tokens, AsyncStorage for cache
- **Navigation**: React Navigation 6.x (Stack + Drawer)
- **Forms**: React Hook Form + Yup validation
- **UI Components**: React Native Paper or Native Base
- **Testing**: Jest + React Native Testing Library


## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────┐
│           React Native Mobile Application           │
├─────────────────────────────────────────────────────┤
│                                                       │
│  ┌────────────────────────────────────────────────┐ │
│  │         Presentation Layer (Screens)           │ │
│  │  • Auth Screens  • Portal Screens  • Forms    │ │
│  └────────────────────────────────────────────────┘ │
│                        ↓                             │
│  ┌────────────────────────────────────────────────┐ │
│  │          Business Logic Layer (Hooks)          │ │
│  │  • useAuth  • useAPI  • useCache  • useRole   │ │
│  └────────────────────────────────────────────────┘ │
│                        ↓                             │
│  ┌────────────────────────────────────────────────┐ │
│  │           Data Layer (Services)                │ │
│  │  • API Client  • Cache Manager  • Token Store │ │
│  └────────────────────────────────────────────────┘ │
│                                                       │
└───────────────────────────┬───────────────────────┘
                             │ HTTPS/JWT
                             ↓
                  ┌─────────────────────┐
                  │   Go Backend API    │
                  │  • /api/auth/login  │
                  │  • /api/me          │
                  │  • /api/merchants   │
                  │  • /api/companies   │
                  │  • ... (REST APIs)  │
                  └─────────────────────┘
```

### Layer Responsibilities

**Presentation Layer**
- Screen components for each feature (Login, Dashboard, Merchants, etc.)
- Role-specific portal components (SuperAdminPortal, MerchantPortal, etc.)
- Reusable UI components (forms, lists, cards, modals)
- Navigation configuration
- Error boundaries

**Business Logic Layer**
- Custom hooks for feature logic (useAuth, useMerchants, useTransactions)
- Role-based access control enforcement
- Data transformation and formatting
- Form validation logic
- Cache invalidation rules

**Data Layer**
- API client with interceptors for auth and errors
- Secure token storage (expo-secure-store)
- Cache management (AsyncStorage)
- Network connectivity monitoring
- Request/response logging


### Authentication Flow

```
┌─────────┐
│  Login  │
│ Screen  │
└────┬────┘
     │ user submits credentials
     ↓
┌─────────────────────┐
│  POST /api/auth/    │
│      login          │
└────┬────────────────┘
     │
     ├─→ 200 Success
     │   ├─ Store JWT in SecureStore
     │   ├─ Extract UserPrincipal (role, tenantId, etc.)
     │   ├─ Navigate to role-specific portal
     │   └─ Fetch initial data
     │
     ├─→ 401 Unauthorized
     │   └─ Display "Invalid credentials"
     │
     ├─→ 403 Forbidden
     │   └─ Display server error message
     │
     └─→ Network Error
         └─ Display "Cannot reach server"
```

### Session Restoration Flow

```
┌────────────┐
│ App Launch │
└─────┬──────┘
      │
      ↓
┌──────────────────┐
│ Check SecureStore│
│   for JWT token  │
└─────┬────────────┘
      │
      ├─→ No Token → Show Login Screen
      │
      └─→ Token Found
          │
          ↓
     ┌────────────────┐
     │ GET /api/me    │
     └────┬───────────┘
          │
          ├─→ 200 Success
          │   ├─ Extract UserPrincipal
          │   └─ Navigate to portal
          │
          └─→ 401/Error
              ├─ Clear token
              └─ Show Login Screen
```


### Role-Based Portal Routing

```typescript
// Portal determination logic
const portalMap = {
  SuperAdmin: SuperAdminPortal,  // Dashboard, Companies, Merchants, Affiliates, Gateways, Banks, Transactions, Settlements, Ledgers, Reports, Leases
  Admin: AdminPortal,             // Dashboard, Companies, Merchants, Affiliates, Gateways, Banks, Transactions, Settlements, Ledgers, Reports
  Company: CompanyPortal,         // Dashboard, Merchants (scoped), Transactions, Settlements, Ledger
  Affiliate: AffiliatePortal,     // Dashboard, Merchants (scoped), Transactions, Ledger
  Merchant: MerchantPortal        // Dashboard, Transactions, Bank Accounts, Ledger
};

// Navigation structure per portal
SuperAdmin/Admin Portal:
  ├─ Drawer Navigation
  │  ├─ Dashboard
  │  ├─ Companies (list → details)
  │  ├─ Merchants (list → details)
  │  ├─ Affiliates (list → details)
  │  ├─ Gateways (list → details)
  │  ├─ Banks (list → details)
  │  ├─ Transactions (list → details)
  │  ├─ Settlements (list → details)
  │  ├─ Ledgers (list → details)
  │  ├─ Reports
  │  └─ Leases (SuperAdmin only)
  └─ User Menu (Profile, Logout)

Company Portal:
  ├─ Drawer Navigation
  │  ├─ Dashboard
  │  ├─ Merchants (scoped to company)
  │  ├─ Transactions
  │  ├─ Settlements
  │  └─ Ledger
  └─ User Menu (Profile, Logout)

Affiliate Portal:
  ├─ Drawer Navigation
  │  ├─ Dashboard
  │  ├─ Merchants (scoped to affiliate)
  │  ├─ Transactions
  │  └─ Ledger
  └─ User Menu (Profile, Logout)

Merchant Portal:
  ├─ Drawer Navigation
  │  ├─ Dashboard
  │  ├─ Transactions
  │  ├─ Bank Accounts
  │  └─ Ledger
  └─ User Menu (Profile, Logout)
```


## Components and Interfaces

### Core Services

#### 1. API Client Service

```typescript
// services/apiClient.ts
interface ApiClientConfig {
  baseURL: string;
  timeout: number;
}

interface ApiResponse<T> {
  data: T;
  status: number;
  headers: Record<string, string>;
}

interface ApiError {
  code: string;
  message: string;
  fields?: Record<string, string>;
}

class ApiClient {
  private axiosInstance: AxiosInstance;
  
  constructor(config: ApiClientConfig);
  
  // Interceptors
  attachAuthToken(token: string): void;
  handleUnauthorized(onUnauthorized: () => void): void;
  handleNetworkError(onError: (error: ApiError) => void): void;
  
  // HTTP methods
  get<T>(endpoint: string, params?: object): Promise<ApiResponse<T>>;
  post<T>(endpoint: string, data: object): Promise<ApiResponse<T>>;
  put<T>(endpoint: string, data: object): Promise<ApiResponse<T>>;
  delete<T>(endpoint: string): Promise<ApiResponse<T>>;
}
```

#### 2. Token Storage Service

```typescript
// services/tokenStore.ts
interface TokenStore {
  saveToken(token: string): Promise<void>;
  getToken(): Promise<string | null>;
  clearToken(): Promise<void>;
  hasToken(): Promise<boolean>;
}

// Implementation uses expo-secure-store
class SecureTokenStore implements TokenStore {
  private readonly TOKEN_KEY = 'auth_token';
  
  async saveToken(token: string): Promise<void>;
  async getToken(): Promise<string | null>;
  async clearToken(): Promise<void>;
  async hasToken(): Promise<boolean>;
}
```


#### 3. Cache Manager Service

```typescript
// services/cacheManager.ts
interface CacheEntry<T> {
  data: T;
  timestamp: number;
  key: string;
}

interface CacheConfig {
  ttl: number; // Time-to-live in milliseconds (default: 5 minutes)
}

class CacheManager {
  private readonly storage: AsyncStorage;
  private readonly ttl: number;
  
  constructor(config: CacheConfig);
  
  // Cache operations
  set<T>(key: string, data: T): Promise<void>;
  get<T>(key: string): Promise<T | null>;
  has(key: string): Promise<boolean>;
  isStale(key: string): Promise<boolean>;
  invalidate(key: string): Promise<void>;
  invalidatePattern(pattern: string): Promise<void>; // e.g., 'merchants*'
  clear(): Promise<void>;
}

// Cache key structure
// Format: "{collection}:{id?}:{filter?}"
// Examples:
//   - "merchants:all"
//   - "merchants:123"
//   - "transactions:all:status=completed"
//   - "dashboard:superadmin"
```

#### 4. Network Monitor Service

```typescript
// services/networkMonitor.ts
interface NetworkState {
  isConnected: boolean;
  isInternetReachable: boolean;
  type: 'wifi' | 'cellular' | 'unknown' | 'none';
}

class NetworkMonitor {
  getCurrentState(): Promise<NetworkState>;
  subscribe(listener: (state: NetworkState) => void): () => void;
  isOnline(): Promise<boolean>;
}
```


### Custom Hooks

#### 1. useAuth Hook

```typescript
// hooks/useAuth.ts
interface UserPrincipal {
  accountId: string;
  role: 'SuperAdmin' | 'Admin' | 'Company' | 'Affiliate' | 'Merchant';
  tenantId: string | null;
  ownerType: string | null;
  ownerId: string | null;
  email: string;
  name: string;
}

interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  user: UserPrincipal | null;
  error: string | null;
}

interface UseAuthReturn extends AuthState {
  login(userId: string, password: string): Promise<void>;
  logout(): Promise<void>;
  restoreSession(): Promise<void>;
}

function useAuth(): UseAuthReturn {
  // Manages authentication state
  // Handles token storage via TokenStore
  // Calls API client for login/logout
  // Navigates on auth state changes
}
```

#### 2. useAPI Hook

```typescript
// hooks/useAPI.ts
interface UseAPIOptions {
  cacheKey?: string;
  cacheTTL?: number;
  skipCache?: boolean;
  onSuccess?: (data: any) => void;
  onError?: (error: ApiError) => void;
}

interface UseAPIReturn<T> {
  data: T | null;
  loading: boolean;
  error: ApiError | null;
  refetch(): Promise<void>;
  mutate(newData: T): void; // Update local cache
}

function useAPI<T>(
  endpoint: string,
  options?: UseAPIOptions
): UseAPIReturn<T> {
  // Fetches data from API with caching
  // Handles loading and error states
  // Provides refetch and mutate functions
}
```


#### 3. useCollection Hook

```typescript
// hooks/useCollection.ts
interface UseCollectionOptions {
  filters?: Record<string, any>;
  sort?: { field: string; order: 'asc' | 'desc' };
  search?: string;
  pageSize?: number;
}

interface UseCollectionReturn<T> {
  items: T[];
  loading: boolean;
  error: ApiError | null;
  totalCount: number;
  hasMore: boolean;
  
  // Actions
  refresh(): Promise<void>;
  loadMore(): Promise<void>;
  create(item: Partial<T>): Promise<T>;
  update(id: string, item: Partial<T>): Promise<T>;
  remove(id: string): Promise<void>;
  
  // Filter/search
  setFilters(filters: Record<string, any>): void;
  setSort(field: string, order: 'asc' | 'desc'): void;
  setSearch(query: string): void;
}

function useCollection<T>(
  collectionName: string,
  options?: UseCollectionOptions
): UseCollectionReturn<T> {
  // Manages CRUD operations for a collection
  // Handles pagination, filtering, sorting, searching
  // Integrates with cache manager
  // Invalidates related caches on mutations
}
```

#### 4. useForm Hook (React Hook Form wrapper)

```typescript
// hooks/useForm.ts
import { useForm as useRHF } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';

interface UseFormOptions<T> {
  schema: any; // Yup schema
  defaultValues?: Partial<T>;
  onSubmit: (data: T) => Promise<void>;
}

interface UseFormReturn<T> {
  control: any;
  handleSubmit: (e?: any) => void;
  errors: Record<string, any>;
  isSubmitting: boolean;
  isValid: boolean;
  reset(): void;
  setValue(field: keyof T, value: any): void;
}

function useForm<T>(options: UseFormOptions<T>): UseFormReturn<T> {
  // Wraps react-hook-form with Yup validation
  // Provides simplified interface for forms
  // Handles submission loading states
}
```


### Screen Components

#### 1. Login Screen

```typescript
// screens/LoginScreen.tsx
interface LoginScreenProps {
  navigation: NavigationProp<any>;
}

function LoginScreen({ navigation }: LoginScreenProps) {
  // Form: userId (text), password (secure)
  // Password visibility toggle
  // Submit button with loading state
  // Error message display
  // Uses useAuth hook for login
  // Navigates to portal on success
}
```

#### 2. Portal Wrapper

```typescript
// screens/PortalWrapper.tsx
interface PortalWrapperProps {
  role: UserRole;
  user: UserPrincipal;
}

function PortalWrapper({ role, user }: PortalWrapperProps) {
  // Renders appropriate portal based on role
  // Sets up drawer navigation
  // Includes header with user info and logout
  // Displays offline indicator when disconnected
  // Handles deep link routing
}
```

#### 3. Dashboard Screen

```typescript
// screens/DashboardScreen.tsx
interface DashboardData {
  metrics: Array<{ label: string; value: string; trend?: string }>;
  recentActivity: Array<any>;
  summary: Record<string, any>;
}

function DashboardScreen() {
  // Fetches role-specific dashboard data
  // Displays key metrics in cards
  // Shows recent activity list
  // Pull-to-refresh enabled
  // Uses useAPI hook
}
```


#### 4. List Screen (Generic)

```typescript
// screens/ListScreen.tsx
interface ListScreenProps<T> {
  collectionName: string;
  title: string;
  renderItem: (item: T) => React.ReactNode;
  onItemPress: (item: T) => void;
  filterOptions?: FilterConfig[];
  searchPlaceholder?: string;
  canCreate?: boolean;
  onCreatePress?: () => void;
}

function ListScreen<T>({
  collectionName,
  title,
  renderItem,
  onItemPress,
  filterOptions,
  searchPlaceholder,
  canCreate,
  onCreatePress
}: ListScreenProps<T>) {
  // Uses useCollection hook
  // Displays FlatList with renderItem
  // Search bar at top
  // Filter buttons
  // Pull-to-refresh
  // Infinite scroll pagination
  // Empty state when no items
  // Loading skeleton
}
```

#### 5. Details Screen (Generic)

```typescript
// screens/DetailsScreen.tsx
interface DetailsScreenProps<T> {
  itemId: string;
  collectionName: string;
  renderDetails: (item: T) => React.ReactNode;
  actions?: Array<{
    label: string;
    onPress: (item: T) => void;
    variant?: 'primary' | 'danger';
  }>;
  canEdit?: boolean;
  onEditPress?: (item: T) => void;
}

function DetailsScreen<T>({
  itemId,
  collectionName,
  renderDetails,
  actions,
  canEdit,
  onEditPress
}: DetailsScreenProps<T>) {
  // Fetches single item from API
  // Displays item details using renderDetails
  // Action buttons at bottom
  // Edit button in header if canEdit
  // Loading state
  // Error handling with retry
}
```


#### 6. Form Screen (Generic)

```typescript
// screens/FormScreen.tsx
interface FormField {
  name: string;
  label: string;
  type: 'text' | 'email' | 'number' | 'select' | 'date' | 'switch';
  required?: boolean;
  options?: Array<{ label: string; value: any }>; // for select
  placeholder?: string;
  keyboardType?: 'default' | 'numeric' | 'email-address';
}

interface FormScreenProps<T> {
  title: string;
  fields: FormField[];
  schema: any; // Yup schema
  defaultValues?: Partial<T>;
  onSubmit: (data: T) => Promise<void>;
  submitLabel?: string;
}

function FormScreen<T>({
  title,
  fields,
  schema,
  defaultValues,
  onSubmit,
  submitLabel = 'Submit'
}: FormScreenProps<T>) {
  // Uses useForm hook
  // Renders fields dynamically based on type
  // Shows validation errors inline
  // Submit button with loading state
  // Dismisses keyboard on submit
}
```

### Reusable UI Components

#### 1. ListItem Component

```typescript
// components/ListItem.tsx
interface ListItemProps {
  title: string;
  subtitle?: string;
  rightText?: string;
  status?: 'success' | 'pending' | 'failed' | 'active' | 'inactive';
  onPress: () => void;
  leftIcon?: string;
}

function ListItem({
  title,
  subtitle,
  rightText,
  status,
  onPress,
  leftIcon
}: ListItemProps) {
  // Touchable row with title, subtitle, right text
  // Status badge with color coding
  // Optional left icon
  // Chevron right indicator
}
```


#### 2. StatusBadge Component

```typescript
// components/StatusBadge.tsx
type BadgeStatus = 'success' | 'pending' | 'failed' | 'active' | 'inactive' | 'suspended' | 'expired' | 'revoked';

interface StatusBadgeProps {
  status: BadgeStatus;
  label?: string; // Override default label
}

function StatusBadge({ status, label }: StatusBadgeProps) {
  // Color-coded pill badge
  // Green: success, active
  // Yellow: pending
  // Red: failed, revoked
  // Gray: inactive, expired, suspended
}
```

#### 3. FormInput Component

```typescript
// components/FormInput.tsx
interface FormInputProps {
  control: any; // React Hook Form control
  name: string;
  label: string;
  error?: string;
  type?: 'text' | 'email' | 'password' | 'number';
  placeholder?: string;
  keyboardType?: KeyboardTypeOptions;
  secureTextEntry?: boolean;
  multiline?: boolean;
  disabled?: boolean;
}

function FormInput({
  control,
  name,
  label,
  error,
  type,
  placeholder,
  keyboardType,
  secureTextEntry,
  multiline,
  disabled
}: FormInputProps) {
  // Controlled input via React Hook Form
  // Label above input
  // Error message below input in red
  // Appropriate keyboard type
  // Password visibility toggle if secureTextEntry
}
```


#### 4. EmptyState Component

```typescript
// components/EmptyState.tsx
interface EmptyStateProps {
  icon?: string;
  title: string;
  message: string;
  actionLabel?: string;
  onActionPress?: () => void;
}

function EmptyState({
  icon,
  title,
  message,
  actionLabel,
  onActionPress
}: EmptyStateProps) {
  // Centered content with icon, title, message
  // Optional action button
  // Used when lists are empty or no results found
}
```

#### 5. ErrorDisplay Component

```typescript
// components/ErrorDisplay.tsx
interface ErrorDisplayProps {
  error: ApiError | Error;
  onRetry?: () => void;
  retryLabel?: string;
}

function ErrorDisplay({
  error,
  onRetry,
  retryLabel = 'Retry'
}: ErrorDisplayProps) {
  // Error icon and message
  // Retry button if onRetry provided
  // Parses ApiError vs generic Error
}
```

#### 6. LoadingSkeleton Component

```typescript
// components/LoadingSkeleton.tsx
interface LoadingSkeletonProps {
  type: 'list' | 'details' | 'dashboard';
  count?: number; // For list type
}

function LoadingSkeleton({ type, count = 5 }: LoadingSkeletonProps) {
  // Placeholder UI while data loads
  // Shimmering animation effect
  // Different layouts for list vs details vs dashboard
}
```


## Data Models

### Authentication Models

```typescript
// models/auth.ts
interface LoginRequest {
  userId: string;
  password: string;
}

interface LoginResponse {
  token: string;
  user: UserPrincipal;
}

interface UserPrincipal {
  accountId: string;
  role: 'SuperAdmin' | 'Admin' | 'Company' | 'Affiliate' | 'Merchant';
  tenantId: string | null;
  ownerType: string | null;
  ownerId: string | null;
  email: string;
  name: string;
}
```

### Business Entity Models

```typescript
// models/entities.ts
interface Company {
  id: string;
  name: string;
  email: string;
  phone: string;
  address: string;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}

interface Merchant {
  id: string;
  name: string;
  companyId: string;
  companyName?: string;
  affiliateId?: string;
  affiliateName?: string;
  email: string;
  phone: string;
  status: 'active' | 'inactive';
  commissionRate: number;
  gateways: Gateway[];
  bankAccounts: BankAccount[];
  createdAt: string;
  updatedAt: string;
}

interface Affiliate {
  id: string;
  name: string;
  email: string;
  phone: string;
  commissionRate: number;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}

interface Gateway {
  id: string;
  name: string;
  type: string;
  status: 'active' | 'inactive';
  credentials?: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

interface Bank {
  id: string;
  name: string;
  code: string;
  swiftCode?: string;
  createdAt: string;
  updatedAt: string;
}

interface BankAccount {
  id: string;
  merchantId: string;
  bankId: string;
  bankName?: string;
  accountNumber: string;
  accountHolderName: string;
  ifscCode?: string;
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
}
```


### Transaction Models

```typescript
// models/transaction.ts
interface Transaction {
  id: string;
  merchantId: string;
  merchantName?: string;
  companyId?: string;
  companyName?: string;
  affiliateId?: string;
  affiliateName?: string;
  gatewayId: string;
  gatewayName?: string;
  amount: number;
  currency: string;
  status: 'pending' | 'completed' | 'failed';
  commission: CommissionBreakdown;
  transactionDate: string;
  reference: string;
  createdAt: string;
  updatedAt: string;
}

interface CommissionBreakdown {
  baseAmount: number;
  merchantCommission: number;
  affiliateCommission: number;
  companyCommission: number;
  gatewayFee: number;
  netAmount: number;
}

interface Settlement {
  id: string;
  merchantId: string;
  merchantName?: string;
  companyId?: string;
  companyName?: string;
  amount: number;
  currency: string;
  status: 'pending' | 'completed' | 'failed';
  settlementDate: string;
  bankAccountId?: string;
  reference: string;
  createdAt: string;
  updatedAt: string;
}

interface LedgerEntry {
  id: string;
  entityType: 'merchant' | 'company' | 'affiliate';
  entityId: string;
  entityName?: string;
  transactionId?: string;
  settlementId?: string;
  type: 'debit' | 'credit';
  amount: number;
  balance: number;
  description: string;
  entryDate: string;
  createdAt: string;
}
```


### Administrative Models

```typescript
// models/admin.ts
interface Lease {
  id: string;
  tenantId: string;
  tenantName?: string;
  status: 'active' | 'suspended' | 'expired' | 'revoked';
  startDate: string;
  endDate: string;
  accessLevel: string;
  createdAt: string;
  updatedAt: string;
}

interface Report {
  id: string;
  type: 'transaction' | 'commission' | 'settlement' | 'ledger';
  parameters: ReportParameters;
  generatedAt: string;
  data: any;
}

interface ReportParameters {
  startDate: string;
  endDate: string;
  merchantId?: string;
  companyId?: string;
  affiliateId?: string;
  gatewayId?: string;
  status?: string;
}

interface DashboardMetrics {
  role: string;
  metrics: Array<{
    label: string;
    value: string | number;
    trend?: number; // Percentage change
    icon?: string;
  }>;
  recentActivity: Array<any>;
  summary: Record<string, any>;
}
```

### API Request/Response Models

```typescript
// models/api.ts
interface PaginatedRequest {
  page?: number;
  pageSize?: number;
  sort?: string;
  order?: 'asc' | 'desc';
  search?: string;
  filters?: Record<string, any>;
}

interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

interface ApiError {
  code: string;
  message: string;
  fields?: Record<string, string>; // Field-level validation errors
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Applicability of Property-Based Testing

This React Native mobile application is primarily a UI-focused system with extensive API integration and side effects. Most of the requirements involve:
- UI rendering and navigation (screens, portals, menus)
- External API integration (network calls, response handling)
- Side effects (storage operations, navigation, state updates)
- User interactions (touch events, form submissions)

These are **NOT appropriate for property-based testing**. Instead, they should be tested using:
- **Unit tests** for component rendering and behavior
- **Integration tests** for API client interactions (with mocks)
- **UI component tests** for user interactions
- **Snapshot tests** for UI consistency

However, there is a **limited set of pure utility functions** where property-based testing IS appropriate:
- **Formatting utilities** (currency, numbers, dates, percentages)
- **Validation functions** (email, phone, numeric, positive number validation)
- **Universal interceptor behaviors** (401 handling across all API endpoints)

The following properties target these pure functions and universal behaviors:


### Property 1: Currency Formatting Consistency

*For any* numeric value, when formatted as Indian Rupee currency, the output SHALL include the ₹ symbol, use Indian comma separators (lakhs and crores), and display exactly two decimal places.

**Validates: Requirements 37.1, 37.2, 37.3**

### Property 2: Percentage Formatting

*For any* numeric value representing a percentage, when formatted, the output SHALL include the % symbol and preserve the numeric value.

**Validates: Requirements 37.4**

### Property 3: Email Validation Consistency

*For any* string input to an email validation function, the function SHALL accept all strings matching valid email format (RFC 5322 simplified pattern) and reject all strings that do not match the pattern.

**Validates: Requirements 43.2**

### Property 4: Phone Validation Consistency

*For any* string input to a phone validation function, the function SHALL accept valid phone number formats and reject invalid formats consistently.

**Validates: Requirements 43.3**

### Property 5: Numeric Input Validation

*For any* string input to an amount field validator, the validator SHALL accept only strings that represent valid numeric values (integers or decimals) and reject all non-numeric strings.

**Validates: Requirements 43.4, 61.1**

### Property 6: Positive Number Validation

*For any* numeric input to an amount field validator, the validator SHALL accept only positive numbers (> 0) and reject zero and negative numbers.

**Validates: Requirements 43.5, 61.2**

### Property 7: Decimal Place Limitation

*For any* currency amount input with arbitrary decimal places, the formatting function SHALL truncate or round the value to exactly 2 decimal places.

**Validates: Requirements 61.4**

### Property 8: Required Field Validation

*For any* form field marked as required, the validation function SHALL prevent form submission when the field is empty or contains only whitespace.

**Validates: Requirements 43.1**


### Property 9: 401 Response Triggers Token Clearing

*For any* API endpoint call that returns a 401 Unauthorized status, the API client interceptor SHALL clear the stored session token, clear all cached data, and navigate to the login screen.

**Validates: Requirements 4.1, 4.2, 4.3**

### Property 10: Bearer Token Attachment

*For any* authenticated API request to any endpoint, the API client SHALL attach the session token as a Bearer token in the Authorization header.

**Validates: Requirements 3.2**

### Property 11: Role-Based Portal Routing

*For any* valid user role (SuperAdmin, Admin, Company, Affiliate, Merchant), after successful authentication, the application SHALL navigate to the portal component corresponding to that specific role.

**Validates: Requirements 2.10**

## Property Testing Strategy

These 11 properties represent the pure functional logic and universal behaviors in the mobile application that benefit from property-based testing. They will be implemented using **fast-check** (JavaScript property-based testing library).

**Testing approach:**
- **Properties 1-8** test pure formatting and validation functions in isolation
- **Properties 9-11** test universal behaviors of the API client and routing system using mocks

**Test configuration:**
- Minimum **100 iterations** per property test
- Each property test includes tag: `Feature: react-native-mobile-app, Property {N}: {description}`
- Fast-check generators will produce:
  - Random numbers (integers, decimals, negative, positive, zero)
  - Random strings (valid emails, invalid emails, phone numbers, alphanumeric, special characters)
  - Random API responses with various status codes
  - Random user roles

**Complementary testing:**
The remaining acceptance criteria (UI rendering, navigation flows, API integration, offline behavior, etc.) will be tested using:
- **Jest unit tests** for component logic and utility functions
- **React Native Testing Library** for component rendering and user interactions
- **Mock Service Worker (MSW)** or similar for API integration tests
- **Example-based tests** for specific scenarios (login success, 401 handling, offline mode, etc.)


## Error Handling

### Error Classification

The mobile app handles four categories of errors:

1. **Network Errors**: Cannot reach server, timeout, DNS failure
2. **API Errors**: Structured errors from backend (400, 401, 403, 404, 409, 500)
3. **Validation Errors**: Client-side form validation failures
4. **Application Errors**: Unhandled exceptions, React errors, storage failures

### Error Handling Strategy

#### 1. Network Errors

```typescript
// Handled by API client interceptor
catch (error) {
  if (error.code === 'ECONNABORTED' || error.code === 'ETIMEDOUT') {
    return {
      message: 'Request timed out. Please try again.',
      retryable: true
    };
  }
  if (error.message === 'Network Error') {
    return {
      message: 'Cannot reach the server. Please check your connection.',
      retryable: true
    };
  }
}
```

- Display user-friendly message
- Provide retry button
- Show cached data if available
- Display offline indicator in UI

#### 2. API Errors (Structured)

```typescript
// API Error Response Format
interface ApiError {
  code: string;        // e.g., "VALIDATION_ERROR", "NOT_FOUND"
  message: string;     // User-friendly message
  fields?: Record<string, string>; // Field-level errors for 400
}

// Error handling by status code
- 400: Display field-level errors inline in forms
- 401: Clear session, navigate to login (no user message needed)
- 403: Display "You don't have permission to perform this action"
- 404: Display "The requested resource was not found"
- 409: Display conflict-specific message from server
- 500: Display "Server error. Please try again later"
```

#### 3. Validation Errors

```typescript
// Handled by React Hook Form + Yup
- Display inline below input field
- Red text and icon
- Prevent form submission until resolved
- Clear on input change

Example validation messages:
- "This field is required"
- "Please enter a valid email address"
- "Phone number must be 10 digits"
- "Amount must be a positive number"
- "Amount cannot have more than 2 decimal places"
```


#### 4. Application Errors

```typescript
// Error Boundary Component
class ErrorBoundary extends React.Component {
  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    // Log error with stack trace
    logError(error, errorInfo);
    
    // Display fallback UI
    this.setState({ hasError: true });
  }
  
  render() {
    if (this.state.hasError) {
      return <ErrorFallbackScreen onReset={this.reset} />;
    }
    return this.props.children;
  }
}

// Wrap each portal in ErrorBoundary
<ErrorBoundary>
  <PortalWrapper />
</ErrorBoundary>
```

### Error Display Components

1. **Inline Field Error**: Below form input (validation errors)
2. **Toast/Snackbar**: Brief confirmation or error message (3-5 seconds)
3. **Alert Dialog**: Critical errors requiring acknowledgment
4. **Error Screen**: Full-screen error for fatal errors with retry/reset option
5. **Empty State with Error**: For list screens when data fetch fails

### Retry Strategy

- **Automatic Retry**: None (user must manually retry)
- **Retry Button**: Provided for all transient errors (network, timeout, 500)
- **No Retry for**: 400, 401, 403, 404 (client errors, not transient)
- **Exponential Backoff**: Not implemented (single manual retry only)

### Error Logging

```typescript
// Development: console.log/error
// Production: Send to error tracking service (e.g., Sentry)

interface ErrorLog {
  timestamp: string;
  level: 'error' | 'warning' | 'info';
  category: 'network' | 'api' | 'validation' | 'application';
  message: string;
  stack?: string;
  context?: {
    userId?: string;
    route?: string;
    apiEndpoint?: string;
    statusCode?: number;
  };
}

// Never log sensitive data:
- Passwords
- Tokens
- Credit card numbers
- Personal identifiable information (beyond userId for tracking)
```


## Testing Strategy

### Testing Approach

The mobile application testing strategy combines multiple testing methodologies appropriate to different layers of the application:

#### 1. Property-Based Testing (Limited Scope)

**What to test:** Pure utility functions and universal behaviors
- Currency formatting functions
- Number formatting functions
- Validation functions (email, phone, amount)
- Universal API interceptor behaviors (401 handling, Bearer token attachment)
- Role-based routing logic

**Library:** fast-check (JavaScript property-based testing)

**Configuration:**
- Minimum 100 iterations per property
- Tag format: `Feature: react-native-mobile-app, Property {N}: {description}`
- Generators for: numbers (integers, decimals, edge cases), strings (valid/invalid formats), roles, API responses

**Coverage:** 11 properties (see Correctness Properties section)

#### 2. Unit Testing

**What to test:** Individual functions and components in isolation
- Utility functions not covered by PBT
- React hooks (useAuth, useAPI, useCollection) with mocked dependencies
- Service classes (ApiClient, CacheManager, TokenStore)
- Custom validation logic
- Data transformation functions
- State management logic

**Library:** Jest + React Native Testing Library

**Examples:**
```typescript
describe('useAuth', () => {
  it('should set loading state during login', async () => {
    // Mock API client
    // Call login
    // Assert loading states
  });
  
  it('should store token on successful login', async () => {
    // Mock successful API response
    // Call login
    // Assert token was stored in SecureStore
  });
});
```


#### 3. Component Testing

**What to test:** React components with user interactions
- Screen components (LoginScreen, DashboardScreen, ListScreen, etc.)
- Reusable UI components (ListItem, StatusBadge, FormInput, etc.)
- User interactions (button clicks, form submissions, navigation)
- Rendering with different props
- Conditional rendering based on state/role

**Library:** React Native Testing Library + Jest

**Examples:**
```typescript
describe('LoginScreen', () => {
  it('should display error message on invalid credentials', async () => {
    const { getByText, getByPlaceholder } = render(<LoginScreen />);
    
    fireEvent.changeText(getByPlaceholder('User ID'), 'testuser');
    fireEvent.changeText(getByPlaceholder('Password'), 'wrongpass');
    fireEvent.press(getByText('Login'));
    
    await waitFor(() => {
      expect(getByText(/Invalid credentials/)).toBeTruthy();
    });
  });
});

describe('StatusBadge', () => {
  it('should render green badge for success status', () => {
    const { getByText } = render(<StatusBadge status="success" />);
    const badge = getByText('Success');
    
    expect(badge.props.style.backgroundColor).toBe('#4CAF50');
  });
});
```

#### 4. Integration Testing

**What to test:** API client integration with backend (using mocks)
- API client methods (get, post, put, delete)
- Request/response interceptors
- Error handling for various status codes
- Token attachment to requests
- Cache integration with API responses

**Library:** Jest + Mock Service Worker (MSW) or manual mocks

**Examples:**
```typescript
describe('ApiClient', () => {
  it('should attach Bearer token to authenticated requests', async () => {
    const client = new ApiClient(config);
    client.attachAuthToken('test-token');
    
    const request = await client.get('/api/merchants');
    
    expect(request.headers.Authorization).toBe('Bearer test-token');
  });
  
  it('should handle 401 by calling onUnauthorized callback', async () => {
    const onUnauthorized = jest.fn();
    const client = new ApiClient(config);
    client.handleUnauthorized(onUnauthorized);
    
    // Mock 401 response
    mockApi.get('/api/merchants').reply(401);
    
    await client.get('/api/merchants');
    
    expect(onUnauthorized).toHaveBeenCalled();
  });
});
```


#### 5. End-to-End Testing (Optional)

**What to test:** Full user workflows through the app
- Login flow (enter credentials → navigate to portal)
- CRUD operations (create merchant → view merchant → update merchant)
- Offline behavior (go offline → view cached data → go online → sync)
- Navigation flows (navigate between screens → back button behavior)

**Library:** Detox (React Native E2E testing)

**Note:** E2E tests are optional and should be limited to critical user journeys due to their brittleness and maintenance cost.

### Test Coverage Goals

- **Property-based tests**: 11 properties covering pure functions and universal behaviors
- **Unit tests**: 70%+ code coverage of utility functions, hooks, and services
- **Component tests**: All screen components and reusable UI components
- **Integration tests**: All API client methods and error scenarios

### Test Execution

```json
// package.json scripts
{
  "test": "jest",
  "test:watch": "jest --watch",
  "test:coverage": "jest --coverage",
  "test:pbt": "jest --testPathPattern=property",
  "test:e2e": "detox test"
}
```

### Continuous Integration

- Run all tests (unit + component + integration + PBT) on every pull request
- Fail build if any test fails
- Fail build if code coverage drops below 70%
- Generate coverage report and property test execution logs

### Testing Philosophy

**Property-based testing** is used for pure functions where input variation reveals edge cases (formatting, validation, universal behaviors). This provides high-confidence correctness across a wide input space.

**Unit and component testing** are used for UI components, hooks, and integration points. These use example-based tests with specific scenarios.

**Integration testing** verifies the API client and external dependencies work correctly with mocked responses.

Together, these strategies provide comprehensive coverage while using the right tool for each layer of the application.


## Security Considerations

### 1. Token Storage

- **Mechanism**: expo-secure-store (uses Keychain on iOS, EncryptedSharedPreferences on Android)
- **Token Type**: JWT Bearer token
- **Storage Key**: `auth_token` (encrypted at rest)
- **Access**: Only accessible by the app, not shared with other apps
- **Clearing**: Token cleared on logout, 401 responses, and invalid session

### 2. API Communication

- **Protocol**: HTTPS only (enforced in production builds)
- **Certificate Validation**: Enabled (no insecure connections allowed)
- **Certificate Pinning**: Implemented in production builds to prevent MITM attacks
- **Timeout**: 30 seconds per request

### 3. Data Security

**In Transit:**
- All API communication over HTTPS/TLS 1.2+
- Bearer token in Authorization header (not in URL or body)

**At Rest:**
- JWT token: encrypted via SecureStore
- Cached data: encrypted using AsyncStorage with encryption wrapper
- No sensitive data in plain text logs
- No sensitive data in crash reports

**In Memory:**
- Clear sensitive data when app backgrounds (iOS memory dump protection)
- Obfuscate sensitive UI when app switcher is active

### 4. Input Validation

- Client-side validation for all form inputs (defense in depth)
- Server-side validation is authoritative
- Sanitize user inputs before display to prevent injection
- Use parameterized API requests (not string concatenation)

### 5. Logging and Debugging

**Development:**
- Full request/response logging
- Console logs for debugging
- Stack traces displayed

**Production:**
- No sensitive data in logs (passwords, tokens, PII)
- Obfuscated stack traces sent to error tracking
- API errors logged without request payloads containing sensitive data

### 6. App Permissions

**Required Permissions:**
- Internet access (INTERNET)
- Network state (ACCESS_NETWORK_STATE)

**Optional Permissions:**
- Biometric authentication (USE_FINGERPRINT, USE_BIOMETRIC)
- Notifications (POST_NOTIFICATIONS)

**Not Requested:**
- Location, Camera, Microphone, Contacts, Storage (not needed for app functionality)


## Performance Optimization

### 1. List Rendering

**Problem**: Large lists (1000+ transactions) cause lag and memory issues

**Solution**:
- Use `FlatList` with `windowSize` optimization
- Implement `getItemLayout` for fixed-height items (faster scrolling)
- Use `removeClippedSubviews` for Android performance
- Limit initial render to 20 items, load more on scroll
- Memoize list items with `React.memo`

```typescript
<FlatList
  data={items}
  renderItem={renderItem}
  keyExtractor={keyExtractor}
  windowSize={10}
  maxToRenderPerBatch={10}
  initialNumToRender={20}
  removeClippedSubviews={true}
  onEndReached={loadMore}
  onEndReachedThreshold={0.5}
/>
```

### 2. Search and Filter Debouncing

**Problem**: Typing in search triggers API call on every keystroke

**Solution**:
- Debounce search input with 300ms delay
- Cancel pending requests when new search is initiated
- Use local filtering when data is already cached

```typescript
const debouncedSearch = useMemo(
  () => debounce((query) => performSearch(query), 300),
  []
);
```

### 3. Image Optimization

**Problem**: Large images slow down list scrolling

**Solution**:
- Use `react-native-fast-image` for caching and priority loading
- Lazy load images in lists
- Use thumbnail URLs for list views, full res for details
- Implement progressive image loading

### 4. Cache Strategy

**Problem**: Redundant API calls waste bandwidth and slow down app

**Solution**:
- Cache GET responses with 5-minute TTL
- Invalidate cache on mutations (create, update, delete)
- Use stale-while-revalidate pattern (show cache, fetch in background)
- Prefetch data for likely next screens

```typescript
// Cache key structure
"merchants:all" -> cached for 5 minutes
"merchants:123" -> cached for 5 minutes
"transactions:all:page=1" -> cached for 5 minutes

// Invalidation rules
- On create merchant: invalidate "merchants:all"
- On update merchant: invalidate "merchants:all" and "merchants:{id}"
- On delete merchant: invalidate "merchants:all" and "merchants:{id}"
```


### 5. Component Optimization

**Problem**: Unnecessary re-renders slow down UI

**Solution**:
- Use `React.memo` for pure components
- Use `useMemo` for expensive computations
- Use `useCallback` for callback props to prevent re-renders
- Split large components into smaller, focused components

```typescript
const MerchantListItem = React.memo(({ merchant, onPress }) => {
  return (
    <TouchableOpacity onPress={() => onPress(merchant.id)}>
      {/* Item content */}
    </TouchableOpacity>
  );
});
```

### 6. Bundle Size Optimization

**Problem**: Large APK size

**Solution**:
- Enable ProGuard/R8 for code minification and obfuscation
- Remove unused dependencies
- Use dynamic imports for large libraries
- Enable Hermes JavaScript engine (faster startup, smaller bundle)
- Split bundles by portal (code splitting)

### 7. Startup Time Optimization

**Problem**: Slow app launch

**Solution**:
- Lazy load non-critical screens
- Defer heavy computations until after initial render
- Use splash screen while app initializes
- Preload authentication state from SecureStore in parallel with UI render

```typescript
// App initialization sequence
1. Show splash screen
2. Parallel: Restore token from SecureStore + Render app shell
3. If token exists, validate with /api/me
4. Navigate to appropriate portal or login
5. Fetch initial data after navigation
```

### 8. Memory Management

**Problem**: Memory leaks cause crashes on low-end devices

**Solution**:
- Unsubscribe from listeners in useEffect cleanup
- Cancel pending API requests when component unmounts
- Clear large data structures when navigating away
- Limit cache size (max 100 entries per collection)
- Use pagination instead of loading all data at once


## Offline Support Strategy

### Cache-First Architecture

The mobile app implements a **cache-first** strategy for viewing data while offline, but requires online connectivity for mutations.

### What Works Offline

1. **View Cached Data**
   - Recently loaded lists (merchants, transactions, companies, etc.)
   - Previously viewed detail screens
   - Dashboard metrics (if cached)
   - User profile information

2. **Navigation**
   - Navigate between cached screens
   - Open detail screens for cached items
   - Use drawer navigation

3. **UI Interactions**
   - Search and filter cached data locally
   - Sort cached lists
   - View cached reports

### What Requires Online

1. **Mutations**: Create, update, delete operations
2. **Fresh Data**: Pull-to-refresh, load more pages
3. **Authentication**: Login, logout (but session restoration works offline)
4. **Reports**: Generate new reports
5. **Uncached Screens**: Viewing items not in cache

### Offline Indicator

```typescript
// Display persistent banner when offline
<OfflineBanner visible={!isConnected}>
  You are offline. Viewing cached data.
</OfflineBanner>

// Disable mutation buttons
<Button
  disabled={!isConnected}
  onPress={createMerchant}
>
  Create Merchant
</Button>

// Show tooltip on disabled actions
"This action requires internet connection"
```

### Cache Invalidation

- **On Logout**: Clear all cached data
- **On 401 Response**: Clear all cached data
- **On Mutation Success**: Invalidate affected collections
- **On TTL Expiry**: Mark cache as stale, fetch fresh data when online
- **Manual Refresh**: Pull-to-refresh bypasses cache

### Sync Strategy

**No Background Sync**: The app does not queue mutations for later sync. All mutations must be performed while online and complete immediately.

**Conflict Resolution**: Not applicable (no offline mutations)


## Build and Deployment

### Development Environment Setup

```bash
# Prerequisites
- Node.js 18+ and npm/yarn
- Expo CLI (npm install -g expo-cli)
- Android Studio (for Android emulator)
- Java JDK 11+

# Project setup
npx create-expo-app payment-gateway-mobile --template
cd payment-gateway-mobile
npm install

# Install dependencies
npm install @react-navigation/native @react-navigation/stack @react-navigation/drawer
npm install react-hook-form yup @hookform/resolvers
npm install axios
npm install @react-native-async-storage/async-storage
npm install expo-secure-store
npm install @react-native-community/netinfo
npm install fast-check --save-dev
npm install @testing-library/react-native jest --save-dev

# Run development server
npx expo start
```

### Build Configuration

**app.json** / **app.config.js**:
```json
{
  "expo": {
    "name": "Payment Gateway Ops",
    "slug": "payment-gateway-mobile",
    "version": "1.0.0",
    "orientation": "portrait",
    "icon": "./assets/icon.png",
    "userInterfaceStyle": "automatic",
    "splash": {
      "image": "./assets/splash.png",
      "resizeMode": "contain",
      "backgroundColor": "#ffffff"
    },
    "android": {
      "package": "com.dashhold.paymentgateway",
      "versionCode": 1,
      "adaptiveIcon": {
        "foregroundImage": "./assets/adaptive-icon.png",
        "backgroundColor": "#FFFFFF"
      },
      "permissions": [
        "INTERNET",
        "ACCESS_NETWORK_STATE",
        "USE_FINGERPRINT",
        "USE_BIOMETRIC"
      ]
    },
    "extra": {
      "apiBaseUrl": process.env.API_BASE_URL || "https://api.paymentgateway.com"
    },
    "plugins": [
      "expo-secure-store"
    ]
  }
}
```


### APK Build Process

#### Option 1: EAS Build (Recommended)

```bash
# Install EAS CLI
npm install -g eas-cli

# Login to Expo account
eas login

# Configure EAS
eas build:configure

# Build APK for Android
eas build --platform android --profile production

# Output: APK file downloadable from Expo dashboard
```

**eas.json**:
```json
{
  "build": {
    "development": {
      "android": {
        "buildType": "apk",
        "gradleCommand": ":app:assembleDebug"
      }
    },
    "production": {
      "android": {
        "buildType": "apk",
        "gradleCommand": ":app:assembleRelease"
      }
    }
  }
}
```

#### Option 2: Local Build

```bash
# Generate Android project
npx expo prebuild --platform android

# Build APK locally
cd android
./gradlew assembleRelease

# Output: android/app/build/outputs/apk/release/app-release.apk
```

### Code Signing

**Development:**
- Debug keystore (auto-generated)
- No manual signing required

**Production:**
- Generate release keystore:
  ```bash
  keytool -genkeypair -v -keystore payment-gateway-release.keystore \
    -alias payment-gateway -keyalg RSA -keysize 2048 -validity 10000
  ```
- Store keystore credentials in `android/gradle.properties` (not committed)
- Configure signing in `android/app/build.gradle`


### Environment Configuration

**Development:**
- API Base URL: `http://10.0.2.2:8080` (Android emulator) or `http://localhost:8080`
- Logging: Full request/response logs
- Error display: Stack traces visible
- Certificate pinning: Disabled

**Staging:**
- API Base URL: `https://staging-api.paymentgateway.com`
- Logging: Error logs only
- Error display: User-friendly messages
- Certificate pinning: Enabled

**Production:**
- API Base URL: `https://api.paymentgateway.com`
- Logging: Error logs only (sent to Sentry)
- Error display: User-friendly messages only
- Certificate pinning: Enabled
- Code obfuscation: ProGuard/R8 enabled

### Distribution

**Internal Testing:**
- Share APK file directly via email, cloud storage, or internal portal
- Install on devices with "Install from Unknown Sources" enabled

**Closed Beta:**
- Distribute via Firebase App Distribution
- Invite testers by email
- Collect crash reports and feedback

**Production:**
- Publish to Google Play Store (internal track → closed testing → production)
- Or distribute APK directly to users (enterprise distribution)

### Version Management

**Version Scheme**: `MAJOR.MINOR.PATCH` (Semantic Versioning)
- MAJOR: Breaking changes to user workflow or data model
- MINOR: New features, backwards compatible
- PATCH: Bug fixes, no new features

**Version Code**: Auto-incremented integer (Android requirement)

**Example:**
- Version Name: `1.2.3`
- Version Code: `10203` (calculated as MAJOR * 10000 + MINOR * 100 + PATCH)


## Accessibility

### WCAG AA Compliance

The mobile app targets **WCAG 2.1 Level AA** compliance for accessibility.

### 1. Text and Color

- **Minimum font size**: 14pt for body text, 12pt for secondary text
- **Color contrast**: 4.5:1 for normal text, 3:1 for large text (18pt+)
- **No color-only indicators**: Status badges use color + text label
- **High contrast mode**: Support system-level high contrast settings

### 2. Touch Targets

- **Minimum size**: 44x44 points (iOS), 48x48dp (Android)
- **Spacing**: 8pt minimum between touch targets
- **Active area**: Full button/list item is tappable, not just text

### 3. Screen Reader Support

**Accessibility Labels:**
```typescript
// All interactive elements have labels
<TouchableOpacity
  accessible={true}
  accessibilityLabel="Create new merchant"
  accessibilityRole="button"
  accessibilityHint="Opens the merchant creation form"
>
  <Text>Create Merchant</Text>
</TouchableOpacity>

// Images have alt text
<Image
  source={logo}
  accessible={true}
  accessibilityLabel="Payment Gateway Ops logo"
/>

// Form inputs have labels
<TextInput
  accessible={true}
  accessibilityLabel="User ID"
  accessibilityHint="Enter your user ID to log in"
/>
```

**Navigation Announcements:**
- Announce screen titles when navigating
- Announce loading states ("Loading merchants")
- Announce error messages
- Announce success confirmations

### 4. Keyboard Navigation

- All interactive elements focusable in logical order
- Focus indicators visible for selected element
- Skip to content option for screen reader users

### 5. Dynamic Text Sizing

- Support iOS Dynamic Type and Android Font Scaling
- Test at 200% text size
- Avoid fixed heights that break with large text
- Use scalable units (sp on Android, points on iOS)


### 6. Motion and Animation

- **Respect reduced motion preference**: Disable animations if user enables "Reduce Motion"
- **No auto-playing animations**: User-triggered only
- **Timeout warnings**: For timed actions (e.g., session timeout), provide warning and option to extend

### 7. Form Accessibility

- **Labels associated with inputs**: Use accessible={true} and accessibilityLabel
- **Error announcements**: Screen reader announces validation errors
- **Required field indicators**: Marked visually and programmatically
- **Input types**: Appropriate keyboard type (numeric, email, etc.)

### Testing for Accessibility

**Tools:**
- React Native Accessibility Inspector (included in React Native)
- Android TalkBack (screen reader)
- iOS VoiceOver (screen reader)
- Manual testing with screen readers

**Checklist:**
- [ ] All interactive elements have accessibility labels
- [ ] All images have alt text or are marked decorative
- [ ] Touch targets meet minimum size requirements
- [ ] Color contrast meets WCAG AA standards
- [ ] App works with TalkBack/VoiceOver enabled
- [ ] App works with 200% text scaling
- [ ] Animations respect reduced motion preference

**Note:** Full WCAG compliance requires manual testing with assistive technologies and expert accessibility review. These design guidelines provide a foundation, but validation is essential.


## Project Structure

```
payment-gateway-mobile/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── FormInput.tsx
│   │   ├── StatusBadge.tsx
│   │   ├── ListItem.tsx
│   │   ├── EmptyState.tsx
│   │   ├── ErrorDisplay.tsx
│   │   ├── LoadingSkeleton.tsx
│   │   └── OfflineBanner.tsx
│   │
│   ├── screens/             # Screen components
│   │   ├── auth/
│   │   │   ├── LoginScreen.tsx
│   │   │   └── BiometricSetupScreen.tsx
│   │   ├── dashboard/
│   │   │   └── DashboardScreen.tsx
│   │   ├── merchants/
│   │   │   ├── MerchantListScreen.tsx
│   │   │   ├── MerchantDetailsScreen.tsx
│   │   │   └── MerchantFormScreen.tsx
│   │   ├── transactions/
│   │   │   ├── TransactionListScreen.tsx
│   │   │   └── TransactionDetailsScreen.tsx
│   │   ├── companies/
│   │   ├── affiliates/
│   │   ├── gateways/
│   │   ├── banks/
│   │   ├── settlements/
│   │   ├── ledgers/
│   │   ├── reports/
│   │   └── leases/
│   │
│   ├── navigation/          # Navigation configuration
│   │   ├── RootNavigator.tsx
│   │   ├── AuthNavigator.tsx
│   │   ├── SuperAdminPortal.tsx
│   │   ├── AdminPortal.tsx
│   │   ├── CompanyPortal.tsx
│   │   ├── AffiliatePortal.tsx
│   │   └── MerchantPortal.tsx
│   │
│   ├── hooks/               # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useAPI.ts
│   │   ├── useCollection.ts
│   │   ├── useForm.ts
│   │   ├── useNetworkStatus.ts
│   │   └── useCache.ts
│   │
│   ├── services/            # Business logic and external services
│   │   ├── apiClient.ts
│   │   ├── tokenStore.ts
│   │   ├── cacheManager.ts
│   │   ├── networkMonitor.ts
│   │   └── logger.ts
│   │
│   ├── models/              # TypeScript type definitions
│   │   ├── auth.ts
│   │   ├── entities.ts
│   │   ├── transaction.ts
│   │   ├── admin.ts
│   │   └── api.ts
│   │
│   ├── utils/               # Utility functions
│   │   ├── formatters.ts    # Currency, date, number formatting
│   │   ├── validators.ts    # Email, phone, amount validation
│   │   ├── constants.ts     # App constants
│   │   └── helpers.ts       # General helper functions
│   │
│   ├── context/             # React Context providers
│   │   ├── AuthContext.tsx
│   │   └── ThemeContext.tsx
│   │
│   ├── config/              # Configuration
│   │   ├── api.config.ts
│   │   ├── theme.config.ts
│   │   └── env.ts
│   │
│   └── __tests__/           # Tests
│       ├── property/        # Property-based tests
│       │   ├── formatters.property.test.ts
│       │   └── validators.property.test.ts
│       ├── unit/            # Unit tests
│       ├── integration/     # Integration tests
│       └── components/      # Component tests
│
├── assets/                  # Images, fonts, icons
│   ├── icon.png
│   ├── splash.png
│   └── adaptive-icon.png
│
├── app.json                 # Expo configuration
├── babel.config.js
├── tsconfig.json
├── package.json
├── jest.config.js
└── README.md
```


## Implementation Notes

### Development Workflow

1. **Phase 1: Foundation** (Week 1-2)
   - Set up Expo project with TypeScript
   - Configure navigation (React Navigation)
   - Implement API client service
   - Implement token storage service
   - Implement cache manager service
   - Create base authentication flow (login screen, session restoration)

2. **Phase 2: Core Infrastructure** (Week 2-3)
   - Implement custom hooks (useAuth, useAPI, useCollection)
   - Create reusable UI components (FormInput, StatusBadge, ListItem, etc.)
   - Implement error handling system
   - Set up offline support and network monitoring
   - Create portal wrapper and navigation structure

3. **Phase 3: Role-Based Portals** (Week 3-5)
   - Implement SuperAdmin portal with all screens
   - Implement Admin portal (similar to SuperAdmin, without Leases)
   - Implement Company portal (scoped views)
   - Implement Affiliate portal (scoped views)
   - Implement Merchant portal (minimal views)

4. **Phase 4: CRUD Operations** (Week 5-7)
   - Implement list screens for all collections
   - Implement detail screens for all entities
   - Implement form screens for create/update
   - Implement delete functionality with confirmations
   - Implement search, filter, and sort on list screens

5. **Phase 5: Advanced Features** (Week 7-8)
   - Implement dashboard with metrics
   - Implement reports generation and export
   - Implement leases management (SuperAdmin only)
   - Implement commission breakdown display
   - Implement biometric authentication
   - Implement deep linking

6. **Phase 6: Polish and Testing** (Week 8-10)
   - Write property-based tests for pure functions
   - Write unit tests for hooks and services
   - Write component tests for screens
   - Write integration tests for API client
   - Accessibility testing and fixes
   - Performance optimization
   - Security audit

7. **Phase 7: Build and Deployment** (Week 10-11)
   - Configure production environment
   - Generate release keystore
   - Build production APK
   - Internal testing and bug fixes
   - Documentation for deployment
   - Handoff to stakeholders


### Key Implementation Decisions

#### 1. Why Expo?

**Pros:**
- Simplified setup and configuration
- Built-in access to native APIs (SecureStore, NetInfo, etc.)
- Easy APK builds with EAS Build
- Over-the-air updates
- Good developer experience

**Cons:**
- Larger bundle size than bare React Native
- Limited customization of native modules
- Dependency on Expo SDK versions

**Decision**: Use Expo Managed Workflow for faster development and simpler build process. The app doesn't require custom native modules, so Expo's limitations are acceptable.

#### 2. Why React Navigation?

- De facto standard for React Native navigation
- Supports drawer, stack, and tab navigation
- Good TypeScript support
- Active maintenance and community

#### 3. Why React Hook Form + Yup?

- Declarative form validation
- Minimal re-renders (performance)
- TypeScript support
- Integrates well with Yup schema validation
- Simpler than Formik for React Native

#### 4. Why Axios over Fetch?

- Interceptors for request/response transformation
- Automatic request cancellation
- Timeout configuration
- Better error handling
- TypeScript support

#### 5. Why AsyncStorage for Cache?

- Simple key-value storage for non-sensitive data
- Async operations don't block UI
- Sufficient for cache use case
- Can be encrypted with wrapper if needed

#### 6. Why fast-check for PBT?

- Mature JavaScript property-based testing library
- Good TypeScript support
- Rich set of built-in generators
- Integrates seamlessly with Jest
- Active maintenance


### Potential Risks and Mitigations

#### Risk 1: Backend API Changes

**Risk**: Backend API evolves and breaks mobile app compatibility

**Mitigation**:
- Version API endpoints (/api/v1/, /api/v2/)
- Implement graceful degradation for missing fields
- Backend maintains backwards compatibility for mobile app
- Coordinate mobile releases with backend changes

#### Risk 2: Performance on Low-End Devices

**Risk**: App is slow or crashes on budget Android devices

**Mitigation**:
- Test on low-end devices (2GB RAM, older processors)
- Implement pagination and lazy loading
- Use FlatList optimizations
- Monitor memory usage
- Profile performance with React DevTools

#### Risk 3: Network Reliability

**Risk**: Users have poor network, leading to timeouts and errors

**Mitigation**:
- Implement robust offline caching
- Show cached data immediately
- Provide clear offline indicators
- Implement retry logic for transient failures
- Test on throttled/unstable network connections

#### Risk 4: Token Expiry While Using App

**Risk**: JWT expires while user is actively using the app

**Mitigation**:
- Handle 401 responses gracefully
- Implement token refresh if backend supports it
- Clear session and redirect to login
- Show user-friendly message ("Session expired, please log in again")

#### Risk 5: Breaking Changes in Expo SDK

**Risk**: Upgrading Expo SDK breaks existing functionality

**Mitigation**:
- Lock Expo SDK version in package.json
- Test thoroughly before upgrading SDK
- Review Expo release notes for breaking changes
- Maintain version compatibility matrix


## Future Enhancements

### Phase 2 Features (Post-MVP)

1. **Push Notifications**
   - Real-time alerts for transaction status changes
   - Settlement completion notifications
   - Critical system alerts

2. **Biometric Quick Actions**
   - Fingerprint approval for sensitive operations
   - Face recognition for transaction authorization

3. **Advanced Reporting**
   - Interactive charts and graphs
   - Custom date range selection
   - Multi-format export (PDF, CSV, Excel)

4. **Bulk Operations**
   - Multi-select for batch updates
   - Bulk status changes
   - Batch export

5. **Search Improvements**
   - Full-text search across all collections
   - Search history and suggestions
   - Saved search filters

6. **iOS Support**
   - Build IPA for iOS devices
   - iOS-specific UI adjustments
   - TestFlight distribution

7. **Tablet Optimization**
   - Multi-column layouts for larger screens
   - Side-by-side master-detail views
   - Landscape-optimized UI

8. **Offline Mutations**
   - Queue mutations while offline
   - Sync when connectivity restored
   - Conflict resolution strategy

9. **Theme Customization**
   - User-selectable color themes
   - Dark mode toggle (independent of system)
   - Font size preferences

10. **Advanced Analytics**
    - User behavior tracking
    - Performance monitoring (Sentry, Firebase Performance)
    - Crash reporting with context

---

**End of Design Document**

