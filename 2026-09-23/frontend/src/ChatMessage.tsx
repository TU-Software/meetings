import type { ChatMessage as ChatMessageType } from './Client.ts'
import './ChatMessage.css'

interface Props {
  message: ChatMessageType
  currentUser: string
}

export function ChatMessage({ message, currentUser }: Props) {
  const isOwn = message.sender === currentUser

  return (
    <div className={`chat-message ${isOwn ? 'own' : 'other'}`}>
      {!isOwn && <span className="sender">{message.sender}</span>}
      <span className="bubble">{message.content}</span>
    </div>
  )
}
