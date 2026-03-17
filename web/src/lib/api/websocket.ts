// WebSocket connection manager for real-time messaging.
// Provides auto-reconnect with exponential backoff, message queueing,
// and reactive stores for connection status and typing indicators.

import { getAccessToken } from './client';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type WSMessageType =
	| 'new_message'
	| 'message_read'
	| 'typing_indicator'
	| 'user_online'
	| 'user_offline'
	| 'ping'
	| 'pong';

export interface WSMessage {
	type: WSMessageType;
	payload: unknown;
}

export interface WSNewMessagePayload {
	id: string;
	conversation_id: string;
	sender_id: string;
	content: string;
	message_type: string;
	created_at: string;
}

export interface WSMessageReadPayload {
	conversation_id: string;
	reader_id: string;
	read_at: string;
}

export interface WSTypingPayload {
	conversation_id: string;
	user_id: string;
	is_typing: boolean;
}

export interface WSUserStatusPayload {
	user_id: string;
}

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'reconnecting';

// ---------------------------------------------------------------------------
// Subscriber types
// ---------------------------------------------------------------------------

type StatusSubscriber = (status: ConnectionStatus) => void;
type MessageHandler<T = unknown> = (payload: T) => void;
type TypingSubscriber = (typingUsers: Map<string, Set<string>>) => void;
type OnlineSubscriber = (onlineUsers: Set<string>) => void;

// ---------------------------------------------------------------------------
// WebSocket Manager
// ---------------------------------------------------------------------------

class WebSocketManager {
	private ws: WebSocket | null = null;
	private url: string = '';
	private reconnectAttempt = 0;
	private maxReconnectDelay = 30000; // 30s
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
	private messageQueue: string[] = [];
	private intentionalClose = false;

	// State
	private _status: ConnectionStatus = 'disconnected';
	private _typingUsers: Map<string, Set<string>> = new Map(); // conversationId -> set of userIds
	private _onlineUsers: Set<string> = new Set();

	// Subscribers
	private statusSubscribers: Set<StatusSubscriber> = new Set();
	private messageHandlers: Map<WSMessageType, Set<MessageHandler>> = new Map();
	private typingSubscribers: Set<TypingSubscriber> = new Set();
	private onlineSubscribers: Set<OnlineSubscriber> = new Set();

	// Typing indicator debounce
	private typingTimers: Map<string, ReturnType<typeof setTimeout>> = new Map();

	/**
	 * Connect to the WebSocket server.
	 * @param baseUrl The API base URL (e.g. http://localhost:3000)
	 */
	connect(baseUrl?: string): void {
		if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
			return;
		}

		const token = getAccessToken();
		if (!token) {
			console.warn('[WS] No auth token available, skipping connection');
			return;
		}

		// Build WebSocket URL
		const base = baseUrl || this.getDefaultBaseUrl();
		const wsProtocol = base.startsWith('https') ? 'wss' : 'ws';
		const host = base.replace(/^https?:\/\//, '');
		this.url = `${wsProtocol}://${host}/ws?token=${encodeURIComponent(token)}`;

		this.setStatus('connecting');
		this.intentionalClose = false;

		try {
			this.ws = new WebSocket(this.url);
			this.ws.onopen = this.onOpen.bind(this);
			this.ws.onmessage = this.onMessage.bind(this);
			this.ws.onclose = this.onClose.bind(this);
			this.ws.onerror = this.onError.bind(this);
		} catch (err) {
			console.error('[WS] Failed to create WebSocket:', err);
			this.scheduleReconnect();
		}
	}

	/**
	 * Disconnect from the WebSocket server.
	 */
	disconnect(): void {
		this.intentionalClose = true;
		this.clearTimers();
		if (this.ws) {
			this.ws.close(1000, 'client disconnect');
			this.ws = null;
		}
		this.setStatus('disconnected');
	}

	/**
	 * Send a message through the WebSocket.
	 */
	send(msg: WSMessage): void {
		const data = JSON.stringify(msg);

		if (this.ws?.readyState === WebSocket.OPEN) {
			this.ws.send(data);
		} else {
			// Queue for when connection is restored
			this.messageQueue.push(data);
		}
	}

	/**
	 * Send a typing indicator for a conversation.
	 */
	sendTyping(conversationId: string, isTyping: boolean): void {
		this.send({
			type: 'typing_indicator',
			payload: {
				conversation_id: conversationId,
				is_typing: isTyping
			}
		});
	}

	/**
	 * Get current connection status.
	 */
	getStatus(): ConnectionStatus {
		return this._status;
	}

	/**
	 * Check if a user is currently online.
	 */
	isUserOnline(userId: string): boolean {
		return this._onlineUsers.has(userId);
	}

	/**
	 * Get typing users for a conversation.
	 */
	getTypingUsers(conversationId: string): Set<string> {
		return this._typingUsers.get(conversationId) || new Set();
	}

	// ---------------------------------------------------------------------------
	// Subscriptions
	// ---------------------------------------------------------------------------

	/**
	 * Subscribe to connection status changes.
	 */
	subscribeStatus(fn: StatusSubscriber): () => void {
		this.statusSubscribers.add(fn);
		fn(this._status);
		return () => this.statusSubscribers.delete(fn);
	}

	/**
	 * Subscribe to a specific message type.
	 */
	on<T = unknown>(type: WSMessageType, handler: MessageHandler<T>): () => void {
		if (!this.messageHandlers.has(type)) {
			this.messageHandlers.set(type, new Set());
		}
		const handlers = this.messageHandlers.get(type)!;
		handlers.add(handler as MessageHandler);
		return () => handlers.delete(handler as MessageHandler);
	}

	/**
	 * Subscribe to typing indicator changes.
	 */
	subscribeTyping(fn: TypingSubscriber): () => void {
		this.typingSubscribers.add(fn);
		fn(this._typingUsers);
		return () => this.typingSubscribers.delete(fn);
	}

	/**
	 * Subscribe to online user changes.
	 */
	subscribeOnline(fn: OnlineSubscriber): () => void {
		this.onlineSubscribers.add(fn);
		fn(this._onlineUsers);
		return () => this.onlineSubscribers.delete(fn);
	}

	// ---------------------------------------------------------------------------
	// Internal handlers
	// ---------------------------------------------------------------------------

	private onOpen(): void {
		console.log('[WS] Connected');
		this.reconnectAttempt = 0;
		this.setStatus('connected');
		this.startHeartbeat();
		this.flushQueue();
	}

	private onMessage(event: MessageEvent): void {
		try {
			const msg: WSMessage = JSON.parse(event.data);
			this.handleMessage(msg);
		} catch (err) {
			console.warn('[WS] Invalid message received:', err);
		}
	}

	private onClose(event: CloseEvent): void {
		console.log('[WS] Closed:', event.code, event.reason);
		this.clearTimers();
		this.ws = null;

		if (!this.intentionalClose) {
			this.setStatus('reconnecting');
			this.scheduleReconnect();
		} else {
			this.setStatus('disconnected');
		}
	}

	private onError(event: Event): void {
		console.error('[WS] Error:', event);
	}

	private handleMessage(msg: WSMessage): void {
		switch (msg.type) {
			case 'new_message':
			case 'message_read':
				// Forward to subscribers
				break;

			case 'typing_indicator': {
				const payload = msg.payload as WSTypingPayload;
				this.handleTypingIndicator(payload);
				break;
			}

			case 'user_online': {
				const payload = msg.payload as WSUserStatusPayload;
				this._onlineUsers.add(payload.user_id);
				this.notifyOnlineSubscribers();
				break;
			}

			case 'user_offline': {
				const payload = msg.payload as WSUserStatusPayload;
				this._onlineUsers.delete(payload.user_id);
				this.notifyOnlineSubscribers();
				break;
			}

			case 'pong':
				// Heartbeat response received
				break;

			default:
				console.debug('[WS] Unhandled message type:', msg.type);
		}

		// Notify type-specific handlers
		const handlers = this.messageHandlers.get(msg.type);
		if (handlers) {
			for (const handler of handlers) {
				try {
					handler(msg.payload);
				} catch (err) {
					console.error('[WS] Handler error:', err);
				}
			}
		}
	}

	private handleTypingIndicator(payload: WSTypingPayload): void {
		const convId = payload.conversation_id;

		if (payload.is_typing) {
			if (!this._typingUsers.has(convId)) {
				this._typingUsers.set(convId, new Set());
			}
			this._typingUsers.get(convId)!.add(payload.user_id);

			// Auto-clear typing after 5 seconds
			const key = `${convId}:${payload.user_id}`;
			if (this.typingTimers.has(key)) {
				clearTimeout(this.typingTimers.get(key)!);
			}
			this.typingTimers.set(
				key,
				setTimeout(() => {
					this._typingUsers.get(convId)?.delete(payload.user_id);
					if (this._typingUsers.get(convId)?.size === 0) {
						this._typingUsers.delete(convId);
					}
					this.notifyTypingSubscribers();
				}, 5000)
			);
		} else {
			this._typingUsers.get(convId)?.delete(payload.user_id);
			if (this._typingUsers.get(convId)?.size === 0) {
				this._typingUsers.delete(convId);
			}
		}

		this.notifyTypingSubscribers();
	}

	// ---------------------------------------------------------------------------
	// Reconnect with exponential backoff
	// ---------------------------------------------------------------------------

	private scheduleReconnect(): void {
		if (this.intentionalClose) return;

		const baseDelay = 1000; // 1s
		const delay = Math.min(baseDelay * Math.pow(2, this.reconnectAttempt), this.maxReconnectDelay);
		this.reconnectAttempt++;

		console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempt})`);

		this.reconnectTimer = setTimeout(() => {
			this.connect();
		}, delay);
	}

	// ---------------------------------------------------------------------------
	// Heartbeat
	// ---------------------------------------------------------------------------

	private startHeartbeat(): void {
		this.heartbeatTimer = setInterval(() => {
			if (this.ws?.readyState === WebSocket.OPEN) {
				this.send({ type: 'ping', payload: {} });
			}
		}, 30000);
	}

	// ---------------------------------------------------------------------------
	// Helpers
	// ---------------------------------------------------------------------------

	private setStatus(status: ConnectionStatus): void {
		this._status = status;
		for (const fn of this.statusSubscribers) {
			fn(status);
		}
	}

	private notifyTypingSubscribers(): void {
		for (const fn of this.typingSubscribers) {
			fn(this._typingUsers);
		}
	}

	private notifyOnlineSubscribers(): void {
		for (const fn of this.onlineSubscribers) {
			fn(this._onlineUsers);
		}
	}

	private flushQueue(): void {
		while (this.messageQueue.length > 0 && this.ws?.readyState === WebSocket.OPEN) {
			const msg = this.messageQueue.shift()!;
			this.ws.send(msg);
		}
	}

	private clearTimers(): void {
		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
		if (this.heartbeatTimer) {
			clearInterval(this.heartbeatTimer);
			this.heartbeatTimer = null;
		}
	}

	private getDefaultBaseUrl(): string {
		if (typeof window !== 'undefined') {
			// In development, the API is typically on a different port
			const hostname = window.location.hostname;
			return `http://${hostname}:3000`;
		}
		return 'http://localhost:3000';
	}
}

// ---------------------------------------------------------------------------
// Singleton instance
// ---------------------------------------------------------------------------

export const wsManager = new WebSocketManager();
