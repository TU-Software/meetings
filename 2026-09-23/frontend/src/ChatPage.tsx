import { useEffect, useRef, useState } from 'react'
import type { ChatMessage as ChatMessageType } from './Client.ts'
import { useChatSession } from './ChatContext.ts'
import { RoomSidebar, type Room } from './RoomSidebar.tsx'
import { ChatMessage } from './ChatMessage.tsx'
import { ChatInput } from './ChatInput.tsx'
import { useRoomMessages } from './useRoomMessages.ts'
import './ChatPage.css'

export function ChatPage() {
  const { client, username } = useChatSession()
  const [connected, setConnected] = useState(client.connected)
  const [rooms, setRooms] = useState<Room[]>([])
  const [activeRoom, setActiveRoom] = useState<Room | null>(null)
  // Track which rooms have had their history fetched this session.
  const fetchedRooms = useRef<Set<string>>(new Set())
  const activeRoomRef = useRef<Room | null>(null)

  const { messages, addMessage, seedMessages } = useRoomMessages(activeRoom?.id ?? null)
  const bottomRef = useRef<HTMLDivElement | null>(null)

  // Fetch the full room list once on mount and wire up client handlers.
  useEffect(() => {
    fetch('/api/rooms')
      .then(r => r.json())
      .then((data: Room[]) => setRooms(data))
      .catch(err => console.error('Failed to fetch rooms:', err))

    client.options.onMessage = (msg) => {
      if (activeRoomRef.current?.id === msg.room_id) {
        addMessage(msg)
      }
    }
    client.options.onRoomCreated = (event) => {
      setRooms(prev => {
        if (prev.some(r => r.id === event.room.id)) return prev
        return [...prev, event.room].sort((a, b) => a.name.localeCompare(b.name))
      })
    }
    client.options.onClose = () => setConnected(false)
  }, [client, addMessage])

  // Join/leave as the active room changes, and fetch history on first visit.
  useEffect(() => {
    const prev = activeRoomRef.current
    activeRoomRef.current = activeRoom

    if (prev) client.leaveRoom(prev.id)
    if (!activeRoom) return

    client.joinRoom(activeRoom.id)

    // Fetch history from the server the first time this room is selected.
    if (!fetchedRooms.current.has(activeRoom.id)) {
      fetchedRooms.current.add(activeRoom.id)
      fetch(`/api/rooms/${activeRoom.id}`)
        .then(r => r.json())
        .then((history: ChatMessageType[]) => seedMessages(history))
        .catch(err => console.error('Failed to fetch room history:', err))
    }

    return () => {
      client.leaveRoom(activeRoom.id)
    }
  }, [client, activeRoom, seedMessages])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function handleCreateRoom(name: string) {
    const res = await fetch('/api/rooms', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    })
    if (!res.ok) {
      console.error('Failed to create room', await res.text())
      return
    }
    const room: Room = await res.json()
    // The room_created WebSocket event will update the sidebar for all other
    // clients; add it locally immediately for the creator.
    setRooms(prev => {
      if (prev.some(r => r.id === room.id)) return prev
      return [...prev, room].sort((a, b) => a.name.localeCompare(b.name))
    })
    setActiveRoom(room)
  }

  function handleSend(content: string) {
    if (activeRoom) client.sendMessage(activeRoom.id, content)
  }

  return (
    <div className="chat-page">
      <RoomSidebar
        rooms={rooms}
        activeRoomId={activeRoom?.id ?? null}
        onSelect={setActiveRoom}
        onCreate={handleCreateRoom}
      />

      <div className="chat-main">
        <header className="chat-header">
          {activeRoom
            ? <h2># {activeRoom.name}</h2>
            : <h2 className="muted">Select a room</h2>}
          <span className="you">{username}</span>
        </header>

        <div className="chat-messages">
          {activeRoom
            ? messages.map(msg => (
                <ChatMessage key={msg.id} message={msg} currentUser={username} />
              ))
            : <p className="chat-empty">Choose a room from the sidebar to start chatting.</p>}
          <div ref={bottomRef} />
        </div>

        <ChatInput
          onSend={handleSend}
          roomName={activeRoom?.name}
          disabled={!connected || !activeRoom}
        />
      </div>
    </div>
  )
}
