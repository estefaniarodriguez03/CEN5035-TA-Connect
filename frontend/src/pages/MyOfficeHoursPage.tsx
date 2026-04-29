import { useState } from 'react';
import { toast } from 'sonner';
import {
  createOfficeHour,
  updateOfficeHour,
  deleteOfficeHour,
  DAY_NAMES,
  type OfficeHour,
} from '../api/officeHours';
import { type TACourse } from '../api/courses';

function formatTime(time: string): string {
  const [hours, minutes] = time.split(':');
  const hour = parseInt(hours);
  const ampm = hour >= 12 ? 'PM' : 'AM';
  const displayHour = hour > 12 ? hour - 12 : hour === 0 ? 12 : hour;
  return `${displayHour}:${minutes} ${ampm}`;
}

function toTimeInput(t: string): string {
  return t.slice(0, 5);
}

interface FormState {
  day_of_week: number;
  start_time: string;
  end_time: string;
  course_id: number;
  location: string;
}

const EMPTY_FORM: FormState = {
  day_of_week: 1,
  start_time: '',
  end_time: '',
  course_id: 0,
  location: '',
};

export default function MyOfficeHoursPage({
  taCourses = [],
  officeHours,
  onOfficeHoursChange,
}: {
  taCourses?: TACourse[];
  officeHours: OfficeHour[];
  onOfficeHoursChange: (hours: OfficeHour[]) => void;
}) {
  const [showAddModal, setShowAddModal] = useState(false);
  const [editingSchedule, setEditingSchedule] = useState<OfficeHour | null>(null);
  const [formData, setFormData] = useState<FormState>({ ...EMPTY_FORM });

  const groupedSchedules = DAY_NAMES.reduce((acc, day, idx) => {
    acc[day] = officeHours.filter((s) => s.day_of_week === idx);
    return acc;
  }, {} as Record<string, OfficeHour[]>);

  const getCourseLabel = (courseID: number): string => {
    const course = taCourses.find((c) => c.id === courseID);
    return course
      ? `${course.code}${course.name ? ` - ${course.name}` : ''}`
      : `Course ${courseID}`;
  };

  const handleAddSchedule = async () => {
    if (!formData.start_time || !formData.end_time || !formData.course_id) {
      toast.error('Please fill in all required fields');
      return;
    }
    try {
      const created = await createOfficeHour({
        course_id: formData.course_id,
        day_of_week: formData.day_of_week,
        start_time: formData.start_time,
        end_time: formData.end_time,
        location: formData.location,
      });
      onOfficeHoursChange(
        [...officeHours, created].sort(
          (a, b) => a.day_of_week - b.day_of_week || a.start_time.localeCompare(b.start_time)
        )
      );
      toast.success('Office hours added');
      setShowAddModal(false);
      setFormData({ ...EMPTY_FORM });
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to add office hours');
    }
  };

  const handleUpdateSchedule = async () => {
    if (!editingSchedule) return;
    try {
      const updated = await updateOfficeHour(editingSchedule.id, {
        course_id: editingSchedule.course_id,
        day_of_week: editingSchedule.day_of_week,
        start_time: editingSchedule.start_time,
        end_time: editingSchedule.end_time,
        location: editingSchedule.location,
      });
      onOfficeHoursChange(
        officeHours
          .map((s) => (s.id === updated.id ? updated : s))
          .sort(
            (a, b) => a.day_of_week - b.day_of_week || a.start_time.localeCompare(b.start_time)
          )
      );
      toast.success('Office hours updated');
      setEditingSchedule(null);
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to update office hours');
    }
  };

  const handleDeleteSchedule = async (id: number) => {
    try {
      await deleteOfficeHour(id);
      onOfficeHoursChange(officeHours.filter((s) => s.id !== id));
      toast.success('Office hours deleted');
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to delete office hours');
    }
  };

  return (
    <div className="oh-page">

      {/* Page Header */}
      <div className="oh-page-header">
        <div>
          <h1 className="welcome-title oh-page-title">My Office Hours</h1>
          <p className="oh-page-subtitle">Manage your office hour schedule and availability</p>
        </div>
        <button
          className="start-queue-btn oh-add-btn"
          onClick={() => setShowAddModal(true)}
          disabled={taCourses.length === 0}
          title={taCourses.length === 0 ? 'Add courses from the Dashboard first' : undefined}
        >
          + Add Office Hours
        </button>
      </div>

      {taCourses.length === 0 && (
        <p style={{ color: '#6b7280', fontSize: '0.9rem', marginBottom: '1rem' }}>
          No courses linked yet. Add courses from the Dashboard before scheduling office hours.
        </p>
      )}

      {/* Schedule — one card per day */}
      <div className="oh-schedule-grid">
        {DAY_NAMES.map((day) => (
          <div key={day} className="oh-day-card">
            <div className="oh-day-header">
              <h2 className="oh-day-heading">{day}</h2>
            </div>
            <div className="oh-day-body">
              {groupedSchedules[day].length > 0 ? (
                groupedSchedules[day].map((schedule) => (
                  <div key={schedule.id} className="office-hour-card oh-schedule-item">
                    <div className="oh-schedule-info">
                      <div className="office-hour-time oh-time-row">
                        <span className="oh-time-label">
                          {formatTime(schedule.start_time)} – {formatTime(schedule.end_time)}
                        </span>
                      </div>
                      <p className="office-hour-course oh-course-label">
                        {getCourseLabel(schedule.course_id)}
                      </p>
                      {schedule.location && (
                        <p className="oh-course-label" style={{ color: '#6b7280', fontSize: '0.85rem' }}>
                          {schedule.location}
                        </p>
                      )}
                    </div>
                    <div className="office-hour-actions oh-item-actions">
                      <button
                        className="modify-btn"
                        onClick={() => setEditingSchedule(schedule)}
                      >
                        Edit
                      </button>
                      <button
                        className="cancel-btn"
                        onClick={() => void handleDeleteSchedule(schedule.id)}
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                ))
              ) : (
                <p className="oh-empty-day">No office hours scheduled</p>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Add Modal */}
      {showAddModal && (
        <OfficeHoursModal
          title="Add Office Hours"
          formData={formData}
          taCourses={taCourses}
          onChange={setFormData}
          onSubmit={() => void handleAddSchedule()}
          onClose={() => {
            setShowAddModal(false);
            setFormData({ ...EMPTY_FORM });
          }}
          submitLabel="Add Schedule"
        />
      )}

      {/* Edit Modal */}
      {editingSchedule && (
        <OfficeHoursModal
          title="Edit Office Hours"
          formData={{
            day_of_week: editingSchedule.day_of_week,
            start_time: toTimeInput(editingSchedule.start_time),
            end_time: toTimeInput(editingSchedule.end_time),
            course_id: editingSchedule.course_id,
            location: editingSchedule.location,
          }}
          taCourses={taCourses}
          onChange={(updated) => setEditingSchedule({ ...editingSchedule, ...updated })}
          onSubmit={() => void handleUpdateSchedule()}
          onClose={() => setEditingSchedule(null)}
          submitLabel="Update Schedule"
        />
      )}
    </div>
  );
}

interface ModalProps {
  title: string;
  formData: FormState;
  taCourses: TACourse[];
  onChange: (data: FormState) => void;
  onSubmit: () => void;
  onClose: () => void;
  submitLabel: string;
}

function OfficeHoursModal({ title, formData, taCourses, onChange, onSubmit, onClose, submitLabel }: ModalProps) {
  return (
    <div className="announcement-modal-overlay">
      <div className="announcement-modal-card">
        <div className="announcement-modal-header">
          <h3 className="announcement-modal-title">{title}</h3>
          <button className="announcement-modal-close-btn" onClick={onClose}>✕</button>
        </div>
        <div className="announcement-modal-body">

          <div className="oh-modal-field">
            <label className="oh-modal-label">Day of Week</label>
            <select
              className="oh-modal-input"
              value={formData.day_of_week}
              onChange={(e) => onChange({ ...formData, day_of_week: parseInt(e.target.value) })}
            >
              {DAY_NAMES.map((day, idx) => (
                <option key={day} value={idx}>{day}</option>
              ))}
            </select>
          </div>

          <div className="oh-modal-time-row">
            <div className="oh-modal-field">
              <label className="oh-modal-label">Start Time</label>
              <input
                type="time"
                className="oh-modal-input"
                value={formData.start_time}
                onChange={(e) => onChange({ ...formData, start_time: e.target.value })}
              />
            </div>
            <div className="oh-modal-field">
              <label className="oh-modal-label">End Time</label>
              <input
                type="time"
                className="oh-modal-input"
                value={formData.end_time}
                onChange={(e) => onChange({ ...formData, end_time: e.target.value })}
              />
            </div>
          </div>

          <div className="oh-modal-field">
            <label className="oh-modal-label">Course</label>
            {taCourses.length > 0 ? (
              <select
                className="oh-modal-input"
                value={formData.course_id === 0 ? '' : formData.course_id}
                onChange={(e) => onChange({ ...formData, course_id: parseInt(e.target.value) })}
              >
                <option value="">Select a course</option>
                {taCourses.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.code}{c.name ? ` - ${c.name}` : ''}
                  </option>
                ))}
              </select>
            ) : (
              <p style={{ color: '#6b7280', fontSize: '0.875rem', margin: 0 }}>
                No courses linked yet. Add courses from the Dashboard first.
              </p>
            )}
          </div>

          <div className="oh-modal-field">
            <label className="oh-modal-label">Location (optional)</label>
            <input
              type="text"
              className="oh-modal-input"
              placeholder="e.g. CSE E221"
              value={formData.location}
              onChange={(e) => onChange({ ...formData, location: e.target.value })}
            />
          </div>

          <div className="announcement-modal-actions">
            <button className="announcement-cancel-btn" onClick={onClose}>Cancel</button>
            <button
              className="start-queue-btn announcement-send-all-btn"
              onClick={onSubmit}
              disabled={!formData.course_id || !formData.start_time || !formData.end_time}
            >
              {submitLabel}
            </button>
          </div>

        </div>
      </div>
    </div>
  );
}