// place files you want to import through the `$lib` alias in this folder.

export function extractFromSearchParams(key: string): string | undefined {
    if (!globalThis.location) return undefined;
    const searchParams = new URLSearchParams(globalThis.location.search);
    return searchParams.get(key) || undefined;
}
