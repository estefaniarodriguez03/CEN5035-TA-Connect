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
    updateUser: vi.fn(),
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
    useLocation: () => ({ state: hoisted.locationState, pathname: '/student' }),
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
  updateQueueState: vi.fn(),
  nextQueueStudent: vi.fn(),
  setActiveQueueForCourse: vi.fn(),
  clearActiveQueueForCourse: vi.fn(),
  setActiveQueueForOfficeHour: vi.fn(),
  clearActiveQueueForOfficeHour: vi.fn(),
  getActiveQueueForOfficeHour: vi.fn(),
  joinQueue: vi.fn(),
  leaveQueue: vi.fn(),
  postQueueAnnouncement: vi.fn(),
  startSession: vi.fn(),
  browseWaitDisplay: vi.fn(),
  myWaitMinutesFromEntry: vi.fn(),
}));

vi.mock('../api/courses', () => ({
  listMyTACourses: vi.fn(),
  addMyTACourse: vi.fn(),
  removeMyTACourse: vi.fn(),
  listAllCourses: vi.fn(),
}));

vi.mock('../api/officeHours', () => ({
  getOfficeHoursByTA: vi.fn(),
  listOfficeHoursByTA: vi.fn(),
  listOfficeHoursByCourse: vi.fn(),
  createOfficeHour: vi.fn(),
  updateOfficeHour: vi.fn(),
  deleteOfficeHour: vi.fn(),
  DAY_NAMES: ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],
}));

vi.mock('../api/studentSchedule', () => ({
  listStudentSchedule: vi.fn(),
  addToStudentSchedule: vi.fn(),
  removeFromStudentSchedule: vi.fn(),
}));

vi.mock('../api/profile', () => ({
  updateUserProfile: vi.fn(),
}));

vi.mock('../api/studentCourses', () => ({
  getStudentCourses: vi.fn(),
  updateStudentCourses: vi.fn(),
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
import MyCoursesPage from './MyCoursesPage';
import MyOfficeHoursPage from './MyOfficeHoursPage';

import { login, register } from '../api/auth';
import {
  getActiveQueueByCourse,
  createQueue,
  getQueueOrNull,
  subscribeToQueueEvents,
  updateQueueStatus,
  updateQueueState,
  nextQueueStudent,
  getActiveQueueForOfficeHour,
  joinQueue,
  leaveQueue,
  postQueueAnnouncement,
  startSession,
  browseWaitDisplay,
  myWaitMinutesFromEntry,
} from '../api/queue';
import {
  listMyTACourses,
  addMyTACourse,
  listAllCourses,
} from '../api/courses';
import {
  listOfficeHoursByTA,
  listOfficeHoursByCourse,
  createOfficeHour,
  updateOfficeHour,
  deleteOfficeHour,
} from '../api/officeHours';
import {
  listStudentSchedule,
  addToStudentSchedule,
  removeFromStudentSchedule,
} from '../api/studentSchedule';
import { updateUserProfile } from '../api/profile';
import { getStudentCourses } from '../api/studentCourses';

// ─── Shared mock data ─────────────────────────────────────────────────────────

const mockOfficeHour = {
  id: 1,
  ta_id: 10,
  ta_username: 'Test TA',
  course_id: 2,
  day_of_week: new Date().getDay(),
  start_time: '11:00:00',
  end_time: '13:00:00',
  location: 'CSE E222',
};

const mockScheduleEntry = {
  id: 1,
  office_hour_id: 1,
  course_id: 2,
  course_code: 'COP3530',
  course_name: 'Data Structures',
  course_color: 'orange',
  ta_username: 'Test TA',
  ta_id: 10,
  day_of_week: new Date().getDay(),
  start_time: '11:00:00',
  end_time: '13:00:00',
  location: 'CSE E222',
};

const mockCourse = { id: 2, code: 'COP3530', name: 'Data Structures' };

// The full option label as it appears in the dropdown
const MOCK_COURSE_LABEL = 'COP3530 – Data Structures – Test TA (11:00 AM - 1:00 PM)';

// Mock fetch to simulate the active queue poll in StudentDashboard
function mockActiveFetch() {
  vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
    ok: true,
    json: async () => ({ id: 7, status: 'open' }),
  }));
}

// Helper to click the schedule card without ambiguity — the time text
// appears both in the schedule card (.time-text) and the sidebar (<span>).
async function clickScheduleCard() {
  const timeTexts = await screen.findAllByText('11:00 AM - 1:00 PM');
  const card = timeTexts.find(el => el.classList.contains('time-text'));
  expect(card).toBeTruthy();
  fireEvent.click(card!);
}

// ─── beforeEach / afterEach ───────────────────────────────────────────────────

beforeEach(() => {
  vi.clearAllMocks();
  hoisted.locationState = undefined;
  hoisted.auth.user = { id: 10, username: 'Test TA', email: 'ta@test.com', role: 'ta' };

  // Queue mocks
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
  vi.mocked(updateQueueState).mockResolvedValue({ id: 10, status: 'open' });
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
  vi.mocked(startSession).mockResolvedValue({
    id: 1,
    queue_id: 10,
    ta_id: 10,
    student_id: 101,
    zoom_meeting_id: '123456789',
    zoom_join_url: 'https://zoom.us/j/123456789',
    zoom_start_url: 'https://zoom.us/s/123456789',
    zoom_passcode: 'abc123',
    started_at: new Date().toISOString(),
  });
  vi.mocked(getActiveQueueForOfficeHour).mockReturnValue(7);
  vi.mocked(joinQueue).mockResolvedValue({
    id: 99,
    queue_id: 7,
    position: 2,
    joined_at: new Date().toISOString(),
  });
  vi.mocked(leaveQueue).mockResolvedValue();
  vi.mocked(postQueueAnnouncement).mockResolvedValue(undefined);
  vi.mocked(browseWaitDisplay).mockReturnValue({ line: '~4 min', sub: '1 student in queue' });
  vi.mocked(myWaitMinutesFromEntry).mockReturnValue(4);

  // Course mocks
  vi.mocked(listMyTACourses).mockResolvedValue([mockCourse]);
  vi.mocked(addMyTACourse).mockResolvedValue(mockCourse);
  vi.mocked(listAllCourses).mockResolvedValue([
    mockCourse,
    { id: 3, code: 'CEN3031', name: 'Software Engineering' },
  ]);
  vi.mocked(getStudentCourses).mockResolvedValue([
    { id: 2, code: 'COP3530', name: 'Data Structures', color: 'orange' },
  ]);

  // Office hours mocks
  vi.mocked(listOfficeHoursByTA).mockResolvedValue([mockOfficeHour]);
  vi.mocked(listOfficeHoursByCourse).mockResolvedValue([mockOfficeHour]);
  vi.mocked(createOfficeHour).mockResolvedValue(mockOfficeHour);
  vi.mocked(updateOfficeHour).mockResolvedValue({
    ...mockOfficeHour,
    start_time: '14:00:00',
    end_time: '15:00:00',
  });
  vi.mocked(deleteOfficeHour).mockResolvedValue();

  // Student schedule mocks
  vi.mocked(listStudentSchedule).mockResolvedValue([mockScheduleEntry]);
  vi.mocked(addToStudentSchedule).mockResolvedValue(undefined);
  vi.mocked(removeFromStudentSchedule).mockResolvedValue(undefined);

  // Profile mocks
  vi.mocked(updateUserProfile).mockResolvedValue({
    id: 10,
    username: 'Updated Name',
    email: 'ta@test.com',
    role: 'ta',
  });
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

// ─── Login ────────────────────────────────────────────────────────────────────

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

// ─── Register ─────────────────────────────────────────────────────────────────

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

  it('shows fallback registration error message', async () => {
    vi.mocked(register).mockRejectedValue(new Error('unknown'));

    render(<Register />);
    fireEvent.click(screen.getByRole('button', { name: 'Register' }));

    expect(
      await screen.findByText('Registration failed: unknown')
    ).toBeInTheDocument();
  });
});

// ─── Student Dashboard ────────────────────────────────────────────────────────

describe('StudentDashboard page', () => {
  beforeEach(() => {
    hoisted.auth.user = { id: 42, username: 'Student User', email: 'student@test.com', role: 'student' };
  });

  it('renders student greeting', () => {
    render(<StudentDashboard />);
    expect(screen.getByText(/Welcome Back, Student User!/i)).toBeInTheDocument();
  });

  it('renders the course dropdown', () => {
    render(<StudentDashboard />);
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('populates dropdown with schedule entries from API', async () => {
    render(<StudentDashboard />);
    await waitFor(() => {
      expect(listStudentSchedule).toHaveBeenCalled();
    });
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('shows placeholder when no schedule entries exist', async () => {
    vi.mocked(listStudentSchedule).mockResolvedValue([]);
    render(<StudentDashboard />);
    await waitFor(() => {
      const matches = screen.getAllByText(/No courses added yet/i);
      expect(matches.length).toBeGreaterThan(0);
    });
  });

  it('auto-selects course from navigation state', async () => {
    hoisted.locationState = {
      autoSelectLabel: MOCK_COURSE_LABEL,
    } as any;
    render(<StudentDashboard />);
    await waitFor(() => {
      expect(hoisted.navigate).toHaveBeenCalledWith(
        '/student',
        expect.objectContaining({ state: null })
      );
    });
  });

  it('displays all five days in the weekly schedule', () => {
    render(<StudentDashboard />);
    ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday'].forEach((day) => {
      expect(screen.getByText(day)).toBeInTheDocument();
    });
  });

  it('Join Queue button is disabled when no course is selected', async () => {
    vi.mocked(listStudentSchedule).mockResolvedValue([]);
    render(<StudentDashboard />);
    const joinButton = screen.getByRole('button', { name: 'Join Queue' });
    expect(joinButton).toBeDisabled();
  });

  it('joins queue and shows real-time section', async () => {
    mockActiveFetch();

    vi.mocked(getQueueOrNull).mockResolvedValue({
      id: 7,
      course_id: 2,
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
    const dropdown = screen.getByRole('combobox');
    await waitFor(() => {
      expect(dropdown.querySelectorAll('option').length).toBeGreaterThan(1);
    });
    fireEvent.change(dropdown, { target: { value: MOCK_COURSE_LABEL } });

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Join Queue' })).not.toBeDisabled();
    });

    fireEvent.click(screen.getByRole('button', { name: 'Join Queue' }));
    expect(await screen.findByText('Real-Time Queue Status')).toBeInTheDocument();
    expect(joinQueue).toHaveBeenCalled();
  });

  it('leaves queue and hides real-time section', async () => {
    mockActiveFetch();

    vi.mocked(getQueueOrNull).mockResolvedValue({
      id: 7,
      course_id: 2,
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
    const dropdown = screen.getByRole('combobox');
    await waitFor(() => expect(dropdown.querySelectorAll('option').length).toBeGreaterThan(1));
    fireEvent.change(dropdown, { target: { value: MOCK_COURSE_LABEL } });

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Join Queue' })).not.toBeDisabled();
    });

    fireEvent.click(screen.getByRole('button', { name: 'Join Queue' }));
    await screen.findByText('Real-Time Queue Status');

    fireEvent.click(screen.getByRole('button', { name: /Cancel & Leave Queue/i }));

    await waitFor(() => {
      expect(leaveQueue).toHaveBeenCalled();
      expect(screen.queryByText('Real-Time Queue Status')).not.toBeInTheDocument();
    });
  });
});

// ─── TA Dashboard ─────────────────────────────────────────────────────────────

describe('TADashboard page', () => {
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

  it('loads TA courses from API on mount', async () => {
    render(<TADashboard />);
    await waitFor(() => {
      expect(listMyTACourses).toHaveBeenCalled();
    });
  });

  it('loads office hours from API on mount', async () => {
    render(<TADashboard />);
    await waitFor(() => {
      expect(listOfficeHoursByTA).toHaveBeenCalledWith(10);
    });
  });

  it('displays Send Announcement button on closed dashboard', () => {
    render(<TADashboard />);
    expect(screen.getByRole('button', { name: 'Send Announcement' })).toBeInTheDocument();
  });

  it('disables start button when no office hour is selected', async () => {
    vi.mocked(listOfficeHoursByTA).mockResolvedValue([]);
    render(<TADashboard />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Select Time to Start Live Queue/i })).toBeDisabled();
    });
  });

  it('enables start button after selecting an office-hour card', async () => {
    render(<TADashboard />);
    await clickScheduleCard();
    expect(screen.getByRole('button', { name: /Start Office Hours Live Queue/i })).not.toBeDisabled();
  });

  it('opens live queue and calls backend queue creation flow', async () => {
    render(<TADashboard />);
    await clickScheduleCard();
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));

    expect(await screen.findByText('Live Queue')).toBeInTheDocument();
    expect(createQueue).toHaveBeenCalledWith(2);
    expect(updateQueueState).toHaveBeenCalledWith(10, 'open');
  });

  it('closes queue and returns to closed dashboard state', async () => {
    render(<TADashboard />);
    await clickScheduleCard();
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));
    await screen.findByText('Live Queue');

    vi.mocked(updateQueueState).mockResolvedValueOnce({ id: 10, status: 'closed' });
    fireEvent.click(screen.getByText('Close Queue'));

    await waitFor(() => {
      expect(screen.queryByText('Live Queue')).not.toBeInTheDocument();
    });
    expect(updateQueueState).toHaveBeenCalledWith(10, 'closed');
  });

  it('opens announcement modal and sends announcement', async () => {
    render(<TADashboard />);
    await clickScheduleCard();
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));
    await screen.findByText('Live Queue');

    fireEvent.click(screen.getAllByText('Send Announcement')[0]);
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
  });

  it('calls startSession when Start Session is clicked on first queued student', async () => {
    vi.mocked(getQueueOrNull).mockResolvedValue({
      id: 10,
      course_id: 2,
      ta_id: 10,
      status: 'open',
      created_at: new Date().toISOString(),
      entries: [{
        id: 1,
        queue_id: 10,
        student_id: 101,
        position: 1,
        joined_at: new Date(Date.now() - 8 * 60 * 1000).toISOString(),
        username: 'student1',
      }],
    });

    render(<TADashboard />);
    await clickScheduleCard();
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));

    expect(await screen.findByText('student1')).toBeInTheDocument();
    fireEvent.click(screen.getByText('Start Session'));

    await waitFor(() => {
      expect(startSession).toHaveBeenCalledWith(10, 101);
    });
  });

  it('shows session modal with Zoom details after starting session', async () => {
    vi.mocked(getQueueOrNull).mockResolvedValue({
      id: 10,
      course_id: 2,
      ta_id: 10,
      status: 'open',
      created_at: new Date().toISOString(),
      entries: [{
        id: 1,
        queue_id: 10,
        student_id: 101,
        position: 1,
        joined_at: new Date().toISOString(),
        username: 'student1',
      }],
    });

    render(<TADashboard />);
    await clickScheduleCard();
    fireEvent.click(screen.getByRole('button', { name: /Start Office Hours Live Queue/i }));
    await screen.findByText('student1');
    fireEvent.click(screen.getByText('Start Session'));

    await waitFor(() => {
      expect(screen.getByText('Session with student1')).toBeInTheDocument();
    });
    expect(screen.getByText('123456789')).toBeInTheDocument();
  });
});

// ─── TA Dashboard — My Office Hours tab ──────────────────────────────────────

describe('TADashboard — My Office Hours tab', () => {
  it('switches to My Office Hours tab when clicked', async () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByRole('button', { name: 'My Office Hours' }));
    expect(await screen.findByText('Manage your office hour schedule and availability')).toBeInTheDocument();
  });

  it('switches back to Dashboard tab from My Office Hours', async () => {
    render(<TADashboard />);
    fireEvent.click(screen.getByRole('button', { name: 'My Office Hours' }));
    await screen.findByText('Manage your office hour schedule and availability');

    fireEvent.click(screen.getByRole('button', { name: 'Dashboard' }));
    expect(await screen.findByText(/Welcome Back, Test TA/i)).toBeInTheDocument();
    expect(screen.queryByText('Manage your office hour schedule and availability')).not.toBeInTheDocument();
  });
});

// ─── MyOfficeHoursPage ────────────────────────────────────────────────────────

describe('MyOfficeHoursPage', () => {
  const mockOfficeHours = [mockOfficeHour];
  const mockOnChange = vi.fn();

  it('renders all seven days of the week', () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={mockOfficeHours}
        onOfficeHoursChange={mockOnChange}
      />
    );
    ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'].forEach((day) => {
      expect(screen.getByText(day)).toBeInTheDocument();
    });
  });

  it('displays existing office hours', () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={mockOfficeHours}
        onOfficeHoursChange={mockOnChange}
      />
    );
    expect(screen.getByText(/11:00 AM/)).toBeInTheDocument();
  });

  it('shows empty state for days with no office hours', () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={[]}
        onOfficeHoursChange={mockOnChange}
      />
    );
    expect(screen.getAllByText('No office hours scheduled').length).toBeGreaterThan(0);
  });

  it('opens the Add Office Hours modal when button is clicked', async () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={[]}
        onOfficeHoursChange={mockOnChange}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: '+ Add Office Hours' }));
    expect(await screen.findByText('Add Office Hours')).toBeInTheDocument();
    expect(screen.getAllByRole('combobox').length).toBeGreaterThan(0);
  });

  it('shows course dropdown in modal with linked courses', async () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={[]}
        onOfficeHoursChange={mockOnChange}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: '+ Add Office Hours' }));
    await waitFor(() => {
      expect(screen.getByText('COP3530 - Data Structures')).toBeInTheDocument();
    });
  });

  it('disables Add Office Hours button when no courses are linked', () => {
    render(
      <MyOfficeHoursPage
        taCourses={[]}
        officeHours={[]}
        onOfficeHoursChange={mockOnChange}
      />
    );
    expect(screen.getByRole('button', { name: '+ Add Office Hours' })).toBeDisabled();
  });

  it('calls createOfficeHour and onOfficeHoursChange when form is submitted', async () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={[]}
        onOfficeHoursChange={mockOnChange}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: '+ Add Office Hours' }));

    const timeInputs = document.querySelectorAll('input[type="time"]');
    expect(timeInputs.length).toBe(2);
    fireEvent.change(timeInputs[0], { target: { value: '10:00' } });
    fireEvent.change(timeInputs[1], { target: { value: '11:00' } });

    const courseSelect = screen.getAllByRole('combobox')[1];
    fireEvent.change(courseSelect, { target: { value: '2' } });

    const submitBtn = screen.getByRole('button', { name: 'Add Schedule' });
    expect(submitBtn).not.toBeDisabled();
    fireEvent.click(submitBtn);

    await waitFor(() => {
      expect(createOfficeHour).toHaveBeenCalledWith(
        expect.objectContaining({
          course_id: 2,
          start_time: '10:00',
          end_time: '11:00',
        })
      );
    });
  });

  it('calls deleteOfficeHour and onOfficeHoursChange when Delete is clicked', async () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={mockOfficeHours}
        onOfficeHoursChange={mockOnChange}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));

    await waitFor(() => {
      expect(deleteOfficeHour).toHaveBeenCalledWith(1);
      expect(mockOnChange).toHaveBeenCalled();
    });
  });

  it('opens edit modal pre-filled when Edit is clicked', async () => {
    render(
      <MyOfficeHoursPage
        taCourses={[mockCourse]}
        officeHours={mockOfficeHours}
        onOfficeHoursChange={mockOnChange}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: 'Edit' }));
    expect(await screen.findByText('Edit Office Hours')).toBeInTheDocument();
  });
});

// ─── MyCoursesPage ────────────────────────────────────────────────────────────

describe('MyCoursesPage', () => {
  beforeEach(() => {
    hoisted.auth.user = { id: 42, username: 'Student User', email: 'student@test.com', role: 'student' };
  });

  it('renders my courses page with weekly view by default', async () => {
    render(<MyCoursesPage />);
    expect(await screen.findByText("This Week's Office Hours")).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Weekly' })).toHaveClass('active');
  });

  it('loads and displays student schedule from API', async () => {
    render(<MyCoursesPage />);
    await waitFor(() => {
      expect(listStudentSchedule).toHaveBeenCalled();
    });
    expect(await screen.findByText('COP3530')).toBeInTheDocument();
  });

  it('switches to today view when clicking Today button', async () => {
    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: 'Today' }));
    await waitFor(() => {
      expect(screen.getByText("Today's Office Hours")).toBeInTheDocument();
    });
  });

  it('opens Add Office Hours modal when clicking add button', async () => {
    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: /Add Office Hours/i }));
    expect(await screen.findByText('Course')).toBeInTheDocument();
  });

  it('loads student courses when modal is opened', async () => {
    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: /Add Office Hours/i }));

    await waitFor(() => {
      expect(getStudentCourses).toHaveBeenCalled();
    });

    const options = screen.getAllByRole('option');
    expect(options.some(o => o.textContent?.includes('COP3530'))).toBe(true);
  });

  it('shows TA dropdown after selecting a course', async () => {
    vi.mocked(getStudentCourses).mockResolvedValue([
      { id: 2, code: 'COP3530', name: 'Data Structures', color: 'orange' },
    ]);
    vi.mocked(listOfficeHoursByCourse).mockResolvedValue([mockOfficeHour]);

    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: /Add Office Hours/i }));

    await waitFor(() => {
      expect(getStudentCourses).toHaveBeenCalled();
    });

    const modalSelect = document.querySelector('select.form-select') as HTMLElement;
    expect(modalSelect).toBeTruthy();
    fireEvent.change(modalSelect, { target: { value: '2' } });

    await waitFor(() => {
      expect(listOfficeHoursByCourse).toHaveBeenCalledWith(2);
    });
  });

  it('Add to Schedule button is disabled when no slot is selected', async () => {
    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: /Add Office Hours/i }));
    expect(screen.getByRole('button', { name: 'Add to Schedule' })).toBeDisabled();
  });

  it('closes modal when Cancel is clicked', async () => {
    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: /Add Office Hours/i }));
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }));
    await waitFor(() => {
      expect(screen.queryByRole('button', { name: 'Add to Schedule' })).not.toBeInTheDocument();
    });
  });

  it('calls removeFromStudentSchedule when × delete button is clicked', async () => {
    render(<MyCoursesPage />);
    await screen.findByText('COP3530');
    const deleteBtn = screen.getAllByRole('button', { name: '×' })[0];
    fireEvent.click(deleteBtn);
    await waitFor(() => {
      expect(removeFromStudentSchedule).toHaveBeenCalledWith(1);
    });
  });

  it('navigates to /student when Dashboard nav is clicked', async () => {
    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: 'Dashboard' }));
    expect(hoisted.navigate).toHaveBeenCalledWith('/student');
  });

  it('navigates to /student with autoSelectLabel when Join Queue is clicked in today view', async () => {
    render(<MyCoursesPage />);
    await screen.findByText("This Week's Office Hours");
    fireEvent.click(screen.getByRole('button', { name: 'Today' }));

    const joinBtn = await screen.findByRole('button', { name: /Join Queue/i });
    fireEvent.click(joinBtn);

    await waitFor(() => {
      expect(hoisted.navigate).toHaveBeenCalledWith(
        '/student',
        expect.objectContaining({
          state: expect.objectContaining({ autoSelectLabel: expect.stringContaining('COP3530') }),
        })
      );
    });
  });
});