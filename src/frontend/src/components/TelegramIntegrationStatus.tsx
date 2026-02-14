import type { TelegramStatus } from '../api/telegram'

type TelegramIntegrationStatusProps = {
  status: TelegramStatus | null
  isLoadingStatus: boolean
}

function formatDate(value: string | null): string {
  if (!value) {
    return 'Никогда'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString()
}

export default function TelegramIntegrationStatus({
  status,
  isLoadingStatus,
}: TelegramIntegrationStatusProps) {
  return (
    <section className="space-y-4 rounded-2xl border border-slate-200 bg-slate-50/80 p-4 sm:p-5">
      <div className="space-y-1">
        <h2 className="text-base font-semibold text-slate-900">Статус (за последние 7 дней)</h2>
      </div>

      {isLoadingStatus ? (
        <p className="text-sm text-slate-600">Загрузка...</p>
      ) : (
        <dl className="grid gap-3 text-sm sm:grid-cols-2">
          <div className="rounded-xl border border-slate-200 bg-white p-3">
            <dt className="text-xs font-medium uppercase tracking-wide text-slate-500">Включен</dt>
            <dd className="mt-1">
              <span
                className={[
                  'inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold ring-1 ring-inset',
                  status?.enabled
                    ? 'bg-emerald-50 text-emerald-700 ring-emerald-200'
                    : 'bg-rose-50 text-rose-700 ring-rose-200',
                ].join(' ')}
              >
                <span className="mr-1.5 h-1.5 w-1.5 rounded-full bg-current opacity-80" />
                {status?.enabled ? 'Да' : 'Нет'}
              </span>
            </dd>
          </div>
          <div className="rounded-xl border border-slate-200 bg-white p-3">
            <dt className="text-xs font-medium uppercase tracking-wide text-slate-500">Chat ID</dt>
            <dd className="mt-1 break-all text-base font-semibold text-slate-900">{status?.chatId || '-'}</dd>
          </div>
          <div className="rounded-xl border border-slate-200 bg-white p-3">
            <dt className="text-xs font-medium uppercase tracking-wide text-slate-500">Последняя отправка</dt>
            <dd className="mt-1 text-base font-semibold text-slate-900">
              {formatDate(status?.lastSentAt ?? null)}
            </dd>
          </div>
          <div className="rounded-xl border border-slate-200 bg-white p-3">
            <dt className="text-xs font-medium uppercase tracking-wide text-slate-500">Отправлено / Не отправлено</dt>
            <dd className="mt-1 text-base font-semibold text-slate-900">
              {status?.sentCount7d ?? 0} / {status?.failedCount7d ?? 0}
            </dd>
          </div>
        </dl>
      )}
    </section>
  )
}
