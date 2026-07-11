export interface Category {
  id: number
  name: string
  color: string
  created_at: string
}

export interface Recurrence {
  id: number
  title: string
  notes: string
  category_id: number | null
  priority: Priority
  estimated_seconds: number
  scheduled_start: string | null
  freq: 'daily' | 'weekly' | 'monthly'
  interval: number
  days_of_week: string
  day_of_month: number | null
  start_date: string
  end_date: string | null
  active: boolean
  created_at: string
  updated_at: string
}

export type Priority = 'low' | 'medium' | 'high' | 'urgent'
export type TaskStatus = 'todo' | 'in_progress' | 'paused' | 'done'

export interface Task {
  id: number
  title: string
  notes: string
  category_id: number | null
  priority: Priority
  status: TaskStatus
  estimated_seconds: number
  actual_seconds: number
  scheduled_date: string | null
  scheduled_start: string | null
  active_started_at: string | null
  completed_at: string | null
  recurrence_id: number | null
  sort_order: number
  created_at: string
  updated_at: string
  elapsed_seconds: number
  is_active: boolean
}

export interface JournalEntry {
  id: number
  date: string
  title: string
  content: string
  mood: string | null
  created_at: string
  updated_at: string
}

export type MealTag = 'fasting' | 'pre_meal' | 'post_meal' | 'bedtime' | 'random' | ''

export interface BloodSugarReading {
  id: number
  value_mgdl: number
  taken_at: string
  meal_tag: MealTag
  notes: string
  source: 'manual' | 'meter' | 'import'
  seq_number: number | null
  created_at: string
}

export interface CalendarDay {
  date: string
  total: number
  done: number
  ratio: number
  state: 'empty' | 'complete' | 'partial' | 'missed'
}

export interface EstVsActualDay {
  date: string
  estimated_seconds: number
  actual_seconds: number
  total: number
  done: number
}

export interface CategoryTime {
  category_id: number | null
  category_name: string
  seconds: number
}

export interface Stats {
  tasks_completed: number
  tasks_total: number
  completion_rate: number
  estimated_seconds_total: number
  actual_seconds_total: number
  est_vs_actual: EstVsActualDay[]
  time_by_category: CategoryTime[]
  current_streak: number
  busiest_category: string
  avg_daily_completion: number
}

export interface AISummary {
  id: number
  period_type: string
  period_key: string
  content: string
  model: string
  created_at: string
}

export interface SummaryResponse {
  enabled: boolean
  cached?: boolean
  summary?: AISummary
}
