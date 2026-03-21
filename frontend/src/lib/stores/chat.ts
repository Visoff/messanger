import { writable } from "svelte/store";

export const selectedChatId = writable<string | undefined>(undefined);
export const selectedTopicId = writable<string | undefined>(undefined);
export const selectedChatWithTopcis = writable<boolean>(false);
