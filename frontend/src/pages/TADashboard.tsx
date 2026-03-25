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
import { clearActiveQueueForCourse, clearActiveQueueForOfficeHour, createQueue, getActiveQueueByCourse, getQueueOrNull, nextQueueStudent, setActiveQueueForCourse, setActiveQueueForOfficeHour, subscribeToQueueEvents, updateQueueStatus } from "../api/queue";

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
  courseCode: string;
  courseID: number;
}

export default function TADashboard() {
  const { user, logout } = useAuth();

  const [queueStatus, setQueueStatus] = useState<'closed' | 'open' | 'paused'>('closed');
  const [showAnnouncementModal, setShowAnnouncementModal] = useState(false);
  const [announcementText, setAnnouncementText] = useState('');
  const [queueStudents, setQueueStudents] = useState<QueueStudent[]>([]);
  const [activeQueueID, setActiveQueueID] = useState<number | null>(null);
  const [activeCourseCode, setActiveCourseCode] = useState<string | null>(null);
  const [activeOfficeHourTime, setActiveOfficeHourTime] = useState<string | null>(null);
  const [selectedOfficeHourID, setSelectedOfficeHourID] = useState<number | null>(null);

  // TA dashboard currently operates on one selected office hour at a time.

  const getWaitMinutes = (joinedAt: Date): number => {
    return Math.max(0, Math.floor((Date.now() - joinedAt.getTime()) / 60000));
  };

  const formatWait = (minutes: number): string => `${minutes} min`;

  const mapQueueEntriesToStudents = (entries: Array<{
    id: number;
    position: number;
    username: string;
    joined_at: string;
  }>): QueueStudent[] => {
    return entries.map((entry) => {
      const joinedAt = new Date(entry.joined_at);
      return {
        id: entry.id,
        position: entry.position,
        name: entry.username,
        joinedAt,
        waitTime: formatWait(getWaitMinutes(joinedAt)),
      };
    });
  };

  const todaySchedule: OfficeHour[] = [
    { id: 1, day: "Today", time: "11:00 AM - 1:00 PM", course: "COP3530 - Data Structures", courseCode: "COP3530", courseID: 2 },
    { id: 2, day: "Today", time: "3:30 PM - 4:30 PM", course: "COP3530 - Data Structures", courseCode: "COP3530", courseID: 2 },
  ];

  const weeklyOfficeHours: OfficeHour[] = [
    { id: 1, day: "Monday", time: "11:00 AM - 1:00 PM", course: "COP3530 - Data Structures", courseCode: "COP3530", courseID: 2 },
    { id: 2, day: "Monday", time: "3:30 PM - 4:30 PM", course: "COP3530 - Data Structures", courseCode: "COP3530", courseID: 2 },
    { id: 3, day: "Wednesday", time: "2:00 PM - 3:00 PM", course: "COP3530 - Data Structures", courseCode: "COP3530", courseID: 2 },
    { id: 4, day: "Friday", time: "10:00 AM - 11:30 AM", course: "COP3530 - Data Structures", courseCode: "COP3530", courseID: 2 },
  ];

  const selectedOfficeHour = todaySchedule.find((slot) => slot.id === selectedOfficeHourID) ?? null;

  const refreshQueueData = async () => {
    if (!activeQueueID) {
      setQueueStudents([]);
      return;
    }
    try {
      const queue = await getQueueOrNull(activeQueueID);
      if (!queue) {
        setQueueStudents([]);
        return;
      }
      setQueueStudents(mapQueueEntriesToStudents(queue.entries));
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to load queue';
      toast.info(message);
    }
  };

  const stats = {
    studentsHelped: 12,
    avgWaitTime: queueStudents.length === 0
      ? '0 min'
      : formatWait(
          Math.round(
            queueStudents.reduce((sum, s) => sum + getWaitMinutes(s.joinedAt), 0) / queueStudents.length
          )
        ),
    currentQueueLength: queueStatus === 'closed' ? 0 : queueStudents.length,
    longestWaitTime: queueStatus === 'closed' || queueStudents.length === 0
      ? '0 min'
      : formatWait(Math.max(...queueStudents.map((s) => getWaitMinutes(s.joinedAt)))),
    mostCommonTopic: "Data Structures",
    avgSessionDuration: "12 min",
  };

  const sortedQueue = [...queueStudents].sort((a, b) => a.joinedAt.getTime() - b.joinedAt.getTime());

  // Recompute live wait-time labels every 15 seconds.
  useEffect(() => {
    const interval = setInterval(() => {
      setQueueStudents((prev) => prev.map((student) => ({
        ...student,
        waitTime: formatWait(getWaitMinutes(student.joinedAt)),
      })));
    }, 15000);
    return () => clearInterval(interval);
  }, []);

  // Load queue state when TA opens/pauses queue view.
  useEffect(() => {
    if (queueStatus === 'closed' || !activeQueueID) {
      return;
    }
    void refreshQueueData();
  }, [queueStatus, activeQueueID]);

  // Subscribe to backend queue events for real-time updates while live queue is visible.
  useEffect(() => {
    if (queueStatus === 'closed' || !activeQueueID) {
      return;
    }

    const unsubscribe = subscribeToQueueEvents(
      activeQueueID,
      () => {
        void refreshQueueData();
      },
      () => {
        // silent reconnect fallback: just poll snapshot once on error
        void refreshQueueData();
      }
    );

    return () => unsubscribe();
  }, [queueStatus, activeQueueID]);

  // Poll snapshot as a reliability fallback in case SSE drops.
  useEffect(() => {
    if (queueStatus === 'closed' || !activeQueueID) {
      return;
    }

    const interval = setInterval(() => {
      void refreshQueueData();
    }, 3000);

    return () => clearInterval(interval);
  }, [queueStatus, activeQueueID]);

  const handleOpenQueue = async () => {
    if (!selectedOfficeHour) {
      toast.info('Select an office hour time before starting the live queue.');
      return;
    }

    try {
      const existing = await getActiveQueueByCourse(selectedOfficeHour.courseID);
      const queueID = existing?.id ?? (await createQueue(selectedOfficeHour.courseID)).id;
      setActiveQueueForCourse(selectedOfficeHour.courseCode, queueID);
      setActiveQueueForOfficeHour(selectedOfficeHour.courseCode, selectedOfficeHour.time, queueID);
      setActiveQueueID(queueID);
      setActiveCourseCode(selectedOfficeHour.courseCode);
      setActiveOfficeHourTime(selectedOfficeHour.time);
      await updateQueueStatus(queueID, 'open');
      setQueueStatus('open');
      void refreshQueueData();
      toast.success(`Live Queue started for ${selectedOfficeHour.courseCode} ${selectedOfficeHour.time}`);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to start queue';
      toast.info(message);
    }
  };

  const handlePauseQueue = async () => {
    if (!activeQueueID) {
      toast.info('No active queue found.');
      return;
    }
    try {
      const nextStatus = queueStatus === 'paused' ? 'open' : 'paused';
      await updateQueueStatus(activeQueueID, nextStatus);
      if (activeCourseCode && activeOfficeHourTime) {
        if (nextStatus === 'open') {
          setActiveQueueForOfficeHour(activeCourseCode, activeOfficeHourTime, activeQueueID);
        } else {
          clearActiveQueueForOfficeHour(activeCourseCode, activeOfficeHourTime);
        }
      }
      setQueueStatus(nextStatus);
      toast.info(nextStatus === 'open' ? 'Queue reopened' : 'Queue paused');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to update queue status';
      toast.info(message);
    }
  };

  const handleCloseQueue = async () => {
    try {
      if (activeQueueID) {
        await updateQueueStatus(activeQueueID, 'closed');
      }
      if (activeCourseCode) {
        clearActiveQueueForCourse(activeCourseCode);
      }
      if (activeCourseCode && activeOfficeHourTime) {
        clearActiveQueueForOfficeHour(activeCourseCode, activeOfficeHourTime);
      }
      setQueueStatus('closed');
      setActiveQueueID(null);
      setActiveCourseCode(null);
      setActiveOfficeHourTime(null);
      setQueueStudents([]);
      toast.warning('Queue closed');
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to close queue';
      toast.info(message);
    }
  };

  const handleStartSession = async (student: QueueStudent) => {
    if (!activeQueueID) {
      toast.info('No active queue found.');
      return;
    }
    try {
      await nextQueueStudent(activeQueueID);
      await refreshQueueData();
      toast.success(`Starting session with ${student.name}`);
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Failed to start session';
      toast.info(message);
    }
  };

  const handleRemoveStudent = async (student: QueueStudent) => {
    const firstStudent = sortedQueue[0];
    if (!firstStudent || firstStudent.id !== student.id) {
      toast.info('Only the first student can be advanced at this time.');
      return;
    }
    await handleStartSession(student);
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
    <div className="ta-dashboard ta-dashboard-root">

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
        <>
        <div className="dashboard-content live-queue-dashboard-content">
          <div className="main-section queue-live-layout">

            {/* Welcome Section */}
            <div className="welcome-section">
              <h1 className="welcome-title">
                Welcome Back, {user?.username || "Lovely TA"}! Here is your Schedule for the Day
              </h1>
              <div className="schedule-cards">
                {todaySchedule.map((slot) => (
                  <div
                    key={slot.id}
                    className={`schedule-card ${selectedOfficeHourID === slot.id ? 'selected-office-hour' : ''}`}
                    onClick={() => setSelectedOfficeHourID(slot.id)}
                    role="button"
                    tabIndex={0}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        setSelectedOfficeHourID(slot.id);
                      }
                    }}
                  >
                    <div className="schedule-time">
                      <img src={orangeClockIcon} alt="Time" className="time-icon" />
                      <span className="time-text">{slot.time}</span>
                    </div>
                    <div className="schedule-course">{slot.course}</div>
                  </div>
                ))}
              </div>
              <button onClick={() => void handleOpenQueue()} className="start-queue-btn" disabled={!selectedOfficeHourID}>
                <span className="play-icon">▶</span>
                {selectedOfficeHourID ? 'Start Office Hours Live Queue' : 'Select Time to Start Live Queue'}
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

          </div>

          {/* Sidebar */}
          <aside className="sidebar">
            <div className="sidebar-content">
              <h2 className="sidebar-title">This Week's Office Hours</h2>
              {officeHoursSidebar}
            </div>
          </aside>
        </div>

        {/* Announcements - Full Width */}
        <div className="announcements-section dashboard-announcements-full-width">
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
        </>
      ) : (
        // OPEN/PAUSED STATE - Live queue view
        <div className="dashboard-content">
          <div className="main-section">

            {/* Queue Header */}
            <div className="welcome-section live-queue-header-section">
              <div className="live-queue-header-row">
                <div>
                  <h1 className="welcome-title live-queue-title">Live Queue</h1>
                  <div className="live-queue-status-row">
                    <div className={`live-queue-status-dot ${queueStatus === 'open' ? 'is-open' : 'is-paused'}`} />
                    <span className="live-queue-status-text">
                      {queueStatus === 'open' ? 'Open - Accepting Students' : 'Paused - No New Students'}
                    </span>
                  </div>
                </div>
                <div className="live-queue-action-row">
                  <button
                    onClick={handlePauseQueue}
                    className={`queue-action-btn pause-btn ${queueStatus === 'paused' ? 'paused' : 'open'}`}
                  >
                    {queueStatus === 'paused' ? 'Resume Queue' : 'Pause Queue'}
                  </button>
                  <button
                    onClick={handleCloseQueue}
                    className="queue-action-btn close-btn"
                  >
                    Close Queue
                  </button>
                  <button
                    className="send-announcement-btn live-send-announcement-btn"
                    onClick={() => setShowAnnouncementModal(true)}
                  >
                    Send Announcement
                  </button>
                </div>
              </div>
            </div>

            {/* Queue Stats */}
            <div className="stats-grid live-queue-stats-grid">
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
            <div className="sidebar-content queue-list-panel">
              <h2 className="sidebar-title">Students in Queue ({sortedQueue.length})</h2>
              <div className="queue-list-scroll-region">
                {sortedQueue.length === 0 ? (
                  <div className="queue-empty-state">
                    <p className="queue-empty-title">No students in queue</p>
                    <p className="queue-empty-subtitle">
                      Students will appear here when they join
                    </p>
                  </div>
                ) : (
                  <div className="queue-students-list">
                  {sortedQueue.map((student, index) => (
                    <div key={student.id} className="office-hour-card">
                      <div className="queue-student-row">
                        <div className="queue-student-left">
                          <div className="queue-student-position-badge">
                            {index + 1}
                          </div>
                          <div>
                            <div className="queue-student-name">
                              {student.name}
                            </div>
                            <div className="queue-student-meta">
                              <span className="queue-student-wait">
                                Waiting {student.waitTime}
                              </span>
                              {student.topic && (
                                <>
                                  <span className="queue-topic-separator">•</span>
                                  <span className="queue-student-topic">
                                    {student.topic}
                                  </span>
                                </>
                              )}
                            </div>
                          </div>
                        </div>
                        <div className="office-hour-actions queue-student-actions">
                          {index === 0 && (
                            <button
                              onClick={() => handleStartSession(student)}
                              className="start-queue-btn queue-start-session-btn"
                            >
                              Start Session
                            </button>
                          )}
                          <button
                            onClick={() => handleRemoveStudent(student)}
                            className="cancel-btn queue-remove-btn"
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
          </div>

          {/* Sidebar */}
          <aside className="sidebar">
            <div className="sidebar-content live-sidebar-office-hours">
              <h2 className="sidebar-title">This Week's Office Hours</h2>
              {officeHoursSidebar}
            </div>
            <div className="announcements-section live-sidebar-announcements">
              <div className="live-announcements-header">
                <h3 className="live-announcements-title">
                  Announcements
                </h3>
                <p className="live-announcements-subtitle">
                  Send updates to all students in the queue
                </p>
              </div>
              <button className="send-announcement-btn full-width-announcement-btn" onClick={() => setShowAnnouncementModal(true)}>
                Send Announcement
              </button>
            </div>
          </aside>
        </div>
      )}

      {/* Announcement Modal */}
      {showAnnouncementModal && (
        <div className="announcement-modal-overlay">
          <div className="announcement-modal-card">
            <div className="announcement-modal-header">
              <h3 className="announcement-modal-title">
                Send Announcement
              </h3>
              <button onClick={() => setShowAnnouncementModal(false)}
                className="announcement-modal-close-btn">
                ✕
              </button>
            </div>
            <div className="announcement-modal-body">
              <p className="announcement-modal-description">
                {queueStatus === 'closed'
                  ? 'Send a message to all students in your course'
                  : `Send a message to all ${queueStudents.length} students currently in the queue`}
              </p>
              <textarea
                value={announcementText}
                onChange={(e) => setAnnouncementText(e.target.value)}
                placeholder="e.g., Running 10 minutes late, extending office hours by 30 minutes..."
                className="announcement-modal-textarea"
              />
              <div className="announcement-example-box">
                <p className="announcement-example-title">
                  Example announcements:
                </p>
                <ul className="announcement-example-list">
                  <li>Running 10 minutes late</li>
                  <li>Office hours canceled today</li>
                  <li>Extending office hours by 30 minutes</li>
                </ul>
              </div>
              <div className="announcement-modal-actions">
                <button onClick={() => setShowAnnouncementModal(false)}
                  className="announcement-cancel-btn">
                  Cancel
                </button>
                <button onClick={handleSendAnnouncement}
                  className="start-queue-btn announcement-send-all-btn">
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