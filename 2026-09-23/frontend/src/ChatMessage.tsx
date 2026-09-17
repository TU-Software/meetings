import type { OutboundMessage } from './Client.ts'
import './ChatMessage.css'

interface Props {
  message: OutboundMessage
  /** The local user's username, used to style own messages differently. */
  currentUser: string
}

export function ChatMessage({ message, currentUser }: Props) {
  if (message.system) {
    return <div className="chat-message system">{message.content}</div>
  }

  const isOwn = message.sender === currentUser

  return (
    <div className={`chat-message ${isOwn ? 'own' : 'other'}`}>
      {!isOwn && <span className="sender">{message.sender}</span>}
      <span className="bubble">{message.content}</span>
    </div>
  )
}
