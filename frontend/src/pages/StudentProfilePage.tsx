import { useAuth } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { toast } from "sonner";
import { updateUserProfile } from "../api/profile";
import { listAllCourses } from "../api/courses";
import { getStudentCourses, updateStudentCourses } from "../api/studentCourses";
import ufLogo from "../images/UF Logo.png";
import whiteNotificationIcon from "../images/White Notification Icon.png";
import whiteProfileIcon from "../images/White Profile Icon.png";

const COURSE_COLORS = ['green', 'purple', 'yellow', 'red', 'orange'];
const YEAR_OPTIONS = ['Freshman', 'Sophomore', 'Junior', 'Senior', 'Graduate', 'PhD'];

function getCourseColor(course: Course): string {
  if (course.color) {
    return course.color;
  }
  return COURSE_COLORS[course.id % COURSE_COLORS.length];
}

interface Course {
  id: number;
  code: string;
  name: string;
  color?: string;
}

export default function StudentProfilePage() {
  const { user, logout, updateUser } = useAuth();
  const navigate = useNavigate();

  // Profile state
  const [isEditMode, setIsEditMode] = useState(false);
  const [isAddCourseModalOpen, setIsAddCourseModalOpen] = useState(false);
  const [availableCourses, setAvailableCourses] = useState<Course[]>([]);
  const [selectedCourseId, setSelectedCourseId] = useState<number | "">(0);
  const [editName, setEditName] = useState(user?.username || "Student");
  const [editMajor, setEditMajor] = useState(user?.major || "Computer Science");
  const [editYear, setEditYear] = useState(user?.year || "Junior");
  const [editCourses, setEditCourses] = useState<Course[]>([]);

  // Initialize state from user context when component mounts or user changes
  useEffect(() => {
    if (user) {
      setEditName(user.username);
      setEditMajor(user.major || "Computer Science");
      setEditYear(user.year || "Junior");
    }
  }, [user]);

  // Fetch available courses from the API
  useEffect(() => {
    const fetchCourses = async () => {
      try {
        const courses = await listAllCourses();
        setAvailableCourses(courses);
        if (courses.length > 0) {
          setSelectedCourseId(courses[0].id);
        }
      } catch (error) {
        toast.error("Failed to load available courses");
      }
    };
    fetchCourses();
  }, []);

  // Fetch student's enrolled courses from the API
  useEffect(() => {
    const fetchStudentCourses = async () => {
      try {
        const courses = await getStudentCourses();
        setEditCourses(courses);
      } catch (error) {
        toast.error("Failed to load your courses");
      }
    };
    fetchStudentCourses();
  }, []);

  // Mock student data - in the future, fetch from API
  const studentData = {
    email: user?.email || "student@ufl.edu",
    major: editMajor,
    year: editYear,
    name: editName,
    courses: editCourses,
  };

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const handleEditProfile = () => {
    setIsEditMode(true);
  };

  const handleAddCourse = () => {
    setIsAddCourseModalOpen(true);
  };

  const handleConfirmAddCourse = () => {
    if (selectedCourseId !== "") {
      const selectedCourse = availableCourses.find(c => c.id === selectedCourseId);
      if (selectedCourse && !editCourses.some(c => c.id === selectedCourse.id)) {
        setEditCourses([...editCourses, selectedCourse]);
      }
    }
    setIsAddCourseModalOpen(false);
    if (availableCourses.length > 0) {
      setSelectedCourseId(availableCourses[0].id);
    }
  };

  const handleCloseAddCourseModal = () => {
    setIsAddCourseModalOpen(false);
    if (availableCourses.length > 0) {
      setSelectedCourseId(availableCourses[0].id);
    }
  };

  const handleRemoveCourse = (id: number) => {
    setEditCourses(editCourses.filter(c => c.id !== id));
  };

  const handleSave = async () => {
    // Update user with new name, major, and year
    if (user) {
      try {
        const response = await updateUserProfile(user.id, {
          username: editName,
          major: editMajor,
          year: editYear,
        });

        updateUser({
          ...user,
          username: response.username,
          major: response.major,
          year: response.year,
        });

        // Save courses to database
        const courseIds = editCourses.map(c => c.id);
        await updateStudentCourses(courseIds);

        toast.success("Profile updated successfully!");
      } catch (error) {
        const message = error instanceof Error ? error.message : "Failed to update profile";
        toast.error(message);
      }
    }
    setIsEditMode(false);
  };

  const handleCancel = async () => {
    // Reset to current user context values
    if (user) {
      setEditName(user.username);
      setEditMajor(user.major || "Computer Science");
      setEditYear(user.year || "Junior");
    }
    // Reload student courses from database
    try {
      const courses = await getStudentCourses();
      setEditCourses(courses);
    } catch (error) {
      toast.error("Failed to reload courses");
    }
    setIsEditMode(false);
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
            <button className="nav-tab" onClick={() => navigate("/student")}>
              Dashboard
            </button>
            <button className="nav-tab" onClick={() => navigate("/student/my-courses")}>
              My Courses
            </button>
          </div>
        </div>
        <div className="navbar-right">
          <div className="notification-icon">
            <img src={whiteNotificationIcon} alt="Notifications" />
          </div>
          <button className="profile-icon" onClick={() => {/* Already on profile page */}}>
            <img src={whiteProfileIcon} alt="Profile" />
          </button>
        </div>
      </nav>

      {/* Main Content */}
      <div className="student-profile-content">
        <div className="profile-header">
          <h1 className="profile-title">Welcome to your Profile!</h1>
        </div>

        {!isEditMode ? (
          // View Mode
          <div className="profile-card">
            <div className="profile-info-section">
              <h2 className="profile-name">{studentData.name}</h2>

              <div className="profile-details">
                <div className="profile-detail-item">
                  <span className="profile-detail-label">Email:</span>
                  <span className="profile-detail-value">{studentData.email}</span>
                </div>

                <div className="profile-detail-item">
                  <span className="profile-detail-label">Major:</span>
                  <span className="profile-detail-value">{studentData.major}</span>
                </div>

                <div className="profile-detail-item">
                  <span className="profile-detail-label">Year:</span>
                  <span className="profile-detail-value">{studentData.year}</span>
                </div>
              </div>

              <div className="profile-courses-section">
                <h3 className="profile-courses-title">Courses:</h3>
                <div className="profile-courses-list">
                  {studentData.courses.map((course) => (
                    <div key={course.id} className="profile-course-item">
                      <div
                        className="profile-course-dot"
                        data-color={getCourseColor(course)}
                      ></div>
                      <span className="profile-course-text">
                        {course.code} - {course.name}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="profile-actions">
              <button className="profile-edit-btn" onClick={handleEditProfile}>
                Edit Profile
              </button>
              <button className="profile-logout-btn" onClick={handleLogout}>
                Log Out
              </button>
            </div>
          </div>
        ) : (
          // Edit Mode
          <div className="profile-card profile-card-edit">
            <div className="profile-edit-form">
              <div className="profile-form-section">
                <label className="profile-form-label">Name</label>
                <input
                  type="text"
                  className="profile-form-input"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                  placeholder="Enter your name"
                />
              </div>

              <div className="profile-form-section">
                <label className="profile-form-label">Email</label>
                <input
                  type="email"
                  className="profile-form-input"
                  value={studentData.email}
                  disabled
                  placeholder="Email (read-only)"
                />
              </div>

              <div className="profile-form-section">
                <label className="profile-form-label">Major</label>
                <input
                  type="text"
                  className="profile-form-input"
                  value={editMajor}
                  onChange={(e) => setEditMajor(e.target.value)}
                  placeholder="Enter your major"
                />
              </div>

              <div className="profile-form-section">
                <label className="profile-form-label">Year</label>
                <select
                  className="profile-form-input profile-form-select"
                  value={editYear}
                  onChange={(e) => setEditYear(e.target.value)}
                >
                  {YEAR_OPTIONS.map((year) => (
                    <option key={year} value={year}>
                      {year}
                    </option>
                  ))}
                </select>
              </div>

              <div className="profile-form-section">
                <div className="profile-courses-header">
                  <label className="profile-form-label">Courses:</label>
                  <button className="profile-add-course-btn" onClick={handleAddCourse}>
                    + Add Course
                  </button>
                </div>
                <div className="profile-courses-list">
                  {editCourses.map((course) => (
                    <div key={course.id} className="profile-edit-course-item">
                      <div className="profile-edit-course-info">
                        <div
                          className="profile-course-dot"
                          data-color={getCourseColor(course)}
                        ></div>
                        <span className="profile-course-text">
                          {course.code} - {course.name}
                        </span>
                      </div>
                      <button
                        className="profile-remove-course-btn"
                        onClick={() => handleRemoveCourse(course.id)}
                      >
                        Remove
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="profile-edit-buttons">
              <button className="profile-save-btn" onClick={handleSave}>
                Save
              </button>
              <button className="profile-cancel-btn" onClick={handleCancel}>
                Cancel
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Add Course Modal */}
      {isAddCourseModalOpen && (
        <div className="modal-overlay">
          <div className="add-course-modal">
            <div className="modal-header-blue">
              <h2 className="modal-title-white">Add Course</h2>
              <button className="modal-close-btn-white" onClick={handleCloseAddCourseModal}>
                ✕
              </button>
            </div>
            <div className="modal-body">
              <label className="modal-subtitle">Select Available Course</label>
              <select
                className="modal-select-blue"
                value={selectedCourseId}
                onChange={(e) => setSelectedCourseId(Number(e.target.value))}
              >
                {availableCourses.filter(course => !editCourses.some(c => c.id === course.id)).map((course) => (
                  <option key={course.id} value={course.id}>
                    {course.code} - {course.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="modal-footer-center">
              <button className="modal-add-btn-blue" onClick={handleConfirmAddCourse}>
                + Add Course
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
