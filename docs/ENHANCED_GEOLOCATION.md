# Enhanced Geolocation System - Technical Documentation

## Masalah Yang Dipecahkan

### Problem Statement
Karyawan yang sudah berada di kantor sering **ditolak** saat attendance karena GPS inaccuracy, meskipun secara fisik mereka sudah di lokasi yang benar.

### Root Causes
1. **GPS Accuracy Varies** - GPS bisa memiliki error 5-50+ meter tergantung kondisi
2. **Indoor GPS Degradation** - Signal GPS lemah di dalam gedung
3. **Urban Canyon Effect** - Gedung tinggi menyebabkan GPS multipath
4. **Fixed Threshold Problem** - System lama menggunakan radius fixed (mis: 200m) tanpa consider GPS accuracy

## Solusi: Adaptive GPS-Aware Threshold

### Core Innovation: Dynamic Threshold Calculation

```
Adaptive Threshold = Base Radius + (GPS Accuracy × Multiplier)
```

**Multiplier Table:**
| GPS Accuracy | Quality    | Multiplier | Example Calculation              |
|--------------|------------|------------|----------------------------------|
| < 10m        | Excellent  | 0.5        | 200m + (8m × 0.5) = 204m        |
| 10-25m       | Good       | 1.0        | 200m + (15m × 1.0) = 215m       |
| 25-50m       | Fair       | 1.5        | 200m + (40m × 1.5) = 260m       |
| > 50m        | Poor       | 2.0        | 200m + (60m × 2.0) = 320m       |

**Safety Cap:** Maximum threshold = Base Radius × 3

### How It Works

#### Scenario 1: Excellent GPS (Outdoor, Clear Sky)
```
Employee Location: 195m from office
GPS Accuracy: 8m (excellent)
Base Radius: 200m
Adaptive Threshold: 200m + (8m × 0.5) = 204m
Result: ✅ APPROVED (195m < 204m)
```

#### Scenario 2: Poor GPS (Indoor, Weak Signal)
```
Employee Location: 220m from office (GPS drift)
GPS Accuracy: 45m (fair)
Base Radius: 200m
Adaptive Threshold: 200m + (45m × 1.5) = 267.5m
Result: ✅ APPROVED (220m < 267.5m)
```

#### Scenario 3: Too Far (Actual Out of Range)
```
Employee Location: 400m from office
GPS Accuracy: 15m (good)
Base Radius: 200m
Adaptive Threshold: 200m + (15m × 1.0) = 215m
Result: ❌ REJECTED (400m > 215m) - Legitimately out of range
```

## Backend Implementation

### 1. Enhanced Location Utils (`pkg/utils/location.go`)

#### New Functions

**`IsWithinOfficeRangeAdaptive(lat, lon, accuracy float64)`**
- Returns: `(inRange bool, distance float64, adaptiveRadius float64)`
- Replaces old `IsWithinOfficeRange` with GPS-aware checking
- Auto-calculates optimal threshold based on GPS quality

**`ValidateLocationAdaptive(lat, lon, accuracy float64)`**
- Returns: `LocationValidationResponse` with rich metadata
- Provides detailed feedback including:
  - GPS quality assessment ("excellent", "good", "fair", "poor")
  - User-friendly recommendations
  - Detailed distance breakdown
  - Auto-approval suggestions for edge cases

**`AdaptiveThreshold(baseRadius, gpsAccuracy float64)`**
- Core algorithm for dynamic threshold calculation
- Implements multiplier logic based on GPS accuracy
- Includes safety caps to prevent abuse

**`GetGPSQuality(accuracy float64)`**
- Maps accuracy values to human-readable quality levels
- Used for UI feedback and logging

#### Auto-Approval for Edge Cases

```go
overagePercentage := ((distance - adaptiveRadius) / adaptiveRadius) * 100

if overagePercentage < 10 {
    // Very close - likely GPS error
    inRange = true  // Auto-approve!
    message = "Kemungkinan error GPS signal, attendance di-approve otomatis"
}
```

### 2. Updated Request Models

```go
type ClockInRequest struct {
    Latitude    float64 `json:"latitude"`
    Longitude   float64 `json:"longitude"`
    PhotoSelfie string  `json:"photo_selfie"`
    Force       bool    `json:"force"`
    Accuracy    float64 `json:"accuracy"`  // NEW: GPS accuracy in meters
}
```

### 3. Enhanced API Responses

**Check Location Response:**
```json
{
  "status": "success",
  "data": {
    "in_range": true,
    "message": "Lokasi valid, dalam jangkauan kantor",
    "detailed_message": "Jarak: 195m | Jangkauan: 215m | Akurasi GPS: good (15m)",
    "need_force": false,
    "distance": 195.2,
    "max_radius": 200,
    "adaptive_radius": 215,
    "gps_accuracy": 15.0,
    "gps_quality": "good",
    "recommendation": ""
  }
}
```

**Clock-In with Poor GPS:**
```json
{
  "status": "success",
  "data": {
    "in_range": true,
    "message": "Lokasi valid, dalam jangkauan kantor",
    "detailed_message": "Jarak: 220m | Jangkauan: 268m | Akurasi GPS: fair (45m)",
    "need_force": false,
    "distance": 220.5,
    "max_radius": 200,
    "adaptive_radius": 267.5,
    "gps_accuracy": 45.0,
    "gps_quality": "fair",
    "recommendation": "Untuk akurasi lebih baik, coba pindah ke area dengan sinyal GPS lebih kuat (dekat jendela atau outdoor)"
  }
}
```

## Frontend Integration Guide

### 1. Get GPS Position with Accuracy

```javascript
const getLocationWithAccuracy = () => {
  return new Promise((resolve, reject) => {
    navigator.geolocation.getCurrentPosition(
      (position) => {
        resolve({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
          accuracy: position.coords.accuracy, // Key field!
          altitude: position.coords.altitude,
          heading: position.coords.heading,
          speed: position.coords.speed,
        });
      },
      (error) => reject(error),
      {
        enableHighAccuracy: true,  // Request best accuracy
        timeout: 10000,
        maximumAge: 0  // Don't use cached position
      }
    );
  });
};
```

### 2. Check Location Before Clock-In

```javascript
const checkLocation = async () => {
  try {
    const location = await getLocationWithAccuracy();
    
    const response = await fetch('/api/attendance/check-location', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        latitude: location.latitude,
        longitude: location.longitude,
        accuracy: location.accuracy,  // Send GPS accuracy!
      }),
    });

    const result = await response.json();
    
    // Show feedback to user
    if (result.data.in_range) {
      showSuccess(result.data.message);
      if (result.data.recommendation) {
        showInfo(result.data.recommendation);
      }
    } else {
      showWarning(result.data.message);
      showDetails(result.data.detailed_message);
      if (result.data.need_force) {
        showForceButton();
      }
    }
    
    return result.data;
  } catch (error) {
    console.error('Location check failed:', error);
    throw error;
  }
};
```

### 3. Clock-In with GPS Accuracy

```javascript
const clockIn = async (photoSelfie, force = false) => {
  try {
    const location = await getLocationWithAccuracy();
    
    const response = await fetch('/api/attendance/clock-in', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        latitude: location.latitude,
        longitude: location.longitude,
        accuracy: location.accuracy,  // Important!
        photo_selfie: photoSelfie,
        force: force,
      }),
    });

    const result = await response.json();
    
    if (response.ok) {
      showSuccess('Clock-in successful!');
      // Show GPS quality feedback
      console.log(`GPS Quality: ${result.data.gps_quality}`);
      console.log(`Distance: ${result.data.distance}m`);
    } else {
      showError(result.message);
    }
    
    return result;
  } catch (error) {
    console.error('Clock-in failed:', error);
    throw error;
  }
};
```

### 4. UI Feedback for GPS Quality

```javascript
const renderGPSQualityIndicator = (quality, accuracy) => {
  const qualityConfig = {
    excellent: { color: 'green', icon: '✅', text: 'Sinyal GPS Sangat Baik' },
    good: { color: 'blue', icon: '👍', text: 'Sinyal GPS Baik' },
    fair: { color: 'orange', icon: '⚠️', text: 'Sinyal GPS Cukup' },
    poor: { color: 'red', icon: '❌', text: 'Sinyal GPS Lemah' },
  };
  
  const config = qualityConfig[quality] || qualityConfig.poor;
  
  return `
    <div class="gps-indicator" style="color: ${config.color}">
      <span class="icon">${config.icon}</span>
      <span class="text">${config.text}</span>
      <span class="accuracy">(${accuracy.toFixed(0)}m accuracy)</span>
    </div>
  `;
};
```

## Configuration

### Environment Variables

```env
# Office location
OFFICE_LATITUDE=-6.2088
OFFICE_LONGITUDE=106.8456

# Base radius in meters (sebelum adaptive adjustment)
ATTENDANCE_RADIUS_METERS=200

# Enable/disable location checking
ENABLE_LOCATION_CHECK=true
```

### Recommended Base Radius Settings

| Office Type | Recommended Radius | Reasoning |
|-------------|-------|-----------|
| Small office (< 1000m²) | 100-150m | Limited area, more precision needed |
| Medium office (1000-5000m²) | 150-250m | Balance between accuracy and tolerance |
| Large campus (> 5000m²) | 250-500m | Wide area, more flexibility needed |
| Multiple buildings | 300-1000m | Large perimeter, high flexibility |

**Note:** Adaptive system will auto-adjust based on GPS quality, so you can use tighter base radius.

## Testing Guide

### Test Scenarios

#### 1. **Excellent GPS - Outdoor**
```bash
curl -X POST http://localhost:8080/api/attendance/check-location \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": -6.2088,
    "longitude": 106.8456,
    "accuracy": 8
  }'
```
Expected: Approved with excellent GPS quality

#### 2. **Poor GPS - Indoor**
```bash
curl -X POST http://localhost:8080/api/attendance/check-location \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": -6.2092,
    "longitude": 106.8460,
    "accuracy": 60
  }'
```
Expected: Approved with poor GPS quality, larger adaptive radius

#### 3. **Edge Case - Just Outside Range**
```bash
curl -X POST http://localhost:8080/api/attendance/check-location \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": -6.2105,
    "longitude": 106.8456,
    "accuracy": 15
  }'
```
Expected: May auto-approve if within 10% overage

## Monitoring & Logging

### Backend Logs

```
📍 [GPS] Excellent accuracy (8.0m) - threshold: 204.0m
📍 [Location] Distance: 195.2m | Base Radius: 200.0m | Adaptive Radius: 204.0m | In Range: true

⚠️ [GPS] Fair accuracy (45.0m) - threshold: 267.5m
📍 [Location] Distance: 220.5m | Base Radius: 200.0m | Adaptive Radius: 267.5m | In Range: true

⚠️ [GPS] Poor accuracy (75.0m) - threshold: 350.0m
⚠️ [GPS] Threshold capped from 350.0m to 600.0m
📍 [Location] Distance: 185.5m | Base Radius: 200.0m | Adaptive Radius: 600.0m | In Range: true
```

## Benefits Summary

✅ **Reduced False Rejections**: Employees dengan GPS lemah tidak ditolak
✅ **Better User Experience**: Clear feedback tentang GPS quality
✅ **Maintained Security**: Still reject legitimate out-of-range attempts
✅ **Adaptive**: Auto-adjusts untuk berbagai kondisi GPS
✅ **Transparent**: User dapat lihat GPS quality mereka
✅ **Intelligent**: Auto-approval untuk edge cases yang jelas GPS error

## Migration Notes

### Backward Compatibility

- Old endpoints tetap berfungsi (tanpa accuracy field)
- Jika `accuracy` not provided, system assumes moderate accuracy (30m)
- Frontend dapat gradual update untuk include accuracy

### Rollout Strategy

1. **Phase 1**: Deploy backend changes
2. **Phase 2**: Update frontend untuk send GPS accuracy
3. **Phase 3**: Monitor logs untuk fine-tune thresholds
4. **Phase 4**: Adjust base radius if needed based on data

## Troubleshooting

### Issue: Masih banyak false rejections

**Solution:**
1. Check logs untuk GPS accuracy distribution
2. Jika umumnya poor (>50m), pertimbangkan increase base radius
3. Verify frontend mengirim accuracy correctly

### Issue: Too many approvals from far away

**Solution:**
1. Check cap setting (max 3× base radius)
2. Consider reducing multipliers in `AdaptiveThreshold`
3. Add additional security checks (geofencing, IP, etc)

### Issue: GPS accuracy selalu 0

**Solution:**
1. Frontend tidak mengirim accuracy field
2. Browser tidak provide accuracy data
3. Implement fallback estimation based on device/context
