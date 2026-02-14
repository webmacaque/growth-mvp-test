const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api'

export type ConnectTelegramPayload = {
  botToken: string
  chatId: string
  enabled: boolean
}

export type TelegramStatus = {
  enabled: boolean
  chatId: string
  lastSentAt: string | null
  sentCount7d: number
  failedCount7d: number
}

async function parseErrorMessage(response: Response): Promise<string> {
  try {
    const data = (await response.json()) as { error?: string }
    if (data.error) {
      return data.error
    }
  } catch {
    console.error('Failed to parse error message', response)
  }

  return response.statusText || 'Request failed'
}

export async function getTelegramStatus(shopId: string): Promise<TelegramStatus> {
  const response = await fetch(`${API_BASE_URL}/shops/${shopId}/telegram/status`)

  if (!response.ok) {
    const message = await parseErrorMessage(response)
    throw new Error(message)
  }

  return (await response.json()) as TelegramStatus
}

export async function connectTelegram(
  shopId: string,
  payload: ConnectTelegramPayload,
): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/shops/${shopId}/telegram/connect`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    const message = await parseErrorMessage(response)
    throw new Error(message)
  }
}
