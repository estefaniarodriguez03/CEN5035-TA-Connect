import { useAuth } from "../context/AuthContext";
import { useState, useMemo } from "react";
import ufLogo from "../images/UF Logo.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";
import orangeDateIcon from "../images/Orange Date Icon.png";
import { useNavigate } from "react-router-dom";

interface CourseOfficeHour {
  id: number;
  courseCode: string;
  courseName: string;
  taName: string;
  day: string;
  time: string;
  color: string;
  timeStart: string;
  timeEnd: string;
}

interface TodayOfficeHour {
  id: number;
  courseCode: string;
  courseName: string;
  taName: string;
  time: string;
  color: string;
  timeStart: string;
  timeEnd: string;
}

export default function MyCoursesPage() {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const [viewMode, setViewMode] = useState<"weekly" | "today">("weekly");
  const [filterCourse, setFilterCourse] = useState<string>("all");
  const [filterTA, setFilterTA] = useState<string>("all");
  const [showAddModal, setShowAddModal] = useState(false);
  const [modalCourse, setModalCourse] = useState<string>("");
  const [modalDay, setModalDay] = useState<string>("Monday");
  const [modalStartTime, setModalStartTime] = useState<string>("09");
  const [modalStartPeriod, setModalStartPeriod] = useState<string>("AM");
  const [modalEndTime, setModalEndTime] = useState<string>("11");
  const [modalEndPeriod, setModalEndPeriod] = useState<string>("AM");

  const defaultOfficeHours: CourseOfficeHour[] = [
    { id: 1, courseCode: "CEN3031", courseName: "Software Engineering", taName: "Estefania Rodriguez", day: "Monday", time: "9:00 AM - 11:00 AM", color: "green", timeStart: "09:00", timeEnd: "11:00" },
    { id: 2, courseCode: "COP4020", courseName: "Programming Languages", taName: "Raghav Nanjappan", day: "Monday", time: "11:00 AM - 1:00 PM", color: "yellow", timeStart: "11:00", timeEnd: "13:00" },
    { id: 3, courseCode: "COP3530", courseName: "Data Structures", taName: "Sara Waters", day: "Monday", time: "2:00 PM - 4:00 PM", color: "purple", timeStart: "14:00", timeEnd: "16:00" },
    { id: 4, courseCode: "COP4600", courseName: "Operating Systems", taName: "John Spurrier", day: "Tuesday", time: "10:00 AM - 12:00 PM", color: "red", timeStart: "10:00", timeEnd: "12:00" },
    { id: 5, courseCode: "COP3530", courseName: "Data Structures", taName: "Sara Waters", day: "Tuesday", time: "1:00 PM - 3:00 PM", color: "purple", timeStart: "13:00", timeEnd: "15:00" },
    { id: 6, courseCode: "CEN3031", courseName: "Software Engineering", taName: "Estefania Rodriguez", day: "Tuesday", time: "3:00 PM - 5:00 PM", color: "green", timeStart: "15:00", timeEnd: "17:00" },
    { id: 7, courseCode: "COP4020", courseName: "Programming Languages", taName: "Raghav Nanjappan", day: "Wednesday", time: "9:00 AM - 11:00 AM", color: "yellow", timeStart: "09:00", timeEnd: "11:00" },
    { id: 8, courseCode: "COP4600", courseName: "Operating Systems", taName: "John Spurrier", day: "Wednesday", time: "11:00 AM - 1:00 PM", color: "red", timeStart: "11:00", timeEnd: "13:00" },
    { id: 9, courseCode: "COP3530", courseName: "Data Structures", taName: "Sara Waters", day: "Wednesday", time: "2:00 PM - 4:00 PM", color: "purple", timeStart: "14:00", timeEnd: "16:00" },
    { id: 10, courseCode: "CEN3031", courseName: "Software Engineering", taName: "Estefania Rodriguez", day: "Thursday", time: "10:00 AM - 12:00 PM", color: "green", timeStart: "10:00", timeEnd: "12:00" },
    { id: 11, courseCode: "COP4020", courseName: "Programming Languages", taName: "Raghav Nanjappan", day: "Thursday", time: "1:00 PM - 3:00 PM", color: "yellow", timeStart: "13:00", timeEnd: "15:00" },
    { id: 12, courseCode: "COP4600", courseName: "Operating Systems", taName: "John Spurrier", day: "Thursday", time: "3:00 PM - 5:00 PM", color: "red", timeStart: "15:00", timeEnd: "17:00" },
    
    { id: 13, courseCode: "COP3530", courseName: "Data Structures", taName: "Sara Waters", day: "Friday", time: "9:00 AM - 11:00 AM", color: "purple", timeStart: "09:00", timeEnd: "11:00" },
    { id: 14, courseCode: "CEN3031", courseName: "Software Engineering", taName: "Estefania Rodriguez", day: "Friday", time: "11:00 AM - 1:00 PM", color: "green", timeStart: "11:00", timeEnd: "13:00" },
    { id: 15, courseCode: "COP4020", courseName: "Programming Languages", taName: "Raghav Nanjappan", day: "Friday", time: "2:00 PM - 4:00 PM", color: "yellow", timeStart: "14:00", timeEnd: "16:00" },
  ];

  const [allOfficeHours, setAllOfficeHours] = useState<CourseOfficeHour[]>(defaultOfficeHours);
  
  const defaultTodayOfficeHours: TodayOfficeHour[] = [
    { id: 1, courseCode: "CEN3031", courseName: "Software Engineering", taName: "Estefania Rodriguez", time: "3:00 PM - 5:00 PM", color: "green", timeStart: "15:00", timeEnd: "17:00" },
    { id: 2, courseCode: "COP3530", courseName: "Data Structures", taName: "Sara Waters", time: "1:00 PM - 3:00 PM", color: "purple", timeStart: "13:00", timeEnd: "15:00" },
    { id: 3, courseCode: "COP4020", courseName: "Programming Languages", taName: "Raghav Nanjappan", time: "No office hours today", color: "yellow", timeStart: "", timeEnd: "" },
    { id: 4, courseCode: "COP4600", courseName: "Operating Systems", taName: "John Spurrier", time: "10:00 AM - 12:00 PM", color: "red", timeStart: "10:00", timeEnd: "12:00" },
  ];

  const [todayOfficeHours, setTodayOfficeHours] = useState<TodayOfficeHour[]>(defaultTodayOfficeHours);

  const days = ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"];

  // Get unique courses and TAs for filter dropdowns
  const uniqueCourses = Array.from(new Set(allOfficeHours.map(oh => `${oh.courseCode} - ${oh.courseName}`)));
  const uniqueTAs = Array.from(new Set(allOfficeHours.map(oh => oh.taName)));

  // Filter office hours based on selected filters
  const filteredOfficeHours = useMemo(() => {
    return allOfficeHours.filter(oh => {
      const courseMatch = filterCourse === "all" || `${oh.courseCode} - ${oh.courseName}` === filterCourse;
      const taMatch = filterTA === "all" || oh.taName === filterTA;
      return courseMatch && taMatch;
    });
  }, [allOfficeHours, filterCourse, filterTA]);

  // Filter today's office hours based on selected filters
  const filteredTodayOfficeHours = useMemo(() => {
    return todayOfficeHours.filter(oh => {
      const courseMatch = filterCourse === "all" || `${oh.courseCode} - ${oh.courseName}` === filterCourse;
      const taMatch = filterTA === "all" || oh.taName === filterTA;
      return courseMatch && taMatch;
    });
  }, [todayOfficeHours, filterCourse, filterTA]);

  const getWeekRange = () => {
    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysToMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const monday = new Date(today);
    monday.setDate(monday.getDate() - daysToMonday);
    
    const friday = new Date(monday);
    friday.setDate(friday.getDate() + 4);
    
    const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
    const mondayMonth = monthNames[monday.getMonth()];
    const fridayMonth = monthNames[friday.getMonth()];
    
    return `Week of ${mondayMonth} ${monday.getDate()} - ${fridayMonth} ${friday.getDate()}`;
  };

  const getTodayDate = () => {
    const today = new Date();
    const monthNames = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];
    const dayNames = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
    return `${dayNames[today.getDay()]}, ${monthNames[today.getMonth()]} ${today.getDate()}`;
  };

  const handleNavigation = (page: string) => {
    if (page === "dashboard") {
      navigate("/student");
    }
  };

  const getCourseColor = (courseCode: string): string => {
    switch (courseCode) {
      case "CEN3031":
        return "green";
      case "COP3530":
        return "purple";
      case "COP4020":
        return "yellow";
      case "COP4600":
        return "red";
      default:
        return "green";
    }
  };

  const getTAName = (courseCode: string): string => {
    switch (courseCode) {
      case "CEN3031":
        return "Estefania Rodriguez";
      case "COP3530":
        return "Sara Waters";
      case "COP4020":
        return "Raghav Nanjappan";
      case "COP4600":
        return "John Spurrier";
      default:
        return "You";
    }
  };

  const formatTimeDisplay = (hour: string, period: string): string => {
    return `${hour}:00 ${period}`;
  };

  const convertTo24Hour = (hour: string, period: string): string => {
    let hourNum = parseInt(hour);
    if (period === "PM" && hourNum !== 12) {
      hourNum += 12;
    } else if (period === "AM" && hourNum === 12) {
      hourNum = 0;
    }
    return hourNum.toString().padStart(2, "0") + ":00";
  };

  const handleAddOfficeHours = () => {
    if (!modalCourse || !modalDay) {
      alert("Please select a course and day");
      return;
    }

    const [courseCode, courseName] = modalCourse.split(" - ");
    const color = getCourseColor(courseCode);
    const taName = getTAName(courseCode);
    const startTimeDisplay = formatTimeDisplay(modalStartTime, modalStartPeriod);
    const endTimeDisplay = formatTimeDisplay(modalEndTime, modalEndPeriod);
    const timeDisplay = `${startTimeDisplay} - ${endTimeDisplay}`;
    const timeStart = convertTo24Hour(modalStartTime, modalStartPeriod);
    const timeEnd = convertTo24Hour(modalEndTime, modalEndPeriod);

    const newId = Math.max(...allOfficeHours.map(oh => oh.id), ...todayOfficeHours.map(oh => oh.id)) + 1;

    // Add to weekly view
    const newWeeklyEntry: CourseOfficeHour = {
      id: newId,
      courseCode,
      courseName,
      taName,
      day: modalDay,
      time: timeDisplay,
      color,
      timeStart,
      timeEnd,
    };

    setAllOfficeHours([...allOfficeHours, newWeeklyEntry]);

    // Check if the day is today
    const today = new Date();
    const dayNames = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
    const todayName = dayNames[today.getDay()];

    if (modalDay === todayName) {
      const newTodayEntry: TodayOfficeHour = {
        id: newId,
        courseCode,
        courseName,
        taName,
        time: timeDisplay,
        color,
        timeStart,
        timeEnd,
      };
      setTodayOfficeHours([...todayOfficeHours, newTodayEntry]);
    }

    // Reset and close modal
    setModalCourse("");
    setModalDay("Monday");
    setModalStartTime("09");
    setModalStartPeriod("AM");
    setModalEndTime("11");
    setModalEndPeriod("AM");
    setShowAddModal(false);
  };

  const handleDeleteOfficeHours = (id: number) => {
    // Remove from weekly view
    setAllOfficeHours(allOfficeHours.filter(oh => oh.id !== id));
    
    // In today view, update to "No office hours today" instead of removing
    setTodayOfficeHours(todayOfficeHours.map(oh => 
      oh.id === id 
        ? { ...oh, time: "No office hours today", timeStart: "", timeEnd: "" }
        : oh
    ));
  };

  return (
    <div className="my-courses-page">
      {/* Navigation Bar */}
      <nav className="navbar">
        <div className="navbar-left">
          <div className="logo">
            <img src={ufLogo} alt="UF Logo" className="logo-icon" />
            <span className="logo-text">TA Connect</span>
          </div>
          <div className="nav-tabs">
            <button className="nav-tab" onClick={() => handleNavigation("dashboard")}>Dashboard</button>
            <button className="nav-tab active">My Courses</button>
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
      <div className="my-courses-content">
        <div className="my-courses-header">
          <h1 className="my-courses-title">My Courses - Office Hours</h1>
          <button className="add-office-hours-btn" onClick={() => setShowAddModal(true)}>+ Add Office Hours</button>
        </div>

        {/* View and Filter Controls */}
        <div className="controls-section">
          <div className="view-controls">
            <span className="controls-label">View</span>
            <div className="view-buttons">
              <button 
                className={`view-btn ${viewMode === "weekly" ? "active" : ""}`}
                onClick={() => setViewMode("weekly")}
              >
                Weekly
              </button>
              <button 
                className={`view-btn ${viewMode === "today" ? "active" : ""}`}
                onClick={() => setViewMode("today")}
              >
                Today
              </button>
            </div>
          </div>

          <div className="filter-controls">
            <div className="filter-group">
              <label className="filter-label">Filter by Course</label>
              <select 
                className="filter-dropdown"
                value={filterCourse}
                onChange={(e) => setFilterCourse(e.target.value)}
              >
                <option value="all">All Courses</option>
                {uniqueCourses.map((course, idx) => (
                  <option key={idx} value={course}>{course}</option>
                ))}
              </select>
            </div>

            <div className="filter-group">
              <label className="filter-label">Filter by TA</label>
              <select 
                className="filter-dropdown"
                value={filterTA}
                onChange={(e) => setFilterTA(e.target.value)}
              >
                <option value="all">All TAs</option>
                {uniqueTAs.map((ta, idx) => (
                  <option key={idx} value={ta}>{ta}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Weekly View */}
        {viewMode === "weekly" && (
          <div className="weekly-view-section">
            <div className="view-header">
              <h2 className="view-title">This Week's Office Hours</h2>
              <div className="view-date">
                <img src={orangeDateIcon} alt="Calendar" className="calendar-icon" />
                {getWeekRange()}
              </div>
            </div>

            <div className="schedule-grid-container">
              {days.map((day) => (
                <div key={day} className="schedule-column">
                  <div className="day-header">{day}</div>
                  <div className="day-slots">
                    {filteredOfficeHours
                      .filter((oh) => oh.day === day)
                      .map((oh) => (
                        <div key={oh.id} className={`time-slot slot-${oh.color}`}>
                          <button 
                            className="delete-btn"
                            onClick={() => handleDeleteOfficeHours(oh.id)}
                            title="Delete office hours"
                          >
                            ×
                          </button>
                          <div className="slot-time">{oh.time}</div>
                          <div className="slot-ta">{oh.taName}</div>
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
        )}

        {/* Today View */}
        {viewMode === "today" && (
          <div className="today-view-section">
            <div className="view-header">
              <h2 className="view-title">Today's Office Hours</h2>
              <div className="view-date">
                <img src={orangeDateIcon} alt="Calendar" className="calendar-icon" />
                {getTodayDate()}
              </div>
            </div>

            <div className="today-courses-grid">
              {filteredTodayOfficeHours.map((oh) => (
                <div key={oh.id} className="today-course-card">
                  <button 
                    className="delete-btn-today"
                    onClick={() => handleDeleteOfficeHours(oh.id)}
                    title="Delete office hours"
                  >
                    ×
                  </button>
                  <div className={`course-badge badge-${oh.color}`}>{oh.courseCode}</div>
                  <h3 className="course-card-title">{oh.courseName}</h3>
                  <div className="course-ta-label">TA: {oh.taName}</div>
                  
                  <div className="course-office-hours">
                    <div className="office-hours-label">Today's Office Hours:</div>
                    <div className={`office-hours-time ${oh.time.includes("No office") ? "no-hours" : ""}`}>
                      {oh.time}
                    </div>
                  </div>

                  <button className={`join-queue-today-btn ${oh.time.includes("No office") ? "disabled" : ""}`}
                    disabled={oh.time.includes("No office")}
                  >
                    {oh.time.includes("No office") ? "No queue today" : "Join Queue for " + oh.courseCode}
                  </button>
                </div>
              ))}
            </div>

            {filteredTodayOfficeHours.length === 0 && (
              <div className="no-results-message">
                <p>No office hours match your filters for today.</p>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Add Office Hours Modal */}
      {showAddModal && (
        <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">Add Office Hours</h2>
              <button className="modal-close-btn" onClick={() => setShowAddModal(false)}>✕</button>
            </div>

            <div className="modal-body">
              <div className="form-group">
                <label className="form-label">Course</label>
                <select 
                  className="form-select"
                  value={modalCourse}
                  onChange={(e) => setModalCourse(e.target.value)}
                >
                  <option value="">Select a course...</option>
                  {uniqueCourses.map((course, idx) => (
                    <option key={idx} value={course}>{course}</option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <label className="form-label">Day</label>
                <select 
                  className="form-select"
                  value={modalDay}
                  onChange={(e) => setModalDay(e.target.value)}
                >
                  {days.map((day) => (
                    <option key={day} value={day}>{day}</option>
                  ))}
                </select>
              </div>

              <div className="form-group">
                <label className="form-label">Start Time</label>
                <div className="time-inputs">
                  <select 
                    className="form-select time-select"
                    value={modalStartTime}
                    onChange={(e) => setModalStartTime(e.target.value)}
                  >
                    {Array.from({ length: 12 }, (_, i) => {
                      const hour = (i + 1).toString();
                      return (
                        <option key={hour} value={hour}>{hour}:00</option>
                      );
                    })}
                  </select>
                  <select 
                    className="form-select time-select"
                    value={modalStartPeriod}
                    onChange={(e) => setModalStartPeriod(e.target.value)}
                  >
                    <option value="AM">AM</option>
                    <option value="PM">PM</option>
                  </select>
                </div>
              </div>

              <div className="form-group">
                <label className="form-label">End Time</label>
                <div className="time-inputs">
                  <select 
                    className="form-select time-select"
                    value={modalEndTime}
                    onChange={(e) => setModalEndTime(e.target.value)}
                  >
                    {Array.from({ length: 12 }, (_, i) => {
                      const hour = (i + 1).toString();
                      return (
                        <option key={hour} value={hour}>{hour}:00</option>
                      );
                    })}
                  </select>
                  <select 
                    className="form-select time-select"
                    value={modalEndPeriod}
                    onChange={(e) => setModalEndPeriod(e.target.value)}
                  >
                    <option value="AM">AM</option>
                    <option value="PM">PM</option>
                  </select>
                </div>
              </div>
            </div>

            <div className="modal-footer">
              <button className="modal-cancel-btn" onClick={() => setShowAddModal(false)}>Cancel</button>
              <button className="modal-submit-btn" onClick={handleAddOfficeHours}>Add Office Hours</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
