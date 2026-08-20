# Pagination Testing Guide

## Quick Start

```bash
cd "c:\gateway provider\frontend"
npm run dev
```

Then open http://localhost:5173 in your browser.

---

## Testing Checklist

### 1. Affiliates Page (5 min)

**Navigate to:** Affiliates

- [ ] **Initial Load**
  - Pagination appears if >10 affiliates
  - Shows "Showing 1–25 of X"
  - Page 1 is highlighted

- [ ] **Navigation**
  - Click page 2 → shows items 26-50
  - Click "Next" → advances to next page
  - Click "Prev" → goes back
  - Click specific page number → jumps correctly

- [ ] **Per Page Selector**
  - Change to "10 per page" → shows 10 items
  - Change to "50 per page" → shows 50 items
  - Change to "100 per page" → shows 100 items
  - Each change resets to page 1

- [ ] **CRUD Operations**
  - Add new affiliate → stays on current page
  - Edit affiliate → returns to same page
  - Delete affiliate → pagination adjusts

### 2. Gateways Page (3 min)

**Navigate to:** Gateways

- [ ] Pagination appears if >10 gateways
- [ ] Page navigation works
- [ ] Per-page selector works
- [ ] Add/Edit/Delete maintains pagination

### 3. Transactions Page (10 min)

**Navigate to:** Transactions

- [ ] **Without Filter**
  - Shows all transactions paginated
  - Default 25 per page
  - Navigation works

- [ ] **With Company Filter**
  - Select a company from dropdown
  - Pagination resets to page 1
  - Shows filtered count
  - Pagination works on filtered results

- [ ] **Add Transaction**
  - Click "+ Add Transaction"
  - Fill form and save
  - Resets to page 1 (shows new transaction)

- [ ] **Delete Transaction**
  - Delete a transaction
  - Pagination adjusts automatically

### 4. Settlements Page (5 min)

**Navigate to:** Settlements

- [ ] Shows settlements paginated (newest first)
- [ ] Page navigation works
- [ ] Add settlement → resets to page 1
- [ ] Delete settlement → pagination adjusts

### 5. Ledgers Page (15 min)

**Navigate to:** Ledgers

#### **Company Ledger Tab**

- [ ] **List View**
  - Shows paginated list of companies
  - Navigation works

- [ ] **Detail View**
  - Click "View" on a company
  - Shows detailed ledger entries paginated
  - Running balance calculates correctly across pages
  - "Back to companies" button works

#### **Affiliate Ledger Tab**

- [ ] Shows paginated list of affiliates
- [ ] Navigation works
- [ ] Commission calculations correct

#### **Merchant Ledger Tab**

- [ ] Shows paginated list of merchants
- [ ] Navigation works
- [ ] Direct/Affiliate distinction shown correctly

### 6. Reports Page (15 min)

**Navigate to:** Reports

Test each tab with pagination:

#### **Company Wise Tab**

- [ ] Shows paginated company report
- [ ] Navigation works
- [ ] Calculations correct across pages

#### **Merchant Wise Tab**

- [ ] Shows paginated merchant report
- [ ] Switch from previous tab resets to page 1

#### **Affiliate Wise Tab**

- [ ] Shows paginated affiliate report
- [ ] Tab switch resets pagination

#### **Gateway Wise Tab**

- [ ] Shows paginated gateway report
- [ ] Tab switch resets pagination

#### **Settlement Tab**

- [ ] Shows paginated settlement report
- [ ] Tab switch resets pagination

#### **Outstanding Tab**

- [ ] Shows paginated outstanding balances
- [ ] Tab switch resets pagination

#### **Date Filters**

- [ ] Set "From Date" → resets to page 1
- [ ] Set "To Date" → resets to page 1
- [ ] Clear filters → resets to page 1
- [ ] Pagination works on filtered results

#### **Export Functions**

- [ ] Click "⬇ Excel (CSV)" → exports ALL data (not just current page)
- [ ] Click "⬇ PDF" → exports ALL data (not just current page)

---

## Edge Cases to Test

### Empty States

- [ ] **Zero Items**
  - No pagination shown
  - Empty state message appears

- [ ] **Exactly 10 Items**
  - No pagination shown (threshold = 10)

- [ ] **11 Items**
  - Pagination appears
  - Shows 2 pages

### Single Page

- [ ] **1-10 items**
  - No pagination
  
- [ ] **11-25 items** (default page size)
  - Shows pagination
  - Only page 1 exists
  - Prev/Next buttons disabled

### Large Datasets

- [ ] **100+ items**
  - Page numbers show with ellipsis
  - Example: `1 ... 4 5 6 ... 20`
  - Navigation still works correctly

### Boundary Cases

- [ ] **First Page**
  - "Prev" button disabled
  - Page 1 highlighted
  
- [ ] **Last Page**
  - "Next" button disabled
  - Last page highlighted

- [ ] **Middle Page**
  - Both Prev and Next enabled
  - Current page highlighted

---

## Visual/UX Tests

### Desktop (>768px)

- [ ] Pagination bar spans full width
- [ ] Info on left, controls on right
- [ ] Buttons properly sized and spaced
- [ ] Hover effects work

### Mobile (<768px)

- [ ] Pagination stacks vertically
- [ ] Info section full width
- [ ] Controls centered
- [ ] Buttons still clickable
- [ ] Dropdown works on touch

### Scroll Behavior

- [ ] Click any page number → smoothly scrolls to top
- [ ] Smooth scroll animation visible
- [ ] Focus maintained after scroll

### Accessibility

- [ ] Tab navigation works
- [ ] Buttons have proper labels
- [ ] Current page announced by screen readers
- [ ] Disabled states clearly indicated

---

## Performance Tests

### Large Lists

Create test data:

1. Add 200+ affiliates (or use seed data)
2. Navigate to Affiliates page
3. Check:
   - [ ] Page loads quickly (<1s)
   - [ ] No lag when changing pages
   - [ ] No lag when changing items per page

### Filter Performance

1. Go to Transactions with 500+ records
2. Apply company filter
3. Check:
   - [ ] Filter applies instantly
   - [ ] Pagination updates immediately
   - [ ] No performance degradation

---

## Browser Compatibility

Test on each browser:

### Chrome/Edge

- [ ] All features work
- [ ] Smooth scrolling works
- [ ] Styles render correctly

### Firefox

- [ ] All features work
- [ ] Smooth scrolling works
- [ ] Styles render correctly

### Safari

- [ ] All features work
- [ ] Smooth scrolling works
- [ ] Styles render correctly

### Mobile Safari (iOS)

- [ ] Touch interactions work
- [ ] Responsive layout correct
- [ ] Dropdowns work on mobile

### Mobile Chrome (Android)

- [ ] Touch interactions work
- [ ] Responsive layout correct
- [ ] Dropdowns work on mobile

---

## Common Issues & Solutions

### Issue: Pagination not showing

**Check:**
- Is data count > 10?
- Is the data array empty?
- Check console for errors

### Issue: Page numbers not updating

**Check:**
- Is currentPage state being updated?
- Is onPageChange callback being called?
- Check React DevTools for state

### Issue: Running balance wrong in ledgers

**Check:**
- Balance calculation in paginated view
- Verify it matches non-paginated total
- Check startIndex calculation

### Issue: Export includes only current page

**Check:**
- Export function should use `rows`, not `paginatedRows`
- Verify it's exporting full dataset

### Issue: Styles not applied

**Check:**
- CSS file imported in main entry point?
- Browser cache cleared?
- Check browser DevTools for CSS errors

---

## Automated Testing (Optional)

### Unit Tests

```bash
npm run test
```

Test cases to add:

```javascript
describe('Pagination Component', () => {
  it('should display correct page numbers', () => {});
  it('should handle page changes', () => {});
  it('should handle items per page changes', () => {});
  it('should disable prev on first page', () => {});
  it('should disable next on last page', () => {});
});
```

### E2E Tests (Cypress/Playwright)

```javascript
describe('Affiliates Pagination', () => {
  it('should paginate affiliates list', () => {
    cy.visit('/affiliates');
    cy.get('.pagination-container').should('be.visible');
    cy.get('[data-page="2"]').click();
    cy.get('.pagination-container').should('contain', 'Showing 26–50');
  });
});
```

---

## Sign-off Checklist

Before marking as complete:

- [ ] All 6 pages tested manually
- [ ] All edge cases verified
- [ ] Mobile responsive tested
- [ ] Browser compatibility confirmed
- [ ] Performance acceptable
- [ ] No console errors
- [ ] Documentation reviewed
- [ ] Code reviewed by team

---

## Success Criteria

✅ **Passed** if:
- All functionality works as expected
- No visual glitches or layout issues
- Performance is acceptable (no lag)
- Works on all supported browsers
- Mobile experience is good
- No console errors or warnings

❌ **Failed** if:
- Pagination doesn't appear when it should
- Page navigation doesn't work
- Calculations are incorrect
- Performance is poor
- Browser compatibility issues
- Mobile layout broken

---

## Estimated Testing Time

- **Quick Test:** 15 minutes (basic functionality)
- **Full Test:** 60 minutes (all features + edge cases)
- **Comprehensive Test:** 90 minutes (includes mobile + browsers)

---

**Priority:** High  
**Complexity:** Medium  
**Risk:** Low (non-breaking change)

**Status:** Ready for testing  
**Tester:** [Your Name]  
**Date:** [Test Date]
