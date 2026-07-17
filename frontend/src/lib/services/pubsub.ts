import { API_URL } from "$lib/api/env";

type EventHandler = (data: unknown) => void;

class PubSubService {
    private eventSource: EventSource | null = null;
    private handlers: Map<string, Set<EventHandler>> = new Map();
    private _onDestroy: (() => void) | null = null;

    private constructor() {}

    static create(token: string): PubSubService {
        const service = new PubSubService();
        const es = new EventSource(`${API_URL}/pubsub/sse?token=${token}`);

        es.addEventListener("message", (e) => {
            try {
                const data = JSON.parse(e.data);
                const type = (data as Record<string, unknown>).type as string | undefined;
                if (type && service.handlers.has(type)) {
                    service.handlers.get(type)!.forEach((h) => h(data));
                }
                if (service.handlers.has("*")) {
                    service.handlers.get("*")!.forEach((h) => h(data));
                }
            } catch {
                // skip malformed events
            }
        });

        service.eventSource = es;
        return service;
    }

    on(eventType: string, handler: EventHandler): void {
        if (!this.handlers.has(eventType)) {
            this.handlers.set(eventType, new Set());
        }
        this.handlers.get(eventType)!.add(handler);
    }

    off(eventType: string, handler: EventHandler): void {
        this.handlers.get(eventType)?.delete(handler);
    }

    destroy(): void {
        this.eventSource?.close();
        this.eventSource = null;
        this.handlers.clear();
    }
}

let sharedInstance: PubSubService | null = null;

export function getPubSub(token: string): PubSubService {
    if (!sharedInstance) {
        sharedInstance = PubSubService.create(token);
    }
    return sharedInstance;
}

export function destroyPubSub(): void {
    sharedInstance?.destroy();
    sharedInstance = null;
}
