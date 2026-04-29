import DOMPurify from 'dompurify';
import { IsDevelopment } from '../../wailsjs/go/app/App';

export function sanitizeUGC (userInput: string): string {
    // sanitize to only allow plain tagless strings
    return DOMPurify.sanitize(userInput, { ALLOWED_TAGS: ["#text"] })
}

export async function log(msg: string, ...rest: any[]): Promise<void> {
    const dev = await IsDevelopment()
    if (dev) {
        console.log(msg, ...rest)
    }
}
