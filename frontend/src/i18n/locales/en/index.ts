import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import admin from './admin'
import misc from './misc'
import custom from './custom'
import { deepMerge } from '../../deepMerge'

// 二开：以二开自定义语言包为基底，上游模块化语言包覆盖同名键。
// 保留二开独有键，同时上游共享键以官方文案为准。
const upstream = {
  ...landing,
  ...common,
  ...dashboard,
  admin,
  ...misc,
}

export default deepMerge(custom as Record<string, unknown>, upstream as Record<string, unknown>)
