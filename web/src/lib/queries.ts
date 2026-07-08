import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './api'
import type {
  AISummary,
  BloodSugarReading,
  CalendarDay,
  Category,
  JournalEntry,
  Recurrence,
  Stats,
  SummaryResponse,
  Task,
} from './types'

export interface TaskFilters {
  date?: string
  from?: string
  to?: string
  status?: string
  priority?: string
  category_id?: string
}

function taskQueryString(filters: TaskFilters): string {
  const params = new URLSearchParams()
  for (const [k, v] of Object.entries(filters)) {
    if (v) params.set(k, v)
  }
  const qs = params.toString()
  return qs ? `?${qs}` : ''
}

export function useTasks(filters: TaskFilters = {}) {
  return useQuery({
    queryKey: ['tasks', filters],
    queryFn: () => api.get<Task[]>(`/tasks${taskQueryString(filters)}`),
  })
}

export function useCreateTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Task> & { title: string }) => api.post<Task>('/tasks', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useUpdateTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: Partial<Task> }) =>
      api.patch<Task>(`/tasks/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useDeleteTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.del(`/tasks/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

function invalidateTimerAndTasks(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ['tasks'] })
  qc.invalidateQueries({ queryKey: ['timer'] })
  qc.invalidateQueries({ queryKey: ['calendar'] })
  qc.invalidateQueries({ queryKey: ['stats'] })
}

export function useStartTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.post<Task>(`/tasks/${id}/start`),
    onSuccess: () => invalidateTimerAndTasks(qc),
  })
}

export function usePauseTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.post<Task>(`/tasks/${id}/pause`),
    onSuccess: () => invalidateTimerAndTasks(qc),
  })
}

export function useCompleteTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.post<Task>(`/tasks/${id}/complete`),
    onSuccess: () => invalidateTimerAndTasks(qc),
  })
}

export function useCurrentTimer() {
  return useQuery({
    queryKey: ['timer', 'current'],
    queryFn: () => api.get<{ current: Task | null }>('/timer/current'),
    refetchInterval: 5000,
  })
}

export function useNextTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ current: Task | null }>('/timer/next'),
    onSuccess: () => invalidateTimerAndTasks(qc),
  })
}

export function usePauseCurrent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ current: Task | null }>('/timer/pause'),
    onSuccess: () => invalidateTimerAndTasks(qc),
  })
}

export function useCategories() {
  return useQuery({
    queryKey: ['categories'],
    queryFn: () => api.get<Category[]>('/categories'),
  })
}

export function useCreateCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { name: string; color: string }) => api.post<Category>('/categories', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
  })
}

export function useUpdateCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: Partial<Category> }) =>
      api.patch<Category>(`/categories/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
  })
}

export function useDeleteCategory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.del(`/categories/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['categories'] }),
  })
}

export function useRecurrences() {
  return useQuery({
    queryKey: ['recurrences'],
    queryFn: () => api.get<Recurrence[]>('/recurrences'),
  })
}

export function useCreateRecurrence() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: Partial<Recurrence> & { title: string; freq: string; start_date: string }) =>
      api.post<Recurrence>('/recurrences', input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['recurrences'] })
      qc.invalidateQueries({ queryKey: ['tasks'] })
    },
  })
}

export function useUpdateRecurrence() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: Partial<Recurrence> }) =>
      api.patch<Recurrence>(`/recurrences/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recurrences'] }),
  })
}

export function useDeleteRecurrence() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, hard }: { id: number; hard?: boolean }) =>
      api.del(`/recurrences/${id}${hard ? '?hard=1' : ''}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['recurrences'] })
      qc.invalidateQueries({ queryKey: ['tasks'] })
    },
  })
}

export function useMonth(year: number, month: number, categoryId?: string) {
  return useQuery({
    queryKey: ['calendar', 'month', year, month, categoryId],
    queryFn: () =>
      api.get<CalendarDay[]>(
        `/calendar/month?year=${year}&month=${month}${categoryId ? `&category_id=${categoryId}` : ''}`,
      ),
  })
}

export function useJournal(params: { date?: string; from?: string; to?: string } = {}) {
  const qs = new URLSearchParams(
    Object.entries(params).filter(([, v]) => v) as [string, string][],
  ).toString()
  return useQuery({
    queryKey: ['journal', params],
    queryFn: () => api.get<JournalEntry[]>(`/journal${qs ? `?${qs}` : ''}`),
  })
}

export function useCreateJournal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { date: string; title?: string; content: string; mood?: string | null }) =>
      api.post<JournalEntry>('/journal', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['journal'] }),
  })
}

export function useUpdateJournal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: Partial<JournalEntry> }) =>
      api.patch<JournalEntry>(`/journal/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['journal'] }),
  })
}

export function useDeleteJournal() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.del(`/journal/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['journal'] }),
  })
}

export function useStats(from?: string, to?: string, categoryId?: string) {
  const qs = new URLSearchParams({
    ...(from && { from }),
    ...(to && { to }),
    ...(categoryId && { category_id: categoryId }),
  }).toString()
  return useQuery({
    queryKey: ['stats', from, to, categoryId],
    queryFn: () => api.get<Stats>(`/stats${qs ? `?${qs}` : ''}`),
  })
}

export function useBloodSugar(from?: string, to?: string) {
  const qs = new URLSearchParams({
    ...(from && { from }),
    ...(to && { to }),
  }).toString()
  return useQuery({
    queryKey: ['bloodsugar', from, to],
    queryFn: () => api.get<BloodSugarReading[]>(`/bloodsugar${qs ? `?${qs}` : ''}`),
  })
}

export function useCreateReading() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: { value_mgdl: number; taken_at?: string; meal_tag?: string; notes?: string }) =>
      api.post<BloodSugarReading>('/bloodsugar', input),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bloodsugar'] }),
  })
}

export function useUpdateReading() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: number; patch: Partial<BloodSugarReading> }) =>
      api.patch<BloodSugarReading>(`/bloodsugar/${id}`, patch),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bloodsugar'] }),
  })
}

export function useDeleteReading() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.del(`/bloodsugar/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bloodsugar'] }),
  })
}

export function useSyncMeter() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ synced: number; fetched: number }>('/bloodsugar/sync'),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['bloodsugar'] }),
  })
}

export function useGenerateBloodSugarSummary() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ from, to }: { from: string; to: string }) =>
      api.post<{ enabled: boolean; cached: boolean; summary: AISummary }>(
        `/summary/bloodsugar?from=${from}&to=${to}`,
      ),
    onSuccess: (_data, vars) =>
      qc.invalidateQueries({ queryKey: ['summary', 'blood_sugar', `${vars.from}..${vars.to}`] }),
  })
}

export function useSummary(period: string, key: string) {
  return useQuery({
    queryKey: ['summary', period, key],
    queryFn: () => api.get<SummaryResponse>(`/summary?period=${period}&key=${key}`),
  })
}

export function useGenerateSummary() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ period, key }: { period: string; key: string }) =>
      api.post<{ enabled: boolean; cached: boolean; summary: AISummary }>(
        `/summary/generate?period=${period}&key=${key}`,
      ),
    onSuccess: (_data, vars) =>
      qc.invalidateQueries({ queryKey: ['summary', vars.period, vars.key] }),
  })
}

export function useGenerateJournalSummary() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ from, to }: { from: string; to: string }) =>
      api.post<{ enabled: boolean; cached: boolean; summary: AISummary }>(
        `/summary/journal?from=${from}&to=${to}`,
      ),
    onSuccess: (_data, vars) =>
      qc.invalidateQueries({ queryKey: ['summary', 'journal', `${vars.from}..${vars.to}`] }),
  })
}
