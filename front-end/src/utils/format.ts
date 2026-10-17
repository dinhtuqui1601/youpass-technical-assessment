import dayjs from 'dayjs'

export function formatDateTime(value?: string | null): string {
  if (!value) {
    return '—'
  }
  const d = dayjs(value)
  return d.isValid() ? d.format('DD/MM/YYYY HH:mm') : '—'
}
