# Student Attendance System - Integration Guide

## 📍 Where to Find Components

### 1. Admin Dashboard - Student Management Tab

**Location:** `/components/AdminDashboard.tsx`

The admin dashboard now includes a **"Students"** tab that contains all student attendance management features:

- **Scanner Tab**: QR code scanning for attendance
- **History Tab**: View all attendance records
- **Bulk Import Tab**: Upload CSV to add multiple students

**How to Access:**
1. Login as Admin
2. Navigate to Admin Dashboard
3. Click on "Students" tab (2nd tab)

### 2. Standalone Student Clock-In/Out Module

**Location:** `/components/StudentClockInOut.tsx`

A dedicated, simplified interface just for recording student attendance with:

- **QR Code Scanning**: Camera-based scanning with visual feedback
- **Manual Code Entry**: Type in student codes manually
- **Large Photo Display**: Shows student photo for 3 seconds after successful scan
- **Today's Summary**: Real-time statistics (clock-ins, clock-outs, currently in school)
- **Recent Scans**: Last 5 attendance records

**Features:**
- ✅ Toggle between Clock In / Clock Out modes
- ✅ Two input methods: Camera scan or manual entry
- ✅ Real-time validation and feedback
- ✅ Success sound on successful scan
- ✅ Large, centered student photo display (3 seconds)
- ✅ Live attendance statistics
- ✅ Recent scans sidebar

## 🎯 Integration Examples

### Option 1: Add to Main App Routes

```tsx
import { StudentClockInOut } from './components/StudentClockInOut';
import { StudentQRDisplay } from './components/StudentQRDisplay';

function App() {
  const [user, setUser] = useState(null);

  return (
    <Router>
      <Routes>
        {/* For Staff/Admin - Dedicated Clock Station */}
        <Route 
          path="/attendance/clock" 
          element={
            user?.role === 'admin' || user?.role === 'staff'
              ? <StudentClockInOut /> 
              : <Navigate to="/login" />
          } 
        />

        {/* For Students - View QR Code */}
        <Route 
          path="/student/qr-code" 
          element={
            user?.role === 'student'
              ? <StudentQRDisplay /> 
              : <Navigate to="/login" />
          } 
        />

        {/* For Admin - Full Management */}
        <Route 
          path="/admin" 
          element={
            user?.role === 'admin'
              ? <AdminDashboard /> 
              : <Navigate to="/login" />
          } 
        />
      </Routes>
    </Router>
  );
}
```

### Option 2: Add to Navigation Menu

```tsx
// In your sidebar/navigation component

{user?.role === 'admin' || user?.role === 'staff' ? (
  <nav>
    {/* Other menu items */}
    
    <NavLink to="/attendance/clock">
      <Clock className="w-5 h-5" />
      <span>Student Clock</span>
    </NavLink>
    
    {user?.role === 'admin' && (
      <NavLink to="/admin">
        <Users className="w-5 h-5" />
        <span>Admin Dashboard</span>
      </NavLink>
    )}
  </nav>
) : null}

{user?.role === 'student' && (
  <NavLink to="/student/qr-code">
    <QrCode className="w-5 h-5" />
    <span>My QR Code</span>
  </NavLink>
)}
```

### Option 3: Dedicated Attendance Station

Create a kiosk-style attendance station:

```tsx
// AttendanceStation.tsx
import { StudentClockInOut } from './components/StudentClockInOut';

export function AttendanceStation() {
  return (
    <div className="min-h-screen bg-gray-50">
      {/* Full-screen attendance clock */}
      <StudentClockInOut />
    </div>
  );
}

// In App.tsx
<Route path="/station" element={<AttendanceStation />} />
```

### Option 4: Staff Quick Access Dashboard

```tsx
function StaffDashboard() {
  const [activeView, setActiveView] = useState('overview');

  return (
    <div className="p-6">
      <div className="mb-6">
        <Button 
          onClick={() => setActiveView('clock')}
          size="lg"
          className="bg-green-600 hover:bg-green-700"
        >
          <Clock className="w-5 h-5 mr-2" />
          Open Attendance Clock
        </Button>
      </div>

      {activeView === 'clock' ? (
        <StudentClockInOut />
      ) : (
        // Your other staff dashboard content
        <div>...</div>
      )}
    </div>
  );
}
```

## 🔧 Component Props and Customization

### StudentClockInOut

No props required - fully self-contained.

**Customization Options:**

```tsx
// Modify the component to accept custom props
interface StudentClockInOutProps {
  defaultMode?: 'clock_in' | 'clock_out';
  defaultTab?: 'scan' | 'manual';
  onSuccess?: (student: Student) => void;
  showSummary?: boolean;
  showRecentScans?: boolean;
}

// Usage
<StudentClockInOut 
  defaultMode="clock_in"
  defaultTab="scan"
  onSuccess={(student) => console.log('Student scanned:', student)}
  showSummary={true}
  showRecentScans={true}
/>
```

### StudentQRDisplay

No props required - automatically fetches student's QR code.

**Features:**
- Auto-generated QR code
- Download QR code as PNG
- Regenerate QR code
- Usage statistics

### StudentAttendanceManagement (Admin)

Located in `/components/admin/StudentAttendanceManagement.tsx`

**Features:**
- Three tabs: Scanner, History, Bulk Import
- Full attendance management
- CSV bulk import
- Comprehensive reporting

## 📱 Use Cases

### Use Case 1: School Entrance/Exit Gate

**Setup:**
- Tablet/iPad at school entrance
- Runs StudentClockInOut component
- Staff mode: Clock In (morning)
- Staff mode: Clock Out (afternoon)

**URL:** `https://yourapp.com/attendance/clock`

### Use Case 2: Classroom Attendance

**Setup:**
- Teacher's device
- Quick scan students as they enter class
- Manual entry for students without QR codes

### Use Case 3: Student Self-Service

**Setup:**
- Students view/download their QR codes
- Print QR codes for ID cards
- Students can regenerate if lost

**URL:** `https://yourapp.com/student/qr-code`

### Use Case 4: Admin Management

**Setup:**
- Full access to all features
- Bulk student import
- Attendance history and reports
- QR code management

**URL:** `https://yourapp.com/admin` → Students tab

## 🎨 Styling and Branding

### Customize Colors

```tsx
// In StudentClockInOut.tsx, update these classes:

// Clock In button
<Button className="bg-green-600 hover:bg-green-700"> 
  Clock In
</Button>

// Clock Out button
<Button className="bg-blue-600 hover:bg-blue-700">
  Clock Out
</Button>

// Success photo border
<Card className="border-4 border-green-500">
```

### Customize Photo Display Duration

```tsx
// In StudentClockInOut.tsx, change timeout from 3000ms to desired duration

setTimeout(() => {
  setShowStudentPhoto(false);
  setLastScannedStudent(null);
}, 3000); // Change to 5000 for 5 seconds, etc.
```

### Add Custom Sound

```tsx
// Replace the success sound URL with your own audio file

const audio = new Audio('/sounds/success.mp3');
audio.play().catch(() => {});
```

## 📊 Data Flow

### Clock-In/Out Process

```
User selects Clock In/Clock Out mode
    ↓
Scans QR Code (or enters manually)
    ↓
Frontend → POST /api/student-attendance/process
    ↓
Backend verifies QR code
    ↓
Backend creates attendance record
    ↓
Backend sends SMS to parent (async)
    ↓
Frontend receives success response
    ↓
Frontend displays student photo (3 seconds)
    ↓
Frontend refreshes statistics
    ↓
Done
```

## 🔒 Security Considerations

### Access Control

```tsx
// Protect routes based on user role

const ProtectedRoute = ({ children, allowedRoles }) => {
  const user = useAuth();
  
  if (!user) {
    return <Navigate to="/login" />;
  }
  
  if (!allowedRoles.includes(user.role)) {
    return <Navigate to="/unauthorized" />;
  }
  
  return children;
};

// Usage
<Route path="/attendance/clock" element={
  <ProtectedRoute allowedRoles={['admin', 'staff']}>
    <StudentClockInOut />
  </ProtectedRoute>
} />
```

### API Authentication

All API calls use Bearer token authentication:

```tsx
fetch('/api/student-attendance/process', {
  headers: {
    'Authorization': `Bearer ${localStorage.getItem('token')}`
  }
})
```

## 🚀 Quick Start

### Step 1: Install Dependencies

```bash
npm install jsqr
```

### Step 2: Add Components to Your App

```tsx
import { StudentClockInOut } from './components/StudentClockInOut';
import { StudentQRDisplay } from './components/StudentQRDisplay';
import { AdminDashboard } from './components/AdminDashboard';
```

### Step 3: Add Routes

```tsx
<Route path="/attendance/clock" element={<StudentClockInOut />} />
<Route path="/student/qr-code" element={<StudentQRDisplay />} />
<Route path="/admin" element={<AdminDashboard />} />
```

### Step 4: Test the Flow

1. **As Admin:**
   - Go to Admin Dashboard
   - Click "Students" tab
   - Use "Bulk Import" to add students
   - Or manually add students

2. **As Student:**
   - Go to `/student/qr-code`
   - View/download QR code

3. **As Staff:**
   - Go to `/attendance/clock`
   - Select Clock In/Clock Out
   - Scan student QR code or enter manually
   - See student photo and confirmation

## 📝 API Endpoints Used

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/student-attendance/process` | POST | Record attendance |
| `/api/student-attendance/qr-code` | GET | Get student QR code |
| `/api/student-attendance/today-summary` | GET | Get today's stats |
| `/api/bulk-import/template` | GET | Download CSV template |
| `/api/bulk-import/students` | POST | Upload students |

## 💡 Tips

1. **For Best QR Scanning:**
   - Use good lighting
   - Hold phone steady
   - Keep QR code within frame
   - Use rear camera when possible

2. **For Kiosk Setup:**
   - Use tablet in portrait mode
   - Enable auto-lock prevention
   - Use fullscreen mode
   - Consider a stand or mount

3. **For Performance:**
   - QR scanning runs at 300ms intervals
   - Photo display is 3 seconds
   - Summary refreshes every 30 seconds
   - Camera stops when switching tabs

## 🐛 Troubleshooting

**Camera not working?**
- Check browser permissions
- Ensure HTTPS in production
- Try different browser
- Check if camera is already in use

**QR code not scanning?**
- Ensure good lighting
- Clean camera lens
- Check QR code quality
- Try manual entry instead

**Photo not displaying?**
- Check if student has profile_photo_url
- Verify image URL is accessible
- Check console for errors

---

**Ready to use!** All components are now integrated and ready for testing.
