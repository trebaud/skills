export type SectionStatus = 'todo' | 'in_progress' | 'done'

export type Section = {
  id: string
  title: string
  status: SectionStatus
  note?: string
}

declare module 'claude-code' {
  interface PluginState {
    'spec-progress': { sections: Shaped<Section[]> }
  }
}
