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

export function formatHash(hashString: string): string {
    const input = [];
    for (let i = 0; i <= hashString.length; i += 4) {
        // grab groups of 4 characters each
        let entry = hashString.slice(i, i+4) + " ";
        // output a newline after four groups of 4
        if (((i+4) % 16) === 0) {
            entry += "\n";
        }
        input.push(entry);
    }
    // trim the last newline and return as a string
    return input.join("").trim();
}
