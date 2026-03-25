import { useState } from 'react';
import { toast } from 'sonner';

interface QueueStudent {
  id: number;
  position: number;
  name: string;
  waitTime: string;
  joinedAt: Date;
  topic?: string;
  email?: string;
}

interface QueuePageProps {
  onEndSession?: () => void;
}

export function QueuePage({ onEndSession }: QueuePageProps) {
  const [queueStudents, setQueueStudents] = useState<QueueStudent[]>([
    { id: 1, position: 1, name: 'Sarah Johnson', waitTime: '35 min', joinedAt: new Date(Date.now() - 35 * 60000), topic: 'Linked Lists', email: 'sarah.j@ufl.edu' },
    { id: 2, position: 2, name: 'Michael Chen', waitTime: '28 min', joinedAt: new Date(Date.now() - 28 * 60000), topic: 'Binary Trees', email: 'mchen@ufl.edu' },
    { id: 3, position: 3, name: 'Emily Rodriguez', waitTime: '22 min', joinedAt: new Date(Date.now() - 22 * 60000), topic: 'Hash Tables', email: 'emily.r@ufl.edu' },
    { id: 4, position: 4, name: 'David Park', waitTime: '18 min', joinedAt: new Date(Date.now() - 18 * 60000), topic: 'Graph Algorithms', email: 'dpark@ufl.edu' },
    { id: 5, position: 5, name: 'Jessica Williams', waitTime: '12 min', joinedAt: new Date(Date.now() - 12 * 60000), topic: 'Dynamic Programming', email: 'jwilliams@ufl.edu' },
    { id: 6, position: 6, name: 'Alex Thompson', waitTime: '5 min', joinedAt: new Date(Date.now() - 5 * 60000), topic: 'Recursion', email: 'athompson@ufl.edu' },
  ]);

  const [queueStatus, setQueueStatus] = useState<'open' | 'paused' | 'closed'>('open');
  const [showAnnouncementModal, setShowAnnouncementModal] = useState(false);
  const [announcementText, setAnnouncementText] = useState('');

  const sortedQueue = [...queueStudents].sort((a, b) => a.joinedAt.getTime() - b.joinedAt.getTime());

  const handleStartSession = (student: QueueStudent) => {
    toast.success(`Starting session with ${student.name}`, {
      description: `Topic: ${student.topic}`,
    });
    setQueueStudents(prev => {
      const filtered = prev.filter(s => s.id !== student.id);
      return filtered.map((s, idx) => ({ ...s, position: idx + 1 }));
    });
  };

  const handleRemoveStudent = (student: QueueStudent) => {
    setQueueStudents(prev => {
      const filtered = prev.filter(s => s.id !== student.id);
      return filtered.map((s, idx) => ({ ...s, position: idx + 1 }));
    });
    toast.info(`${student.name} removed from queue`);
  };

  const handlePauseQueue = () => {
    setQueueStatus(queueStatus === 'paused' ? 'open' : 'paused');
    toast.info(queueStatus === 'paused' ? 'Queue reopened' : 'Queue paused - new students cannot join');
  };

  const handleCloseQueue = () => {
    setQueueStatus('closed');
    toast.warning('Queue closed - no new students can join');
    onEndSession?.();
  };

  const handleSendAnnouncement = () => {
    if (announcementText.trim()) {
      toast.success('Announcement sent to all students in queue', {
        description: announcementText,
      });
      setAnnouncementText('');
      setShowAnnouncementModal(false);
    }
  };

  const statusColor =
    queueStatus === 'open' ? 'var(--gator)' :
    queueStatus === 'paused' ? 'var(--alachua)' :
    'var(--bottlebrush)';

  const statusText =
    queueStatus === 'open' ? 'Open - Accepting Students' :
    queueStatus === 'paused' ? 'Paused - Not Accepting New Students' :
    'Closed - No New Students';

  return (
    <div className="main-section">

      {/* Page Header */}
      <div className="welcome-section">
        <h1 className="welcome-title">Queue Management</h1>
        <p style={{ opacity: 0.85, fontFamily: 'IBM Plex Sans, sans-serif' }}>
          View and manage the real-time queue of students
        </p>
      </div>

      {/* Queue Status Controls */}
      <div className="stat-card" style={{ borderColor: 'var(--core-blue)', borderWidth: 2 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <h2 style={{ fontSize: '1.25rem', fontWeight: 600, marginBottom: '0.75rem', fontFamily: 'IBM Plex Sans, sans-serif', color: 'var(--black)' }}>
              Queue Status
            </h2>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <div style={{ width: 12, height: 12, borderRadius: '50%', backgroundColor: statusColor, flexShrink: 0 }} />
              <span style={{ fontSize: '1rem', fontWeight: 500, fontFamily: 'IBM Plex Sans, sans-serif', color: 'var(--black)' }}>
                {statusText}
              </span>
            </div>
          </div>
          <div style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
            <button
              onClick={handlePauseQueue}
              style={{
                backgroundColor: queueStatus === 'paused' ? 'var(--gator)' : 'var(--alachua)',
                color: queueStatus === 'paused' ? 'var(--white)' : 'var(--black)',
                border: 'none',
                borderRadius: 8,
                padding: '0.625rem 1.25rem',
                fontWeight: 600,
                fontFamily: 'IBM Plex Sans, sans-serif',
                cursor: 'pointer',
              }}
            >
              {queueStatus === 'paused' ? 'Resume Queue' : 'Pause Queue'}
            </button>
            <button
              onClick={handleCloseQueue}
              disabled={queueStatus === 'closed'}
              style={{
                backgroundColor: queueStatus === 'closed' ? 'var(--cool-grey-3)' : 'var(--bottlebrush)',
                color: 'var(--white)',
                border: 'none',
                borderRadius: 8,
                padding: '0.625rem 1.25rem',
                fontWeight: 600,
                fontFamily: 'IBM Plex Sans, sans-serif',
                cursor: queueStatus === 'closed' ? 'not-allowed' : 'pointer',
              }}
            >
              Close Queue
            </button>
            <button
              onClick={() => setShowAnnouncementModal(true)}
              className="send-announcement-btn"
              style={{ padding: '0.625rem 1.25rem', fontSize: '1rem' }}
            >
              Send Announcement
            </button>
          </div>
        </div>
      </div>

      {/* Queue Stats */}
      <div className="stats-grid">
        <div className="stat-card blue">
          <div className="stat-header">
            <span className="stat-label">Total in Queue</span>
          </div>
          <div className="stat-value">{queueStudents.length}</div>
        </div>
        <div className="stat-card orange">
          <div className="stat-header">
            <span className="stat-label">Longest Wait</span>
          </div>
          <div className="stat-value">
            {sortedQueue.length > 0 ? sortedQueue[0].waitTime : '0 min'}
          </div>
        </div>
        <div className="stat-card green">
          <div className="stat-header">
            <span className="stat-label">Avg Wait Time</span>
          </div>
          <div className="stat-value">12 min</div>
        </div>
      </div>

      {/* Queue List */}
      <div className="sidebar-content" style={{ borderRadius: 12 }}>
        <h2 className="sidebar-title">Live Queue ({sortedQueue.length} students)</h2>

        {sortedQueue.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '3rem 0', color: 'var(--cool-grey-3)' }}>
            <p style={{ fontSize: '1.25rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>No students in queue</p>
            <p style={{ fontSize: '0.875rem', marginTop: '0.5rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
              Students will appear here when they join the queue
            </p>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
            {sortedQueue.map((student, index) => (
              <div key={student.id} className="office-hour-card">
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: '1rem', flexWrap: 'wrap' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>

                    {/* Position Badge */}
                    <div style={{
                      width: 48,
                      height: 48,
                      borderRadius: '50%',
                      backgroundColor: 'var(--core-blue)',
                      color: 'var(--white)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      fontWeight: 700,
                      fontSize: '1.25rem',
                      flexShrink: 0,
                      fontFamily: 'Anybody, IBM Plex Sans, sans-serif',
                    }}>
                      {index + 1}
                    </div>

                    {/* Student Info */}
                    <div>
                      <div style={{ fontWeight: 600, fontSize: '1rem', color: 'var(--cool-grey-11)', fontFamily: 'IBM Plex Sans, sans-serif', marginBottom: '0.25rem' }}>
                        {student.name}
                      </div>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', flexWrap: 'wrap' }}>
                        <span style={{ fontSize: '0.875rem', color: 'var(--cool-grey-11)', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                          Waiting {student.waitTime}
                        </span>
                        {student.topic && (
                          <>
                            <span style={{ color: 'var(--cool-grey-3)' }}>•</span>
                            <span style={{ fontSize: '0.875rem', color: 'var(--core-orange)', fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif' }}>
                              {student.topic}
                            </span>
                          </>
                        )}
                        {student.email && (
                          <>
                            <span style={{ color: 'var(--cool-grey-3)' }}>•</span>
                            <span style={{ fontSize: '0.875rem', color: 'var(--cool-grey-11)', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                              {student.email}
                            </span>
                          </>
                        )}
                      </div>
                      <div style={{ fontSize: '0.75rem', color: 'var(--cool-grey-11)', marginTop: '0.25rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                        Joined at {student.joinedAt.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}
                      </div>
                    </div>
                  </div>

                  {/* Actions */}
                  <div className="office-hour-actions" style={{ flexWrap: 'nowrap', width: 'auto' }}>
                    <button
                      onClick={() => handleStartSession(student)}
                      className="start-queue-btn"
                      style={{ width: 'auto', padding: '0.625rem 1.5rem', fontSize: '0.95rem' }}
                    >
                      Start Session
                    </button>
                    <button
                      onClick={() => handleRemoveStudent(student)}
                      className="cancel-btn"
                      style={{ border: '2px solid var(--bottlebrush)', borderRadius: 8, padding: '0.625rem 1rem' }}
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

      {/* Announcement Modal */}
      {showAnnouncementModal && (
        <div style={{
          position: 'fixed', inset: 0,
          backgroundColor: 'rgba(0,0,0,0.5)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          zIndex: 50,
        }}>
          <div style={{ background: 'var(--white)', borderRadius: 12, maxWidth: 480, width: '100%', margin: '0 1rem', overflow: 'hidden', boxShadow: '0 20px 40px rgba(0,0,0,0.3)' }}>
            <div style={{
              background: 'linear-gradient(135deg, var(--core-blue) 0%, var(--dark-blue) 100%)',
              padding: '1rem 1.5rem',
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
            }}>
              <h3 style={{ color: 'var(--white)', fontSize: '1.25rem', fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif' }}>
                Send Announcement
              </h3>
              <button
                onClick={() => setShowAnnouncementModal(false)}
                style={{ background: 'none', border: 'none', color: 'var(--white)', cursor: 'pointer', fontSize: '1.25rem', lineHeight: 1 }}
              >
                ✕
              </button>
            </div>
            <div style={{ padding: '1.5rem' }}>
              <p style={{ color: 'var(--cool-grey-11)', marginBottom: '1rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                Send a message to all {queueStudents.length} students currently in the queue
              </p>
              <textarea
                value={announcementText}
                onChange={(e) => setAnnouncementText(e.target.value)}
                placeholder="e.g., Running 10 minutes late, extending office hours by 30 minutes..."
                style={{
                  width: '100%', border: '1px solid var(--cool-grey-3)', borderRadius: 8,
                  padding: '0.75rem', minHeight: 120, fontFamily: 'IBM Plex Sans, sans-serif',
                  fontSize: '0.95rem', color: 'var(--cool-grey-11)', resize: 'vertical',
                  boxSizing: 'border-box',
                }}
              />
              <div style={{ background: '#e3f2fd', border: '1px solid #bbdefb', borderRadius: 8, padding: '0.75rem', marginTop: '1rem' }}>
                <p style={{ fontSize: '0.875rem', fontWeight: 600, color: 'var(--cool-grey-11)', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                  Example announcements:
                </p>
                <ul style={{ fontSize: '0.875rem', color: 'var(--cool-grey-11)', marginTop: '0.5rem', paddingLeft: '1.25rem', fontFamily: 'IBM Plex Sans, sans-serif' }}>
                  <li>Running 10 minutes late</li>
                  <li>Extending office hours by 30 minutes</li>
                  <li>Taking a 5 minute break, will resume shortly</li>
                </ul>
              </div>
              <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.5rem' }}>
                <button
                  onClick={() => setShowAnnouncementModal(false)}
                  style={{
                    flex: 1, padding: '0.75rem', border: '1px solid var(--cool-grey-3)',
                    borderRadius: 8, fontWeight: 600, fontFamily: 'IBM Plex Sans, sans-serif',
                    cursor: 'pointer', background: 'var(--white)', color: 'var(--cool-grey-11)',
                  }}
                >
                  Cancel
                </button>
                <button
                  onClick={handleSendAnnouncement}
                  className="start-queue-btn"
                  style={{ flex: 1, padding: '0.75rem', fontSize: '1rem' }}
                >
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