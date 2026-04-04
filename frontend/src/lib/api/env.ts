export const API_URL = import.meta.env.VITE_API_URL
// protocol = BASE_URL = API_URL
const IS_SECURE = API_URL.startsWith('https://');
const BASE_URL = API_URL.substring(7 + IS_SECURE);
export function API_URL_WITH_PROTOCOL(protocol: string, secure_protocol?: string) {
    if (IS_SECURE) {
        return (secure_protocol || protocol) + BASE_URL;
    }
    return protocol + BASE_URL;
}
