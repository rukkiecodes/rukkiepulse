# Plan: Clockee Feature Implementation

**TL;DR:** Implement 6 major feature groups across backend, admin panel, and staff mobile app: (1) Staff pickup verification & profile pictures, (2) Admin parent management and pickup actions in Pickup tab, (3) SMS payment flow simplification, (4) Real-time support messaging with Socket.io, (5) Institution coordinate setup/geofencing, (6) Timezone management system. Estimated scope: Backend API additions (~8 endpoints), Admin UI enhancements (~5 pages), Mobile refinements (~2 features), Socket.io infrastructure setup.

---

## **Steps**

### **Phase 1: Backend Infrastructure**

1. **Database Schema Updates** (`server/schema.sql` or migration)
   - Add `timezones` table with timezone data (ID, name, offset, region)
   - Add `timezone_id` to `institutions` table (nullable, default NULL → use UTC)
   - Add `timezone_id` to `users` table (for user-level override, optional)
   - Ensure `institutions.latitude, longitude, radius` are non-nullable with defaults (or add migration)
   - Add `is_active` to `support_tickets` if not exists (for open/resolved filtering)
   - Add `sent_to_admin_email` boolean to `support_tickets` to track if admin was notified

2. **New Backend Endpoints** (`server/routes/` and `server/controllers/`)
   - **Timezone Endpoints** (`/api/timezones`)
     - `GET /` - List all timezones with ID, name, offset, region (public)
     - Used by admin panel and mobile apps
   
   - **Parent Information Endpoints** (`/api/parents` or extend `/api/users`)
     - `GET /:parentId` - Full parent info with assigned children, pickup history (admin only)
     - `PATCH /:parentId` - Update parent status (suspension, approval, deletion) with audit log
     - `PATCH /:parentId/assign-child` - Assign student to parent
     - Response should trigger notification service
   
   - **Pickup Actions Endpoints** (extend `/api/pickups`)
     - `GET /:pickupId/full-details` - Complete pickup info with pickup person data, admin only
     - `PATCH /:pickupId/action` - Perform action on pickup (e.g., cancel, redo verification)
   
   - **Support Messaging Endpoints** (`/api/support` - new or enhanced)
     - `GET /messages/:ticketId` - Get all messages in a conversation
     - `POST /messages/:ticketId` - Admin sends message to staff/parent
     - `GET /tickets` - Admin views their open support tickets
     - These will integrate with Socket.io for real-time delivery
   
   - **SMS Credit Endpoints** (modify `/api/sms/payment`)
     - Remove the external verification endpoint dependency
     - `POST /credit-directly` - Skip external verification, credit institute directly, send receipt email
     - Keep existing Paystack/payment gateway integration for initial purchase

3. **Notification Service Enhancement** (`server/services/NotificationService.js` - new or extend)
   - Create `sendAdminNotification()` method to email institute admin when:
     - Support message received from staff/parent
     - Parent account actions taken (suspension, etc.)
   - Create `formatNotificationEmail()` template helpers (HTML email templates)

4. **Socket.io Setup** (`server/server.js` or new `server/socket.js`)
   - Initialize Socket.io on server (already installed as dependency)
   - Create namespace `/support` for real-time messaging
   - Implement rooms per institution + user (e.g., `/support/inst-{id}:user-{id}`)
   - Middleware to authenticate socket connections via JWT token
   - Events:
     - `admin:message` - Admin sends message (broadcast to staff/parent room)
     - `staff:message` / `parent:message` - Receive message from client
     - `message:received` - Acknowledge receipt back to sender
   - Store messages in DB before broadcasting (idempotency)

5. **Geofencing Logic** (extend `/server/services/LocationService.js`)
   - Validate coordinates against institution radius when staff clocks in
   - Return institution setup status (has coordinates: true/false)
   - Endpoint `/api/institutions/{id}/coordinates` with GET/PATCH

6. **Timezone Service** (`server/services/TimezoneService.js` - new)
   - Load timezones from DB on startup (cache in memory)
   - Utility: `convertToInstituteTimezone(dateUtc, instituteId)` for normalizing all timestamps
   - Apply during API response formatting for all date-sensitive data

---

### **Phase 2: Admin Panel Updates**

1. **Pages: Enhanced Pickup Tab** (`admin/pages/admin/Pickup.vue` - refactor/create)
   - **Parents Subtab** (new view)
     - Table: List all parents with children count, last pickup date
     - Click row → opens drawer/modal with:
       - Parent full info: name, email, phone, photo, address
       - Assigned children list (with assign/remove buttons)
       - Recent pickups (last 10)
       - Action buttons: `Approve` / `Suspend` / `Delete` / `Assign Child`
       - Confirmation dialog before each action with reason input
       - After action: automatic SMS to parent + email + notification broadcast
   
   - **Pickup History Subtab** (enhanced)
     - Table: All pickups with code, date, parent, child, pickup person, status
     - Click row → opens modal with:
       - Full pickup details: timestamps, location, child info, pickup person photo
       - Verification history: who verified and when
       - Action buttons: `Cancel Pickup` / `Force Mark Complete` / `Files/Attachments`
       - Audit trail of all changes to this pickup

2. **Settings Tab > Institutions** (`admin/pages/admin/Settings.vue` - enhance)
   - Add `Timezone Select` dropdown (populated from `/api/timezones`)
   - Integrate with `GeofencingDialog` (see below)
   - Display current timezone offset and sample times in selected timezone

3. **Geofencing Setup Dialog** (new component `admin/components/GeofencingDialog.vue`)
   - Modal triggered on:
     - First admin login if institution has no coordinates
     - Manual request via Settings button
   - Options:
     a) `Get Current Location` - Request browser geolocation, auto-populate lat/lon
     b) `Manual Entry` - Input fields for latitude, longitude, radius (in meters)
   - Validate coordinates (reasonable lat/lon ranges, radius > 0)
   - Save to institution immediately

4. **SMS Tab Refactor** (`admin/pages/admin/SMS.vue` - modify payment section)
   - Remove payment verification step (old logic)
   - Payment flow:
     - Admin selects SMS package (X credits)
     - Click `Buy Credits`
     - Redirect to payment gateway (Paystack)
     - **On payment success:** Backend route `/api/sms/credit-directly` is called with:
       - `institution_id`
       - `amount_in_credits`
       - `payment_proof` (receipt reference)
     - Backend credits institution + sends receipt email immediately

5. **Support Tab Redesign** (`admin/pages/admin/Support.vue` - new/refactor)
   - Split view:
     - **Left panel:** List of open support tickets/conversations
       - Shows participant (staff/parent name), last message preview, timestamp
       - Badge for unread message count
     - **Right panel:** Message thread view
       - Show all messages chronologically (incoming and sent by admin)
       - Sender avatar, name, timestamp
       - Input box at bottom with send button
       - Real-time updates via Socket.io
   - Real-time notifications: Visual indicator when new message arrives
   - CSS styling: Admin messages right-align, staff/parent left-align (like WhatsApp)

---

### **Phase 3: Staff Mobile App Updates**

1. **Profile Tab Enhancement** (`clockee staff/app/(app)/(tabs)/profile.tsx`)
   - Add new section: `Profile Picture`
   - Display current profile picture (if exists) or placeholder
   - Button: `Change Picture`
   - On click → Opens modal with options:
     - `Take Photo` - Launch camera (`expo-camera`)
     - `Choose from Gallery` - Image picker
   - After selection: Show preview with `Upload` / `Cancel` buttons
   - On upload: Send to `/api/uploads/profile` (Cloudinary) via existing redux action
   - Show loading indicator during upload
   - Success toast confirmation
   - Error handling: Show alert on failure

2. **Pickup Verification Verification Flow** (`clockee staff/app/(clockin)/verifyPickup.tsx` - enhance if needed)
   - Confirm already displays pickup person info (from backend)
   - Ensure shows:
     - Pickup person name, ID, photo
     - Child being picked up
     - Parent info (name, phone)
     - 6-character code verification
   - Add visual feedback: Green checkmark on successful verification
   - Real-time sync: If parent account is suspended mid-verification, show alert

3. **Redux Action Updates** (`clockee staff/store/actions/`)
   - Ensure `confirmPickupVerification()` handles server-side suspension checks
   - Add `fetchParentInfo()` action to get parent details for pickup verification screen
   - Error handling: Handle 403 if parent is suspended

---

### **Phase 4: Real-Time Infrastructure (Socket.io)**

1. **Backend Socket.io Implementation** (`server/socket.js` - new file)
   - Connection middleware: Verify JWT token, extract user role and institution
   - Namespace `/support`:
     - Room structure: `support:inst-{institutionId}:admin` (for admins)
     - Room structure: `support:inst-{institutionId}:user-{userId}` (for staff/parent)
   - Events:
     - `admin:message` → Save to DB + broadcast to user's room
     - `user:message` → Save to DB + emit to admin room + email admin
     - `message:read` → Update read_at timestamp
   - Error handling: Disconnect on auth failure, log all socket events
   - Graceful degradation: If socket fails, fallback to HTTP polling (client side)

2. **Admin Panel Socket.io Client** (`admin/services/socket.js` - new)
   - Connect to Socket.io server on app load
   - Subscribe to `/support` namespace
   - Listen for incoming messages and update Pinia store (`support` store)
   - Emit `admin:message` when admin sends a message
   - Implement reconnection logic with exponential backoff

3. **Mobile App Socket.io Client** (`clockee staff/services/socket.ts` - new)
   - Same connection logic for staff/parents
   - Listen for `admin:message` events
   - Trigger push notification on new message (if app in background)
   - Update Redux store with new messages

---

### **Phase 5: Integration & Data Pipeline**

1. **Timezone Propagation**
   - On institution creation/update: Admin selects timezone
   - All API responses for that institution: Format timestamps in selected timezone
   - Staff/parent apps: Display all times in institution's timezone
   - Backend helper: `convertToInstituteTimezone(date, institutionId)` called in serialization layer

2. **Notification Chaining**
   - When admin takes action on parent (suspension, approval):
     - Trigger `sendParentNotification()` SMS via Termii (if phone on file)
     - Trigger `sendAdminNotification()` email to institute admin
     - Pause/cancel active pickups if suspended
     - Broadcast via Socket.io (if client online)

3. **Coordinate Validation**
   - Staff clock-in: Validate current location within institution radius
   - Return geofencing status in `/api/staff/profile` response
   - Admin: On first login, check institution coordinates; if missing, show mandatory dialog

---

## **Verification**

**Backend Testing:**
- Run migration scripts for schema changes (timezone table, institution columns)
- Test new endpoints locally: `/api/timezones`, `/api/parents/{id}`, `/api/pickups/{id}/full-details`, `/api/support/messages`
- Test Socket.io connection: Connect admin/staff clients, send message, verify broadcast
- Test notification pipeline: Trigger parent suspension, verify email + SMS sent
- Test timezone conversion: Select different timezones, verify timestamp formatting in API responses

**Admin Panel Testing:**
- Load Settings page: Verify timezone dropdown populates from API
- Load Pickup tab: Test parents subtab click, verify parent modal opens
- Test parent actions: Suspend parent, verify email/notification sent
- Load Pickup History: Test click pickup, verify full details modal
- Test Support tab: Send message, verify real-time receipt (Socket.io), check admin inbox email
- Test geofencing dialog: Trigger on first login, request device location, verify save

**Mobile App Testing:**
- Load staff profile: Test change picture flow (camera + gallery)
- Test pickup verification: Verify parent info displays correctly
- Test real-time message: Admin sends support message, verify instant notification on staff app
- Verify timezone: Check that all times display in institution timezone

**End-to-End Scenario:**
1. Admin creates institution with coordinates and selects UTC timezone
2. Parent signs up, gets approval
3. Staff picks up child using 6-digit code → sees parent info
4. Parent account suspended by admin → staff sees alert on next pickup attempt
5. Staff messages admin → admin receives Socket.io notification + email
6. Admin responds → staff gets real-time message notification

---

## **Decisions**

- **Profile Picture Scope:** Staff personal profile picture (not separate pickup identification)
- **Parent Actions:** Actions notify parent immediately and pause ongoing pickups
- **Real-Time Tech:** Socket.io full implementation (not polling fallback initially)
- **Timezone Default:** UTC when institution hasn't selected timezone
- **SMS Payment:** Skip external endpoint verification, credit directly with email receipt
- **Geofencing Dialog:** Mandatory on first login if coordinates missing
- **Support Thread UX:** WhatsApp-style layout (left/right align) in admin panel
