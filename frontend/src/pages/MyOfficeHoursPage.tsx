import { useState } from 'react';
import { toast } from 'sonner';

interface OfficeHourSchedule {
  id: number;
  dayOfWeek: string;
  startTime: string;
  endTime: string;
  course: string;
  isRecurring: boolean;
}

const DAYS_OF_WEEK = ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'];

const EMPTY_FORM = {
  dayOfWeek: 'Monday',
  startTime: '',
  endTime: '',
  course: '',
  isRecurring: true,
};

function formatTime(time: string): string {
  const [hours, minutes] = time.split(':');
  const hour = parseInt(hours);
  const ampm = hour >= 12 ? 'PM' : 'AM';
  const displayHour = hour > 12 ? hour - 12 : hour === 0 ? 12 : hour;
  return `${displayHour}:${minutes} ${ampm}`;
}

export default function MyOfficeHoursPage() {
  const [schedules, setSchedules] = useState<OfficeHourSchedule[]>([
    { id: 1, dayOfWeek: 'Monday', startTime: '11:00', endTime: '13:00', course: 'COP3530 - Data Structures', isRecurring: true },
    { id: 2, dayOfWeek: 'Monday', startTime: '15:30', endTime: '16:30', course: 'COP3530 - Data Structures', isRecurring: true },
    { id: 3, dayOfWeek: 'Wednesday', startTime: '14:00', endTime: '15:00', course: 'COP3530 - Data Structures', isRecurring: true },
    { id: 4, dayOfWeek: 'Friday', startTime: '10:00', endTime: '11:30', course: 'COP3530 - Data Structures', isRecurring: true },
  ]);

  const [showAddModal, setShowAddModal] = useState(false);
  const [editingSchedule, setEditingSchedule] = useState<OfficeHourSchedule | null>(null);
  const [formData, setFormData] = useState({ ...EMPTY_FORM });

  const groupedSchedules = DAYS_OF_WEEK.reduce((acc, day) => {
    acc[day] = schedules.filter((s) => s.dayOfWeek === day);
    return acc;
  }, {} as Record<string, OfficeHourSchedule[]>);

  const handleAddSchedule = () => {
    if (!formData.startTime || !formData.endTime || !formData.course) {
      toast.error('Please fill in all fields');
      return;
    }
    setSchedules((prev) => [...prev, { id: Date.now(), ...formData }]);
    toast.success('Office hours added successfully');
    setShowAddModal(false);
    setFormData({ ...EMPTY_FORM });
  };

  const handleUpdateSchedule = () => {
    if (!editingSchedule) return;
    setSchedules((prev) => prev.map((s) => (s.id === editingSchedule.id ? { ...editingSchedule } : s)));
    toast.success('Office hours updated successfully');
    setEditingSchedule(null);
  };

  const handleDeleteSchedule = (id: number) => {
    setSchedules((prev) => prev.filter((s) => s.id !== id));
    toast.success('Office hours deleted');
  };

  return (
    <div className="oh-page">

      {/* Page Header */}
      <div className="oh-page-header">
        <div>
          <h1 className="welcome-title oh-page-title">My Office Hours</h1>
          <p className="oh-page-subtitle">Manage your office hour schedule and availability</p>
        </div>
        <button className="start-queue-btn oh-add-btn" onClick={() => setShowAddModal(true)}>
          + Add Office Hours
        </button>
      </div>

      {/* Schedule — one card per day */}
      <div className="oh-schedule-grid">
        {DAYS_OF_WEEK.map((day) => (
          <div key={day} className="oh-day-card">

            {/* Blue day header — matches sidebar-content gradient */}
            <div className="oh-day-header">
              <h2 className="oh-day-heading">{day}</h2>
            </div>

            <div className="oh-day-body">
              {groupedSchedules[day].length > 0 ? (
                groupedSchedules[day].map((schedule) => (
                  <div key={schedule.id} className="office-hour-card oh-schedule-item">

                    <div className="oh-schedule-info">
                      {/* Time row — reuses office-hour-time colour */}
                      <div className="office-hour-time oh-time-row">
                        <span className="oh-time-label">
                          {formatTime(schedule.startTime)} – {formatTime(schedule.endTime)}
                        </span>
                      </div>

                      {/* Course — reuses office-hour-course */}
                      <p className="office-hour-course oh-course-label">{schedule.course}</p>

                      {schedule.isRecurring && (
                        <span className="oh-recurring-badge">Recurring Weekly</span>
                      )}
                    </div>

                    {/* Actions — reuses existing modify-btn / cancel-btn */}
                    <div className="office-hour-actions oh-item-actions">
                      <button
                        className="modify-btn"
                        onClick={() => setEditingSchedule(schedule)}
                      >
                        Edit
                      </button>
                      <button
                        className="cancel-btn"
                        onClick={() => handleDeleteSchedule(schedule.id)}
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
          onChange={setFormData}
          onSubmit={handleAddSchedule}
          onClose={() => setShowAddModal(false)}
          submitLabel="Add Schedule"
        />
      )}

      {/* Edit Modal */}
      {editingSchedule && (
        <OfficeHoursModal
          title="Edit Office Hours"
          formData={editingSchedule}
          onChange={setEditingSchedule}
          onSubmit={handleUpdateSchedule}
          onClose={() => setEditingSchedule(null)}
          submitLabel="Update Schedule"
        />
      )}
    </div>
  );
}

interface ModalProps {
  title: string;
  formData: {
    dayOfWeek: string;
    startTime: string;
    endTime: string;
    course: string;
    isRecurring: boolean;
  };
  onChange: (data: any) => void;
  onSubmit: () => void;
  onClose: () => void;
  submitLabel: string;
}

function OfficeHoursModal({ title, formData, onChange, onSubmit, onClose, submitLabel }: ModalProps) {
  return (
    <div className="announcement-modal-overlay">
      <div className="announcement-modal-card">

        {/* Header — reuses announcement-modal-header gradient */}
        <div className="announcement-modal-header">
          <h3 className="announcement-modal-title">{title}</h3>
          <button className="announcement-modal-close-btn" onClick={onClose}>✕</button>
        </div>

        <div className="announcement-modal-body">

          {/* Day selector */}
          <div className="oh-modal-field">
            <label className="oh-modal-label">Day of Week</label>
            <select
              className="oh-modal-input"
              value={formData.dayOfWeek}
              onChange={(e) => onChange({ ...formData, dayOfWeek: e.target.value })}
            >
              {DAYS_OF_WEEK.map((day) => (
                <option key={day} value={day}>{day}</option>
              ))}
            </select>
          </div>

          {/* Start / End time */}
          <div className="oh-modal-time-row">
            <div className="oh-modal-field">
              <label className="oh-modal-label">Start Time</label>
              <input
                type="time"
                className="oh-modal-input"
                value={formData.startTime}
                onChange={(e) => onChange({ ...formData, startTime: e.target.value })}
              />
            </div>
            <div className="oh-modal-field">
              <label className="oh-modal-label">End Time</label>
              <input
                type="time"
                className="oh-modal-input"
                value={formData.endTime}
                onChange={(e) => onChange({ ...formData, endTime: e.target.value })}
              />
            </div>
          </div>

          {/* Course */}
          <div className="oh-modal-field">
            <label className="oh-modal-label">Course</label>
            <input
              type="text"
              className="oh-modal-input"
              placeholder="e.g., COP3530 - Data Structures"
              value={formData.course}
              onChange={(e) => onChange({ ...formData, course: e.target.value })}
            />
          </div>

          {/* Recurring checkbox */}
          <div className="oh-modal-recurring-row">
            <input
              type="checkbox"
              id="oh-recurring"
              checked={formData.isRecurring}
              onChange={(e) => onChange({ ...formData, isRecurring: e.target.checked })}
            />
            <label htmlFor="oh-recurring" className="oh-modal-label">Recurring weekly</label>
          </div>

          {/* Actions — reuse announcement button styles */}
          <div className="announcement-modal-actions">
            <button className="announcement-cancel-btn" onClick={onClose}>Cancel</button>
            <button className="start-queue-btn announcement-send-all-btn" onClick={onSubmit}>
              {submitLabel}
            </button>
          </div>

        </div>
      </div>
    </div>
  );
}