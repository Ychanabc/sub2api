import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'
import custom from './custom'
import erkai from './erkai'
import { deepMerge } from '../../deepMerge'

// 二开：以二开自定义语言包为基底，上游模块化语言包覆盖同名键；
// 最后叠加 erkai 补丁，补齐二开自有功能缺失的中文翻译（最高优先级）。
const upstream = {
  ...landing,
  ...common,
  ...dashboard,
  admin,
  ...misc,
}

const merged = deepMerge(custom as Record<string, unknown>, upstream as Record<string, unknown>)
export default deepMerge(merged, erkai as Record<string, unknown>)
