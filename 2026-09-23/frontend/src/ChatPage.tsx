import { useEffect, useRef, useState } from 'react'
import type { OutboundMessage } from './Client.ts'
import { useChatSession } from './ChatContext.ts'
import { ChatMessage } from './ChatMessage.tsx'
import { ChatInput } from './ChatInput.tsx'
import './ChatPage.css'

const ROOM = 'general'

export function ChatPage() {
  const { client, username } = useChatSession()
  const [messages, setMessages] = useState<OutboundMessage[]>([])
  const [connected, setConnected] = useState(client.connected)
  const bottomRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    // Wire up handlers now that we're on the chat page
    client.options.onMessage = (msg) => setMessages(prev => [...prev, msg])
    client.options.onClose = () => setConnected(false)

    client.joinRoom(ROOM)

    return () => {
      client.leaveRoom(ROOM)
    }
  }, [client])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  function handleSend(content: string) {
    client.sendMessage(ROOM, content)
  }

  return (
    <div className="chat-page">
      <header className="chat-header">
        <h2># {ROOM}</h2>
        <span className="you">{username}</span>
      </header>

      <div className="chat-messages">
        {messages.map((msg, i) => (
          <ChatMessage key={i} message={msg} currentUser={username} />
        ))}
        <div ref={bottomRef} />
      </div>

      <ChatInput onSend={handleSend} disabled={!connected} />
    </div>
  )
}
