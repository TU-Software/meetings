import { useState } from 'react'
import './ChatInput.css'

interface Props {
  onSend: (content: string) => void
  roomName?: string
  disabled?: boolean
}

export function ChatInput({ onSend, roomName, disabled = false }: Props) {
  const [value, setValue] = useState('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = value.trim()
    if (!trimmed) return
    onSend(trimmed)
    setValue('')
  }

  return (
    <form className="chat-input" onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder={roomName ? `Message #${roomName}` : 'Select a room…'}
        value={value}
        onChange={e => setValue(e.target.value)}
        disabled={disabled}
        autoComplete="off"
        autoFocus
      />
      <button type="submit" disabled={disabled || !value.trim()}>
        Send
      </button>
    </form>
  )
}
