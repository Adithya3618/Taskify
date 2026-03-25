/** Matches task description meta block used by the board (see BoardComponent). */
const META_SEP = '\n---\n';

export function parseCardMeta(description: string): {
  desc: string;
  due: string;
  priority: string;
  notes: string;
} {
  const idx = description.indexOf(META_SEP);
  let desc = description;
  let due = '';
  let priority = '';
  let notes = '';
  if (idx >= 0) {
    desc = description.slice(0, idx).trim();
    const meta = description.slice(idx + META_SEP.length);
    meta.split('\n').forEach((line) => {
      if (line.startsWith('due:')) due = line.slice(4).trim();
      else if (line.startsWith('priority:')) priority = line.slice(9).trim();
      else if (line.startsWith('notes:')) notes = line.slice(6).trim();
    });
  }
  return { desc, due, priority, notes };
}

function dateKeyLocal(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

/** Build stored description with optional meta block (same as board). */
export function buildCardDescription(desc: string, due: string, priority: string, notes: string): string {
  const baseDesc = desc.trim();
  const parts: string[] = [];
  if (due.trim()) parts.push('due:' + due.trim());
  if (priority.trim()) parts.push('priority:' + priority.trim());
  if (notes.trim()) parts.push('notes:' + notes.trim());
  if (parts.length === 0) return baseDesc;
  return baseDesc + META_SEP + parts.join('\n');
}

/** Normalize due string from meta to `YYYY-MM-DD` for calendar keys, or null. */
export function parseDueToDateKey(due: string): string | null {
  if (!due?.trim()) return null;
  const s = due.trim();
  const ymd = s.match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (ymd) return `${ymd[1]}-${ymd[2]}-${ymd[3]}`;
  const dt = new Date(s);
  if (Number.isNaN(dt.getTime())) return null;
  return dateKeyLocal(dt);
}
