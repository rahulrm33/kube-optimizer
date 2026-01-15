# Testing Guide

This guide walks you through testing the K8s Resource Optimizer dashboard.

## Prerequisites

Before testing, ensure you have:

1. PostgreSQL running
2. Database created and seeded with sample data
3. Web server running on http://localhost:8080

If not, run:

```bash
make db-setup
go run scripts/seed.go
go run cmd/web/main.go
```

## Manual Testing Checklist

### 1. Dashboard Home Page (GET /)

**Test**: Open http://localhost:8080

**Expected Results**:
- ✅ Page loads without errors
- ✅ Four summary cards display at the top:
  - Total Monthly Savings (green, with dollar amount)
  - Over-provisioned count (red)
  - Under-provisioned count (yellow)
  - Optimal count (green)
- ✅ Two charts visible:
  - Pie chart showing pod distribution
  - Bar chart showing top 10 wasteful pods
- ✅ Filters section with 4 dropdowns/inputs
- ✅ Recommendations table with data
- ✅ Last updated timestamp in navbar

**Screenshot**: Take a screenshot for documentation

---

### 2. Summary Cards

**Test**: Verify the summary cards show realistic data

**Expected Results**:
- ✅ Total savings > $0
- ✅ Sum of over + under + optimal = total pods
- ✅ Cards are clickable/hoverable (slight animation)
- ✅ Numbers are formatted correctly (no decimals for counts)

---

### 3. Charts Rendering

**Test**: Verify Chart.js charts render correctly

**Pie Chart**:
- ✅ Shows 3 segments (over/under/optimal)
- ✅ Colors: Red, Yellow, Green
- ✅ Legend displays at bottom
- ✅ Segments are proportional to data

**Bar Chart**:
- ✅ Shows up to 10 bars (horizontal)
- ✅ Bars are blue
- ✅ Pod names on Y-axis
- ✅ Savings amounts on X-axis
- ✅ Bars sorted by savings (highest first)

---

### 4. Namespace Filter

**Test**: Filter by namespace

**Steps**:
1. Click namespace dropdown
2. Select "production" (or any namespace)
3. Wait for table to update

**Expected Results**:
- ✅ Dropdown populates with namespaces from database
- ✅ Table shows only pods from selected namespace
- ✅ Loading indicator appears briefly
- ✅ No JavaScript errors in console

---

### 5. Status Filter

**Test**: Filter by pod status

**Steps**:
1. Select "Over-provisioned" from status dropdown
2. Observe table updates
3. Try "Under-provisioned"
4. Try "Optimal"
5. Select "All Status"

**Expected Results**:
- ✅ Table filters correctly for each status
- ✅ Status badges match filter selection
- ✅ Count matches filtered results
- ✅ Selecting "All" shows all pods again

---

### 6. Sort Functionality

**Test**: Sort recommendations

**Steps**:
1. Select "Savings (High to Low)" - default
2. Select "Waste % (High to Low)"
3. Select "Pod Name (A-Z)"

**Expected Results**:
- ✅ Table re-sorts correctly
- ✅ Savings sort: highest dollar amounts first
- ✅ Waste sort: highest percentages first
- ✅ Name sort: alphabetical order

---

### 7. Search Functionality

**Test**: Search for pods

**Steps**:
1. Type "api" in search box
2. Wait 300ms (debounce)
3. Observe results
4. Clear search box

**Expected Results**:
- ✅ Search filters pods by name
- ✅ Results appear after brief delay (debounce)
- ✅ Partial matches work (e.g., "api" matches "api-server-123")
- ✅ Clearing search shows all pods again
- ✅ Search is case-insensitive

---

### 8. Recommendations Table

**Test**: Verify table displays all required columns

**Expected Columns**:
- ✅ Namespace
- ✅ Pod name
- ✅ Container name
- ✅ Current CPU (formatted: "1.5 cores" or "500m")
- ✅ Recommended CPU
- ✅ Current Memory (formatted: "2.5 GB" or "512 MB")
- ✅ Recommended Memory
- ✅ Savings (formatted: "$45.50")
- ✅ Confidence (badge: high/medium/low)
- ✅ Status (colored badge: over/under/optimal)
- ✅ Actions (2 buttons)

**Expected Results**:
- ✅ All columns display correctly
- ✅ Data is properly formatted
- ✅ Badges have correct colors
- ✅ Table is responsive (scrolls horizontally on mobile)

---

### 9. View Details Button

**Test**: Click the eye icon button

**Steps**:
1. Click the eye (👁) button on any row
2. Check what happens

**Expected Results**:
- ✅ Attempts to open `/pod/{namespace}/{name}` in new tab
- ✅ Currently shows 404 (detail page not implemented in this MVP)
- ✅ No errors in console
- ✅ Original page remains open

**Note**: Detail page implementation would be a future enhancement.

---

### 10. Download YAML Button

**Test**: Download YAML patch

**Steps**:
1. Click the download (⬇) button on any row
2. Wait for download

**Expected Results**:
- ✅ Toast notification appears: "Generating YAML patch..."
- ✅ File downloads automatically
- ✅ Filename format: `patch-{namespace}-{podname}.yaml`
- ✅ File contains valid YAML
- ✅ YAML includes:
  - apiVersion, kind, metadata
  - Container name
  - Resources (requests and limits)
  - CPU and memory values

**Verify YAML Content**:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: api-server-123
  namespace: production
spec:
  containers:
  - name: app
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "600m"
        memory: "614Mi"
```

---

### 11. Refresh Button

**Test**: Manual refresh

**Steps**:
1. Click "Refresh" button in navbar
2. Observe updates

**Expected Results**:
- ✅ Toast notification: "Refreshing data..."
- ✅ Loading indicator appears
- ✅ Summary cards update
- ✅ Charts update
- ✅ Table updates
- ✅ "Last updated" timestamp updates
- ✅ Success toast: "Data refreshed successfully"

---

### 12. Auto-Refresh

**Test**: Wait 5 minutes

**Steps**:
1. Note the current time
2. Leave dashboard open
3. Wait 5 minutes
4. Observe automatic refresh

**Expected Results**:
- ✅ Dashboard refreshes automatically after 5 minutes
- ✅ No user interaction required
- ✅ Toast notification appears
- ✅ Data updates

**Note**: You can modify the interval in the JavaScript to test faster (e.g., 30 seconds).

---

### 13. Dark Mode Toggle

**Test**: Toggle theme

**Steps**:
1. Click moon icon in navbar
2. Observe theme change
3. Click sun icon to toggle back
4. Refresh page

**Expected Results**:
- ✅ Theme switches to dark mode
- ✅ All elements adapt (cards, table, navbar)
- ✅ Icon changes from moon to sun
- ✅ Charts remain visible
- ✅ Theme persists after page refresh (localStorage)
- ✅ Text remains readable in both modes

---

### 14. Responsive Design

**Test**: Mobile/tablet view

**Steps**:
1. Open browser DevTools (F12)
2. Toggle device toolbar (Ctrl+Shift+M)
3. Test various screen sizes:
   - Mobile: 375px
   - Tablet: 768px
   - Desktop: 1920px

**Expected Results**:
- ✅ Summary cards stack vertically on mobile
- ✅ Charts resize appropriately
- ✅ Table scrolls horizontally if needed
- ✅ Filters stack vertically on mobile
- ✅ Navbar remains functional
- ✅ Buttons are touch-friendly (min 44px)
- ✅ No horizontal overflow

---

### 15. Toast Notifications

**Test**: Verify toast messages

**Actions that trigger toasts**:
- Refresh button → "Refreshing data..." then "Data refreshed successfully"
- Download YAML → "Generating YAML patch..."
- Auto-refresh → "Data refreshed successfully"
- Errors → "Error loading recommendations" (simulate by stopping server)

**Expected Results**:
- ✅ Toasts appear in top-right corner
- ✅ Correct color for type (info=blue, success=green, danger=red)
- ✅ Auto-dismiss after 5 seconds
- ✅ Close button works
- ✅ Multiple toasts stack vertically

---

### 16. Loading Indicators

**Test**: Loading states

**Steps**:
1. Apply a filter
2. Observe loading indicator
3. Perform search
4. Click refresh

**Expected Results**:
- ✅ Loading spinner appears during data fetch
- ✅ Message: "Loading data..."
- ✅ Spinner is centered
- ✅ Disappears when data loads
- ✅ Table shows during load (doesn't flash)

---

## API Testing

### Test All API Endpoints

```bash
# 1. Health Check
curl http://localhost:8080/health
# Expected: {"status":"healthy","timestamp":"..."}

# 2. Get Statistics
curl http://localhost:8080/api/stats
# Expected: JSON with total_pods, over_provisioned, etc.

# 3. Get All Pods
curl http://localhost:8080/api/pods
# Expected: {"pods":[...],"total":50,"page":1}

# 4. Filter by Namespace
curl "http://localhost:8080/api/pods?namespace=production"
# Expected: Only production pods

# 5. Filter by Status
curl "http://localhost:8080/api/pods?status=over-provisioned"
# Expected: Only over-provisioned pods

# 6. Sort by Waste
curl "http://localhost:8080/api/pods?sort_by=waste"
# Expected: Pods sorted by waste percentage

# 7. Limit Results
curl "http://localhost:8080/api/pods?limit=10"
# Expected: Only 10 pods

# 8. Search Pods
curl "http://localhost:8080/api/pods?search=api"
# Expected: Pods matching "api"

# 9. Get Pod Detail
curl http://localhost:8080/api/pod/production/api-server-123
# Expected: {"pod":{...},"analysis":{...},"usage_history":[...]}

# 10. Get Recommendations
curl http://localhost:8080/api/recommendations
# Expected: {"recommendations":[...],"total_savings":...}

# 11. Filter Recommendations by Confidence
curl "http://localhost:8080/api/recommendations?confidence=high"
# Expected: Only high confidence recommendations

# 12. Filter by Minimum Savings
curl "http://localhost:8080/api/recommendations?min_savings=50"
# Expected: Only recommendations with $50+ savings

# 13. Get Namespaces
curl http://localhost:8080/api/namespaces
# Expected: {"namespaces":["default","production",...]}

# 14. Mark Recommendation Applied
curl -X POST http://localhost:8080/api/recommendations/1/apply \
  -H "Content-Type: application/json" \
  -d '{"applied":true}'
# Expected: {"success":true,"message":"..."}
```

---

## Performance Testing

### Load Time

**Test**: Measure page load time

**Steps**:
1. Open DevTools → Network tab
2. Hard refresh (Ctrl+Shift+R)
3. Check "Load" time at bottom

**Expected Results**:
- ✅ Page loads in < 2 seconds
- ✅ Charts render in < 1 second
- ✅ API calls complete in < 500ms

### Large Dataset

**Test**: Performance with many pods

**Steps**:
1. Modify seed script to create 500 pods
2. Run seed script
3. Load dashboard
4. Test filters and sorting

**Expected Results**:
- ✅ Page still loads in reasonable time
- ✅ Filters work smoothly
- ✅ No browser lag
- ✅ Consider pagination for 1000+ pods

---

## Browser Compatibility

Test in multiple browsers:

- ✅ Chrome/Edge (Chromium)
- ✅ Firefox
- ✅ Safari
- ✅ Mobile Safari (iOS)
- ✅ Chrome Mobile (Android)

**Expected Results**:
- ✅ All features work in all browsers
- ✅ Charts render correctly
- ✅ CSS displays properly
- ✅ JavaScript functions work

---

## Error Handling

### Test Error Scenarios

**1. Database Down**:
```bash
# Stop PostgreSQL
sudo systemctl stop postgresql
# Reload dashboard
```
Expected: Error page or error message

**2. Invalid API Request**:
```bash
curl http://localhost:8080/api/pod/invalid/invalid
```
Expected: 404 JSON error

**3. Invalid Recommendation ID**:
```bash
curl http://localhost:8080/api/recommendations/99999/yaml
```
Expected: 404 error

**4. Malformed POST Request**:
```bash
curl -X POST http://localhost:8080/api/recommendations/1/apply \
  -H "Content-Type: application/json" \
  -d 'invalid json'
```
Expected: 400 Bad Request

---

## Security Testing

### Basic Security Checks

**1. SQL Injection**:
```bash
curl "http://localhost:8080/api/pods?namespace=production';DROP TABLE pods;--"
```
Expected: No SQL injection (parameterized queries)

**2. XSS**:
```bash
curl "http://localhost:8080/api/pods?search=<script>alert('xss')</script>"
```
Expected: Script tags escaped in output

**3. CORS Headers**:
```bash
curl -I http://localhost:8080/api/stats
```
Expected: CORS headers present

---

## Accessibility Testing

### A11y Checklist

- ✅ All buttons have accessible labels
- ✅ Charts have alt text or aria-labels
- ✅ Color contrast meets WCAG AA standards
- ✅ Keyboard navigation works (Tab through elements)
- ✅ Screen reader compatible
- ✅ Focus indicators visible

**Tools**:
- Chrome Lighthouse audit
- axe DevTools extension
- WAVE browser extension

---

## Test Report Template

After testing, document results:

```markdown
## Test Report - [Date]

### Environment
- OS: macOS/Linux/Windows
- Browser: Chrome 120
- Database: PostgreSQL 15
- Go Version: 1.21

### Summary
- Total Tests: 50
- Passed: 48
- Failed: 2
- Skipped: 0

### Failed Tests
1. **Dark mode persistence**: Theme doesn't persist after refresh
   - Severity: Low
   - Steps to reproduce: ...
   - Expected: ...
   - Actual: ...

### Performance Metrics
- Page Load: 1.2s
- API Response: 250ms avg
- Chart Render: 400ms

### Recommendations
- Add pagination for large datasets
- Implement pod detail page
- Add export to CSV feature
```

---

## Automated Testing (Future)

For production, consider adding:

1. **Unit Tests**:
   ```go
   func TestGetPods(t *testing.T) { ... }
   ```

2. **Integration Tests**:
   - Test API endpoints
   - Test database queries

3. **E2E Tests**:
   - Selenium/Playwright
   - Test user workflows

4. **Load Tests**:
   - Apache Bench
   - k6 load testing

---

## Continuous Testing

Set up CI/CD pipeline:

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: make test
      - run: make lint
```

---

## Need Help?

If tests fail:
1. Check server logs
2. Check browser console (F12)
3. Verify database connection
4. Review PostgreSQL logs
5. Check network tab in DevTools

Happy testing! 🧪

