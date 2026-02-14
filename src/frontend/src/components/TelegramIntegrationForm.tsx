import { useState, type SubmitEvent } from 'react'
import type { ConnectTelegramPayload } from '../api/telegram'

type TelegramIntegrationFormProps = {
  form: ConnectTelegramPayload
  isSaving: boolean
  isLoadingStatus: boolean
  onSubmit: (event: SubmitEvent<HTMLFormElement>) => void
  onChange: (patch: Partial<ConnectTelegramPayload>) => void
}

export default function TelegramIntegrationForm({
  form,
  isSaving,
  isLoadingStatus,
  onSubmit,
  onChange,
}: TelegramIntegrationFormProps) {
  const [isChatIdHelpOpen, setIsChatIdHelpOpen] = useState(false)

  return (
    <section className="space-y-4 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm sm:p-5">
      <div className="space-y-1">
        <h2 className="text-base font-semibold text-slate-900">Данные бота</h2>
      </div>

      <form onSubmit={onSubmit} className="space-y-4">
        <div className="space-y-1.5">
          <label htmlFor="botToken" className="block text-sm font-medium text-slate-700">
            Токен бота
          </label>
          <input
            id="botToken"
            type="text"
            autoComplete="off"
            value={form.botToken}
            onChange={(event) => onChange({ botToken: event.target.value })}
            placeholder="123456:ABC-DEF..."
            className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-900 placeholder:text-slate-400 shadow-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
            disabled={isSaving}
          />
        </div>

        <div className="space-y-1.5">
          <div className="flex items-center gap-2">
            <label htmlFor="chatId" className="block text-sm font-medium text-slate-700">
              Chat ID
            </label>
            <div className="relative">
              <button
                type="button"
                onClick={() => setIsChatIdHelpOpen((previous) => !previous)}
                aria-expanded={isChatIdHelpOpen}
                aria-controls="chat-id-help"
                aria-label="Как узнать Chat ID"
                className="inline-flex h-5 w-5 items-center justify-center rounded-full border border-slate-300 bg-white text-xs font-semibold text-slate-600 hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-blue-300"
              >
                ?
              </button>
            </div>
          </div>
          <input
            id="chatId"
            type="text"
            autoComplete="off"
            value={form.chatId}
            onChange={(event) => onChange({ chatId: event.target.value })}
            placeholder="987654321"
            className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2.5 text-sm text-slate-900 placeholder:text-slate-400 shadow-sm outline-none transition focus:border-blue-500 focus:ring-4 focus:ring-blue-100"
            disabled={isSaving}
          />
          {isChatIdHelpOpen ? (
            <div
              id="chat-id-help"
              className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5 text-sm text-slate-700"
            >
              <p>
                Напишите{' '}
                <a
                  href="https://t.me/userinfobot"
                  target="_blank"
                  rel="noreferrer"
                  className="font-medium text-blue-700 hover:text-blue-800 hover:underline"
                >
                  @userinfobot
                </a>{' '}
                или{' '}
                <a
                  href="https://t.me/RawDataBot"
                  target="_blank"
                  rel="noreferrer"
                  className="font-medium text-blue-700 hover:text-blue-800 hover:underline"
                >
                  @RawDataBot
                </a>{' '}
                в Telegram и используйте{' '}
                <code className="rounded bg-slate-200 px-1 py-0.5 text-xs text-slate-800">chat_id</code> из ответа.
              </p>
              <p className="mt-2">
                Либо откройте URL:{' '}
                <code className="rounded bg-slate-200 px-1 py-0.5 text-xs text-slate-800 break-all">
                  https://api.telegram.org/bot&lt;BOT_TOKEN&gt;/getUpdates
                </code>{' '}
                и скопируйте поле{' '}
                <code className="rounded bg-slate-200 px-1 py-0.5 text-xs text-slate-800">chat.id</code>.
              </p>
            </div>
          ) : null}
        </div>

        <label className="flex items-center justify-between rounded-xl border border-slate-200 bg-slate-50 px-3 py-2.5">
          <span className="text-sm font-medium text-slate-700">Включен</span>
          <span className="relative inline-flex items-center">
            <input
              type="checkbox"
              checked={form.enabled}
              onChange={(event) => onChange({ enabled: event.target.checked })}
              disabled={isSaving}
              aria-label="Включить Telegram интеграцию"
              className="peer sr-only"
            />
            <span
              className="h-6 w-11 rounded-full bg-slate-300 transition-colors peer-checked:bg-blue-600 peer-focus-visible:ring-4 peer-focus-visible:ring-blue-200 peer-disabled:opacity-50 after:absolute after:left-0.5 after:top-0.5 after:h-5 after:w-5 after:rounded-full after:bg-white after:shadow-sm after:transition-transform peer-checked:after:translate-x-5"
              aria-hidden="true"
            />
          </span>
        </label>

        <button
          type="submit"
          disabled={isSaving || isLoadingStatus}
          className="inline-flex items-center justify-center rounded-xl bg-blue-600 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition hover:bg-blue-700 active:translate-y-px disabled:cursor-not-allowed disabled:bg-blue-300"
        >
          {isSaving ? 'Сохранение...' : 'Сохранить'}
        </button>
      </form>
    </section>
  )
}
