export type ID = number;

export interface PaginationMeta {
  total: number;
  limit: number;
  offset: number;
}

export interface Category {
  id: ID;
  name: string;
  description?: string;
}

export interface DatasetMetadata {
  format?: string;
  tags?: string[];
  size?: number;
}

export interface DatasetVersion {
  id: ID;
  dataset_id: ID;
  number: string;
  change_log?: string;
  file_url?: string;
  upload_date?: string;
  metadata?: DatasetMetadata;
}

export interface DatasetRatingSummary {
  average: number;
  count: number;
  synced_at?: string;
}

export interface Dataset {
  id: ID;
  name: string;
  description?: string;
  category_id: ID;
  owner_id: ID;
  is_public: boolean;
  created_at?: string;
  updated_at?: string;
  latest_version?: DatasetVersion;
  metadata?: DatasetMetadata;
  rating_summary?: DatasetRatingSummary;
}

export interface NotificationItem {
  id: ID;
  dataset_id?: ID;
  user_id: ID;
  message: string;
  is_read: boolean;
  created_at?: string;
}

export interface Review {
  id: ID;
  dataset_id: ID;
  user_id: ID;
  rating: number;
  text?: string;
  created_at?: string;
}

export interface Subscription {
  id: ID;
  dataset_id: ID;
  user_id: ID;
  subscribed_at?: string;
}

export type AccessRequestStatus = 'pending' | 'approved' | 'denied';

export interface AccessRequest {
  id: ID;
  dataset_id: ID;
  user_id: ID;
  status: AccessRequestStatus;
  created_at?: string;
}

export interface User {
  id: ID;
  username: string;
  email: string;
  country?: string;
  role: 'guest' | 'user' | 'admin';
  registration_date?: string;
  is_blocked: boolean;
}

export interface AuthenticateRequest {
  email: string;
  password: string;
}

export interface AuthenticateResponse {
  token: string;
  expires_at?: string;
  user: User;
}

export interface PaginatedResponse<T> {
  items: T[];
  meta: PaginationMeta;
}

export interface ErrorResponse {
  code: string;
  message: string;
}
