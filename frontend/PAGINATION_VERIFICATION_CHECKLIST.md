# Pagination Verification Checklist

Use this checklist to verify pagination is working correctly on all pages.

---

## Pre-Testing Setup

- [ ] Ensure you have test data with >10 items on each page type
- [ ] Clear browser cache
- [ ] Test in latest Chrome/Firefox/Safari
- [ ] Have browser console open to catch any errors

---

## Component-Level Testing

### Pagination Component (`Pagination.jsx`)
- [ ] Component renders correctly
- [ ] Shows "Showing X-Y of Z items" correctly
- [ ] Page buttons render (Previous, Next, Page numbers)
- [ ] Items per page dropdown works (10/25/50/100)
- [ ] CSS styling applies correctly (dark theme)
- [ ] Component doesn't show when ≤10 items

---

## Main List Pages (SuperAdmin/Admin)

### 1. Companies Page
- [ ] Pagination appears when >10 companies
- [ ] Navigate to page 2, verify companies 26-50 appear
- [ ] Search for company name, verify reset to page 1
- [ ] Change items per page to 50, verify reset to page 1
- [ ] Verify company count updates with search
- [ ] Edit/Delete buttons work on any page
- [ ] Smooth scroll to top on page change

### 2. Merchants Page
- [ ] Pagination appears when >10 merchants
- [ ] Navigate through pages
- [ ] Search merchants, verify reset to page 1
- [ ] Filter by company (if applicable)
- [ ] Change items per page
- [ ] Actions work on any page

### 3. Affiliates Page
- [ ] Pagination appears when >10 affiliates
- [ ] Navigate through pages
- [ ] Search affiliates
- [ ] Actions work correctly

### 4. Payment Gateways Page
- [ ] Pagination appears when >10 gateways
- [ ] Navigate through pages
- [ ] Search gateways
- [ ] Status filter works with pagination

### 5. Transactions Page
- [ ] Pagination appears when >10 transactions
- [ ] Navigate through pages
- [ ] Company filter works, resets to page 1
- [ ] Transactions show newest first
- [ ] Edit/Delete work on any page

### 6. Settlements Page
- [ ] Pagination appears when >10 settlements
- [ ] Navigate through pages
- [ ] Settlements show newest first
- [ ] Date filter integration

### 7. Banks Page
- [ ] Pagination appears when >10 banks total
- [ ] Banks from all merchants appear
- [ ] Merchant name shows correctly
- [ ] View/Edit/Delete work on any page

### 8. Leases Page (SuperAdmin Only)
- [ ] Pagination appears when >10 leases
- [ ] Navigate through pages
- [ ] View/Extend/Suspend/Revoke actions work
- [ ] Status badges display correctly

---

## Reports Page (All Tabs)

### Transaction Report Tab
- [ ] Pagination with >10 transactions
- [ ] Date filter resets to page 1
- [ ] Export CSV exports ALL data (not just current page)
- [ ] Export PDF exports ALL data

### Settlement Report Tab
- [ ] Pagination works
- [ ] Date filter integration
- [ ] Export functions work correctly

### Company Report Tab
- [ ] Pagination works
- [ ] All company data displays
- [ ] Export functions work

### Merchant Report Tab
- [ ] Pagination works
- [ ] Merchant data accurate
- [ ] Export functions work

### Affiliate Report Tab
- [ ] Pagination works
- [ ] Affiliate data accurate
- [ ] Export functions work

### Gateway Report Tab
- [ ] Pagination works
- [ ] Gateway data accurate
- [ ] Export functions work

---

## Ledgers Page (All Tabs)

### Company Ledger Tab
- [ ] Pagination appears when >10 entries
- [ ] Page 1 running balance starts at 0
- [ ] Navigate to page 2
- [ ] Page 2 running balance continues from page 1
- [ ] Last page final balance matches total outstanding
- [ ] Stats at top (Total Receivable, Received, Outstanding) remain accurate

### Affiliate Ledger Tab
- [ ] Pagination works with >10 entries
- [ ] Running balance calculates correctly across pages
- [ ] Commission tracking accurate

### Merchant Ledger Tab
- [ ] Pagination works with >10 entries
- [ ] Running balance accurate across pages
- [ ] Commission earnings display correctly

---

## Company User Pages

### Login as Company User

#### My Merchants Page
- [ ] Shows only merchants assigned to this company
- [ ] Pagination appears when >10 merchants
- [ ] Navigate through pages
- [ ] Merchant details accurate

#### Transactions Page
- [ ] Shows only transactions for this company
- [ ] Pagination works
- [ ] Commission calculations correct
- [ ] "My Net" column accurate

#### Settlements Page
- [ ] Shows only settlements for this company
- [ ] Pagination works
- [ ] Payment details accurate

#### My Ledger Page
- [ ] Pagination works with >10 entries
- [ ] Running balance correct on page 1
- [ ] Running balance continues on page 2+
- [ ] Final balance matches stats at top

---

## Merchant User Pages

### Login as Merchant User (Direct Merchant)

#### My Transactions Page
- [ ] Shows only merchant's transactions
- [ ] Pagination works
- [ ] Commission column shows values (not "—")
- [ ] Gateway details accurate

#### Commission Ledger Page
- [ ] Pagination works with >10 entries
- [ ] Running balance starts correctly
- [ ] Running balance continues across pages
- [ ] Total Earned stat matches ledger

### Login as Merchant User (Affiliate Merchant)

#### My Transactions Page
- [ ] Shows merchant's transactions
- [ ] Pagination works
- [ ] Commission column shows "—" (paid to affiliate)

#### Commission Ledger Page
- [ ] Shows message: "You operate under an affiliate..."
- [ ] No pagination (no ledger for affiliate merchants)

---

## Edge Case Testing

### Exactly 10 Items
- [ ] No pagination appears
- [ ] All 10 items visible

### Exactly 11 Items
- [ ] Pagination appears
- [ ] Page 1: 10 items (if 10 per page selected)
- [ ] Page 2: 1 item

### Empty Results After Filter
- [ ] "No items found" message
- [ ] No pagination controls
- [ ] No errors in console

### Large Dataset (1000+ items)
- [ ] Pagination renders quickly
- [ ] Page changes are smooth
- [ ] No performance lag
- [ ] Memory usage acceptable

### First Page
- [ ] Previous button disabled
- [ ] First page highlighted
- [ ] Can navigate to next page

### Last Page
- [ ] Next button disabled
- [ ] Last page highlighted
- [ ] Can navigate to previous page

### Middle Pages
- [ ] Both prev/next enabled
- [ ] Current page highlighted
- [ ] Ellipsis shows for distant pages

---

## Items Per Page Selector

- [ ] Default is 25 items per page
- [ ] Can change to 10
- [ ] Can change to 50
- [ ] Can change to 100
- [ ] Changing resets to page 1
- [ ] Total pages recalculate correctly

---

## Search/Filter Integration

### Search Field
- [ ] Type search query
- [ ] Results filter correctly
- [ ] Page resets to 1
- [ ] Pagination recalculates
- [ ] Clear search returns all items

### Company Filter (Transactions)
- [ ] Select company
- [ ] Transactions filter correctly
- [ ] Page resets to 1
- [ ] Pagination recalculates

### Date Filter (Reports)
- [ ] Set date range
- [ ] Results filter correctly
- [ ] Page resets to 1
- [ ] Pagination recalculates

---

## Visual/UX Testing

### Dark Theme Consistency
- [ ] Pagination uses dark background (#0A0A0B)
- [ ] Active page uses green accent (#10B981)
- [ ] Text is readable (light gray)
- [ ] Hover states work
- [ ] Borders are subtle

### Smooth Scrolling
- [ ] Page change scrolls to top
- [ ] Scroll is smooth, not instant
- [ ] User not disoriented after page change

### Button States
- [ ] Active page has green background
- [ ] Hover shows lighter background
- [ ] Disabled buttons are grayed out
- [ ] Cursor changes on hover (pointer)

### Responsive Layout
- [ ] Desktop (1920px): Full layout
- [ ] Tablet (768px): Compact layout
- [ ] Mobile (375px): Stacked if needed
- [ ] No horizontal scroll

---

## Export Functionality

### CSV Export
- [ ] Open page with pagination
- [ ] Navigate to page 2 or 3
- [ ] Click "Export CSV"
- [ ] Verify CSV contains ALL items, not just current page
- [ ] Verify data accuracy

### PDF Export
- [ ] Navigate to page 2 or 3
- [ ] Click "Export PDF"
- [ ] Verify PDF contains ALL items
- [ ] Verify formatting correct

---

## Browser Compatibility

### Chrome
- [ ] Pagination renders correctly
- [ ] All interactions work
- [ ] No console errors

### Firefox
- [ ] Pagination renders correctly
- [ ] All interactions work
- [ ] No console errors

### Safari
- [ ] Pagination renders correctly
- [ ] All interactions work
- [ ] No console errors

### Edge
- [ ] Pagination renders correctly
- [ ] All interactions work
- [ ] No console errors

---

## Performance Testing

### Load Time
- [ ] Page loads quickly with 100 items
- [ ] Page loads quickly with 1000 items
- [ ] No noticeable lag

### Page Change Speed
- [ ] Page changes are instant
- [ ] No loading spinners needed
- [ ] Smooth transition

### Memory Usage
- [ ] No memory leaks on repeated page changes
- [ ] Browser doesn't slow down
- [ ] Console shows no warnings

---

## Console Error Check

- [ ] Open browser console (F12)
- [ ] Navigate through all pages
- [ ] No JavaScript errors
- [ ] No React warnings
- [ ] No missing key warnings

---

## Accessibility (Basic)

- [ ] Tab through pagination controls
- [ ] Buttons are focusable
- [ ] Enter key works on focused button
- [ ] Screen reader can read page numbers
- [ ] Contrast ratio is acceptable

---

## Regression Testing

### Existing Functionality
- [ ] Create new item (company/merchant/etc.)
- [ ] Edit existing item
- [ ] Delete item
- [ ] Search still works
- [ ] Filters still work
- [ ] Modals open correctly
- [ ] Forms submit correctly

---

## Sign-Off

### Developer Testing
- [ ] All code syntax checked
- [ ] All imports correct
- [ ] useState hooks initialized
- [ ] Event handlers bound correctly
- [ ] No linting errors

### QA Testing
- [ ] All pages manually tested
- [ ] All user types tested
- [ ] Edge cases verified
- [ ] No bugs found

### Final Approval
- [ ] Product owner reviewed
- [ ] Design approved
- [ ] Ready for production

---

## Known Issues / Notes

_Document any issues found during testing:_

1. 
2. 
3. 

---

## Test Results Summary

**Date Tested**: _______________
**Tested By**: _______________
**Browser**: _______________
**Total Tests**: 200+
**Passed**: _____
**Failed**: _____
**Notes**: _______________

---

**Overall Status**: 
- [ ] ✅ All Tests Passed - Ready for Production
- [ ] ⚠️ Minor Issues - Fix and Retest
- [ ] ❌ Major Issues - Requires Development Work
