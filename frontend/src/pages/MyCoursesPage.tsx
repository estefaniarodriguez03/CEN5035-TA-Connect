import { useState, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import ufLogo from "../images/UF Logo.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";
import orangeDateIcon from "../images/Orange Date Icon.png";
import { toast } from "sonner";
import { listAllCourses, type TACourse } from "../api/courses";
import { listOfficeHoursByCourse, DAY_NAMES, type OfficeHour } from "../api/officeHours";
import {
  listStudentSchedule,
  addToStudentSchedule,
  removeFromStudentSchedule,
  type ScheduleEntry,
} from "../api/studentSchedule";

const COURSE_COLORS = ['green', 'purple', 'yellow', 'red', 'orange'];

function formatTime(time: string): string {
  const [hours, minutes] = time.split(':');
  const hour = parseInt(hours);
  const ampm = hour >= 12 ? 'PM' : 'AM';
  const displayHour = hour > 12 ? hour - 12 : hour === 0 ? 12 : hour;
  return `${displayHour}:${minutes} ${ampm}`;
}

function getCourseColor(courseID: number): string {
  return COURSE_COLORS[courseID % COURSE_COLORS.length];
}

export default function MyCoursesPage() {
  const navigate = useNavigate();

  const [viewMode, setViewMode] = useState<'weekly' | 'today'>('weekly');
  const [filterCourse, setFilterCourse] = useState<string>('all');
  const [filterTA, setFilterTA] = useState<string>('all');

  const [schedule, setSchedule] = useState<ScheduleEntry[]>([]);
  const [loadingSchedule, setLoadingSchedule] = useState(true);

  const [showAddModal, setShowAddModal] = useState(false);
  const [allCourses, setAllCourses] = useState<TACourse[]>([]);
  const [selectedCourseID, setSelectedCourseID] = useState<number | null>(null);
  const [courseOfficeHours, setCourseOfficeHours] = useState<OfficeHour[]>([]);
  const [selectedOfficeHourID, setSelectedOfficeHourID] = useState<number | null>(null);
  const [selectedTAID, setSelectedTAID] = useState<number | null>(null);
  const [loadingCourses, setLoadingCourses] = useState(false);
  const [loadingOfficeHours, setLoadingOfficeHours] = useState(false);

  const days = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'];

  // Load student schedule on mount
  useEffect(() => {
    void (async () => {
      try {
        const entries = await listStudentSchedule();
        setSchedule(entries);
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to load schedule');
      } finally {
        setLoadingSchedule(false);
      }
    })();
  }, []);

  // Load all courses when modal opens
  useEffect(() => {
    if (!showAddModal) return;
    void (async () => {
      setLoadingCourses(true);
      try {
        const courses = await listAllCourses();
        const real = courses.filter(
          (c) => !c.code.startsWith('AUTO-') && !c.code.startsWith('LEGACY-')
        );
        setAllCourses(real);
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to load courses');
      } finally {
        setLoadingCourses(false);
      }
    })();
  }, [showAddModal]);

  // Load office hours when a course is selected in the modal
  useEffect(() => {
    if (!selectedCourseID) {
      setCourseOfficeHours([]);
      setSelectedOfficeHourID(null);
      return;
    }
    void (async () => {
      setLoadingOfficeHours(true);
      try {
        const hours = await listOfficeHoursByCourse(selectedCourseID);
        setCourseOfficeHours(hours);
        setSelectedOfficeHourID(null);
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to load office hours');
      } finally {
        setLoadingOfficeHours(false);
      }
    })();
  }, [selectedCourseID]);

  const handleAddToSchedule = async () => {
    if (!selectedOfficeHourID) {
      toast.error('Please select a TA and time slot');
      return;
    }
    try {
      await addToStudentSchedule(selectedOfficeHourID);
      const entries = await listStudentSchedule();
      setSchedule(entries);
      toast.success('Added to your schedule');
      setShowAddModal(false);
      setSelectedCourseID(null);
      setSelectedTAID(null);
      setSelectedOfficeHourID(null);
      setCourseOfficeHours([]);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to add to schedule');
    }
  };

  const handleRemoveFromSchedule = async (id: number) => {
    try {
      await removeFromStudentSchedule(id);
      setSchedule((prev) => prev.filter((e) => e.id !== id));
      toast.success('Removed from your schedule');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to remove from schedule');
    }
  };

  const handleJoinQueue = (entry: ScheduleEntry) => {
    const timeRange = `${formatTime(entry.start_time)} - ${formatTime(entry.end_time)}`;
    navigate('/student', {
      state: {
        autoSelectLabel: `${entry.course_code} – ${entry.course_name} – ${entry.ta_username} (${timeRange})`,
      },
    });
  };

  const uniqueCourses = useMemo(() =>
    Array.from(new Map(schedule.map((e) => [e.course_id, `${e.course_code} - ${e.course_name}`])).values()),
    [schedule]
  );

  const uniqueTAs = useMemo(() =>
    Array.from(new Set(schedule.map((e) => e.ta_username))),
    [schedule]
  );

  const todayDayIndex = new Date().getDay();
  const todayDayName = DAY_NAMES[todayDayIndex];

  const filteredSchedule = useMemo(() => {
    return schedule.filter((e) => {
      const courseMatch = filterCourse === 'all' || `${e.course_code} - ${e.course_name}` === filterCourse;
      const taMatch = filterTA === 'all' || e.ta_username === filterTA;
      return courseMatch && taMatch;
    });
  }, [schedule, filterCourse, filterTA]);

  const uniqueTAsForCourse = useMemo(() => {
    const seen = new Map<number, string>();
    courseOfficeHours.forEach((oh) => {
      if (oh.ta_username && !seen.has(oh.ta_id)) {
        seen.set(oh.ta_id, oh.ta_username);
      }
    });
    return Array.from(seen.entries()).map(([id, name]) => ({ id, name }));
  }, [courseOfficeHours]);

  const slotsForSelectedTA = useMemo(() => {
    if (!selectedTAID) return [];
    return courseOfficeHours.filter((oh) => oh.ta_id === selectedTAID);
  }, [courseOfficeHours, selectedTAID]);

  const getWeekRange = () => {
    const today = new Date();
    const dayOfWeek = today.getDay();
    const daysToMonday = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
    const monday = new Date(today);
    monday.setDate(monday.getDate() - daysToMonday);
    const friday = new Date(monday);
    friday.setDate(friday.getDate() + 4);
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    return `Week of ${monthNames[monday.getMonth()]} ${monday.getDate()} - ${monthNames[friday.getMonth()]} ${friday.getDate()}`;
  };

  const getTodayDate = () => {
    const today = new Date();
    const monthNames = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
    return `${todayDayName}, ${monthNames[today.getMonth()]} ${today.getDate()}`;
  };

  const closeModal = () => {
    setShowAddModal(false);
    setSelectedCourseID(null);
    setSelectedTAID(null);
    setSelectedOfficeHourID(null);
    setCourseOfficeHours([]);
  };

  if (loadingSchedule) {
    return (
      <div className="my-courses-page">
        <p style={{ padding: '2rem' }}>Loading your schedule...</p>
      </div>
    );
  }

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
            <button className="nav-tab" onClick={() => navigate('/student')}>Dashboard</button>
            <button className="nav-tab active">My Courses</button>
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
      <div className="my-courses-content">
        <div className="my-courses-header">
          <h1 className="my-courses-title">My Courses - Office Hours</h1>
          <button className="add-office-hours-btn" onClick={() => setShowAddModal(true)}>
            + Add Office Hours
          </button>
        </div>

        {/* View and Filter Controls */}
        <div className="controls-section">
          <div className="view-controls">
            <span className="controls-label">View</span>
            <div className="view-buttons">
              <button
                className={`view-btn ${viewMode === 'weekly' ? 'active' : ''}`}
                onClick={() => setViewMode('weekly')}
              >
                Weekly
              </button>
              <button
                className={`view-btn ${viewMode === 'today' ? 'active' : ''}`}
                onClick={() => setViewMode('today')}
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
        {viewMode === 'weekly' && (
          <div className="weekly-view-section">
            <div className="view-header">
              <h2 className="view-title">This Week's Office Hours</h2>
              <div className="view-date">
                <img src={orangeDateIcon} alt="Calendar" className="calendar-icon" />
                {getWeekRange()}
              </div>
            </div>

            <div className="schedule-grid-container">
              {days.map((day) => {
                const dayIndex = DAY_NAMES.indexOf(day);
                return (
                  <div key={day} className="schedule-column">
                    <div className="day-header">{day}</div>
                    <div className="day-slots">
                      {filteredSchedule
                        .filter((e) => e.day_of_week === dayIndex)
                        .map((e) => (
                          <div key={e.id} className={`time-slot slot-${getCourseColor(e.course_id)}`}>
                            <button
                              className="delete-btn"
                              onClick={() => void handleRemoveFromSchedule(e.id)}
                              title="Remove from schedule"
                            >
                              ×
                            </button>
                            <div className="slot-time">
                              {formatTime(e.start_time)} - {formatTime(e.end_time)}
                            </div>
                            <div className="slot-ta">{e.ta_username}</div>
                            <div className="slot-course">{e.course_code}</div>
                          </div>
                        ))}
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Dynamic Course Legend */}
            <div className="course-legend">
              {Array.from(new Map(schedule.map((e) => [e.course_id, e])).values()).map((e) => (
                <div key={e.course_id} className="legend-item">
                  <span className={`legend-color legend-${getCourseColor(e.course_id)}`}></span>
                  <span>{e.course_code} - {e.course_name}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Today View */}
        {viewMode === 'today' && (
          <div className="today-view-section">
            <div className="view-header">
              <h2 className="view-title">Today's Office Hours</h2>
              <div className="view-date">
                <img src={orangeDateIcon} alt="Calendar" className="calendar-icon" />
                {getTodayDate()}
              </div>
            </div>

            <div className="today-courses-grid">
              {filteredSchedule
                .filter((e) => e.day_of_week === todayDayIndex)
                .map((e) => (
                  <div key={e.id} className="today-course-card">
                    <button
                      className="delete-btn-today"
                      onClick={() => void handleRemoveFromSchedule(e.id)}
                      title="Remove from schedule"
                    >
                      ×
                    </button>
                    <div className={`course-badge badge-${getCourseColor(e.course_id)}`}>
                      {e.course_code}
                    </div>
                    <h3 className="course-card-title">{e.course_name}</h3>
                    <div className="course-ta-label">TA: {e.ta_username}</div>
                    <div className="course-office-hours">
                      <div className="office-hours-label">Today's Office Hours:</div>
                      <div className="office-hours-time">
                        {formatTime(e.start_time)} - {formatTime(e.end_time)}
                      </div>
                    </div>
                    <button
                      className="join-queue-today-btn"
                      onClick={() => handleJoinQueue(e)}
                    >
                      Join Queue for {e.course_code}
                    </button>
                  </div>
                ))}

              {filteredSchedule.filter((e) => e.day_of_week === todayDayIndex).length === 0 && (
                <div className="no-results-message">
                  <p>No office hours scheduled for today.</p>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Add Office Hours Modal */}
      {showAddModal && (
        <div className="modal-overlay" onClick={closeModal}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2 className="modal-title">Add Office Hours</h2>
              <button className="modal-close-btn" onClick={closeModal}>✕</button>
            </div>

            <div className="modal-body">

              {/* Step 1 — Course */}
              <div className="form-group">
                <label className="form-label">Course</label>
                {loadingCourses ? (
                  <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>Loading courses...</p>
                ) : (
                  <select
                    className="form-select"
                    value={selectedCourseID ?? ''}
                    onChange={(e) => {
                      setSelectedCourseID(parseInt(e.target.value) || null);
                      setSelectedTAID(null);
                      setSelectedOfficeHourID(null);
                    }}
                  >
                    <option value="">Select a course...</option>
                    {allCourses.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.code}{c.name ? ` - ${c.name}` : ''}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              {/* Step 2 — TA */}
              {selectedCourseID && (
                <div className="form-group">
                  <label className="form-label">TA</label>
                  {loadingOfficeHours ? (
                    <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>Loading TAs...</p>
                  ) : uniqueTAsForCourse.length === 0 ? (
                    <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>
                      No TAs have scheduled office hours for this course yet.
                    </p>
                  ) : (
                    <select
                      className="form-select"
                      value={selectedTAID ?? ''}
                      onChange={(e) => {
                        setSelectedTAID(parseInt(e.target.value) || null);
                        setSelectedOfficeHourID(null);
                      }}
                    >
                      <option value="">Select a TA...</option>
                      {uniqueTAsForCourse.map((ta) => (
                        <option key={ta.id} value={ta.id}>{ta.name}</option>
                      ))}
                    </select>
                  )}
                </div>
              )}

              {/* Step 3 — Time slot */}
              {selectedTAID && (
                <div className="form-group">
                  <label className="form-label">Time Slot</label>
                  {slotsForSelectedTA.length === 0 ? (
                    <p style={{ color: '#6b7280', fontSize: '0.9rem' }}>
                      No time slots available for this TA.
                    </p>
                  ) : (
                    <select
                      className="form-select"
                      value={selectedOfficeHourID ?? ''}
                      onChange={(e) => setSelectedOfficeHourID(parseInt(e.target.value) || null)}
                    >
                      <option value="">Select a time slot...</option>
                      {slotsForSelectedTA.map((oh) => (
                        <option key={oh.id} value={oh.id}>
                          {DAY_NAMES[oh.day_of_week]} — {formatTime(oh.start_time)} to {formatTime(oh.end_time)}
                          {oh.location ? ` (${oh.location})` : ''}
                        </option>
                      ))}
                    </select>
                  )}
                </div>
              )}
            </div>

            <div className="modal-footer">
              <button className="modal-cancel-btn" onClick={closeModal}>Cancel</button>
              <button
                className="modal-submit-btn"
                onClick={() => void handleAddToSchedule()}
                disabled={!selectedOfficeHourID}
              >
                Add to Schedule
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}