import { atom, read, update } from 'claude-code'
import type { Register } from 'claude-code'

import type { Section, SectionStatus } from '../types'

const PANE = 'spec-progress'
const TITLE = 'Spec progress'
const PLUGIN = 'spec-progress'

// Bump the tag when Section's fields change: it drops what an earlier version stored.
const sections = atom({ plugin: 'spec-progress', key: 'sections' } as const, [], {
  shape: 'sections-2',
})

const STATUSES: SectionStatus[] = ['todo', 'in_progress', 'done']
const MARK: Record<SectionStatus, string> = { todo: '[ ]', in_progress: '[~]', done: '[x]' }
const COLOR: Record<SectionStatus, string | undefined> = {
  todo: undefined,
  in_progress: 'yellow',
  done: 'green',
}

function text(value: unknown): string | undefined {
  return typeof value === 'string' && value ? value : undefined
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    await $.command.register({
      name: 'spec-progress',
      description: 'Show the spec implementation progress pane',
    })
    await $.tool.register({
      name: 'set_sections',
      description:
        'Replace the sections shown in the spec-progress pane with this ordered list, all todo. '
        + 'Call once after splitting a spec into sections.',
      inputSchema: {
        type: 'object',
        properties: {
          sections: {
            type: 'array',
            items: {
              type: 'object',
              properties: { id: { type: 'string' }, title: { type: 'string' } },
              required: ['id', 'title'],
            },
          },
        },
        required: ['sections'],
      },
      isDeferred: false,
    })
    await $.tool.register({
      name: 'update_section',
      description:
        'Set the status of one section in the spec-progress pane. An unknown id adds a new '
        + 'section (pass title). Put the commit sha in note when a section is done.',
      inputSchema: {
        type: 'object',
        properties: {
          id: { type: 'string' },
          status: { type: 'string', enum: STATUSES },
          title: { type: 'string', description: 'Only needed for a new section' },
          note: { type: 'string', description: 'Short note under the section, e.g. a commit sha' },
        },
        required: ['id', 'status'],
      },
      isDeferred: false,
    })
    void $.ui.open({ id: PANE, title: TITLE })

    return next(e)
  })

  on('command.run', { command: 'spec-progress' }, async $ => {
    await $.ui.open({ id: PANE, title: TITLE })

    return { text: 'Spec progress pane opened.' }
  })

  on('tool.call', { tool: `mcp__${PLUGIN}__set_sections` }, async ($, e) => {
    const input = e as unknown as { sections?: unknown }
    const list = Array.isArray(input.sections) ? input.sections : []
    const next: Section[] = []
    for (const item of list) {
      const entry = item as { id?: unknown; title?: unknown }
      const id = text(entry.id)
      if (id) next.push({ id, title: text(entry.title) ?? id, status: 'todo' })
    }
    if (next.length === 0) {
      return { deny: 'sections must be a non-empty list of { id, title }' }
    }
    await update($, sections, () => next)
    void $.ui.open({ id: PANE, title: TITLE })

    // A tool result must be a string: an object fails the tool's output schema.
    return { result: `${next.length} sections set` }
  })

  on('tool.call', { tool: `mcp__${PLUGIN}__update_section` }, async ($, e) => {
    const input = e as unknown as { id?: unknown; status?: unknown; title?: unknown; note?: unknown }
    const id = text(input.id)
    const status = STATUSES.find(s => s === input.status)
    if (!id || !status) {
      return { deny: `id and status (${STATUSES.join(' | ')}) are required` }
    }
    const note = text(input.note)
    const title = text(input.title) ?? id

    await update($, sections, list => {
      if (!list.some(s => s.id === id)) {
        return [...list, { id, title, status, note }]
      }
      return list.map(s => (s.id === id ? { ...s, status, note: note ?? s.note } : s))
    })

    return { result: `${id} is now ${status}` }
  })

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text } = $.ui.resolve(e)
    const list = await read($, sections)
    const done = list.filter(s => s.status === 'done').length

    return (
      <Box flexDirection="column">
        {list.length === 0 && <Text dimColor>No sections yet.</Text>}
        {list.length > 0 && (
          <Text bold>
            {done}/{list.length} done
          </Text>
        )}
        {list.map(section => (
          <Box flexDirection="column">
            <Text color={COLOR[section.status]} dimColor={section.status === 'todo'}>
              {MARK[section.status]} {section.title}
            </Text>
            {section.note && <Text dimColor>{'    '}{section.note}</Text>}
          </Box>
        ))}
      </Box>
    )
  })
}
