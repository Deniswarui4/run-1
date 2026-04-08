// User and Authentication Types
export type UserRole = 'admin' | 'moderator' | 'organizer' | 'attendee';

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role: UserRole;
  is_active: boolean;
  is_verified: boolean;
  two_factor_enabled?: boolean;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role: 'attendee' | 'organizer';
}

export interface LoginRequest {
  email: string;
  password: string;
}

// Event Types
export type EventStatus = 'draft' | 'pending' | 'approved' | 'rejected' | 'published' | 'cancelled';

export interface Category {
  id: string;
  name: string;
  description?: string;
  color?: string;
  icon?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export type FeaturedType = 'none' | 'manual' | 'auto';

export interface Event {
  id: string;
  title: string;
  description: string;
  start_date: string;
  end_date: string;
  venue: string;
  address?: string;
  city: string;
  country?: string;
  category: string;
  category_id?: string;
  category_info?: Category;
  image_url?: string;
  video_url?: string;
  status: 'draft' | 'pending' | 'approved' | 'published' | 'cancelled';
  is_featured: boolean;
  featured_type?: FeaturedType;
  organizer_id: string;
  organizer?: User;
  ticket_types: TicketType[];
  created_at: string;
  updated_at: string;
}

export interface CreateEventRequest {
  title: string;
  description: string;
  category: string;
  venue: string;
  address: string;
  city: string;
  country: string;
  start_date: string;
  end_date: string;
  video_url?: string;
}

// Ticket Types
export interface TicketType {
  id: string;
  event_id: string;
  name: string;
  description: string;
  price: number;
  quantity: number;
  sold: number;
  max_per_order: number;
  sale_start: string;
  sale_end: string;
  created_at: string;
}

export interface CreateTicketTypeRequest {
  name: string;
  description: string;
  price: number;
  quantity: number;
  max_per_order: number;
  sale_start: string;
  sale_end: string;
}

export interface Ticket {
  id: string;
  ticket_number: string;
  event_id: string;
  ticket_type_id: string;
  user_id: string;
  status: 'confirmed' | 'cancelled' | 'used';
  price: number;
  qr_code_url?: string;
  pdf_url?: string;
  checked_in_at?: string;
  checked_in_by?: string;
  event?: Event;
  ticket_type?: TicketType;
  attendee?: User;
  created_at: string;
}

// Cart Types
export interface CartItem {
  ticket_type_id: string;
  quantity: number;
  ticket_type: TicketType;
}

export interface PurchaseTicketRequest {
  event_id: string;
  items: Array<{
    ticket_type_id: string;
    quantity: number;
  }>;
}

export interface PurchaseTicketResponse {
  transaction_id: string;
  payment_reference: string;
  authorization_url: string;
  amount: number;
  currency: string;
}

// Transaction Types
export interface Transaction {
  id: string;
  user_id: string;
  event_id?: string;
  type: string;
  amount: number;
  currency: string;
  platform_fee: number;
  net_amount: number;
  status: 'pending' | 'completed' | 'failed';
  payment_gateway: string;
  payment_reference: string;
  description?: string;
  failure_reason?: string;
  created_at: string;
  updated_at: string;
}

// Balance and Withdrawal Types
export interface OrganizerBalance {
  id: string;
  organizer_id: string;
  total_earnings: number;
  available_balance: number;
  pending_balance: number;
  withdrawn_amount: number;
}

export interface WithdrawalRequest {
  amount: number;
  bank_name: string;
  account_number: string;
  account_name: string;
}

export interface Withdrawal {
  id: string;
  organizer_id: string;
  amount: number;
  withdrawal_fee: number;
  net_amount: number;
  status: 'pending' | 'approved' | 'rejected' | 'processed';
  bank_name: string;
  account_number: string;
  account_name: string;
  transaction_ref?: string;
  reviewed_by?: string;
  processed_by?: string;
  comment?: string;
  created_at: string;
  organizer?: {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
  };
}

// Platform Settings
export interface PlatformSettings {
  id: string;
  platform_fee_percentage: number;
  withdrawal_fee_percentage: number;
  min_withdrawal_amount: number;
  currency: string;
  updated_by: string;
  created_at: string;
  updated_at: string;
}

// Statistics Types
export interface PlatformStats {
  total_users: number;
  total_organizers: number;
  total_events: number;
  total_tickets_sold: number;
  total_revenue: number;
  platform_revenue: number;
}

export interface DailyCount {
  date: string; // "YYYY-MM-DD"
  count: number;
}

export interface EventRevenue {
  event_id: string;
  title: string;
  gross_revenue: number;
  tickets_sold: number;
}

export interface ExtendedPlatformStats extends PlatformStats {
  ticket_type_breakdown: TicketTypeBreakdown[];
  daily_registrations: DailyCount[];
  top_events_by_revenue: EventRevenue[];
}

export interface EventStats {
  total_tickets_sold: number;
  total_revenue: number;
  net_revenue: number;
  checked_in_tickets: number;
}

// Extended event stats types (Requirement 4.1, 4.2, 4.3)
export interface TicketTypeBreakdown {
  ticket_type_id: string;
  name: string;
  sold: number;
  remaining: number;
  gross_revenue: number;
  capacity_pct: number;
}

export interface DailyRevenue {
  date: string; // "YYYY-MM-DD"
  revenue: number;
}

export interface CheckInStats {
  total_confirmed: number;
  checked_in: number;
  check_in_rate: number;
}

export interface ExtendedEventStats extends EventStats {
  ticket_type_breakdown: TicketTypeBreakdown[];
  daily_revenue: DailyRevenue[];
  check_in_stats: CheckInStats;
}

export interface ModeratorStats {
  pending_events: number;
  approved_events: number;
  rejected_events: number;
  my_reviews: number;
}

// Review Types
export interface ReviewRequest {
  action: 'approve' | 'reject';
  comment: string;
}

// Draft Cart Types
export interface DraftCartItem {
  id: string;
  cart_id: string;
  ticket_type_id: string;
  quantity: number;
  ticket_type?: TicketType;
}

export interface DraftCart {
  id: string;
  user_id?: string;
  guest_session?: string;
  event_id: string;
  status: 'active' | 'expired' | 'completed';
  expires_at: string;
  items: DraftCartItem[];
}

export interface CartItemInput {
  ticket_type_id: string;
  quantity: number;
}

export interface GuestCheckoutRequest {
  event_id: string;
  email: string;
  items: CartItemInput[];
}

export interface GuestCheckoutResponse {
  authorization_url: string;
  payment_reference: string;
  amount: number;
}

// Guest Ticket Lookup Types
export interface GuestTicket {
  ticket_number: string;
  event_name: string;
  ticket_type: string;
  qr_code_url?: string;
}

export interface GuestTicketLookupResponse {
  tickets: GuestTicket[];
}

// API Error Response
export interface APIError {
  error: string;
  retry_after?: number;
}
