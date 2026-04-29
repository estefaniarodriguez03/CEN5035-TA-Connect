import { useAuth } from "../context/AuthContext";
import { useState, useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import ufLogo from "../images/UF Logo.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";
import orangeClockIcon from "../images/Orange Clock Icon.png";
import orangeDateIcon from "../images/Orange Date Icon.png";
import { toast } from "sonner";
import {
  joinQueue,
  leaveQueue,
  getQueueOrNull,
  subscribeToQueueEvents,
  browseWaitDisplay,
  myWaitMinutesFromEntry,
} from "../api/queue";
import type {
  QueueEvent,
  QueueStateChangePayload,
  StudentUpNextPayload,
  AnnouncementSentPayload,
  SessionStartedPayload,
} from "../api/queue";
import { listStudentSchedule } from "../api/studentSchedule";

interface CourseOption {
  label: string;
  courseCode: string;
  courseID: number;
  officeHourTimeRange: string;
  dayOfWeek: number;
}

const COURSE_COLORS = ['green', 'purple', 'yellow', 'red', 'blue', 'orange'];

function getCourseColor(courseID: number): string {
  return COURSE_COLORS[courseID % COURSE_COLORS.length];
}

function formatTime12(time: string): string {
  const [hours, minutes] = time.split(':');
  const hour = parseInt(hours);
  const ampm = hour >= 12 ? 'PM' : 'AM';
  const displayHour = hour > 12 ? hour - 12 : hour === 0 ? 12 : hour;
  return `${displayHour}:${minutes} ${ampm}`;
}

function extractUpNextStudentID(payload: QueueEvent["payload"]): number | null {
  const direct = payload as StudentUpNextPayload | undefined;
  if (typeof direct?.student_id === "number") {
    return direct.student_id;
  }
  const legacy = payload as { students?: Array<{ student_id?: number; position?: number }> } | undefined;
  const head = legacy?.students?.find((s) => s.position === 1) ?? legacy?.students?.[0];
  return typeof head?.student_id === "number" ? head.student_id : null;
}

const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

export default function StudentDashboard() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const [selectedCourse, setSelectedCourse] = useState('');
  const [isInQueue, setIsInQueue] = useState(false);
  const [timeJoined, setTimeJoined] = useState<string>("");
  const [studentPosition, setStudentPosition] = useState<number | null>(null);
  const [waitTime, setWaitTime] = useState("0 minutes");
  const [waitTimeSub, setWaitTimeSub] = useState("Select a course to see the queue");
  const [inQueueWaitMinutes, setInQueueWaitMinutes] = useState(0);
  const [queueStudentCount, setQueueStudentCount] = useState(0);
  const [joinedQueueID, setJoinedQueueID] = useState<number | null>(null);
  const [queueStatusForCourse, setQueueStatusForCourse] = useState<string | null>(null);
  const [activeSession, setActiveSession] = useState<SessionStartedPayload | null>(null);
  const [scheduleOptions, setScheduleOptions] = useState<CourseOption[]>([]);
  const [activeQueueID, setActiveQueueID] = useState<number | null>(null);

  // Load student schedule on mount
  useEffect(() => {
    void (async () => {
      try {
        const entries = await listStudentSchedule();
        const options: CourseOption[] = entries.map((e) => ({
          label: `${e.course_code} – ${e.course_name} – ${e.ta_username} (${formatTime12(e.start_time)} - ${formatTime12(e.end_time)})`,
          courseCode: e.course_code,
          courseID: e.course_id,
          officeHourTimeRange: `${formatTime12(e.start_time)} - ${formatTime12(e.end_time)}`,
          dayOfWeek: e.day_of_week,
        }));
        setScheduleOptions(options);
      } catch {
        // silently ignore
      }
    })();
  }, []);

  // Auto-select course from navigation state (coming from My Courses page)
  useEffect(() => {
    const state = location.state as {
      autoSelectLabel?: string;
    } | null;

    if (!state?.autoSelectLabel) return;
    setSelectedCourse(state.autoSelectLabel);
    navigate(location.pathname, { replace: true, state: null });
  }, []);

  const courseOptions = scheduleOptions;
  const selectedCourseOption = courseOptions.find((c) => c.label === selectedCourse) ?? null;

  // Poll backend for active queue when a course is selected
  useEffect(() => {
    if (!selectedCourseOption) {
      setActiveQueueID(null);
      return;
    }

    const checkForActiveQueue = async () => {
      try {
        const res = await fetch(
          `${API_BASE}/api/queues/active?course_id=${selectedCourseOption.courseID}`
        );
        if (res.ok) {
          const data = await res.json();
          setActiveQueueID(data.id ?? null);
          setQueueStatusForCourse(data.status ?? null);
        } else {
          setActiveQueueID(null);
          setQueueStatusForCourse(null);
        }
      } catch {
        setActiveQueueID(null);
      }
    };

    void checkForActiveQueue();

    // Poll every 5 seconds so the student sees the queue open without refreshing
    const interval = setInterval(() => void checkForActiveQueue(), 5000);
    return () => clearInterval(interval);
  }, [selectedCourseOption?.courseID]);

  // Load queue data when selected course changes or when joining/leaving queue
  useEffect(() => {
    const loadQueueData = async () => {
      try {
        if (!selectedCourseOption) {
          setQueueStudentCount(0);
          setWaitTime("—");
          setWaitTimeSub("Select a course to see the queue");
          setInQueueWaitMinutes(0);
          setStudentPosition(null);
          setQueueStatusForCourse(null);
          return;
        }

        const resolvedQueueID = isInQueue && joinedQueueID
          ? joinedQueueID
          : activeQueueID;

        if (!resolvedQueueID) {
          setQueueStudentCount(0);
          setWaitTime("—");
          setWaitTimeSub("No queue is open for this course yet. Your TA will open one when office hours start.");
          setInQueueWaitMinutes(0);
          setStudentPosition(null);
          return;
        }

        const data = await getQueueOrNull(resolvedQueueID);
        if (!data) {
          setQueueStudentCount(0);
          setWaitTime("—");
          setWaitTimeSub("Queue is unavailable or no longer exists");
          setInQueueWaitMinutes(0);
          setStudentPosition(null);
          setQueueStatusForCourse(null);
          if (isInQueue) {
            setIsInQueue(false);
            setJoinedQueueID(null);
          }
          return;
        }
        setQueueStatusForCourse(data.status);
        setQueueStudentCount(data.entries.length);
        const display = browseWaitDisplay(data);
        setWaitTime(display.line);
        setWaitTimeSub(display.sub);

        if (isInQueue) {
          const currentStudent = data.entries.find(
            (entry) => entry.student_id === user?.id
          );
          if (currentStudent) {
            setStudentPosition(currentStudent.position);
            setInQueueWaitMinutes(myWaitMinutesFromEntry(currentStudent));
          } else {
            setStudentPosition(null);
            setInQueueWaitMinutes(0);
            setIsInQueue(false);
            setJoinedQueueID(null);
          }
        } else {
          setInQueueWaitMinutes(0);
        }
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to load queue data";
        toast.error(message);
      }
    };

    void loadQueueData();
  }, [selectedCourse, isInQueue, user?.id, joinedQueueID, activeQueueID]);

  // Subscribe to state changes on the active queue
  useEffect(() => {
    if (isInQueue) return;
    const queueID = activeQueueID;
    if (!queueID) return;

    const unsubscribe = subscribeToQueueEvents(
      queueID,
      (evt: QueueEvent) => {
        if (evt.type === "QUEUE_STATE_CHANGED") {
          const p = evt.payload as QueueStateChangePayload | undefined;
          if (p?.status) {
            setQueueStatusForCourse(p.status);
          }
        }
        if (evt.type === "ANNOUNCEMENT_SENT") {
          const p = evt.payload as AnnouncementSentPayload | undefined;
          if (p?.message) {
            toast.info("Announcement from TA", { description: p.message });
          }
        }
      },
      () => { /* ignore errors for background subscription */ }
    );

    return () => unsubscribe();
  }, [isInQueue, activeQueueID]);

  // Subscribe to real-time queue updates when in queue
  useEffect(() => {
    if (!isInQueue || !joinedQueueID) return;

    const refreshFromServer = async () => {
      try {
        const data = await getQueueOrNull(joinedQueueID);
        if (!data) {
          setQueueStudentCount(0);
          setWaitTime("—");
          setWaitTimeSub("Queue is unavailable or no longer exists");
          setInQueueWaitMinutes(0);
          setStudentPosition(null);
          setIsInQueue(false);
          return;
        }
        setQueueStudentCount(data.entries.length);
        const d = browseWaitDisplay(data);
        setWaitTime(d.line);
        setWaitTimeSub(d.sub);

        const currentStudent = data.entries.find(
          (entry) => entry.student_id === user?.id
        );
        if (currentStudent) {
          setStudentPosition(currentStudent.position);
          setInQueueWaitMinutes(myWaitMinutesFromEntry(currentStudent));
        } else {
          setStudentPosition(null);
          setInQueueWaitMinutes(0);
          setIsInQueue(false);
          setJoinedQueueID(null);
        }
      } catch (err) {
        const message = err instanceof Error ? err.message : "Failed to update queue data";
        toast.error(message);
      }
    };

    const handleQueueEvent = (evt: QueueEvent) => {
      if (evt.type === "QUEUE_STATE_CHANGED") {
        const p = evt.payload as QueueStateChangePayload | undefined;
        if (p?.status) {
          setQueueStatusForCourse(p.status);
        }
        if (p?.status === "paused") {
          toast.info("The queue has been paused by the TA. You are still in line.");
        } else if (p?.status === "closed") {
          toast.warning("The queue has been closed by the TA.");
          setIsInQueue(false);
          setJoinedQueueID(null);
          setStudentPosition(null);
          return;
        } else if (p?.status === "open" && p.previous_status === "paused") {
          toast.success("The queue has been reopened!");
        }
      }
      if (evt.type === "STUDENT_UP_NEXT") {
        const nextStudentID = extractUpNextStudentID(evt.payload);
        if (nextStudentID !== null && nextStudentID === user?.id) {
          toast.success("You're up next. Please get ready!");
        }
      }
      if (evt.type === "SESSION_STARTED") {
        const p = evt.payload as SessionStartedPayload | undefined;
        if (p && p.student_id === user?.id) {
          setActiveSession(p);
          toast.success("Your session is starting now!", {
            description: "Click the Zoom link to join the meeting.",
          });
          setIsInQueue(false);
          setJoinedQueueID(null);
          setStudentPosition(null);
        }
      }
      if (evt.type === "ANNOUNCEMENT_SENT") {
        const p = evt.payload as AnnouncementSentPayload | undefined;
        if (p?.message) {
          toast.info("Announcement from TA", { description: p.message });
        }
      }
      void refreshFromServer();
    };

    const unsubscribe = subscribeToQueueEvents(
      joinedQueueID,
      handleQueueEvent,
      () => { void refreshFromServer(); }
    );

    return () => { unsubscribe(); };
  }, [isInQueue, joinedQueueID, user?.id]);

  // Poll snapshot as reliability fallback
  useEffect(() => {
    if (!isInQueue || !joinedQueueID) return;

    const interval = setInterval(async () => {
      try {
        const data = await getQueueOrNull(joinedQueueID);
        if (!data) {
          setQueueStudentCount(0);
          setWaitTime("—");
          setWaitTimeSub("Queue is unavailable or no longer exists");
          setInQueueWaitMinutes(0);
          setStudentPosition(null);
          setIsInQueue(false);
          setJoinedQueueID(null);
          return;
        }

        setQueueStudentCount(data.entries.length);
        const pol = browseWaitDisplay(data);
        setWaitTime(pol.line);
        setWaitTimeSub(pol.sub);

        const currentStudent = data.entries.find((entry) => entry.student_id === user?.id);
        if (currentStudent) {
          setStudentPosition(currentStudent.position);
          setInQueueWaitMinutes(myWaitMinutesFromEntry(currentStudent));
        } else {
          setStudentPosition(null);
          setInQueueWaitMinutes(0);
          setIsInQueue(false);
          setJoinedQueueID(null);
        }
      } catch {
        // Keep silent to avoid repeated toast spam from periodic poll.
      }
    }, 2000);

    return () => clearInterval(interval);
  }, [isInQueue, joinedQueueID, user?.id]);

  const getWeekRange = () => {
    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysToMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const monday = new Date(today);
    monday.setDate(monday.getDate() - daysToMonday);
    const friday = new Date(monday);
    friday.setDate(friday.getDate() + 4);
    const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    return `Week of ${monthNames[monday.getMonth()]} ${monday.getDate()} - ${monthNames[friday.getMonth()]} ${friday.getDate()}`;
  };

  const handleJoinQueue = async () => {
    try {
      if (!selectedCourseOption) {
        toast.error("Please select a course first");
        return;
      }

      const queueID = activeQueueID;
      if (!queueID) {
        toast.error("The TA has not opened the queue yet. Come back later!");
        return;
      }

      const queueState = await getQueueOrNull(queueID);
      if (!queueState || queueState.status !== "open") {
        if (queueState?.status === "paused") {
          toast.error("The queue is currently paused. Please wait for the TA to resume it.");
        } else if (queueState?.status === "closed") {
          toast.error("The queue is closed. Check back during the next office hours.");
        } else {
          toast.error("The TA has not opened the queue yet. Come back later!");
        }
        return;
      }

      const response = await joinQueue(queueID);
      const joined = new Date(response.joined_at);
      const displayHours = (joined.getHours() % 12) || 12;
      const displayMinutes = joined.getMinutes().toString().padStart(2, "0");
      const ampm = joined.getHours() >= 12 ? "PM" : "AM";
      setTimeJoined(`${displayHours}:${displayMinutes} ${ampm}`);
      setIsInQueue(true);
      setJoinedQueueID(queueID);
      setStudentPosition(response.position);
      if (response.position === 1) {
        toast.success("You're up next. Please get ready!");
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to join queue";
      toast.error(message);
    }
  };

  const handleLeaveQueue = async () => {
    try {
      const queueID = joinedQueueID ?? activeQueueID;
      if (!queueID) {
        toast.error("No active queue to leave.");
        return;
      }

      await leaveQueue(queueID);
      setIsInQueue(false);
      setJoinedQueueID(null);
      setStudentPosition(null);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to leave queue";
      toast.error(message);
    }
  };

  const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

  return (
    <div className="student-dashboard">
      {/* Navigation Bar */}
      <nav className="navbar">
        <div className="navbar-left">
          <div className="logo">
            <img src={ufLogo} alt="UF Logo" className="logo-icon" />
            <span className="logo-text">TA Connect</span>
          </div>
          <div className="nav-tabs">
            <button className="nav-tab active">Dashboard</button>
            <button className="nav-tab" onClick={() => navigate("/student/my-courses")}>My Courses</button>
          </div>
        </div>
        <div className="navbar-right">
          <div className="notification-icon">
            <img src={whiteNotificationIcon} alt="Notifications" />
          </div>
          <button className="profile-icon" onClick={() => navigate("/profile")}>
            <img src={whiteProfileIcon} alt="Profile" />
          </button>
        </div>
      </nav>

      {/* Main Content */}
      <div className="dashboard-content student-content">
        <div className="top-row">
          {/* Welcome Section */}
          <div className="welcome-section student-welcome">
            <h1 className="welcome-title">
              Welcome Back, {user?.username || "Student"}! Join a Queue to get Started
            </h1>

            {/* Course Selection */}
            <div className="course-selection">
              <label className="course-label">Select TA / Course</label>
              <select
                className="course-dropdown"
                value={selectedCourse}
                onChange={(e) => setSelectedCourse(e.target.value)}
                disabled={isInQueue}
              >
                {courseOptions.length === 0 ? (
                  <option value="">No courses added yet — visit My Courses to add some</option>
                ) : (
                  <>
                    <option value="">Select a course and TA...</option>
                    {courseOptions.map((course, index) => (
                      <option key={index} value={course.label}>
                        {course.label}
                      </option>
                    ))}
                  </>
                )}
              </select>
            </div>

            {/* Estimated Wait Time */}
            <div className="wait-time-card">
              <div className="wait-time-header">
                <img src={orangeClockIcon} alt="Clock" className="wait-time-icon" />
                <span className="wait-time-label">Estimated Wait Time</span>
              </div>
              <div className="wait-time-value">{waitTime}</div>
              <div className="wait-time-info">{waitTimeSub}</div>
            </div>

            {/* Queue Status Banner */}
            {!isInQueue && queueStatusForCourse === "paused" && (
              <div className="queue-status-banner queue-status-paused">
                <span className="queue-status-banner-icon">⏸</span>
                <div>
                  <strong>Queue Paused</strong>
                  <div className="queue-status-banner-sub">The TA has temporarily paused the queue. You'll be able to join once it reopens.</div>
                </div>
              </div>
            )}
            {!isInQueue && queueStatusForCourse === "closed" && (
              <div className="queue-status-banner queue-status-closed">
                <span className="queue-status-banner-icon">✕</span>
                <div>
                  <strong>Queue Closed</strong>
                  <div className="queue-status-banner-sub">This queue is no longer accepting students. Check back during the next office hours.</div>
                </div>
              </div>
            )}

            {/* Join Queue Button */}
            <button
              className={`join-queue-btn ${isInQueue ? 'in-queue' : ''} ${!isInQueue && queueStatusForCourse && queueStatusForCourse !== 'open' ? 'queue-blocked' : ''}`}
              onClick={handleJoinQueue}
              disabled={isInQueue || !selectedCourse || (queueStatusForCourse !== null && queueStatusForCourse !== 'open')}
            >
              {isInQueue
                ? 'Already in Queue'
                : queueStatusForCourse === 'paused'
                  ? 'Queue Paused — Cannot Join'
                  : queueStatusForCourse === 'closed'
                    ? 'Queue Closed'
                    : 'Join Queue'}
            </button>
          </div>

          {/* Today's TA Hours */}
          <div className="todays-ta-hours">
            <h2 className="ta-hours-title">Today's TA Hours</h2>
            <div className="ta-hours-list">
              {scheduleOptions.length === 0 ? (
                <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>
                  No courses added yet. Visit My Courses to add office hours to your schedule.
                </p>
              ) : (
                scheduleOptions
                  .filter((opt) => opt.dayOfWeek === new Date().getDay())
                  .map((option, index) => (
                    <div key={index} className="ta-hour-card">
                      <div className="ta-hour-row">
                        <div className="ta-info">
                          <div className="ta-name-time">
                            <span className="status-dot"></span>
                            <span className="ta-name">{option.label}</span>
                          </div>
                          <div className="ta-course-info">{option.courseCode}</div>
                        </div>
                      </div>
                    </div>
                  ))
              )}
              {scheduleOptions.length > 0 &&
                scheduleOptions.filter((opt) => opt.dayOfWeek === new Date().getDay()).length === 0 && (
                  <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>
                    No office hours scheduled for today.
                  </p>
                )}
            </div>
          </div>
        </div>

        {/* Real-Time Queue Status — only shown when in queue */}
        {isInQueue && (
          <div className="real-time-queue-section">
            <div className="queue-header">
              <h2 className="queue-title">Real-Time Queue Status</h2>
              <div className="live-indicator">
                <span className="live-dot"></span>
                <span className="live-text">Live</span>
              </div>
            </div>

            <div className="queue-status-cards">
              <div className="queue-card position-card">
                <div className="card-label">Current Position</div>
                <div className="card-value position-value">#{studentPosition || 0}</div>
                <div className="card-subtext">out of {queueStudentCount} students</div>
              </div>

              <div className="queue-card wait-time-card-alt">
                <div className="card-label">Est. Wait Time</div>
                <div className="card-value wait-value">{inQueueWaitMinutes}</div>
                <div className="card-subtext">minutes remaining</div>
              </div>
            </div>

            <div className="queue-details">
              <div className="queue-detail-row">
                <span className="detail-label">Course</span>
                <span className="detail-value">{selectedCourseOption?.courseCode ?? '—'}</span>
              </div>
              <div className="queue-detail-row">
                <span className="detail-label">Time Joined</span>
                <span className="detail-value">{timeJoined}</span>
              </div>
            </div>

            <div className={`queue-notification ${studentPosition === 1 ? "next-in-line" : ""}`}>
              <div className="notification-icon-circle">
                <span className="notification-icon-text">ⓘ</span>
              </div>
              <div className="notification-text">
                <strong>{studentPosition === 1 ? "Be ready! You're next in line." : "You'll receive a notification when you're next in line."}</strong>
                <div className="notification-subtext">Make sure to stay nearby and keep notifications enabled.</div>
              </div>
            </div>

            <button className="leave-queue-btn" onClick={handleLeaveQueue}>
              <span className="leave-icon">↗</span> Cancel & Leave Queue
            </button>
          </div>
        )}

        {/* Weekly Schedule */}
        <div className="weekly-schedule">
          <div className="schedule-header">
            <h2 className="schedule-title">This Week's Office Hours</h2>
            <div className="schedule-date">
              <img src={orangeDateIcon} alt="Calendar" className="calendar-icon-small" />
              {getWeekRange()}
            </div>
          </div>

          <div className="schedule-grid">
            {["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"].map((day) => {
              const dayIndex = DAY_NAMES.indexOf(day);
              return (
                <div key={day} className="schedule-day">
                  <div className="day-header">{day}</div>
                  <div className="day-slots">
                    {scheduleOptions
                      .filter((opt) => opt.dayOfWeek === dayIndex)
                      .map((opt, idx) => (
                        <div key={idx} className={`time-slot slot-${getCourseColor(opt.courseID)}`}>
                          <div className="slot-time">{opt.officeHourTimeRange}</div>
                          <div className="slot-ta">
                            {opt.label.split(' – ')[2]?.split(' (')[0] ?? ''}
                          </div>
                        </div>
                      ))}
                  </div>
                </div>
              );
            })}
          </div>

          {/* Session Started Modal */}
          {activeSession && (
            <div className="announcement-modal-overlay">
              <div className="announcement-modal-card">
                <div className="announcement-modal-header">
                  <h3 className="announcement-modal-title">
                    Your Session Is Ready
                  </h3>
                  <button
                    onClick={() => setActiveSession(null)}
                    className="announcement-modal-close-btn"
                  >
                    ✕
                  </button>
                </div>
                <div className="announcement-modal-body">
                  <p className="announcement-modal-description">
                    {activeSession.student_name
                      ? `${activeSession.student_name}, your TA has started the session.`
                      : 'Your TA has started the session.'}{' '}
                    Click below to join the Zoom meeting.
                  </p>

                  <div className="queue-detail-row" style={{ marginTop: '0.5rem' }}>
                    <span className="detail-label">Meeting ID</span>
                    <span className="detail-value">{activeSession.zoom_meeting_id || '—'}</span>
                  </div>
                  {activeSession.zoom_passcode && (
                    <div className="queue-detail-row">
                      <span className="detail-label">Passcode</span>
                      <span className="detail-value">{activeSession.zoom_passcode}</span>
                    </div>
                  )}
                  <div className="queue-detail-row">
                    <span className="detail-label">Join URL</span>
                    <a
                      className="detail-value"
                      href={activeSession.zoom_join_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{ wordBreak: 'break-all' }}
                    >
                      {activeSession.zoom_join_url}
                    </a>
                  </div>

                  <div className="announcement-modal-actions">
                    <button
                      onClick={() => setActiveSession(null)}
                      className="announcement-cancel-btn"
                    >
                      Dismiss
                    </button>
                    <a
                      href={activeSession.zoom_join_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="start-queue-btn announcement-send-all-btn"
                      style={{ textAlign: 'center', textDecoration: 'none' }}
                    >
                      Join Zoom Meeting
                    </a>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}