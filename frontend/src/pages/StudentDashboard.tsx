import { useAuth } from "../context/AuthContext";
import { useState } from "react";
import ufLogo from "../images/UF Logo.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";
import orangeClockIcon from "../images/Orange Clock Icon.png";
import orangeDateIcon from "../images/Orange Date Icon.png";

interface TAHour {
  id: number;
  taName: string;
  time: string;
  course: string;
  students: number;
  status: "Live" | "Soon" | "Offline";
}

interface WeeklySlot {
  id: number;
  day: string;
  time: string;
  taName: string;
  course: string;
  color: string;
}

export default function StudentDashboard() {
  const { user, logout } = useAuth();
  const [selectedCourse, setSelectedCourse] = useState(
    "CEN3031 – Software Engineering – Estefania Rodriguez (9:00 AM - 11:00 AM)"
  );

  // Mock data for now we will replace with actual API calls later
  const courseOptions = [
    "CEN3031 – Software Engineering – Estefania Rodriguez (9:00 AM - 11:00 AM)",
    "COP3530 – Data Structures – Sara Waters (11:00 AM - 1:00 PM)",
    "COP4020 – Programming Languages – Raghav Nanjappan (2:00 PM - 4:00 PM)",
    "COP4600 – Operating Systems – John Spurrier (10:00 AM - 12:00 PM)",
  ];

  const queueData: { [key: string]: { students: number; waitTime: string } } = {
    "CEN3031 – Software Engineering – Estefania Rodriguez (9:00 AM - 11:00 AM)": { students: 3, waitTime: "12 minutes" },
    "COP3530 – Data Structures – Sara Waters (11:00 AM - 1:00 PM)": { students: 0, waitTime: "0 minutes" },
    "COP4020 – Programming Languages – Raghav Nanjappan (2:00 PM - 4:00 PM)": { students: 0, waitTime: "0 minutes" },
    "COP4600 – Operating Systems – John Spurrier (10:00 AM - 12:00 PM)": { students: 1, waitTime: "4 minutes" },
  };

  const currentQueueInfo = queueData[selectedCourse] || { students: 0, waitTime: "0 minutes" };

  const todayTAHours: TAHour[] = [
    { id: 1, taName: "Estefania Rodriguez", time: "(9:00 AM - 11:00 AM)", course: "CEN3031", students: 3, status: "Live" },
    { id: 2, taName: "Sara Waters", time: "(11:00 AM - 1:00 PM)", course: "COP3530", students: 0, status: "Soon" },
    { id: 3, taName: "Raghav Nanjappan", time: "(2:00 PM - 4:00 PM)", course: "COP4020", students: 0, status: "Offline" },
    { id: 4, taName: "John Spurrier", time: "(10:00 AM - 12:00 PM)", course: "COP4600", students: 1, status: "Live" },
  ];

  const weeklySchedule: WeeklySlot[] = [
    { id: 1, day: "Monday", time: "9:00 AM - 11:00 AM", taName: "Estefania Rodriguez", course: "CEN3031 - Software Engineering", color: "green" },
    { id: 2, day: "Monday", time: "11:00 AM - 1:00 PM", taName: "Raghav Nanjappan", course: "COP4020 - Programming Languages", color: "yellow" },
    { id: 3, day: "Monday", time: "2:00 PM - 4:00 PM", taName: "Sara Waters", course: "COP3530 - Data Structures", color: "purple" },
    { id: 4, day: "Tuesday", time: "10:00 AM - 12:00 PM", taName: "John Spurrier", course: "COP4600 - Operating Systems", color: "red" },
    { id: 5, day: "Tuesday", time: "1:00 PM - 3:00 PM", taName: "Sara Waters", course: "COP3530 - Data Structures", color: "purple" },
    { id: 6, day: "Tuesday", time: "3:00 PM - 5:00 PM", taName: "Estefania Rodriguez", course: "CEN3031 - Software Engineering", color: "green" },
    { id: 7, day: "Wednesday", time: "9:00 AM - 11:00 AM", taName: "Raghav Nanjappan", course: "COP4020 - Programming Languages", color: "yellow" },
    { id: 8, day: "Wednesday", time: "11:00 AM - 1:00 PM", taName: "John Spurrier", course: "COP4600 - Operating Systems", color: "red" },
    { id: 9, day: "Wednesday", time: "2:00 PM - 4:00 PM", taName: "Sara Waters", course: "COP3530 - Data Structures", color: "purple" },
    { id: 10, day: "Thursday", time: "10:00 AM - 12:00 PM", taName: "Estefania Rodriguez", course: "CEN3031 - Software Engineering", color: "green" },
    { id: 11, day: "Thursday", time: "1:00 PM - 3:00 PM", taName: "Raghav Nanjappan", course: "COP4020 - Programming Languages", color: "yellow" },
    { id: 12, day: "Thursday", time: "3:00 PM - 5:00 PM", taName: "John Spurrier", course: "COP4600 - Operating Systems", color: "red" },
    { id: 13, day: "Friday", time: "9:00 AM - 11:00 AM", taName: "Sara Waters", course: "COP3530 - Data Structures", color: "purple" },
    { id: 14, day: "Friday", time: "11:00 AM - 1:00 PM", taName: "Estefania Rodriguez", course: "CEN3031 - Software Engineering", color: "green" },
    { id: 15, day: "Friday", time: "2:00 PM - 4:00 PM", taName: "Raghav Nanjappan", course: "COP4020 - Programming Languages", color: "yellow" },
  ];

  const days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"];

  const getStatusClass = (status: string) => {
    switch (status) {
      case "Live":
        return "status-live";
      case "Soon":
        return "status-soon";
      case "Offline":
        return "status-offline";
      default:
        return "";
    }
  };

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
            <button className="nav-tab">My Courses</button>
            <button className="nav-tab">My Queue Status</button>
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
      <div className="dashboard-content student-content">
        <div className="top-row">
          {/* Welcome Section */}
          <div className="welcome-section student-welcome">
            <h1 className="welcome-title">
              Welcome Back, {user?.username || "Sarah"}! Join a Queue to get Started
            </h1>

            {/* Course Selection */}
            <div className="course-selection">
              <label className="course-label">Select TA / Course</label>
              <select
                className="course-dropdown"
                value={selectedCourse}
                onChange={(e) => setSelectedCourse(e.target.value)}
              >
                {courseOptions.map((course, index) => (
                  <option key={index} value={course}>
                    {course}
                  </option>
                ))}
              </select>
            </div>

            {/* Estimated Wait Time */}
            <div className="wait-time-card">
              <div className="wait-time-header">
                <img src={orangeClockIcon} alt="Clock" className="wait-time-icon" />
                <span className="wait-time-label">Estimated Wait Time</span>
              </div>
              <div className="wait-time-value">{currentQueueInfo.waitTime}</div>
              <div className="wait-time-info">Based on {currentQueueInfo.students} student{currentQueueInfo.students !== 1 ? "s" : ""} currently in queue</div>
            </div>

            {/* Join Queue Button */}
            <button className="join-queue-btn">
              Join Queue
            </button>
          </div>

          {/* Today's TA Hours */}
          <div className="todays-ta-hours">
            <h2 className="ta-hours-title">Today's TA Hours</h2>
            <div className="ta-hours-list">
              {todayTAHours.map((ta) => (
                <div key={ta.id} className="ta-hour-card">
                  <div className="ta-hour-row">
                    <div className="ta-info">
                      <div className="ta-name-time">
                        <span className="status-dot"></span>
                        <span className="ta-name">{ta.taName} {ta.time}</span>
                      </div>
                      <div className="ta-course-info">
                        {ta.course} - {ta.students} student{ta.students !== 1 ? "s" : ""}
                      </div>
                    </div>
                    <div className={`ta-status-badge ${getStatusClass(ta.status)}`}>
                      {ta.status}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Weekly Schedule */}
        <div className="weekly-schedule">
          <div className="schedule-header">
            <h2 className="schedule-title">This Week's Office Hours</h2>
            <div className="schedule-date">
              <img src={orangeDateIcon} alt="Calendar" className="calendar-icon-small" />
              Week of Feb 16 - Feb 20
            </div>
          </div>

          <div className="schedule-grid">
            {days.map((day) => (
              <div key={day} className="schedule-day">
                <div className="day-header">{day}</div>
                <div className="day-slots">
                  {weeklySchedule
                    .filter((slot) => slot.day === day)
                    .map((slot) => (
                      <div key={slot.id} className={`time-slot slot-${slot.color}`}>
                        <div className="slot-time">{slot.time}</div>
                        <div className="slot-ta">{slot.taName}</div>
                      </div>
                    ))}
                </div>
              </div>
            ))}
          </div>

          {/* Course Legend */}
          <div className="course-legend">
            <div className="legend-item">
              <span className="legend-color legend-green"></span>
              <span>CEN3031 - Software Engineering</span>
            </div>
            <div className="legend-item">
              <span className="legend-color legend-purple"></span>
              <span>COP3530 - Data Structures</span>
            </div>
            <div className="legend-item">
              <span className="legend-color legend-yellow"></span>
              <span>COP4020 - Programming Languages</span>
            </div>
            <div className="legend-item">
              <span className="legend-color legend-red"></span>
              <span>COP4600 - Operating Systems</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
