# Student Attendance System - Integration Guide

## Quick Start

This guide will help you integrate the student attendance system into your Clockee application.

## Prerequisites

### Backend Requirements
- PHP 8.1+
- Laravel 10+
- MariaDB 10.5+
- Composer

### Frontend Requirements
- Node.js 16+
- React 18+
- TypeScript

### External Services (Choose One)
- Twilio account (for SMS)
- Termii account (for SMS)
- Africa's Talking account (for SMS)

## Step-by-Step Integration

### 1. Install Backend Dependencies

```bash
cd laravel-backend

# Install Laravel dependencies
composer install

# Install QR code library (optional but recommended)
composer require simplesoftwareio/simple-qrcode
```

### 2. Run Database Migrations

```bash
# Run all migrations
php artisan migrate

# Or run specific migrations
php artisan migrate --path=/database/migrations/006_create_student_qr_codes_table.php
php artisan migrate --path=/database/migrations/007_create_student_attendance_table.php
php artisan migrate --path=/database/migrations/008_create_bulk_import_logs_table.php
```

### 3. Configure Environment Variables

Edit your `.env` file:

```env
# SMS Provider (choose one: twilio, termii, africastalking)
SMS_PROVIDER=twilio

# Twilio Configuration
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=your_auth_token_here
TWILIO_FROM_NUMBER=+1234567890

# OR Termii Configuration
TERMII_API_KEY=your_termii_api_key_here
TERMII_SENDER_ID=Clockee

# OR Africa's Talking Configuration
AFRICASTALKING_API_KEY=your_africastalking_api_key_here
AFRICASTALKING_USERNAME=your_username_here

# File Storage
FILESYSTEM_DISK=public

# QR Code Settings
QR_CODE_SIZE=300
QR_CODE_FORMAT=png
```

### 4. Create Storage Link

```bash
php artisan storage:link
```

This creates a symbolic link from `public/storage` to `storage/app/public` for storing attendance photos.

### 5. Install Frontend Dependencies

```bash
# Install QR code scanning library
npm install jsqr

# Or with yarn
yarn add jsqr
```

### 6. Update User Model

The User model has already been updated with the necessary relationships:
- `qrCode()` - One-to-one relationship with StudentQRCode
- `studentAttendance()` - One-to-many relationship with StudentAttendance

### 7. Add Routes to Your Application

#### For Students

Add to your student dashboard/profile:

```tsx
import { StudentQRDisplay } from './components/StudentQRDisplay';

function StudentDashboard() {
  return (
    <div>
      {/* Other dashboard content */}
      
      <StudentQRDisplay />
    </div>
  );
}
```

#### For Staff/Admin

Add to your admin panel:

```tsx
import { StudentAttendanceManagement } from './components/admin/StudentAttendanceManagement';

function AdminDashboard() {
  return (
    <div>
      {/* Other admin content */}
      
      <StudentAttendanceManagement />
    </div>
  );
}
```

Or use individual components:

```tsx
import { SchoolQRScanner } from './components/SchoolQRScanner';
import { StudentAttendanceHistory } from './components/StudentAttendanceHistory';
import { BulkStudentRegistration } from './components/BulkStudentRegistration';

function AttendancePage() {
  return (
    <div>
      <SchoolQRScanner />
    </div>
  );
}

function AttendanceHistoryPage() {
  return (
    <div>
      <StudentAttendanceHistory />
    </div>
  );
}

function BulkImportPage() {
  return (
    <div>
      <BulkStudentRegistration />
    </div>
  );
}
```

## SMS Provider Setup

### Option 1: Twilio

1. Sign up at [https://www.twilio.com](https://www.twilio.com)
2. Get your Account SID and Auth Token from the dashboard
3. Purchase a phone number
4. Add credentials to `.env`:

```env
SMS_PROVIDER=twilio
TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=your_auth_token_here
TWILIO_FROM_NUMBER=+1234567890
```

**Pricing:** ~$0.0075 per SMS (varies by country)

### Option 2: Termii (Nigeria-focused)

1. Sign up at [https://termii.com](https://termii.com)
2. Get your API key from the dashboard
3. Configure sender ID
4. Add credentials to `.env`:

```env
SMS_PROVIDER=termii
TERMII_API_KEY=your_termii_api_key_here
TERMII_SENDER_ID=Clockee
```

**Pricing:** ~₦2-4 per SMS

### Option 3: Africa's Talking (Africa-focused)

1. Sign up at [https://africastalking.com](https://africastalking.com)
2. Get your API key and username
3. Add credentials to `.env`:

```env
SMS_PROVIDER=africastalking
AFRICASTALKING_API_KEY=your_africastalking_api_key_here
AFRICASTALKING_USERNAME=your_username_here
```

**Pricing:** ~$0.01 per SMS (varies by country)

## Testing

### Test QR Code Generation

1. Create a test student account
2. Login as the student
3. Navigate to the QR code display page
4. Verify QR code is generated and displayed

### Test Attendance Scanning

1. Login as admin/staff
2. Navigate to QR scanner page
3. Select "Clock In"
4. Scan or manually enter a student's QR code
5. Verify:
   - Student photo displays on screen
   - Success message appears
   - Attendance record is created
   - SMS is sent (if configured)

### Test Bulk Import

1. Login as admin
2. Download CSV template
3. Fill in sample student data
4. Upload CSV file
5. Verify:
   - Students are created
   - QR codes are generated
   - Import results are displayed

## Workflow Examples

### Daily Attendance Workflow

**Morning Clock-In:**
1. Staff opens QR Scanner
2. Selects "Clock In" mode
3. Students arrive and show their QR codes
4. Staff scans each student's QR code
5. Student's photo displays on screen for 3 seconds
6. Parent receives SMS: "Clockee Alert: John Doe clocked in at ABC School on Oct 16, 2025 at 8:00 AM. -Clockee"

**Afternoon Clock-Out:**
1. Staff switches to "Clock Out" mode
2. Students show QR codes at dismissal
3. Staff scans each student's QR code
4. Parent receives SMS: "Clockee Alert: John Doe clocked out at ABC School on Oct 16, 2025 at 3:00 PM. -Clockee"

### New Student Registration Workflow

**Individual Registration:**
1. Admin creates new student account
2. System automatically generates unique QR code
3. Student can view/download QR code from their profile
4. Student can print QR code for daily use

**Bulk Registration:**
1. Admin downloads CSV template
2. Admin fills in student information (can be from existing student database)
3. Admin uploads CSV file
4. System processes file and creates all students
5. System generates unique QR codes for all students
6. Admin can export QR codes for printing

## API Integration Examples

### Get Student QR Code (Frontend)

```typescript
async function getStudentQRCode() {
  const response = await fetch('/api/student-attendance/qr-code', {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  
  const data = await response.json();
  return data.data; // { qr_code, qr_data_url, is_active, ... }
}
```

### Process Attendance (Frontend)

```typescript
async function processAttendance(qrCode: string, type: 'clock_in' | 'clock_out') {
  const response = await fetch('/api/student-attendance/process', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      qr_code: qrCode,
      attendance_type: type,
      location: {
        latitude: 6.5244,
        longitude: 3.3792,
      },
    }),
  });
  
  const data = await response.json();
  return data; // { success, message, data: { attendance, student } }
}
```

### Bulk Import (Frontend)

```typescript
async function bulkImportStudents(file: File) {
  const formData = new FormData();
  formData.append('file', file);
  
  const response = await fetch('/api/admin/bulk-import/students', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
    },
    body: formData,
  });
  
  const data = await response.json();
  return data; // { success, message, data: { total_rows, successful_imports, ... } }
}
```

## Customization

### Custom SMS Messages

Edit `/laravel-backend/app/Services/SMSService.php`:

```php
private function formatAttendanceMessage(
    StudentAttendance $attendance,
    User $student,
    Institution $institution
): string {
    $type = $attendance->attendance_type === 'clock_in' ? 'clocked in' : 'clocked out';
    $time = $attendance->attendance_time->format('h:i A');
    $date = $attendance->attendance_time->format('M d, Y');

    // Customize this message
    return "Your child {$student->full_name} {$type} at {$time}. - {$institution->name}";
}
```

### Custom QR Code Format

Edit `/laravel-backend/app/Services/QRCodeService.php`:

```php
private function generateUniqueCode(User $student): string
{
    // Customize the QR code format here
    $prefix = 'STU';
    $institutionId = str_pad($student->institution_id, 4, '0', STR_PAD_LEFT);
    $studentId = str_pad($student->id, 6, '0', STR_PAD_LEFT);
    $random = strtoupper(Str::random(6));

    return "{$prefix}-{$institutionId}-{$studentId}-{$random}";
}
```

### Custom Photo Display Duration

Edit `/components/SchoolQRScanner.tsx`:

```tsx
// Change the timeout duration (currently 3000ms = 3 seconds)
setTimeout(() => {
  setShowPhotoDisplay(false);
  setCurrentStudent(null);
}, 5000); // Change to 5 seconds
```

## Security Best Practices

1. **HTTPS Required**
   - Camera access requires HTTPS in production
   - Use SSL certificates for your domain

2. **QR Code Protection**
   - Educate students not to share QR codes
   - Implement QR code regeneration if compromised
   - Consider time-limited QR codes for extra security

3. **SMS Security**
   - Store SMS provider credentials securely
   - Use environment variables, never commit to git
   - Rotate API keys regularly

4. **Photo Storage**
   - Photos are stored in `storage/app/public/attendance_photos`
   - Ensure proper permissions (755 for directories, 644 for files)
   - Consider implementing automatic deletion after 30 days

5. **Rate Limiting**
   - Implement rate limiting on QR scan endpoints
   - Prevent abuse of SMS sending

## Troubleshooting

### Common Issues

**1. QR Code Not Displaying**
- Check if migrations ran successfully
- Verify student has a QR code in database
- Check browser console for errors

**2. Camera Not Working**
- Ensure HTTPS connection
- Check browser permissions
- Verify jsQR library is installed
- Test on different browsers

**3. SMS Not Sending**
- Verify SMS provider credentials
- Check institution has SMS units
- Review logs in `storage/logs/laravel.log`
- Test with SMS provider's test console

**4. Bulk Import Failing**
- Check CSV format matches template exactly
- Ensure email addresses are unique
- Verify file size is under 10MB
- Check for special characters in data

**5. Photo Not Displaying**
- Verify student has profile photo uploaded
- Check storage link exists: `php artisan storage:link`
- Verify file permissions on storage directory

### Debug Mode

Enable debug logging for SMS:

```php
// In SMSService.php, temporarily add:
Log::info('SMS Send Attempt', [
    'phone' => $phone,
    'message' => $message,
    'provider' => $provider,
]);
```

## Performance Optimization

### Database Indexes

The migrations already include proper indexes. To verify:

```sql
SHOW INDEX FROM student_qr_codes;
SHOW INDEX FROM student_attendance;
```

### Caching

Consider caching QR codes:

```php
// In QRCodeService.php
public function getQRCode(User $student): StudentQRCode
{
    return Cache::remember(
        "student_qr_{$student->id}",
        now()->addHours(24),
        fn() => $student->qrCode
    );
}
```

### Queue SMS Sending

For better performance, queue SMS sending:

```php
// In StudentAttendanceController.php
dispatch(function () use ($attendance, $student, $phone) {
    $this->smsService->sendAttendanceNotification($attendance, $student, $phone);
});
```

## Next Steps

1. **Test the system thoroughly** with sample students
2. **Train staff** on using the QR scanner
3. **Inform parents** about SMS notifications
4. **Print student QR codes** for daily use
5. **Monitor SMS usage** and costs
6. **Review attendance reports** regularly

## Support

For additional help:
- Review the main README.md
- Check Laravel logs: `storage/logs/laravel.log`
- Review browser console for frontend errors
- Contact development team

## Version History

- **v1.0.0** - Initial release
  - QR code generation
  - Clock-in/out functionality
  - SMS notifications
  - Bulk student import
  - Attendance history
