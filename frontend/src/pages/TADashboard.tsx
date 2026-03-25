import { useAuth } from "../context/AuthContext";
import ufLogo from "../images/UF Logo.png";
import blueMessageIcon from "../images/Blue Message Icon.png";
import orangeClockIcon from "../images/Orange Clock Icon.png";
import orangeDateIcon from "../images/Orange Date Icon.png";
import orangeGroupIcon from "../images/Orange Group Icon.png";
import orangeQuestionIcon from "../images/Orange Question Mark Icon.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";
import { useState, useEffect } from 'react';
import { toast } from 'sonner';

interface QueueStudent {
  id: number;
  position: number;
  name: string;
  waitTime: string;
  joinedAt: Date;
  topic?: string;
}

interface OfficeHour {
  id: number;
  day: string;
  time: string;
  course: string;
}

export default function TADashboard() {
  const { user, logout } = useAuth();

  const [queueStatus, setQueueStatus] = useState<'closed' | 'open' | 'paused'>('closed');
  const [showAnnouncementModal, setShowAnnouncementModal] = useState(false);
  const [announcementText, setAnnouncementText] = useState('');
  const [queueStudents, setQueueStudents] = useState<QueueStudent[]>([
    { id: 1, position: 1, name: 'Sarah Johnson', waitTime: '28 min', joinedAt: new Date(Date.now() - 28 * 60000), topic: 'Linked Lists' },
    { id: 2, position: 2, name: 'Michael Chen', waitTime: '22 min', joinedAt: new Date(Date.now() - 22 * 60000), topic: 'Binary Trees' },
    { id: 3, position: 3, name: 'Emily Rodriguez', waitTime: '18 min', joinedAt: new Date(Date.now() - 18 * 60000), topic: 'Hash Tables' },
    { id: 4, position: 4, name: 'David Park', waitTime: '12 min', joinedAt: new Date(Date.now() - 12 * 60000), topic: 'Graph Algorithms' },
    { id: 5, position: 5, name: 'Jessica Williams', waitTime: '5 min', joinedAt: new Date(Date.now() - 5 * 60000), topic: 'Dynamic Programming' },
  ]);

  const todaySchedule = [
    { id: 1, time: "11:00 AM - 12:00 PM", course: "COP3530 - Data Structures" },
    { id: 2, time: "3:30 PM - 4:30 PM", course: "COP3530 - Data Structures" },
  ];

  const weeklyOfficeHours: OfficeHour[] = [
    { id: 1, day: "Monday", time: "11:00 AM - 12:00 PM", course: "COP3530 - Data Structures" },
    { id: 2, day: "Monday", time: "3:30 PM - 4:30 PM", course: "COP3530 - Data Structures" },
    { id: 3, day: "Wednesday", time: "2:00 PM - 3:00 PM", course: "COP3530 - Data Structures" },
    { id: 4, day: "Friday", time: "10:00 AM - 11:30 AM", course: "COP3530 - Data Structures" },
  ];

  const stats = {
    studentsHelped: 12,
    avgWaitTime: "8 min",
    currentQueueLength: queueStatus === 'closed' ? 0 : queueStudents.length,
    longestWaitTime: queueStatus === 'closed' ? '0 min' : '28 min',
    mostCommonTopic: "Data Structures",
    avgSessionDuration: "12 min",
  };

  const sortedQueue = [...queueStudents].sort((a, b) => a.joinedAt.getTime() - b.joinedAt.getTime());

  // Simulate students joining queue
  useEffect(() => {
    const interval = setInterval(() => {
      if (Math.random() > 0.7 && queueStatus === 'open') {
        const newStudent: QueueStudent = {
          id: Date.now(),
          position: queueStudents.length + 1,
          name: `Student ${Math.floor(Math.random() * 1000)}`,
          waitTime: '0 min',
          joinedAt: new Date(),
          topic: 'General Question',
        };
        setQueueStudents(prev => [...prev, newStudent]);
        toast.success(`${newStudent.name} joined the queue`);
      }
    }, 15000);
    return () => clearInterval(interval);
  }, [queueStudents.length, queueStatus]);

  const handleOpenQueue = () => {
    setQueueStatus('open');
    toast.success('Live Queue is now open - students can join!');
  };

  const handlePauseQueue = () => {
    setQueueStatus(queueStatus === 'paused' ? 'open' : 'paused');
    toast.info(queueStatus === 'paused' ? 'Queue reopened' : 'Queue paused');
  };

  const handleCloseQueue = () => {
    setQueueStatus('closed');
    setQueueStudents([]);
    toast.warning('Queue closed');
  };

  const handleStartSession = (student: QueueStudent) => {
    toast.success(`Starting session with ${student.name}`, {
      description: `Topic: ${student.topic}`,
    });
    setQueueStudents(prev => {
      const filtered = prev.filter(s => s.id !== student.id);
      return filtered.map((s, idx) => ({ ...s, position: idx + 1 }));
    });
  };

  const handleRemoveStudent = (student: QueueStudent) => {
    setQueueStudents(prev => {
      const filtered = prev.filter(s => s.id !== student.id);
      return filtered.map((s, idx) => ({ ...s, position: idx + 1 }));
    });
    toast.info(`${student.name} removed from queue`);
  };

  const handleSendAnnouncement = () => {
    if (announcementText.trim()) {
      toast.success('Announcement sent!', { description: announcementText });
      setAnnouncementText('');
      setShowAnnouncementModal(false);
    }
  };

  const officeHoursSidebar = (
    <div className="office-hours-list">
      {weeklyOfficeHours.map((hour) => (
        <div key={hour.id} className="office-hour-card">
          <div className="office-hour-day">
            <img src={orangeDateIcon} alt="Date" className="calendar-icon" />
            <span>{hour.day}</span>
          </div>
          <div className="office-hour-time">
            <img src={orangeClockIcon} alt="Time" className="time-icon" />
            <span>{hour.time}</span>
          </div>
          <div className="office-hour-course">{hour.course}</div>
          <div className="office-hour-actions">
            <button className="modify-btn">Modify</button>
            <button className="cancel-btn">Cancel</button>
          </div>
        </div>
      ))}
    </div>
  );

  return (
    <div className="ta-dashboard" style={{ minHeight: '100vh', width: '100%', display: 'block' }}>

      {/* Navbar */}
      <nav className="navbar">
        <div className="navbar-left">
          <div className="logo">
            <img src={ufLogo} alt="UF Logo" className="logo-icon" />
            <span className="logo-text">TA Connect</span>
          </div>
          <div className="nav-tabs">
            <button className="nav-tab active">Dashboard</button>
            <button className="nav-tab">My Office Hours</button>
            <button className="nav-tab">Queue</button>
          </div>
        </div>
        <div className="navbar-right">
          <div className="notification-icon">
            <img src={whiteNotificationIcon} alt="Notifications" />
          </div>
          <button className="profile-icon" onClick={logout}>
            <img src={whiteProfileIcon} alt="Profile" />
          </button>
        </div>
      </nav>

      {queueStatus === 'closed' ? (
        // CLOSED STATE - Dashboard view
        <div className="dashboard-content">
          <div className="main-section">

            {/* Welcome Section */}
            <div className="welcome-section">
              <h1 className="welcome-title">
                Welcome Back, {user?.username || "Lovely TA"}! Here is your Schedule for the Day
              </h1>
              <div className="schedule-cards">
                {todaySchedule.map((slot) => (
                  <div key={slot.id} className="schedule-card">
                    <div className="schedule-time">
                      <img src={orangeClockIcon} alt="Time" className="time-icon" />
                      <span className="time-text">{slot.time}</span>
                    </div>
                    <div className="schedule-course">{slot.course}</div>
                  </div>
                ))}
              </div>
              <button onClick={handleOpenQueue} className="start-queue-btn">
                <span className="play-icon">▶</span>
                Start Office Hours Live Queue
              </button>
            </div>

            {/* Stats Grid */}
            <div className="stats-grid">
              <div className="stat-card blue">
                <div className="stat-header">
                  <img src={orangeGroupIcon} alt="Students" className="stat-icon" />
                  <span className="stat-label">Students Helped Today</span>
                </div>
                <div className="stat-value">{stats.studentsHelped}</div>
              </div>
              <div className="stat-card orange">
                <div className="stat-header">
                  <img src={orangeClockIcon} alt="Clock" className="stat-icon" />
                  <span className="stat-label">Avg Wait Time</span>
                </div>
                <div className="stat-value">{stats.avgWaitTime}</div>
              </div>
              <div className="stat-card green">
                <div className="stat-header">
                  <img src={orangeGroupIcon} alt="Queue" className="stat-icon" />
                  <span className="stat-label">Current Queue Length</span>
                </div>
                <div className="stat-value">{stats.currentQueueLength}</div>
              </div>
              <div className="stat-card purple">
                <div className="stat-header">
                  <img src={orangeClockIcon} alt="Time" className="stat-icon" />
                  <span className="stat-label">Longest Wait Time</span>
                </div>
                <div className="stat-value">{stats.longestWaitTime}</div>
              </div>
              <div className="stat-card yellow">
                <div className="stat-header">
                  <img src={orangeQuestionIcon} alt="Topic" className="stat-icon" />
                  <span className="stat-label">Most Common Topic</span>
                </div>
                <div className="stat-value topic">{stats.mostCommonTopic}</div>
              </div>
              <div className="stat-card light-purple">
                <div className="stat-header">
                  <img src={orangeClockIcon} alt="Duration" className="stat-icon" />
                  <span className="stat-label">Avg Session Duration</span>
                </div>
                <div className="stat-value">{stats.avgSessionDuration}</div>
              </div>
            </div>

            {/* Announcements */}
            <div className="announcements-section">
              <div className="announcements-content">
                <div className="announcements-text">
                  <img src={blueMessageIcon} alt="Announcements" className="announcement-icon" />
                  <div>
                    <h3>Send Announcements to Your Students</h3>
                    <p>Broadcast messages to all students enrolled in your courses</p>
                  </div>
                </div>
                <button className="send-announcement-btn" onClick={() => setShowAnnouncementModal(true)}>
                  Send Announcement
                </button>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          <aside className="sidebar">
            <div className="sidebar-content">
              <h2 className="sidebar-title">This Week's Office Hours</h2>
              {officeHoursSidebar}
            </div>
          </aside>
        </div>
      ) : (
        // OPEN/PAUSED STATE - Live queue view
        <div className="dashboard-content">
          <div className="main-section">

            {/* Queue Header */}
            <div className="welcome-section" style={{ padding: '1.5rem' }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem' }}>
                <div>
                  <h1 className="welcome-title" style={{ marginBottom: '0.5rem' }}>Live Queue</h1>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                    <div style={{
                      width: 10, height: 10, borderRadius: '50%',
                      backgroundColor: queueStatus === 'open' ? 'var(--gator)' : 'var(--alachua)',
                      flexShrink: 0
                    }} />
                    <span style={{ color: 'var(--white)', fontFamily: 'IBM Plex Sans, sans-serif', fontSize: '0.95rem' }}>
                      {queueStatus === 'open' ? 'Open - Accepting Students' : 'Paused - Not Accepting New Students'}
                    </span>
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '0.75rem' }}>
                  <button
                    onClick={handlePauseQueue}
                    style={{
                      backgroundColor: queueStatus === 'paused' ? 'var(--gator)' : 'var(--alachua)',
                      color: queueStatus === 'paused' ? 'var(--white)' : 'var(--black)',
                      border: 'none', borderRadius: 8,
                      padding: '0.625rem 1.25rem',
                      fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif', cursor: 'pointer',
                    }}
                  >
                    {queueStatus === 'paused' ? 'Resume Queue' : 'Pause Queue'}
                  </button>
                  <button
                    onClick={handleCloseQueue}
                    style={{
                      backgroundColor: 'var(--bottlebrush)', color: 'var(--white)',
                      border: 'none', borderRadius: 8,
                      padding: '0.625rem 1.25rem',
                      fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif', cursor: 'pointer',
                    }}
                  >
                    Close Queue
                  </button>
                  <button
                    className="send-announcement-btn"
                    onClick={() => setShowAnnouncementModal(true)}
                    style={{ padding: '0.625rem 1.25rem', fontSize: '1rem' }}
                  >
                    Send Announcement
                  </button>
                </div>
              </div>
            </div>

            {/* Queue Stats */}
            <div className="stats-grid" style={{ gridTemplateColumns: 'repeat(3, 1fr)' }}>
              <div className="stat-card blue">
                <div className="stat-header">
                  <img src={orangeGroupIcon} alt="Students" className="stat-icon" />
                  <span className="stat-label">Total in Queue</span>
                </div>
                <div className="stat-value">{queueStudents.length}</div>
              </div>
              <div className="stat-card orange">
                <div className="stat-header">
                  <img src={orangeClockIcon} alt="Clock" className="stat-icon" />
                  <span className="stat-label">Longest Wait</span>
                </div>
                <div className="stat-value">{sortedQueue.length > 0 ? sortedQueue[0].waitTime : '0 min'}</div>
              </div>
              <div className="stat-card green">
                <div className="stat-header">
                  <img src={orangeClockIcon} alt="Avg" className="stat-icon" />
                  <span className="stat-label">Avg Wait Time</span>
                </div>
                <div className="stat-value">{stats.avgWaitTime}</div>
              </div>
            </div>

            {/* Queue List */}
            <div className="sidebar-content" style={{ borderRadius: 12 }}>
              <h2 className="sidebar-title">Students in Queue ({sortedQueue.length})</h2>
              {sortedQueue.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '3rem 0', color: 'var(--cool-grey-3)' }}>
                  <p style={{ fontSize: '1.25rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>No students in queue</p>
                  <p style={{ fontSize: '0.875rem', marginTop: '0.5rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                    Students will appear here when they join
                  </p>
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                  {sortedQueue.map((student, index) => (
                    <div key={student.id} className="office-hour-card">
                      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '1rem', flexWrap: 'wrap' }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                          <div style={{
                            width: 48, height: 48, borderRadius: '50%',
                            backgroundColor: 'var(--core-blue)', color: 'var(--white)',
                            display: 'flex', alignItems: 'center', justifyContent: 'center',
                            fontWeight: 700, fontSize: '1.25rem', flexShrink: 0,
                            fontFamily: 'Anybody, IBM Plex Sans, sans-serif',
                          }}>
                            {index + 1}
                          </div>
                          <div>
                            <div style={{ fontWeight: 600, fontSize: '1rem', color: 'var(--cool-grey-11)', fontFamily: 'IBM Plex Sans, sans-serif', marginBottom: '0.25rem' }}>
                              {student.name}
                            </div>
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                              <span style={{ fontSize: '0.875rem', color: 'var(--cool-grey-11)', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                                Waiting {student.waitTime}
                              </span>
                              {student.topic && (
                                <>
                                  <span style={{ color: 'var(--cool-grey-3)' }}>•</span>
                                  <span style={{ fontSize: '0.875rem', color: 'var(--core-orange)', fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif' }}>
                                    {student.topic}
                                  </span>
                                </>
                              )}
                            </div>
                          </div>
                        </div>
                        <div className="office-hour-actions" style={{ flexWrap: 'nowrap', width: 'auto' }}>
                          {index === 0 && (
                            <button
                              onClick={() => handleStartSession(student)}
                              className="start-queue-btn"
                              style={{ width: 'auto', padding: '0.625rem 1.5rem', fontSize: '0.95rem' }}
                            >
                              Start Session
                            </button>
                          )}
                          <button
                            onClick={() => handleRemoveStudent(student)}
                            className="cancel-btn"
                            style={{ border: '2px solid var(--bottlebrush)', borderRadius: 8, padding: '0.625rem 1rem' }}
                          >
                            Remove
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Sidebar */}
          <aside className="sidebar">
            <div className="sidebar-content" style={{ marginBottom: '1.5rem' }}>
              <h2 className="sidebar-title">This Week's Office Hours</h2>
              {officeHoursSidebar}
            </div>
            <div className="announcements-section" style={{ borderRadius: 12 }}>
              <div style={{ marginBottom: '1rem' }}>
                <h3 style={{ fontFamily: 'IBM Plex Sans, sans-serif', fontWeight: 600, fontSize: '1.125rem', color: 'var(--white)' }}>
                  Announcements
                </h3>
                <p style={{ fontSize: '0.875rem', color: 'var(--white)', opacity: 0.9, marginTop: '0.25rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                  Send updates to all students in the queue
                </p>
              </div>
              <button className="send-announcement-btn" onClick={() => setShowAnnouncementModal(true)}
                style={{ width: '100%' }}>
                Send Announcement
              </button>
            </div>
          </aside>
        </div>
      )}

      {/* Announcement Modal */}
      {showAnnouncementModal && (
        <div style={{
          position: 'fixed', inset: 0, backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50,
        }}>
          <div style={{ background: 'var(--white)', borderRadius: 12, maxWidth: 480, width: '100%', margin: '0 1rem', overflow: 'hidden', boxShadow: '0 20px 40px rgba(0,0,0,0.3)' }}>
            <div style={{
              background: 'linear-gradient(135deg, var(--core-blue) 0%, var(--dark-blue) 100%)',
              padding: '1rem 1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            }}>
              <h3 style={{ color: 'var(--white)', fontSize: '1.25rem', fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif' }}>
                Send Announcement
              </h3>
              <button onClick={() => setShowAnnouncementModal(false)}
                style={{ background: 'none', border: 'none', color: 'var(--white)', cursor: 'pointer', fontSize: '1.25rem' }}>
                ✕
              </button>
            </div>
            <div style={{ padding: '1.5rem' }}>
              <p style={{ color: 'var(--cool-grey-11)', marginBottom: '1rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                {queueStatus === 'closed'
                  ? 'Send a message to all students in your course'
                  : `Send a message to all ${queueStudents.length} students currently in the queue`}
              </p>
              <textarea
                value={announcementText}
                onChange={(e) => setAnnouncementText(e.target.value)}
                placeholder="e.g., Running 10 minutes late, extending office hours by 30 minutes..."
                style={{
                  width: '100%', border: '1px solid var(--cool-grey-3)', borderRadius: 8,
                  padding: '0.75rem', minHeight: 120, fontFamily: 'IBM Plex Sans, sans-serif',
                  fontSize: '0.95rem', color: 'var(--cool-grey-11)', resize: 'vertical', boxSizing: 'border-box',
                }}
              />
              <div style={{ background: '#e3f2fd', border: '1px solid #bbdefb', borderRadius: 8, padding: '0.75rem', marginTop: '1rem' }}>
                <p style={{ fontSize: '0.875rem', fontWeight: 600, color: 'var(--cool-grey-11)', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                  Example announcements:
                </p>
                <ul style={{ fontSize: '0.875rem', color: 'var(--cool-grey-11)', marginTop: '0.5rem', paddingLeft: '1.25rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                  <li>Running 10 minutes late</li>
                  <li>Office hours canceled today</li>
                  <li>Extending office hours by 30 minutes</li>
                </ul>
              </div>
              <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem' }}>
                <button onClick={() => setShowAnnouncementModal(false)}
                  style={{
                    flex: 1, padding: '0.75rem', border: '1px solid var(--cool-grey-3)',
                    borderRadius: 8, fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif',
                    cursor: 'pointer', background: 'var(--white)', color: 'var(--cool-grey-11)',
                  }}>
                  Cancel
                </button>
                <button onClick={handleSendAnnouncement} className="start-queue-btn"
                  style={{ flex: 1, padding: '0.75rem', fontSize: '1rem' }}>
                  Send to All
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}