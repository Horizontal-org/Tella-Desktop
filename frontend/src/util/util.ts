import DOMPurify from 'dompurify';

export function sanitizeUGC (userInput: string): string {
    // sanitize to only allow plain tagless strings
    return DOMPurify.sanitize(userInput, { ALLOWED_TAGS: ["#text"] })
}
