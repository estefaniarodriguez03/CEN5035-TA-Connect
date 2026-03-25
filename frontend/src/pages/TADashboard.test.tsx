// @vitest-environment jsdom
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import TADashboard from './TADashboard';

// Mock the auth context
vi.mock('../context/AuthContext', () => ({
  useAuth: () => ({
    user: { username: 'Test TA' },
    logout: vi.fn(),
  }),
}));

// Mock toast
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    info: vi.fn(),
    warning: vi.fn(),
  },
}));

// Mock images
vi.mock('../images/UF Logo.png', () => ({ default: '' }));
vi.mock('../images/Blue Message Icon.png', () => ({ default: '' }));
vi.mock('../images/Orange Clock Icon.png', () => ({ default: '' }));
vi.mock('../images/Orange Date Icon.png', () => ({ default: '' }));
vi.mock('../images/Orange Group Icon.png', () => ({ default: '' }));
vi.mock('../images/Orange Question Mark Icon.png', () => ({ default: '' }));
vi.mock('../images/White Notification Icon.png', () => ({ default: '' }));
vi.mock('../images/White Profile Icon.png', () => ({ default: '' }));

describe('TADashboard', () => {

  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Renders the dashboard in closed state by default
  it('renders the dashboard in closed state by default', () => {
    render(<TADashboard />);
    expect(screen.getByText('Start Office Hours Live Queue')).toBeInTheDocument();
  });

  // Renders the welcome message with the TA's name
  it('renders the welcome message with the TA username', () => {
    render(<TADashboard />);
    expect(screen.getByText(/Welcome Back, Test TA/i)).toBeInTheDocument();
  });

  // Renders the stats grid
  it('renders all stat cards', () => {
    render(<TADashboard />);
    expect(screen.getByText('Students Helped Today')).toBeInTheDocument();
    expect(screen.getByText('Avg Wait Time')).toBeInTheDocument();
    expect(screen.getByText('Current Queue Length')).toBeInTheDocument();
    expect(screen.getByText('Longest Wait Time')).toBeInTheDocument();
    expect(screen.getByText('Most Common Topic')).toBeInTheDocument();
    expect(screen.getByText('Avg Session Duration')).toBeInTheDocument();
  });

  // handleOpenQueue - clicking Start Office Hours opens the queue
  it('opens the live queue when Start Office Hours Live Queue is clicked', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    expect(screen.getByText('Live Queue')).toBeInTheDocument();
  });

  // Queue shows students when opened
  it('shows students in the queue when the queue is opened', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    expect(screen.getByText('Sarah Johnson')).toBeInTheDocument();
    expect(screen.getByText('Michael Chen')).toBeInTheDocument();
  });

  // handlePauseQueue - pauses the queue
  it('pauses the queue when Pause Queue is clicked', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    fireEvent.click(screen.getByText('Pause Queue'));
    expect(screen.getByText('Paused - Not Accepting New Students')).toBeInTheDocument();
  });

  // handlePauseQueue - resumes the queue
  it('resumes the queue when Resume Queue is clicked', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    fireEvent.click(screen.getByText('Pause Queue'));
    fireEvent.click(screen.getByText('Resume Queue'));
    expect(screen.getByText('Open - Accepting Students')).toBeInTheDocument();
  });

  // handleCloseQueue - closes the queue and returns to dashboard
  it('closes the queue and returns to the dashboard view', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    fireEvent.click(screen.getByText('Close Queue'));
    expect(screen.getByText('Start Office Hours Live Queue')).toBeInTheDocument();
  });

  // handleCloseQueue - clears students from queue
  it('clears all students from the queue when closed', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    expect(screen.getByText('Sarah Johnson')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Close Queue'));
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    expect(screen.queryByText('Sarah Johnson')).not.toBeInTheDocument();
  });

  // handleStartSession - removes first student from queue
  it('removes the first student when Start Session is clicked', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    fireEvent.click(screen.getByText('Start Session'));
    expect(screen.queryByText('Sarah Johnson')).not.toBeInTheDocument();
  });

  // handleRemoveStudent - removes a student from the queue
  it('removes a student when Remove is clicked', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    const removeButtons = screen.getAllByText('Remove');
    fireEvent.click(removeButtons[0]);
    expect(screen.queryByText('Sarah Johnson')).not.toBeInTheDocument();
  });

  // queue count updates after removal
  it('updates the queue count when a student is removed', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    expect(screen.getByText('Students in Queue (5)')).toBeInTheDocument();
    fireEvent.click(screen.getAllByText('Remove')[0]);
    expect(screen.getByText('Students in Queue (4)')).toBeInTheDocument();
  });

  // announcement modal opens from dashboard
  it('opens the announcement modal from the dashboard', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Send Announcement'));
    expect(screen.getByPlaceholderText(/Running 10 minutes late/i)).toBeInTheDocument();
  });

  // announcement modal opens from queue view
  it('opens the announcement modal from the queue view', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    fireEvent.click(screen.getAllByText('Send Announcement')[0]);
    expect(screen.getByPlaceholderText(/Running 10 minutes late/i)).toBeInTheDocument();
  });

  // handleSendAnnouncement - sends and closes modal
  it('sends an announcement and closes the modal', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Send Announcement'));
    fireEvent.change(screen.getByPlaceholderText(/Running 10 minutes late/i), {
      target: { value: 'Office hours extended by 30 minutes!' },
    });
    fireEvent.click(screen.getByText('Send to All'));
    expect(screen.queryByPlaceholderText(/Running 10 minutes late/i)).not.toBeInTheDocument();
  });

  // Empty queue message shows when all students are removed
  it('shows empty queue message when all students are removed', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Start Office Hours Live Queue'));
    const removeButtons = screen.getAllByText('Remove');
    removeButtons.forEach((btn) => fireEvent.click(btn));
    expect(screen.getByText('No students in queue')).toBeInTheDocument();
  });
});