// @vitest-environment jsdom
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

const hoisted = vi.hoisted(() => ({
  navigate: vi.fn(),
  locationState: undefined as { fromRegister?: boolean } | undefined,
  auth: {
    user: { id: 10, username: 'Test TA', email: 'ta@test.com', role: 'ta' as 'student' | 'ta' },
    login: vi.fn(),
    logout: vi.fn(),
  },
  toast: {
    success: vi.fn(),
    info: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
  },
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return {
    ...actual,
    useNavigate: () => hoisted.navigate,
    useLocation: () => ({ state: hoisted.locationState }),
    Link: ({ to, children }: { to: string; children: React.ReactNode }) => <a href={to}>{children}</a>,
  };
});

vi.mock('../context/AuthContext', () => ({
  useAuth: () => hoisted.auth,
}));

vi.mock('sonner', () => ({
  toast: hoisted.toast,
}));

vi.mock('../api/auth', () => ({
  login: vi.fn(),
  register: vi.fn(),
}));

vi.mock('../api/queue', () => ({
  getActiveQueueByCourse: vi.fn(),
  createQueue: vi.fn(),
  getQueueOrNull: vi.fn(),
  subscribeToQueueEvents: vi.fn(),
  updateQueueStatus: vi.fn(),
  nextQueueStudent: vi.fn(),
  setActiveQueueForCourse: vi.fn(),
  clearActiveQueueForCourse: vi.fn(),
  setActiveQueueForOfficeHour: vi.fn(),
  clearActiveQueueForOfficeHour: vi.fn(),
  getActiveQueueForOfficeHour: vi.fn(),
  joinQueue: vi.fn(),
  leaveQueue: vi.fn(),
}));

vi.mock('../images/UF Logo.png', () => ({ default: 'mock-image.png' }));
vi.mock('../images/Blue Message Icon.png', () => ({ default: 'mock-image.png' }));
vi.mock('../images/Orange Clock Icon.png', () => ({ default: 'mock-image.png' }));
vi.mock('../images/Orange Date Icon.png', () => ({ default: 'mock-image.png' }));
vi.mock('../images/Orange Group Icon.png', () => ({ default: 'mock-image.png' }));
vi.mock('../images/Orange Question Mark Icon.png', () => ({ default: 'mock-image.png' }));
vi.mock('../images/White Notification Icon.png', () => ({ default: 'mock-image.png' }));
vi.mock('../images/White Profile Icon.png', () => ({ default: 'mock-image.png' }));

import Login from './Login';
import Register from './Register';
import StudentDashboard from './StudentDashboard';
import TADashboard from './TADashboard';

import { login, register } from '../api/auth';
import {
  getActiveQueueByCourse,
  createQueue,
  getQueueOrNull,
  subscribeToQueueEvents,
  updateQueueStatus,
  nextQueueStudent,
  getActiveQueueForOfficeHour,
  joinQueue,
  leaveQueue,
} from '../api/queue';

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.locationState = undefined;
  hoisted.auth.user = { id: 10, username: 'Test TA', email: 'ta@test.com', role: 'ta' };

  vi.mocked(getActiveQueueByCourse).mockResolvedValue(null);
  vi.mocked(createQueue).mockResolvedValue({
    id: 10,
    course_id: 2,
    ta_id: 10,
    status: 'open',
    created_at: new Date().toISOString(),
  });
  vi.mocked(getQueueOrNull).mockResolvedValue({
    id: 10,
    course_id: 2,
    ta_id: 10,
    status: 'open',
    created_at: new Date().toISOString(),
    entries: [],
  });
  vi.mocked(subscribeToQueueEvents).mockImplementation(() => () => undefined);
  vi.mocked(updateQueueStatus).mockResolvedValue({ id: 10, status: 'open' });
  vi.mocked(nextQueueStudent).mockResolvedValue({
    queue_id: 10,
    status: 'in_session',
    student: {
      id: 1,
      queue_id: 10,
      student_id: 101,
      position: 1,
      joined_at: new Date().toISOString(),
      username: 'Sarah Johnson',
    },
  });

  vi.mocked(getActiveQueueForOfficeHour).mockReturnValue(7);
  vi.mocked(joinQueue).mockResolvedValue({
    id: 99,
    queue_id: 7,
    position: 2,
    joined_at: new Date().toISOString(),
  });
  vi.mocked(leaveQueue).mockResolvedValue();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('Login page', () => {
  it('renders login form fields and submit button', () => {
    render(<Login />);
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Login' })).toBeInTheDocument();
  });

  it('shows registration-success message when redirected from register', () => {
    hoisted.locationState = { fromRegister: true };
    render(<Login />);
    expect(screen.getByText('Registration successful! Please log in below.')).toBeInTheDocument();
  });

  it('logs in a student and navigates to student dashboard', async () => {
    vi.mocked(login).mockResolvedValue({
      token: 'student-token',
      user: { id: 1, username: 'Student One', email: 's@test.com', role: 'student' },
    });

    render(<Login />);
    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 's@test.com' } });
    fireEvent.change(screen.getByPlaceholderText('Password'), { target: { value: 'pass123' } });
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    await waitFor(() => {
      expect(login).toHaveBeenCalledWith('s@test.com', 'pass123');
      expect(hoisted.auth.login).toHaveBeenCalledWith(
        'student-token',
        expect.objectContaining({ role: 'student' })
      );
      expect(hoisted.navigate).toHaveBeenCalledWith('/student');
    });
  });

  it('logs in a TA and navigates to TA dashboard', async () => {
    vi.mocked(login).mockResolvedValue({
      token: 'ta-token',
      user: { id: 10, username: 'TA', email: 'ta@test.com', role: 'ta' },
    });

    render(<Login />);
    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'ta@test.com' } });
    fireEvent.change(screen.getByPlaceholderText('Password'), { target: { value: 'pass123' } });
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    await waitFor(() => {
      expect(hoisted.navigate).toHaveBeenCalledWith('/ta');
    });
  });

  it('shows login error when API call fails', async () => {
    vi.mocked(login).mockRejectedValue(new Error('bad credentials'));

    render(<Login />);
    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'bad@test.com' } });
    fireEvent.change(screen.getByPlaceholderText('Password'), { target: { value: 'wrong' } });
    fireEvent.click(screen.getByRole('button', { name: 'Login' }));

    expect(await screen.findByText('Login failed. Check your email and password.')).toBeInTheDocument();
  });
});

describe('Register page', () => {
  it('renders registration form with default student role', () => {
    render(<Register />);
    expect(screen.getByPlaceholderText('Name')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Email')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Password')).toBeInTheDocument();
    expect(screen.getByDisplayValue('Student')).toBeInTheDocument();
  });

  it('registers successfully and redirects to login with register flag', async () => {
    const setTimeoutSpy = vi.spyOn(globalThis, 'setTimeout');

    vi.mocked(register).mockResolvedValue({
      token: 'unused',
      user: { id: 1, username: 'New User', email: 'new@test.com', role: 'student' },
    });

    render(<Register />);
    fireEvent.change(screen.getByPlaceholderText('Name'), { target: { value: 'New User' } });
    fireEvent.change(screen.getByPlaceholderText('Email'), { target: { value: 'new@test.com' } });
    fireEvent.change(screen.getByPlaceholderText('Password'), { target: { value: 'secret' } });
    fireEvent.change(screen.getByDisplayValue('Student'), { target: { value: 'ta' } });
    fireEvent.click(screen.getByRole('button', { name: 'Register' }));

    expect(await screen.findByText('Registration successful! Redirecting to login...')).toBeInTheDocument();
    expect(register).toHaveBeenCalledWith('New User', 'new@test.com', 'secret', 'ta');
    expect(setTimeoutSpy).toHaveBeenCalled();

    const redirectTimeoutCall = setTimeoutSpy.mock.calls.find((call) => call[1] === 1500);
    const redirectCallback = redirectTimeoutCall?.[0];
    if (typeof redirectCallback === 'function') {
      redirectCallback();
    }

    await waitFor(() => {
      expect(hoisted.navigate).toHaveBeenCalledWith('/login', { state: { fromRegister: true } });
    });

    setTimeoutSpy.mockRestore();
  });

  it('shows specific backend registration error message', async () => {
    vi.mocked(register).mockRejectedValue({
      response: { data: { error: 'Email already exists' } },
    });

    render(<Register />);
    fireEvent.click(screen.getByRole('button', { name: 'Register' }));

    expect(await screen.findByText('Registration failed: Email already exists')).toBeInTheDocument();
  });

  it('shows fallback registration error message', async () => {
    vi.mocked(register).mockRejectedValue(new Error('unknown'));

    render(<Register />);
    fireEvent.click(screen.getByRole('button', { name: 'Register' }));

    expect(
      await screen.findByText('Registration failed. The username or email is already in use.')
    ).toBeInTheDocument();
  });
});

describe('StudentDashboard page', () => {
  beforeEach(() => {
    hoisted.auth.user = { id: 42, username: 'Student User', email: 'student@test.com', role: 'student' };
  });

  it('renders student greeting and queue controls', () => {
    render(<StudentDashboard />);
    expect(screen.getByText(/Welcome Back, Student User!/i)).toBeInTheDocument();
    expect(screen.getByRole('combobox')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Join Queue' })).toBeEnabled();
  });

  it('shows error toast when no active queue exists for selected office hour', async () => {
    vi.mocked(getActiveQueueForOfficeHour).mockReturnValue(null);

    render(<StudentDashboard />);
    fireEvent.click(screen.getByRole('button', { name: 'Join Queue' }));

    await waitFor(() => {
      expect(hoisted.toast.error).toHaveBeenCalledWith(
        'The TA has not opened the queue for this office hours yet. Come back later!'
      );
    });
  });

  it('joins queue, disables selectors, and shows real-time section', async () => {
    vi.mocked(getQueueOrNull).mockResolvedValue({
      id: 7,
      course_id: 1,
      ta_id: 10,
      status: 'open',
      created_at: new Date().toISOString(),
      entries: [{
        id: 200,
        queue_id: 7,
        student_id: 42,
        position: 2,
        joined_at: new Date().toISOString(),
        username: 'Student User',
      }],
    });

    render(<StudentDashboard />);
    fireEvent.click(screen.getByRole('button', { name: 'Join Queue' }));

    expect(await screen.findByText('Real-Time Queue Status')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Already in Queue' })).toBeDisabled();
    expect(screen.getByRole('combobox')).toBeDisabled();
    expect(joinQueue).toHaveBeenCalledWith(7);
  });

  it('leaves queue and hides real-time section', async () => {
    vi.mocked(getQueueOrNull).mockResolvedValue({
      id: 7,
      course_id: 1,
      ta_id: 10,
      status: 'open',
      created_at: new Date().toISOString(),
      entries: [{
        id: 200,
        queue_id: 7,
        student_id: 42,
        position: 1,
        joined_at: new Date().toISOString(),
        username: 'Student User',
      }],
    });

    render(<StudentDashboard />);
    fireEvent.click(screen.getByRole('button', { name: 'Join Queue' }));
    await screen.findByText('Real-Time Queue Status');

    fireEvent.click(screen.getByRole('button', { name: /Cancel & Leave Queue/i }));

    await waitFor(() => {
      expect(leaveQueue).toHaveBeenCalledWith(7);
      expect(screen.queryByText('Real-Time Queue Status')).not.toBeInTheDocument();
    });
  });

  it('auto-exits queue view when queue refresh no longer contains current student', async () => {
    let eventHandler: (() => void) | undefined;

    vi.mocked(subscribeToQueueEvents).mockImplementation((_, onEvent) => {
      eventHandler = () => void onEvent({ type: 'QUEUE_UPDATED', queue_id: 7 });
      return () => undefined;
    });

    vi.mocked(getQueueOrNull)
      .mockResolvedValueOnce({
        id: 7,
        course_id: 1,
        ta_id: 10,
        status: 'open',
        created_at: new Date().toISOString(),
        entries: [],
      })
      .mockResolvedValueOnce({
        id: 7,
        course_id: 1,
        ta_id: 10,
        status: 'open',
        created_at: new Date().toISOString(),
        entries: [{
          id: 200,
          queue_id: 7,
          student_id: 42,
          position: 1,
          joined_at: new Date().toISOString(),
          username: 'Student User',
        }],
      })
      .mockResolvedValueOnce({
        id: 7,
        course_id: 1,
        ta_id: 10,
        status: 'open',
        created_at: new Date().toISOString(),
        entries: [],
      });

    render(<StudentDashboard />);
    fireEvent.click(screen.getByRole('button', { name: 'Join Queue' }));
    await screen.findByText('Real-Time Queue Status');

    eventHandler?.();

    await waitFor(() => {
      expect(screen.queryByText('Real-Time Queue Status')).not.toBeInTheDocument();
      expect(screen.getByRole('button', { name: 'Join Queue' })).toBeEnabled();
    });
  });
});

describe('TADashboard page', () => {
  it('renders closed dashboard and requires selecting office-hour time first', () => {
    render(<TADashboard />);
    expect(screen.getByText('Select Time to Start Live Queue')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Select Time to Start Live Queue/i })).toBeDisabled();
  });

  it('renders welcome message with the TA username', () => {
    render(<TADashboard />);
    expect(screen.getByText(/Welcome Back, Test TA/i)).toBeInTheDocument();
  });

  it('renders all closed-state stat cards', () => {
    render(<TADashboard />);
    expect(screen.getByText('Students Helped Today')).toBeInTheDocument();
    expect(screen.getByText('Avg Wait Time')).toBeInTheDocument();
    expect(screen.getByText('Current Queue Length')).toBeInTheDocument();
    expect(screen.getByText('Longest Wait Time')).toBeInTheDocument();
    expect(screen.getByText('Most Common Topic')).toBeInTheDocument();
    expect(screen.getByText('Avg Session Duration')).toBeInTheDocument();
  });

  it('enables start button after selecting an office-hour card', () => {
    render(<TADashboard />);
    fireEvent.click(screen.getAllByText('11:00 AM - 1:00 PM')[0]);
    expect(screen.getByRole('button', { name: /Start Office Hours Live Queue/i })).not.toBeDisabled();
  });

  it('opens live queue and calls backend queue creation flow', async () => {
    render(<TADashboard />);
    fireEvent.click(screen.getAllByText('11:00 AM - 1:00 PM')[0]);
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));

    expect(await screen.findByText('Live Queue')).toBeInTheDocument();
    expect(createQueue).toHaveBeenCalledWith(2);
    expect(updateQueueStatus).toHaveBeenCalledWith(10, 'open');
  });

  it('pauses and resumes the queue through backend status API', async () => {
    render(<TADashboard />);
    fireEvent.click(screen.getAllByText('11:00 AM - 1:00 PM')[0]);
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));
    await screen.findByText('Live Queue');

    vi.mocked(updateQueueStatus).mockResolvedValueOnce({ id: 10, status: 'paused' });
    fireEvent.click(screen.getByText('Pause Queue'));
    await screen.findByText('Paused - No New Students');
    expect(updateQueueStatus).toHaveBeenCalledWith(10, 'paused');

    vi.mocked(updateQueueStatus).mockResolvedValueOnce({ id: 10, status: 'open' });
    fireEvent.click(screen.getByText('Resume Queue'));
    await screen.findByText('Open - Accepting Students');
    expect(updateQueueStatus).toHaveBeenCalledWith(10, 'open');
  });

  it('closes queue and returns to closed dashboard state', async () => {
    render(<TADashboard />);
    fireEvent.click(screen.getAllByText('11:00 AM - 1:00 PM')[0]);
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));
    await screen.findByText('Live Queue');

    vi.mocked(updateQueueStatus).mockResolvedValueOnce({ id: 10, status: 'closed' });
    fireEvent.click(screen.getByText('Close Queue'));

    await waitFor(() => {
      expect(screen.queryByText('Live Queue')).not.toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: /Start Office Hours Live Queue|Select Time to Start Live Queue/i })).toBeInTheDocument();
    expect(updateQueueStatus).toHaveBeenCalledWith(10, 'closed');
  });

  it('opens announcement modal from dashboard and sends announcement', async () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByText('Send Announcement'));
    expect(screen.getByPlaceholderText(/Running 10 minutes late/i)).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText(/Running 10 minutes late/i), {
      target: { value: 'Office hours extended by 30 minutes!' },
    });
    fireEvent.click(screen.getByText('Send to All'));

    await waitFor(() => {
      expect(hoisted.toast.success).toHaveBeenCalledWith('Announcement sent!', {
        description: 'Office hours extended by 30 minutes!',
      });
    });
    expect(screen.queryByPlaceholderText(/Running 10 minutes late/i)).not.toBeInTheDocument();
  });

  it('calls next endpoint when starting session on first queued student', async () => {
    vi.mocked(getQueueOrNull).mockResolvedValue({
      id: 10,
      course_id: 2,
      ta_id: 10,
      status: 'open',
      created_at: new Date().toISOString(),
      entries: [
        {
          id: 1,
          queue_id: 10,
          student_id: 101,
          position: 1,
          joined_at: new Date(Date.now() - 8 * 60 * 1000).toISOString(),
          username: 'student1',
        },
      ],
    });

    render(<TADashboard />);
    fireEvent.click(screen.getAllByText('11:00 AM - 1:00 PM')[0]);
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));

    expect(await screen.findByText('student1')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Start Session'));

    await waitFor(() => {
      expect(nextQueueStudent).toHaveBeenCalledWith(10);
    });
  });
});
