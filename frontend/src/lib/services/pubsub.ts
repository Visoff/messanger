import { API_URL } from "$lib/api/env";

type EventHandler = (data: unknown) => void;

class PubSubService {
    private eventSource: EventSource | null = null;
    private handlers: Map<string, Set<EventHandler>> = new Map();
    private polling = false;
    private pollingAbort: AbortController | null = null;
    private useLongPolling = false;

    private constructor() {}

    static create(token: string): PubSubService {
        const service = new PubSubService();
        service.useLongPolling = import.meta.env.VITE_USE_LONG_POLLING === "true";

        if (service.useLongPolling) {
            service.startPolling(token);
        } else {
            service.startSSE(token);
        }

        return service;
    }

    private startSSE(token: string): void {
        const es = new EventSource(`${API_URL}/pubsub/sse?token=${token}`);

        es.addEventListener("message", (e) => {
            try {
                const data = JSON.parse(e.data);
                const type = (data as Record<string, unknown>).type as string | undefined;
                if (type && this.handlers.has(type)) {
                    this.handlers.get(type)!.forEach((h) => h(data));
                }
                if (this.handlers.has("*")) {
                    this.handlers.get("*")!.forEach((h) => h(data));
                }
            } catch {
                // skip malformed events
            }
        });

        this.eventSource = es;
    }

    private startPolling(token: string): void {
        this.polling = true;
        this.pollLoop(token);
    }

    private async pollLoop(token: string): Promise<void> {
        while (this.polling) {
            try {
                this.pollingAbort = new AbortController();
                const res = await fetch(
                    `${API_URL}/pubsub/poll?token=${token}&timeout=30`,
                    { signal: this.pollingAbort.signal },
                );
                if (!this.polling) break;
                if (!res.ok) {
                    await new Promise((r) => setTimeout(r, 5000));
                    continue;
                }
                const data: unknown[] = await res.json();
                for (const item of data) {
                    const record = item as Record<string, unknown>;
                    const type = record.type as string | undefined;
                    if (type && this.handlers.has(type)) {
                        this.handlers.get(type)!.forEach((h) => h(item));
                    }
                    if (this.handlers.has("*")) {
                        this.handlers.get("*")!.forEach((h) => h(item));
                    }
                }
            } catch {
                if (!this.polling) break;
                await new Promise((r) => setTimeout(r, 5000));
            }
        }
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
        this.polling = false;
        this.pollingAbort?.abort();
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
