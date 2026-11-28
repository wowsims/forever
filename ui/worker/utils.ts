// These utils are copied to keep workers completely isolated and prevent a circular dependency when building chunks.
// eslint-disable-next-line @typescript-eslint/no-empty-function
export const noop = () => {};
export const sleep = (ms: number) => new Promise(r => setTimeout(r, ms));
