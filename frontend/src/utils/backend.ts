import { config } from '../../wailsjs/go/models'
import type { Case, Collection, FramingConfig, FramingMode, LengthEncoding, Endian, LengthPrefixMode, Charset, CharsetMode, ConnectionSettings, ScriptConfig, Variables } from '@/types'
import { DEFAULT_FRAMING, DEFAULT_CONNECTION_SETTINGS, DEFAULT_SCRIPT_CONFIG } from '@/types'

/**
 * TS FramingConfig -> Go config.FramingConfig
 */
export function toBackendFramingConfig(framing: FramingConfig): config.FramingConfig {
  return new config.FramingConfig({
    mode: framing.mode,
    delimiter: framing.delimiter,
    lengthEncoding: framing.lengthEncoding,
    lengthOffset: framing.lengthOffset,
    lengthBytes: framing.lengthBytes,
    endian: framing.endian,
    includeHeader: framing.includeHeader,
    lengthMode: framing.lengthMode,
    fixedSize: framing.fixedSize
  })
}

/**
 * Go config.FramingConfig -> TS FramingConfig
 */
export function fromBackendFramingConfig(goFraming: config.FramingConfig | undefined): FramingConfig {
  if (!goFraming) {
    return { ...DEFAULT_FRAMING }
  }

  return {
    mode: (goFraming.mode || 'collection') as FramingMode,
    delimiter: goFraming.delimiter,
    lengthEncoding: goFraming.lengthEncoding as LengthEncoding | undefined,
    lengthOffset: goFraming.lengthOffset,
    lengthBytes: goFraming.lengthBytes as 1 | 2 | 4 | undefined,
    endian: goFraming.endian as Endian | undefined,
    includeHeader: goFraming.includeHeader,
    lengthMode: goFraming.lengthMode as LengthPrefixMode | undefined,
    fixedSize: goFraming.fixedSize
  }
}

/**
 * TS ConnectionSettings -> Go config.ConnectionSettings
 */
export function toBackendConnectionSettings(settings: ConnectionSettings | undefined): config.ConnectionSettings | undefined {
  if (!settings) return undefined
  return new config.ConnectionSettings({
    connectTimeout: settings.connectTimeout,
    readTimeout: settings.readTimeout
  })
}

/**
 * Go config.ConnectionSettings -> TS ConnectionSettings
 */
export function fromBackendConnectionSettings(goSettings: config.ConnectionSettings | undefined): ConnectionSettings | undefined {
  if (!goSettings) return undefined
  return {
    connectTimeout: goSettings.connectTimeout || DEFAULT_CONNECTION_SETTINGS.connectTimeout,
    readTimeout: goSettings.readTimeout ?? DEFAULT_CONNECTION_SETTINGS.readTimeout
  }
}

/**
 * TS ScriptConfig -> Go config.ScriptConfig
 */
export function toBackendScriptConfig(scriptConfig: ScriptConfig | undefined): config.ScriptConfig | undefined {
  if (!scriptConfig) return undefined
  return new config.ScriptConfig({
    setupScript: scriptConfig.setupScript,
    setupEnabled: scriptConfig.setupEnabled,
    preSendScript: scriptConfig.preSendScript,
    preSendEnabled: scriptConfig.preSendEnabled,
    postRecvScript: scriptConfig.postRecvScript,
    postRecvEnabled: scriptConfig.postRecvEnabled
  })
}

/**
 * Go config.ScriptConfig -> TS ScriptConfig
 */
export function fromBackendScriptConfig(goConfig: config.ScriptConfig | undefined): ScriptConfig | undefined {
  if (!goConfig) return undefined
  return {
    setupScript: goConfig.setupScript || DEFAULT_SCRIPT_CONFIG.setupScript,
    setupEnabled: goConfig.setupEnabled ?? false,
    preSendScript: goConfig.preSendScript || DEFAULT_SCRIPT_CONFIG.preSendScript,
    preSendEnabled: goConfig.preSendEnabled ?? false,
    postRecvScript: goConfig.postRecvScript || DEFAULT_SCRIPT_CONFIG.postRecvScript,
    postRecvEnabled: goConfig.postRecvEnabled ?? false
  }
}

/**
 * TS Variables -> Go config.Variables (simple pass-through, both are maps)
 */
export function toBackendVariables(vars: Variables | undefined): Record<string, unknown> | undefined {
  if (!vars) return undefined
  return { ...vars }
}

/**
 * Go config.Variables -> TS Variables
 */
export function fromBackendVariables(goVars: Record<string, unknown> | undefined): Variables | undefined {
  if (!goVars || Object.keys(goVars).length === 0) return undefined
  const result: Variables = {}
  for (const [key, value] of Object.entries(goVars)) {
    if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
      result[key] = value
    }
  }
  return result
}

/**
 * TS Case -> Go config.SavedCase
 */
export function toBackendSavedCase(caseItem: Case): config.SavedCase {
  return new config.SavedCase({
    id: caseItem.id,
    name: caseItem.name,
    protocol: caseItem.protocol,
    host: caseItem.host,
    port: caseItem.port,
    createdAt: caseItem.createdAt.toISOString(),
    updatedAt: new Date().toISOString(),
    order: caseItem.order ?? 0,
    framing: toBackendFramingConfig(caseItem.framing),
    charset: caseItem.charset || 'collection',
    connectionSettings: toBackendConnectionSettings(caseItem.connectionSettings),
    draftMessage: caseItem.draftMessage || '',
    draftFormat: caseItem.draftFormat || 'text',
    useVariables: caseItem.useVariables ?? false,
    notes: caseItem.notes || '',
    scriptConfig: toBackendScriptConfig(caseItem.scriptConfig),
    localVariables: toBackendVariables(caseItem.localVariables),
    postRecvSample: caseItem.postRecvSample || ''
  })
}

/**
 * Go config.SavedCase -> Partial TS Case (for loading)
 */
export function fromBackendSavedCase(
  saved: config.SavedCase,
  collectionName: string
): Omit<Case, 'status' | 'messages' | 'isSaved'> {
  return {
    id: saved.id,
    name: saved.name,
    collectionName,
    protocol: saved.protocol as Case['protocol'],
    host: saved.host,
    port: saved.port,
    createdAt: new Date(saved.createdAt),
    order: saved.order ?? 0,
    framing: fromBackendFramingConfig(saved.framing),
    charset: (saved.charset as CharsetMode) || 'collection',
    connectionSettings: fromBackendConnectionSettings(saved.connectionSettings),
    draftMessage: saved.draftMessage || undefined,
    draftFormat: (saved.draftFormat as Case['draftFormat']) || undefined,
    useVariables: saved.useVariables ?? false,
    notes: saved.notes || undefined,
    scriptConfig: fromBackendScriptConfig(saved.scriptConfig),
    localVariables: fromBackendVariables(saved.localVariables),
    postRecvSample: saved.postRecvSample || undefined
  }
}

/**
 * TS Collection -> Go config.Collection (for updating)
 */
export function toBackendCollection(collection: Collection): config.Collection {
  return new config.Collection({
    name: collection.name,
    description: collection.description,
    createdAt: collection.createdAt.toISOString(),
    updatedAt: collection.updatedAt.toISOString(),
    order: collection.order ?? 0,
    sharedFraming: collection.sharedFraming
      ? toBackendFramingConfig(collection.sharedFraming)
      : undefined,
    sharedCharset: collection.sharedCharset || '',
    sharedConnectionSettings: toBackendConnectionSettings(collection.sharedConnectionSettings),
    notes: collection.notes || '',
    sharedScriptConfig: toBackendScriptConfig(collection.sharedScriptConfig),
    sharedVariables: toBackendVariables(collection.sharedVariables)
  })
}

/**
 * Go config.Collection -> TS Collection (for loading)
 */
export function fromBackendCollection(goCol: config.Collection): Omit<Collection, 'cases' | 'isExpanded'> {
  return {
    name: goCol.name,
    description: goCol.description,
    createdAt: new Date(goCol.createdAt),
    updatedAt: new Date(goCol.updatedAt),
    order: goCol.order ?? 0,
    sharedFraming: goCol.sharedFraming
      ? fromBackendFramingConfig(goCol.sharedFraming)
      : undefined,
    sharedCharset: (goCol.sharedCharset as Charset) || undefined,
    sharedConnectionSettings: fromBackendConnectionSettings(goCol.sharedConnectionSettings),
    notes: goCol.notes || undefined,
    sharedScriptConfig: fromBackendScriptConfig(goCol.sharedScriptConfig),
    sharedVariables: fromBackendVariables(goCol.sharedVariables)
  }
}

/**
 * Go config.CollectionWithCases -> TS Collection with Cases
 */
export function fromBackendCollectionWithCases(
  data: config.CollectionWithCases
): Collection {
  const collection = fromBackendCollection(data.collection)
  const cases: Case[] = (data.cases || []).map(c => ({
    ...fromBackendSavedCase(c, collection.name),
    status: 'disconnected' as const,
    messages: [],
    isSaved: true
  }))

  return {
    ...collection,
    cases,
    isExpanded: true
  }
}
