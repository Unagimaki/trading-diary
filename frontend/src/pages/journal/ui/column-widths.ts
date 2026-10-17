export const DEFAULT_COLUMN_WIDTH = 260;
export const MIN_COLUMN_WIDTH = 160;
export const MAX_COLUMN_WIDTH = 1200;
export const clampColumnWidth = (width: number) => Math.min(MAX_COLUMN_WIDTH, Math.max(MIN_COLUMN_WIDTH, Math.round(width)));
