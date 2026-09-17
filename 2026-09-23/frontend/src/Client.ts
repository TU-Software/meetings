// Actions sent to the server
type Action = "join_room" | "leave_room" | "send_message";

interface InboundMessage {
  action: Action;
  room?: string;
  content?: string;
}

// Messages received from the server
export interface OutboundMessage {
  sender: string;
  room: string;
  content: string;
  system?: boolean;
}

export type MessageHandler = (msg: OutboundMessage) => void;

export interface ChatClientOptions {
  /** Base URL of the backend, e.g. "http://localhost:8080". Defaults to the current origin. */
  baseUrl?: string;
  /** Called whenever a message arrives from the server. */
  onMessage: MessageHandler;
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
      .replace(/^http/, "ws") // http -> ws, https -> wss
      .replace(/\/$/, "");
  }

  /**
   * Open a WebSocket connection to the server.
   * Resolves once the connection is open, rejects on error before open.
   */
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
        // The server may batch multiple newline-separated JSON messages
        const lines = (event.data as string).split("\n").filter(Boolean);
        for (const line of lines) {
          try {
            const msg = JSON.parse(line) as OutboundMessage;
            this.options.onMessage(msg);
          } catch {
            console.error("Failed to parse message:", line);
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

  /** Join a room. */
  joinRoom(room: string): void {
    this.send({ action: "join_room", room });
  }

  /** Leave a room. */
  leaveRoom(room: string): void {
    this.send({ action: "leave_room", room });
  }

  /**
   * Send a message to a room.
   * The server ignores the message if the client hasn't joined the room.
   */
  sendMessage(room: string, content: string): void {
    this.send({ action: "send_message", room, content });
  }

  /** Close the WebSocket connection. */
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
