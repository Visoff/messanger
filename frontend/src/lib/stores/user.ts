import { writable } from "svelte/store";
import type { Writable } from "svelte/store";
import type { User } from "../types";

export const user: Writable<User | undefined> = writable(undefined);
