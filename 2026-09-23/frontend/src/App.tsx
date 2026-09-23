import { useState } from 'react'
import { ChatContext, type ChatSession } from './ChatContext.ts'
import { LoginPage } from './LoginPage.tsx'
import { ChatPage } from './ChatPage.tsx'
import { useHashRoute } from './useHashRoute.ts'

export default function App() {
  const { navigate } = useHashRoute()
  const [session, setSession] = useState<ChatSession | null>(null)

  function handleLogin(newSession: ChatSession) {
    setSession(newSession)
    navigate('/chat')
  }

  const route = session ? window.location.hash.slice(1) || '/' : '/'

  if (route === '/chat' && session) {
    return (
      <ChatContext.Provider value={session}>
        <ChatPage />
      </ChatContext.Provider>
    )
  }

  return <LoginPage onLogin={handleLogin} />
}
