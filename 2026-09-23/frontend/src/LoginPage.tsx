import { useState } from 'react'
import { ChatClient } from './Client.ts'
import type { ChatSession } from './ChatContext.ts'
import './LoginPage.css'

interface Props {
  onLogin: (session: ChatSession) => void
}

type Status = 'idle' | 'connecting' | 'error'

export function LoginPage({ onLogin }: Props) {
  const [username, setUsername] = useState('')
  const [status, setStatus] = useState<Status>('idle')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const name = username.trim()
    if (!name) return

    setStatus('connecting')

    const client = new ChatClient({
      onMessage() {},  // placeholder — ChatPage installs its own handler via context
      onOpen() {},
      onClose() {},
    })

    try {
      await client.connect(name)
      onLogin({ client, username: name })
    } catch {
      setStatus('error')
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <h1>Chat</h1>
        <p>Enter a username to join <strong>#general</strong></p>
        <form onSubmit={handleSubmit}>
          <input
            type="text"
            placeholder="Username"
            value={username}
            onChange={e => setUsername(e.target.value)}
            disabled={status === 'connecting'}
            autoFocus
            autoComplete="off"
          />
          <button type="submit" disabled={!username.trim() || status === 'connecting'}>
            {status === 'connecting' ? 'Connecting…' : 'Join'}
          </button>
        </form>
        {status === 'error' && (
          <p className="login-error">Could not connect to server. Is it running?</p>
        )}
      </div>
    </div>
  )
}
