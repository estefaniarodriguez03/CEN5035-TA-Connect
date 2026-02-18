import { useAuth } from "../context/AuthContext";
import ufLogo from "../images/UF Logo.png";
import blueMessageIcon from "../images/Blue Message Icon.png";
import orangeClockIcon from "../images/Orange Clock Icon.png";
import orangeDateIcon from "../images/Orange Date Icon.png";
import orangeGroupIcon from "../images/Orange Group Icon.png";
import orangeQuestionIcon from "../images/Orange Question Mark Icon.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";

interface OfficeHour {
  id: number;
  day: string;
  time: string;
  course: string;
}

interface TimeSlot {
  id: number;
  time: string;
  course: string;
}

export default function TADashboard() {
  const { user, logout } = useAuth();

  // Mock data for now we will replace with actual API calls later
  const todaySchedule: TimeSlot[] = [
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
    avgWaitTime: "12 min",
    currentQueueLength: 0,
    longestWaitTime: "0 min",
    mostCommonTopic: "Data Structures",
    avgSessionDuration: "5 min",
  };

  return (
    <div className="ta-dashboard">
      {/* Navigation Bar */}
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

      {/* Main Content */}
      <div className="dashboard-content">
        <div className="main-section">
          {/* Welcome Header */}
          <div className="welcome-section">
            <h1 className="welcome-title">
              Welcome Back, {user?.username || "Lovely TA"}! Here is your Schedule for the Day
            </h1>

            {/* Today's Schedule */}
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

            {/* Start Queue Button */}
            <button className="start-queue-btn">
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

          {/* Announcements Section */}
          <div className="announcements-section">
            <div className="announcements-content">
              <div className="announcements-text">
                <img src={blueMessageIcon} alt="Announcements" className="announcement-icon" />
                <div>
                  <h3>Send Announcements to Your Students</h3>
                  <p>Broadcast messages to all students enrolled in your courses</p>
                </div>
              </div>
              <button className="send-announcement-btn">
                Send Announcement
              </button>
            </div>
          </div>
        </div>

        {/* Right Sidebar */}
        <aside className="sidebar">
          <div className="sidebar-content">
            <h2 className="sidebar-title">This Week's Office Hours</h2>
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
                    <button
                      className="modify-btn"
                    >
                      Modify
                    </button>
                    <button
                      className="cancel-btn"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
}
