package station

import (
	"sync"
	"time"
)

// StationMetrics holds various performance and operational metrics
type StationMetrics struct {
	// Announcement metrics
	AnnouncementsSent     uint64    `json:"announcements_sent"`
	AnnouncementsReceived uint64    `json:"announcements_received"`
	LastAnnouncementSent  time.Time `json:"last_announcement_sent"`
	LastAnnouncementRecv  time.Time `json:"last_announcement_received"`

	// Message metrics
	MessagesSent          uint64    `json:"messages_sent"`
	MessagesReceived      uint64    `json:"messages_received"`
	MessagesSentBytes     uint64    `json:"messages_sent_bytes"`
	MessagesReceivedBytes uint64    `json:"messages_received_bytes"`
	LastMessageSent       time.Time `json:"last_message_sent"`
	LastMessageReceived   time.Time `json:"last_message_received"`

	// Error metrics
	ErrorCount uint64    `json:"error_count"`
	LastError  time.Time `json:"last_error"`
	ErrorRate  float64   `json:"error_rate"` // errors per minute

	// Device metrics
	DeviceCount      int    `json:"device_count"`
	ActiveDevices    int    `json:"active_devices"`
	DeviceErrorCount uint64 `json:"device_error_count"`

	// Network metrics
	NetworkInterfaceCount int `json:"network_interface_count"`
	IPAddressCount        int `json:"ip_address_count"`

	// Timing metrics
	StartTime       time.Time     `json:"start_time"`
	Uptime          time.Duration `json:"uptime"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	MaxResponseTime time.Duration `json:"max_response_time"`
	MinResponseTime time.Duration `json:"min_response_time"`

	// Health metrics
	HealthCheckCount uint64  `json:"health_check_count"`
	HealthyChecks    uint64  `json:"healthy_checks"`
	UnhealthyChecks  uint64  `json:"unhealthy_checks"`
	HealthScore      float64 `json:"health_score"` // percentage

	// Resource metrics
	MemoryUsage uint64  `json:"memory_usage_bytes"`
	CPUUsage    float64 `json:"cpu_usage_percent"`
	DiskUsage   uint64  `json:"disk_usage_bytes"`

	mu sync.RWMutex `json:"-"`
}

// NewStationMetrics creates a new metrics instance
func NewStationMetrics() *StationMetrics {
	return &StationMetrics{
		StartTime:       time.Now(),
		MinResponseTime: time.Hour, // Initialize to high value
	}
}

// UpdateMetrics updates the station's uptime and other time-based metrics
func (m *StationMetrics) UpdateMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Uptime = time.Since(m.StartTime)

	// Calculate error rate (errors per minute)
	if m.Uptime > 0 {
		minutes := m.Uptime.Minutes()
		if minutes > 0 {
			m.ErrorRate = float64(m.ErrorCount) / minutes
		}
	}

	// Calculate health score
	if m.HealthCheckCount > 0 {
		m.HealthScore = (float64(m.HealthyChecks) / float64(m.HealthCheckCount)) * 100.0
	}
}

// RecordAnnouncement records an outgoing announcement
func (m *StationMetrics) RecordAnnouncement() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.AnnouncementsSent++
	m.LastAnnouncementSent = time.Now()
}

// RecordAnnouncementReceived records an incoming announcement
func (m *StationMetrics) RecordAnnouncementReceived() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.AnnouncementsReceived++
	m.LastAnnouncementRecv = time.Now()
}

// RecordMessageSent records an outgoing message
func (m *StationMetrics) RecordMessageSent(size uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.MessagesSent++
	m.MessagesSentBytes += size
	m.LastMessageSent = time.Now()
}

// RecordMessageReceived records an incoming message
func (m *StationMetrics) RecordMessageReceived(size uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.MessagesReceived++
	m.MessagesReceivedBytes += size
	m.LastMessageReceived = time.Now()
}

// RecordError records an error occurrence
func (m *StationMetrics) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ErrorCount++
	m.LastError = time.Now()
}

// RecordResponseTime records a response time measurement
func (m *StationMetrics) RecordResponseTime(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update min/max response times
	if duration < m.MinResponseTime {
		m.MinResponseTime = duration
	}
	if duration > m.MaxResponseTime {
		m.MaxResponseTime = duration
	}

	// Calculate running average (exponential moving average for simplicity)
	if m.AvgResponseTime == 0 {
		m.AvgResponseTime = duration
	} else {
		m.AvgResponseTime = (m.AvgResponseTime + duration) / 2
	}
}

// RecordHealthCheck records a health check result
func (m *StationMetrics) RecordHealthCheck(healthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HealthCheckCount++
	if healthy {
		m.HealthyChecks++
	} else {
		m.UnhealthyChecks++
	}
}

// UpdateDeviceMetrics updates device-related metrics
func (m *StationMetrics) UpdateDeviceMetrics(deviceCount, activeDevices int, deviceErrors uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DeviceCount = deviceCount
	m.ActiveDevices = activeDevices
	m.DeviceErrorCount = deviceErrors
}

// UpdateNetworkMetrics updates network-related metrics
func (m *StationMetrics) UpdateNetworkMetrics(interfaceCount, ipCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.NetworkInterfaceCount = interfaceCount
	m.IPAddressCount = ipCount
}

// GetMetrics returns a copy of current metrics
func (m *StationMetrics) GetMetrics() StationMetrics {
	// Update dynamic metrics before returning
	m.UpdateMetrics()

	m.mu.RLock()
	defer m.mu.RUnlock()

	// return a copy
	return StationMetrics{
		AnnouncementsSent:     m.AnnouncementsSent,
		AnnouncementsReceived: m.AnnouncementsReceived,
		LastAnnouncementSent:  m.LastAnnouncementSent,
		LastAnnouncementRecv:  m.LastAnnouncementRecv,
		MessagesSent:          m.MessagesSent,
		MessagesReceived:      m.MessagesReceived,
		MessagesSentBytes:     m.MessagesSentBytes,
		MessagesReceivedBytes: m.MessagesReceivedBytes,
		LastMessageSent:       m.LastMessageSent,
		LastMessageReceived:   m.LastMessageReceived,
		ErrorCount:            m.ErrorCount,
		LastError:             m.LastError,
		ErrorRate:             m.ErrorRate,
		DeviceCount:           m.DeviceCount,
		ActiveDevices:         m.ActiveDevices,
		DeviceErrorCount:      m.DeviceErrorCount,
		NetworkInterfaceCount: m.NetworkInterfaceCount,
		IPAddressCount:        m.IPAddressCount,
		StartTime:             m.StartTime,
		Uptime:                m.Uptime,
		AvgResponseTime:       m.AvgResponseTime,
		MaxResponseTime:       m.MaxResponseTime,
		MinResponseTime:       m.MinResponseTime,
		HealthCheckCount:      m.HealthCheckCount,
		HealthyChecks:         m.HealthyChecks,
		UnhealthyChecks:       m.UnhealthyChecks,
		HealthScore:           m.HealthScore,
		MemoryUsage:           m.MemoryUsage,
		CPUUsage:              m.CPUUsage,
		DiskUsage:             m.DiskUsage,
	}
}
