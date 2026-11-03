import json, sys, xml.etree.ElementTree as ET
src, out = sys.argv[1], sys.argv[2]
d = json.load(open(src)); t = ''.join(x['text'] for x in d); j = json.loads(t[t.find('{'):])
root = ET.fromstring(j['data']['xml'])

def inline(el):
    parts = []
    for c in el:
        tag = c.tag
        if tag == 'text':
            parts.append(inline_text(c))
        elif tag == 'date':
            parts.append(c.get('value', ''))
        elif tag == 'mention':
            parts.append(c.get('name', ''))
        elif tag in ('dropdown', 'comment', 'embed'):
            pass
        else:
            parts.append(inline(c))
    return ''.join(parts)

def inline_text(t):
    s = t.text or ''
    for c in t:
        inner = (c.text or '') + ''.join(inline_text(g) if g.tag == 'text' else (g.text or '') for g in c)
        if c.tag == 'code': s += '`' + inner + '`'
        elif c.tag == 'bold': s += '**' + inner + '**'
        elif c.tag == 'link': s += '[' + inner + '](' + c.get('href', '') + ')'
        else: s += inner
        s += c.tail or ''
    return s

def merge_code(s):
    # adjacent code runs split by edits: `a``b` -> `ab`
    return s.replace('``', '')

def block(el, depth=0):
    tag = el.tag
    if tag == 'paragraph':
        h = el.get('heading')
        txt = merge_code(inline(el))
        return ('#' * int(h) + ' ' + txt) if h else txt
    if tag == 'list':
        kind = el.get('kind'); lines = []
        for n, li in enumerate([c for c in el if c.tag == 'listItem'], 1):
            marker = f'{n}.' if kind == 'ordered' else ('- [ ]' if kind == 'check' else '-')
            if kind == 'check' and li.get('checked') == 'true': marker = '- [x]'
            sub = [block(c, depth + 1) for c in li]
            first, rest = sub[0] if sub else '', sub[1:]
            item = '    ' * depth + marker + ' ' + first
            for r in rest:
                item += '\n' + '\n'.join(('    ' * (depth + 1)) + l if l else l for l in r.split('\n'))
            lines.append(item)
        return '\n'.join(lines)
    if tag == 'table':
        rows = []
        for r in [c for c in el if c.tag == 'row']:
            cells = [merge_code(' '.join(block(p) for p in cell)).replace('|', '\\|').replace('\n', ' ') for cell in r if cell.tag == 'cell']
            rows.append('| ' + ' | '.join(cells) + ' |')
        if rows:
            ncols = rows[0].count(' | ') + 1
            rows.insert(1, '|' + '---|' * ncols)
        return '\n'.join(rows)
    if tag == 'codeBlock':
        code = ''.join((c.text or '') for c in el if c.tag == 'text')
        return '```' + (el.get('language') or '') + '\n' + code + '\n```'
    if tag == 'embed':
        return ''
    return merge_code(inline(el))

blocks = [b for b in (block(c) for c in root) if b != '']
open(out, 'w').write('\n\n'.join(blocks) + '\n')
