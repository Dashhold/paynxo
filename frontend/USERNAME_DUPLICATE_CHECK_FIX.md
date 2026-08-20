# Username Duplicate Check Fix

## Issue Description

When creating new Companies, Merchants, or Affiliates, the system was not checking if the User ID already existed in the database. This allowed duplicate usernames across different user types, which could cause login issues and data integrity problems.

## Root Cause

The save validation functions in Companies.jsx, Merchants.jsx, and Affiliates.jsx were only validating:
- Required fields (name, userId, password)
- Field format

But **NOT** checking if the userId was already in use by another entity.

## Solution Implemented

Added username duplicate checking across **all user types** (companies, merchants, affiliates) before saving new or edited records.

### Implementation Details

#### 1. Companies.jsx
```jsx
// Check for duplicate userId across all user types
const userIdLower = form.userId.trim().toLowerCase();
const isDuplicate = [
  ...db.companies.filter(c => editing === 'new' || c.id !== editing.id),
  ...db.merchants,
  ...db.affiliates
].some(entity => entity.userId && entity.userId.toLowerCase() === userIdLower);

if (isDuplicate) {
  alert('This User ID is already taken. Please choose a different User ID.');
  return;
}
```

#### 2. Merchants.jsx
```jsx
// Check for duplicate userId across all user types
const userIdLower = form.userId.trim().toLowerCase();
const isDuplicate = [
  ...db.companies,
  ...db.merchants.filter(m => editing === 'new' || m.id !== editing.id),
  ...db.affiliates
].some(entity => entity.userId && entity.userId.toLowerCase() === userIdLower);

if (isDuplicate) {
  alert('This User ID is already taken. Please choose a different User ID.');
  return;
}
```

#### 3. Affiliates.jsx
```jsx
// Check for duplicate userId across all user types
const userIdLower = form.userId.trim().toLowerCase();
const isDuplicate = [
  ...db.companies,
  ...db.merchants,
  ...db.affiliates.filter(a => editing === 'new' || a.id !== editing.id)
].some(entity => entity.userId && entity.userId.toLowerCase() === userIdLower);

if (isDuplicate) {
  alert('This User ID is already taken. Please choose a different User ID.');
  return;
}
```

## Key Features

### ✅ Case-Insensitive Comparison
- Converts userId to lowercase before comparison
- Prevents "user1" and "User1" from being created as separate accounts

### ✅ Cross-Entity Checking
- Checks across companies, merchants, AND affiliates
- Ensures no duplicate userIds exist in the entire system

### ✅ Edit Mode Support
- When editing an existing entity, excludes the current entity from duplicate check
- Allows editing other fields without triggering false duplicate warning

### ✅ Clear User Feedback
- Displays clear alert message: "This User ID is already taken. Please choose a different User ID."
- User understands the issue and what to do

## Logic Breakdown

### When Creating New Entity (editing === 'new')
```javascript
[
  ...db.companies,           // Check all companies
  ...db.merchants,           // Check all merchants
  ...db.affiliates           // Check all affiliates
].some(entity => entity.userId.toLowerCase() === userIdLower)
```

### When Editing Existing Entity (editing !== 'new')
```javascript
[
  ...db.companies.filter(c => c.id !== editing.id),  // Exclude current company
  ...db.merchants,                                    // Check all merchants
  ...db.affiliates                                    // Check all affiliates
].some(entity => entity.userId.toLowerCase() === userIdLower)
```

This allows the user to edit a company and keep the same userId, but prevents changing to a userId that's already taken by another entity.

## Test Scenarios

### Scenario 1: Create Company with Existing Merchant UserId
**Steps:**
1. Create merchant with userId "merchant1"
2. Try to create company with userId "merchant1"

**Expected Result:**
❌ Alert: "This User ID is already taken. Please choose a different User ID."

### Scenario 2: Create Merchant with Existing Company UserId (Case Insensitive)
**Steps:**
1. Create company with userId "company1"
2. Try to create merchant with userId "COMPANY1"

**Expected Result:**
❌ Alert: "This User ID is already taken. Please choose a different User ID."

### Scenario 3: Create Affiliate with Existing Affiliate UserId
**Steps:**
1. Create affiliate with userId "affiliate1"
2. Try to create another affiliate with userId "affiliate1"

**Expected Result:**
❌ Alert: "This User ID is already taken. Please choose a different User ID."

### Scenario 4: Edit Company Keeping Same UserId
**Steps:**
1. Create company with userId "company1"
2. Edit company, change name but keep userId "company1"
3. Save

**Expected Result:**
✅ Saves successfully (no duplicate warning for own userId)

### Scenario 5: Edit Company Changing to Existing UserId
**Steps:**
1. Create company with userId "company1"
2. Create merchant with userId "merchant1"
3. Edit company, change userId to "merchant1"
4. Try to save

**Expected Result:**
❌ Alert: "This User ID is already taken. Please choose a different User ID."

### Scenario 6: Create with Unique UserId
**Steps:**
1. Try to create company with userId "uniqueuser123"
2. No other entity has this userId

**Expected Result:**
✅ Saves successfully

## Files Modified

1. `frontend/src/pages/Companies.jsx` - Added duplicate check in save()
2. `frontend/src/pages/Merchants.jsx` - Added duplicate check in save()
3. `frontend/src/pages/Affiliates.jsx` - Added duplicate check in save()

## Benefits

### Before Fix
❌ Could create duplicate userIds
❌ Login confusion (which account logs in?)
❌ Data integrity issues
❌ No user feedback on duplicates

### After Fix
✅ Prevents duplicate userIds across all user types
✅ Case-insensitive checking
✅ Clear error message to user
✅ Allows editing without false positives
✅ Maintains data integrity

## Technical Notes

### Why Check All Three Collections?
Companies, Merchants, and Affiliates all have login credentials (userId + password). If userIds are duplicated:
- Login becomes ambiguous - which account should authenticate?
- Security risk - wrong account might get access
- Data confusion - unclear which entity a transaction belongs to

### Why Case-Insensitive?
- Most login systems treat "User1" and "user1" as the same
- Prevents confusion from case variations
- Consistent with industry standards

### Why Filter Current Entity in Edit Mode?
```javascript
db.companies.filter(c => c.id !== editing.id)
```
When editing company with id "comp123" and userId "company1":
- Without filter: Would find itself as duplicate → false positive
- With filter: Excludes itself from check → allows keeping same userId

## Edge Cases Handled

### ✅ Empty String UserId
Handled by existing validation before duplicate check:
```javascript
if (!form.userId.trim()) { 
  alert('Login User ID is required.'); 
  return; 
}
```

### ✅ Null or Undefined UserId in Database
```javascript
entity.userId && entity.userId.toLowerCase()
```
Only checks entities that have a userId field.

### ✅ Whitespace Differences
```javascript
form.userId.trim().toLowerCase()
```
Trims whitespace before comparison.

## Performance

### Impact: Negligible
- Checks done in-memory on cached `db` object
- Typical database sizes: <1000 entities per collection
- Array operations complete in <1ms
- No API calls during validation

### Time Complexity
- O(n) where n = total companies + merchants + affiliates
- For 100 companies + 500 merchants + 50 affiliates = 650 checks
- Modern JavaScript engines handle this instantly

## Future Enhancements (Optional)

### 1. Real-Time Validation
Instead of alert on save, show error message as user types:
```jsx
<Field label="User ID">
  <input 
    value={form.userId} 
    onChange={handleUserIdChange}
  />
  {userIdError && <span className="error">{userIdError}</span>}
</Field>
```

### 2. Server-Side Validation
Add duplicate check in backend API as well:
```go
// In account creation handler
if accountExists(userId) {
  return 409, "User ID already exists"
}
```

### 3. Username Suggestions
If duplicate found, suggest alternatives:
```javascript
if (isDuplicate) {
  alert(`User ID "${form.userId}" is taken. Try: ${form.userId}1, ${form.userId}_2`);
}
```

## Status: ✅ COMPLETE

The username duplicate checking is now implemented and working correctly across all user types.

---

**Issue**: Username duplicates not detected
**Solution**: Added cross-entity duplicate validation
**Status**: Fixed and tested
**Files**: 3 modified
**Risk**: Low (validation only, no breaking changes)
