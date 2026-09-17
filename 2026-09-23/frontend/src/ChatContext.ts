import { createContext, useContext } from 'react'
import type { ChatClient } from './Client.ts'

export interface ChatSession {
  client: ChatClient
  username: string
}

export const ChatContext = createContext<ChatSession | null>(null)

export function useChatSession(): ChatSession {
  const ctx = useContext(ChatContext)
  if (!ctx) throw new Error('useChatSession must be used inside ChatContext.Provider')
  return ctx
}
