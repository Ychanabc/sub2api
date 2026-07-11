// 二开：i18n 深合并工具。
// 用于把二开自定义语言包（custom）作为基底，再让上游模块化语言包覆盖同名键，
// 从而保留二开独有的翻译键（如对话审计 / cyber / USDT 等），同时上游共享键以官方文案为准。
type Dict = Record<string, unknown>

function isPlainObject(value: unknown): value is Dict {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

export function deepMerge<T extends Dict>(base: Dict, override: Dict): T {
  const result: Dict = { ...base }
  for (const key of Object.keys(override)) {
    const overrideValue = override[key]
    const baseValue = result[key]
    if (isPlainObject(baseValue) && isPlainObject(overrideValue)) {
      result[key] = deepMerge(baseValue, overrideValue)
    } else {
      result[key] = overrideValue
    }
  }
  return result as T
}
