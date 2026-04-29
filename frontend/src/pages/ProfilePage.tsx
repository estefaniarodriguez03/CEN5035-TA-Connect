import { useAuth } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { toast } from "sonner";
import TALayout from "../components/TALayout";
import { updateUserProfile } from "../api/profile";
import { addMyTACourse, listMyTACourses, removeMyTACourse, type TACourse } from "../api/courses";

interface Course extends TACourse {
  color: string;
}

const COURSE_COLORS = ["#FA4616", "#0021A5", "#E21C3D", "#F7A600", "#2D6A4F", "#6A2A60"];

const HEX_TO_COLOR: Record<string, string> = {
  "#FA4616": "orange",
  "#0021A5": "blue",
  "#E21C3D": "red",
  "#F7A600": "yellow",
  "#2D6A4F": "green",
  "#6A2A60": "purple",
};

const COLOR_NAME_TO_HEX: Record<string, string> = {
  "orange": "#FA4616",
  "blue": "#0021A5",
  "red": "#E21C3D",
  "yellow": "#F7A600",
  "green": "#2D6A4F",
  "purple": "#6A2A60",
};

function getColorName(hexColor: string): string {
  return HEX_TO_COLOR[hexColor.toUpperCase()] || "orange";
}

export default function ProfilePage() {
  const { user, logout, updateUser } = useAuth();
  const navigate = useNavigate();
  const [isEditMode, setIsEditMode] = useState(false);
  const [editName, setEditName] = useState(user?.username || "");
  const [editCourses, setEditCourses] = useState<Course[]>([]);
  const [originalCourses, setOriginalCourses] = useState<Course[]>([]);
  const [showAddCourseModal, setShowAddCourseModal] = useState(false);
  const [newCourseCode, setNewCourseCode] = useState("");
  const [newCourseName, setNewCourseName] = useState("");
  const [selectedColor, setSelectedColor] = useState("#FA4616");
  const [isSaving, setIsSaving] = useState(false);

  // Load courses from database on mount
  useEffect(() => {
    if (user?.role !== "ta") return;
    void (async () => {
      try {
        const courses = await listMyTACourses();
        // Convert database colors (orange, blue, etc.) to hex colors for UI
        const coursesWithHexColors = courses.map(course => {
          const hexColor = course.color ? COLOR_NAME_TO_HEX[course.color.toLowerCase()] ?? "#FA4616" : "#FA4616";
          return {
            ...course,
            color: hexColor
          };
        });
        setEditCourses(coursesWithHexColors);
        setOriginalCourses(coursesWithHexColors);
      } catch (error) {
        const message = error instanceof Error ? error.message : 'Failed to load courses';
        toast.error(message);
      }
    })();
  }, [user?.role]);

  // Only TAs can access this profile page
  if (user?.role !== "ta") {
    navigate("/student");
    return null;
  }

  const handleLogout = () => {
    logout();
  };

  const handleEditProfile = () => {
    setIsEditMode(true);
  };

  const handleSave = async () => {
    if (!user) return;

    setIsSaving(true);
    try {
      // Call API to update profile
      const response = await updateUserProfile(user.id, {
        username: editName,
      });

      // Identify removed courses
      const removedCourses = originalCourses.filter(
        (original) => !editCourses.find((current) => current.id === original.id)
      );

      // Delete removed courses from database
      for (const course of removedCourses) {
        try {
          await removeMyTACourse(course.id);
        } catch (error) {
          const message = error instanceof Error ? error.message : "Failed to remove course";
          toast.error(message);
        }
      }

      // Update local auth context with new user data
      updateUser({
        ...user,
        username: response.username,
      });

      // Update original courses to current state
      setOriginalCourses(editCourses);

      // Close edit mode
      setIsEditMode(false);

      toast.success("Profile updated successfully!");
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to update profile";
      toast.error(message);
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    // Reset to original values
    setEditName(user?.username || "");
    setEditCourses([...originalCourses]);
    setShowAddCourseModal(false);
    setNewCourseCode("");
    setNewCourseName("");
    setSelectedColor("#FA4616");
    setIsEditMode(false);
  };

  const handleAddCourse = async () => {
    const code = newCourseCode.trim().toUpperCase();
    const name = newCourseName.trim();
    
    if (!code) {
      toast.info("Enter a course code");
      return;
    }

    try {
      const colorName = getColorName(selectedColor);
      const course = await addMyTACourse(code, name, colorName);
      setEditCourses([...editCourses, { ...course, color: selectedColor }]);
      setNewCourseCode("");
      setNewCourseName("");
      setSelectedColor("#FA4616");
      setShowAddCourseModal(false);
      toast.success(`Added ${course.code} to your courses`);
    } catch (error) {
      const message = error instanceof Error ? error.message : "Failed to add course";
      toast.error(message);
    }
  };

  const handleRemoveCourse = (id: number) => {
    setEditCourses(editCourses.filter(c => c.id !== id));
  };

  const handleTabChange = (tab: 'dashboard' | 'office-hours' | 'queue') => {
    navigate("/ta", { state: { activeTab: tab } });
  };

  return (
    <TALayout activeTab="profile" onTabChange={handleTabChange}>
      <div className="profile-page-container">
        <h1 className="profile-welcome-title">Welcome to your Profile!</h1>
        
        <div className="profile-card">
          {!isEditMode ? (
            <>
              <div className="profile-card-content">
                <h2 className="profile-name">{user?.username || "User"}</h2>
                
                <div className="profile-info-row">
                  <span className="profile-label">Email:</span>
                  <span className="profile-value">{user?.email || "—"}</span>
                </div>

                <div className="profile-courses-section">
                  <span className="profile-label">Courses:</span>
                  <div className="profile-courses-list">
                    {editCourses.length === 0 ? (
                      <span style={{ color: '#6b7280' }}>No courses linked yet.</span>
                    ) : (
                      editCourses.map((course) => (
                        <div key={course.id} className="course-item">
                          <span className="course-dot" style={{ backgroundColor: course.color }}></span>
                          <span>{course.code}{course.name ? ` - ${course.name}` : ''}</span>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>

              <div className="profile-card-buttons">
                <button className="profile-edit-btn" onClick={handleEditProfile}>
                  Edit Profile
                </button>
                <button className="profile-logout-btn" onClick={handleLogout}>
                  Log Out
                </button>
              </div>
            </>
          ) : (
            <>
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
                  <div className="profile-form-email-disabled">
                    {user?.email || "—"}
                  </div>
                </div>

                <div className="profile-form-section">
                  <div className="profile-courses-header">
                    <label className="profile-form-label">Courses</label>
                    <button 
                      className="profile-add-course-btn" 
                      onClick={() => setShowAddCourseModal(true)}
                    >
                      + Add Course
                    </button>
                  </div>
                  <div className="profile-edit-courses-list">
                    {editCourses.length === 0 ? (
                      <span style={{ color: '#6b7280' }}>No courses linked yet.</span>
                    ) : (
                      editCourses.map((course) => (
                        <div key={course.id} className="profile-edit-course-item">
                          <span className="course-dot" style={{ backgroundColor: course.color }}></span>
                          <span className="course-name">{course.code}{course.name ? ` - ${course.name}` : ''}</span>
                          <button
                            className="profile-remove-course-btn"
                            onClick={() => handleRemoveCourse(course.id)}
                          >
                            Remove
                          </button>
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </div>

              <div className="profile-card-edit-buttons">
                <button 
                  className="profile-save-btn" 
                  onClick={handleSave}
                  disabled={isSaving}
                >
                  {isSaving ? "Saving..." : "Save"}
                </button>
                <button 
                  className="profile-cancel-btn" 
                  onClick={handleCancel}
                  disabled={isSaving}
                >
                  Cancel
                </button>
              </div>
            </>
          )}
        </div>

        {/* Add Course Modal */}
        {showAddCourseModal && (
          <div className="add-course-modal-overlay">
            <div className="add-course-modal-card">
              <div className="add-course-modal-header">
                <h3 className="add-course-modal-title">Add Course</h3>
                <button 
                  className="add-course-modal-close-btn"
                  onClick={() => setShowAddCourseModal(false)}
                >
                  ✕
                </button>
              </div>

              <div className="add-course-modal-body">
                <div className="add-course-form-section">
                  <label className="add-course-form-label">Course Code</label>
                  <input
                    type="text"
                    className="add-course-form-input"
                    placeholder="e.g., COP 3530"
                    value={newCourseCode}
                    onChange={(e) => setNewCourseCode(e.target.value)}
                  />
                </div>

                <div className="add-course-form-section">
                  <label className="add-course-form-label">Course Name</label>
                  <input
                    type="text"
                    className="add-course-form-input"
                    placeholder="e.g., Data Structures"
                    value={newCourseName}
                    onChange={(e) => setNewCourseName(e.target.value)}
                  />
                </div>

                <div className="add-course-color-section">
                  <label className="add-course-form-label">Color</label>
                  <div className="add-course-color-picker">
                    {COURSE_COLORS.map((color) => (
                      <button
                        key={color}
                        className={`add-course-color-option ${selectedColor === color ? 'selected' : ''}`}
                        style={{ backgroundColor: color }}
                        onClick={() => setSelectedColor(color)}
                        aria-label={`Select color ${color}`}
                      />
                    ))}
                  </div>
                </div>
              </div>

              <div className="add-course-modal-actions">
                <button
                  className="add-course-modal-add-btn"
                  onClick={handleAddCourse}
                >
                  Add Course
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </TALayout>
  );
}
