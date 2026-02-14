import { useEffect, useMemo, useState } from 'react'
import type { SubmitEvent } from 'react'
import { useParams } from 'react-router-dom'
import {
  connectTelegram,
  getTelegramStatus,
  type ConnectTelegramPayload,
  type TelegramStatus,
} from '../api/telegram'
import TelegramIntegrationForm from '../components/TelegramIntegrationForm'
import TelegramIntegrationStatus from '../components/TelegramIntegrationStatus'

type FormState = ConnectTelegramPayload

const INITIAL_FORM: FormState = {
  botToken: '',
  chatId: '',
  enabled: true,
}

export default function TelegramIntegrationPage() {
  const { shopId } = useParams<{ shopId: string }>()

  const [form, setForm] = useState<FormState>(INITIAL_FORM)
  const [status, setStatus] = useState<TelegramStatus | null>(null)
  const [isLoadingStatus, setIsLoadingStatus] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)

  const currentShopId = useMemo(() => shopId?.trim() ?? '', [shopId])

  async function loadStatus() {
    setError(null)
    setIsLoadingStatus(true)

    try {
      const data = await getTelegramStatus(currentShopId)
      setStatus(data)
      setForm((previous) => ({
        ...previous,
        chatId: data.chatId || previous.chatId,
        enabled: data.enabled,
      }))
    } catch (loadError) {
      const message = loadError instanceof Error ? loadError.message : 'Ошибка при загрузке статуса.'
      setError(message)
    } finally {
      setIsLoadingStatus(false)
    }
  }

  useEffect(() => {
    void loadStatus()
  }, [currentShopId])

  async function onSubmit(event: SubmitEvent<HTMLFormElement>) {
    event.preventDefault()
    setSuccess(null)

    if (!form.botToken.trim() || !form.chatId.trim()) {
      setError('Токен и Chat ID являются обязательными.')
      return
    }

    setError(null)
    setIsSaving(true)

    try {
      await connectTelegram(currentShopId, {
        botToken: form.botToken.trim(),
        chatId: form.chatId.trim(),
        enabled: form.enabled,
      })
      setSuccess('Настройки сохранены.')
      setForm((previous) => ({ ...previous, botToken: '' }))
      await loadStatus()
    } catch (saveError) {
      const message = saveError instanceof Error ? saveError.message : 'Ошибка при сохранении настроек.'
      setError(message)
    } finally {
      setIsSaving(false)
    }
  }

  function handleFormChange(patch: Partial<FormState>) {
    setForm((previous) => ({ ...previous, ...patch }))
  }

  return (
    <div className="space-y-6 sm:space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-slate-900 sm:text-3xl">Growth MVP</h1>
      </header>

      <TelegramIntegrationForm
        form={form}
        isSaving={isSaving}
        isLoadingStatus={isLoadingStatus}
        onSubmit={onSubmit}
        onChange={handleFormChange}
      />

      {error ? (
        <p className="rounded-xl border border-red-200 border-l-4 border-l-red-400 bg-red-50 px-3 py-2.5 text-sm text-red-700">
          {error}
        </p>
      ) : null}
      {success ? (
        <p className="rounded-xl border border-emerald-200 border-l-4 border-l-emerald-400 bg-emerald-50 px-3 py-2.5 text-sm text-emerald-700">
          {success}
        </p>
      ) : null}

      <TelegramIntegrationStatus status={status} isLoadingStatus={isLoadingStatus} />
    </div>
  )
}
