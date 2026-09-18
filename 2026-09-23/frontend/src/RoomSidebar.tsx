import { useState } from 'react'
import './RoomSidebar.css'

export interface Room {
  id: string
  name: string
}

interface Props {
  rooms: Room[]
  activeRoomId: string | null
  onSelect: (room: Room) => void
  onCreate: (name: string) => Promise<void>
}

export function RoomSidebar({ rooms, activeRoomId, onSelect, onCreate }: Props) {
  const [newName, setNewName] = useState('')
  const [creating, setCreating] = useState(false)

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    const name = newName.trim()
    if (!name) return
    setCreating(true)
    try {
      await onCreate(name)
      setNewName('')
    } finally {
      setCreating(false)
    }
  }

  return (
    <nav className="room-sidebar">
      <div className="sidebar-heading">Rooms</div>

      <ul className="room-list">
        {rooms.map(room => (
          <li key={room.id}>
            <button
              className={`room-item ${room.id === activeRoomId ? 'active' : ''}`}
              onClick={() => onSelect(room)}
            >
              # {room.name}
            </button>
          </li>
        ))}
      </ul>

      <form className="new-room-form" onSubmit={handleCreate}>
        <input
          type="text"
          placeholder="New room…"
          value={newName}
          onChange={e => setNewName(e.target.value)}
          disabled={creating}
        />
        <button type="submit" disabled={creating || !newName.trim()}>+</button>
      </form>
    </nav>
  )
}
