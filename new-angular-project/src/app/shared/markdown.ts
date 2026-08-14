/**
 * Minimal markdown renderer for server-generated firmware READMEs.
 *
 * The environment is offline (npm registry unreachable), so instead of
 * pulling in a full markdown library this renders exactly the subset the
 * backend's README template produces: ATX headings, GFM tables, fenced
 * code blocks, blockquotes, ordered/unordered lists (with lazy
 * continuation), bold, and inline code.
 *
 * All input is HTML-escaped first, so device names (user-controlled text
 * baked into the README) can never inject markup; Angular's [innerHTML]
 * sanitizer runs on the output as a second layer. Note the escape happens
 * before parsing, so blockquote detection below matches the escaped
 * "&gt;" prefix rather than a raw ">".
 */

/** Blockquote marker after HTML escaping. */
const QUOTE_PREFIX = '&gt;';

function escapeHtml(s: string): string {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

/** Bold + inline code (after escaping). */
function renderInline(s: string): string {
  return s.replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
}

function splitRow(line: string): string[] {
  return line
    .trim()
    .replace(/^\|/, '')
    .replace(/\|$/, '')
    .split('|')
    .map((c) => c.trim());
}

function isSeparatorRow(line: string): boolean {
  const t = line.trim();
  return t.startsWith('|') && /^\|[\s:|-]+\|?$/.test(t) && t.includes('-');
}

function isSpecialLine(line: string): boolean {
  const t = line.trim();
  return (
    t === '' ||
    t.startsWith('```') ||
    /^#{1,6}\s/.test(t) ||
    t.startsWith('|') ||
    t.startsWith(QUOTE_PREFIX) ||
    /^\s*-\s+/.test(t) ||
    /^\s*\d+\.\s+/.test(t)
  );
}

function renderTable(head: string[], rows: string[][]): string {
  const hasHeader = head.some((c) => c !== '');
  const parts: string[] = ['<table>'];
  if (hasHeader) {
    parts.push('<thead><tr>', ...head.map((c) => `<th>${renderInline(c)}</th>`), '</tr></thead>');
  }
  parts.push('<tbody>');
  for (const row of rows) {
    parts.push('<tr>', ...row.map((c) => `<td>${renderInline(c)}</td>`), '</tr>');
  }
  parts.push('</tbody></table>');
  return parts.join('');
}

export function renderMarkdown(md: string): string {
  const lines = escapeHtml(md).split('\n');
  const out: string[] = [];
  let i = 0;

  // Absorbs wrapped continuation lines into the preceding list item
  // (markdown lazy continuation): non-blank, non-structural lines roll up
  // into the item being collected.
  const takeLazyContinuation = (): string => {
    let text = '';
    while (i < lines.length && !isSpecialLine(lines[i])) {
      text += ' ' + lines[i].trim();
      i++;
    }
    return text;
  };

  while (i < lines.length) {
    const line = lines[i];

    if (!line.trim()) {
      i++;
      continue;
    }

    // Fenced code block: ``` ... ```
    if (line.trim().startsWith('```')) {
      i++;
      const code: string[] = [];
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        code.push(lines[i]);
        i++;
      }
      i++; // closing fence
      out.push(`<pre><code>${code.join('\n')}</code></pre>`);
      continue;
    }

    // ATX heading: # ... ######
    const heading = /^(#{1,6})\s+(.*)$/.exec(line);
    if (heading) {
      const level = heading[1].length;
      out.push(`<h${level}>${renderInline(heading[2].trim())}</h${level}>`);
      i++;
      continue;
    }

    // GFM table: header row + separator row + body rows.
    if (line.trim().startsWith('|') && i + 1 < lines.length && isSeparatorRow(lines[i + 1])) {
      const head = splitRow(line);
      i += 2;
      const rows: string[][] = [];
      while (i < lines.length && lines[i].trim().startsWith('|')) {
        rows.push(splitRow(lines[i]));
        i++;
      }
      out.push(renderTable(head, rows));
      continue;
    }

    // Blockquote: consecutive "> " lines merged into one paragraph.
    if (line.trim().startsWith(QUOTE_PREFIX)) {
      const quote: string[] = [];
      while (i < lines.length && lines[i].trim().startsWith(QUOTE_PREFIX)) {
        quote.push(lines[i].trim().replace(/^&gt;\s?/, ''));
        i++;
      }
      out.push(`<blockquote><p>${renderInline(quote.join(' '))}</p></blockquote>`);
      continue;
    }

    // Unordered list: consecutive "- " items.
    if (/^\s*-\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^\s*-\s+/.test(lines[i])) {
        let text = lines[i].replace(/^\s*-\s+/, '');
        i++;
        text += takeLazyContinuation();
        items.push(`<li>${renderInline(text.trim())}</li>`);
      }
      out.push(`<ul>${items.join('')}</ul>`);
      continue;
    }

    // Ordered list: consecutive "N. " items; a list interrupted by e.g. a
    // code block restarts via the <ol start> attribute so numbering reads on.
    if (/^\s*\d+\.\s+/.test(line)) {
      const items: string[] = [];
      let start = 1;
      const first = /^\s*(\d+)\.\s+/.exec(line);
      if (first) {
        start = parseInt(first[1], 10);
      }
      while (i < lines.length && /^\s*\d+\.\s+/.test(lines[i])) {
        let text = lines[i].replace(/^\s*\d+\.\s+/, '');
        i++;
        text += takeLazyContinuation();
        items.push(`<li>${renderInline(text.trim())}</li>`);
      }
      const startAttr = start !== 1 ? ` start="${start}"` : '';
      out.push(`<ol${startAttr}>${items.join('')}</ol>`);
      continue;
    }

    // Paragraph: consecutive plain lines merged.
    const para: string[] = [];
    while (i < lines.length && !isSpecialLine(lines[i])) {
      para.push(lines[i].trim());
      i++;
    }
    out.push(`<p>${renderInline(para.join(' '))}</p>`);
  }

  return out.join('\n');
}
