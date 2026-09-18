// Actions sent to the server
type Action = "join_room" | "leave_room" | "send_message";

interface InboundMessage {
  action: Action;
  room_id: string;
  content?: string;
}

// Payloads received from the server — discriminated by `type`.

export interface ChatMessage {
  type: "message";
  id: string;
  room_id: string;
  room_name: string;
  sender: string;
  content: string;
}

export interface RoomCreatedEvent {
  type: "room_created";
  room: { id: string; name: string };
}

export type ServerMessage = ChatMessage | RoomCreatedEvent;

export type ChatMessageHandler = (msg: ChatMessage) => void;
export type RoomCreatedHandler = (event: RoomCreatedEvent) => void;

export interface ChatClientOptions {
  /** Base URL of the backend, e.g. "http://localhost:8080". Defaults to the current origin. */
  baseUrl?: string;
  /** Called for each chat message received. */
  onMessage: ChatMessageHandler;
  /** Called when a room_created event is received. */
  onRoomCreated?: RoomCreatedHandler;
  /** Called when the connection is established. */
  onOpen?: () => void;
  /** Called when the connection closes. */
  onClose?: () => void;
  /** Called on a WebSocket error. */
  onError?: (event: Event) => void;
}

export class ChatClient {
  private ws: WebSocket | null = null;
  private readonly baseUrl: string;
  readonly options: ChatClientOptions;

  constructor(options: ChatClientOptions) {
    this.options = options;
    this.baseUrl = (options.baseUrl ?? window.location.origin)
      .replace(/^http/, "ws")
      .replace(/\/$/, "");
  }

  connect(username: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const url = `${this.baseUrl}/ws?username=${encodeURIComponent(username)}`;
      const ws = new WebSocket(url);
      this.ws = ws;

      ws.addEventListener("open", () => {
        this.options.onOpen?.();
        resolve();
      });

      ws.addEventListener("message", (event: MessageEvent<string>) => {
        const lines = (event.data as string).split("\n").filter(Boolean);
        for (const line of lines) {
          try {
            const msg = JSON.parse(line) as ServerMessage;
            if (msg.type === "room_created") {
              this.options.onRoomCreated?.(msg);
            } else {
              this.options.onMessage(msg);
            }
          } catch {
            console.error("Failed to parse server message:", line);
          }
        }
      });

      ws.addEventListener("close", () => {
        this.options.onClose?.();
      });

      ws.addEventListener("error", (event) => {
        this.options.onError?.(event);
        reject(event);
      });
    });
  }

  /** Join a room by UUID. */
  joinRoom(roomId: string): void {
    this.send({ action: "join_room", room_id: roomId });
  }

  /** Leave a room by UUID. */
  leaveRoom(roomId: string): void {
    this.send({ action: "leave_room", room_id: roomId });
  }

  /** Send a message to a room by UUID. */
  sendMessage(roomId: string, content: string): void {
    this.send({ action: "send_message", room_id: roomId, content });
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }

  get connected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  private send(msg: InboundMessage): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn("ChatClient: attempted to send while not connected", msg);
      return;
    }
    this.ws.send(JSON.stringify(msg));
  }
}
