import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { Message } from "../types";

export const newMessageEvent: Writable<Message | null> = writable(null);
