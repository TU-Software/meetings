import { useCallback, useEffect, useState } from 'react'
import type { ChatMessage } from './Client.ts'

function storageKey(roomId: string) {
  return `chat:messages:${roomId}`
}

function loadMessages(roomId: string): ChatMessage[] {
  try {
    const raw = localStorage.getItem(storageKey(roomId))
    return raw ? (JSON.parse(raw) as ChatMessage[]) : []
  } catch {
    return []
  }
}

function saveMessages(roomId: string, messages: ChatMessage[]) {
  try {
    // Keep at most 200 messages per room to avoid unbounded storage growth.
    localStorage.setItem(storageKey(roomId), JSON.stringify(messages.slice(-200)))
  } catch {
    // Quota exceeded or private browsing — silently ignore.
  }
}

/**
 * Manages messages for a single room, backed by localStorage.
 * Each room gets its own isolated key so messages never bleed across rooms.
 */
export function useRoomMessages(roomId: string | null) {
  const [messages, setMessages] = useState<ChatMessage[]>(() =>
    roomId ? loadMessages(roomId) : []
  )

  // When the active room changes, load from storage immediately.
  useEffect(() => {
    setMessages(roomId ? loadMessages(roomId) : [])
  }, [roomId])

  // Seed from the API: merge fetched history with any messages already in
  // storage, deduplicate by ID, then sort by ID (UUIDv7 is time-ordered).
  const seedMessages = useCallback((fetched: ChatMessage[]) => {
    if (!roomId) return
    setMessages(prev => {
      const byId = new Map(prev.map(m => [m.id, m]))
      for (const m of fetched) byId.set(m.id, m)
      const merged = [...byId.values()].sort((a, b) => a.id < b.id ? -1 : a.id > b.id ? 1 : 0)
      saveMessages(roomId, merged)
      return merged
    })
  }, [roomId])

  const addMessage = useCallback((msg: ChatMessage) => {
    setMessages(prev => {
      // Deduplicate in case a live message arrived while we were fetching history.
      if (prev.some(m => m.id === msg.id)) return prev
      const next = [...prev, msg]
      saveMessages(msg.room_id, next)
      return next
    })
  }, [])

  return { messages, addMessage, seedMessages }
}
