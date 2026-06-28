<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div v-else class="cats-home">
    <header class="cats-header">
      <nav class="cats-nav">
        <router-link to="/" class="brand">
          <span class="brand-logo">
            <img v-if="siteLogo" :src="siteLogo" alt="Cats AI" />
            <span v-else>CA</span>
          </span>
          <span class="brand-text">
            <strong>{{ displaySiteName }}</strong>
            <small>{{ copy.navSubtitle }}</small>
          </span>
        </router-link>

        <div class="nav-links">
          <a href="#platform">{{ copy.navPlatform }}</a>
          <a href="#models">{{ copy.navModels }}</a>
          <a href="#billing">{{ copy.navBilling }}</a>
        </div>

        <div class="nav-actions">
          <LocaleSwitcher />
          <button class="round-action" :title="copy.themeTitle" @click="toggleTheme">
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
          <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="nav-cta">
            {{ isAuthenticated ? copy.dashboard : copy.login }}
          </router-link>
        </div>
      </nav>
    </header>

    <main>
      <section class="hero">
        <div class="hero-copy">
          <p class="eyebrow">{{ copy.eyebrow }}</p>
          <h1>{{ copy.heroTitle }}</h1>
          <p class="hero-lede">{{ copy.heroLead }}</p>

          <div class="hero-actions">
            <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="primary-cta">
              {{ isAuthenticated ? copy.openConsole : copy.startNow }}
              <Icon name="arrowRight" size="sm" />
            </router-link>
            <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer" class="secondary-cta">
              <Icon name="book" size="sm" />
              {{ copy.docs }}
            </a>
          </div>

          <div class="trust-row">
            <span v-for="item in copy.trustItems" :key="item">
              <Icon name="checkCircle" size="sm" />
              {{ item }}
            </span>
          </div>
        </div>

        <aside class="console-shell" aria-label="Cats AI routing console preview">
          <div class="console-chrome">
            <span></span>
            <span></span>
            <span></span>
          </div>

          <div class="console-header">
            <div>
              <p>{{ copy.consoleTitle }}</p>
              <strong>{{ copy.consoleSubtitle }}</strong>
            </div>
            <span class="live-pill">
              <i></i>
              {{ copy.liveRouting }}
            </span>
          </div>

          <div class="console-grid">
            <article class="routing-card">
              <div class="mini-title">
                <Icon name="swap" size="sm" />
                {{ copy.routingTitle }}
              </div>
              <div v-for="route in copy.routes" :key="route.name" class="route-row">
                <span>{{ route.name }}</span>
                <strong>{{ route.status }}</strong>
              </div>
            </article>

            <article class="latency-card">
              <div class="mini-title">
                <Icon name="bolt" size="sm" />
                {{ copy.edgeTitle }}
              </div>
              <strong>{{ copy.edgeValue }}</strong>
              <div class="traffic-bars">
                <i v-for="i in 18" :key="i" :style="{ height: `${20 + ((i * 13) % 50)}px` }"></i>
              </div>
            </article>
          </div>

          <div class="metric-grid">
            <div v-for="metric in copy.metrics" :key="metric.label">
              <strong>{{ metric.value }}</strong>
              <span>{{ metric.label }}</span>
            </div>
          </div>
        </aside>
      </section>

      <section id="platform" class="signal-strip">
        <div v-for="item in copy.capabilityStrip" :key="item.label">
          <strong>{{ item.value }}</strong>
          <span>{{ item.label }}</span>
        </div>
      </section>

      <section class="section-head">
        <p class="eyebrow">{{ copy.platformEyebrow }}</p>
        <h2>{{ copy.platformTitle }}</h2>
        <p>{{ copy.platformLead }}</p>
      </section>

      <section class="feature-board">
        <article v-for="feature in copy.features" :key="feature.title">
          <span class="feature-icon"><Icon :name="feature.icon" size="lg" /></span>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.desc }}</p>
        </article>
      </section>

      <section id="models" class="split-panel">
        <div>
          <p class="eyebrow">{{ copy.modelEyebrow }}</p>
          <h2>{{ copy.modelTitle }}</h2>
          <p>{{ copy.modelLead }}</p>
        </div>
        <div class="model-cloud">
          <span v-for="model in copy.models" :key="model">{{ model }}</span>
        </div>
      </section>

      <section id="billing" class="billing-panel">
        <div>
          <p class="eyebrow">{{ copy.billingEyebrow }}</p>
          <h2>{{ copy.billingTitle }}</h2>
          <p>{{ copy.billingLead }}</p>
        </div>
        <div class="billing-methods">
          <span v-for="method in copy.paymentMethods" :key="method">{{ method }}</span>
        </div>
      </section>

      <section class="enterprise-callout">
        <div>
          <p class="eyebrow">{{ copy.enterpriseEyebrow }}</p>
          <h2>{{ copy.enterpriseTitle }}</h2>
          <p>{{ copy.enterpriseLead }}</p>
        </div>
        <router-link :to="isAuthenticated ? dashboardPath : '/login'" class="primary-cta">
          {{ copy.contactAction }}
          <Icon name="arrowRight" size="sm" />
        </router-link>
      </section>
    </main>

    <footer class="cats-footer">
      <span>{{ copy.footer }}</span>
      <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ copy.docsShort }}</a>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

type FeatureIcon = 'globe' | 'server' | 'chart' | 'key' | 'creditCard' | 'cpu'

interface HomeCopy {
  navSubtitle: string
  navPlatform: string
  navModels: string
  navBilling: string
  themeTitle: string
  dashboard: string
  login: string
  eyebrow: string
  heroTitle: string
  heroLead: string
  openConsole: string
  startNow: string
  docs: string
  docsShort: string
  trustItems: string[]
  consoleTitle: string
  consoleSubtitle: string
  liveRouting: string
  routingTitle: string
  edgeTitle: string
  edgeValue: string
  routes: Array<{ name: string; status: string }>
  metrics: Array<{ value: string; label: string }>
  capabilityStrip: Array<{ value: string; label: string }>
  platformEyebrow: string
  platformTitle: string
  platformLead: string
  features: Array<{ icon: FeatureIcon; title: string; desc: string }>
  modelEyebrow: string
  modelTitle: string
  modelLead: string
  models: string[]
  billingEyebrow: string
  billingTitle: string
  billingLead: string
  paymentMethods: string[]
  enterpriseEyebrow: string
  enterpriseTitle: string
  enterpriseLead: string
  contactAction: string
  footer: string
}

const currentYear = new Date().getFullYear()

const copies: Record<'zh' | 'en', HomeCopy> = {
  zh: {
    navSubtitle: '企业级 AI API 控制台',
    navPlatform: '平台能力',
    navModels: '模型覆盖',
    navBilling: '计费方式',
    themeTitle: '切换主题',
    dashboard: '进入控制台',
    login: '登录控制台',
    eyebrow: '官方 Key 渠道 · 多模型聚合 · SLA 可用性承诺',
    heroTitle: 'Cats AI，把全球大模型能力收进一个稳定入口',
    heroLead: '面向开发者和企业团队的一站式 AI API 网关。统一接入 Gemini、Antigravity、OpenAI、GPT 与 Claude 全模型，聚合 Pro 5x / 20x Team 多渠道，异常自动切换，无需等待。',
    openConsole: '打开工作台',
    startNow: '立即开始',
    docs: '接入文档',
    docsShort: '文档',
    trustItems: ['充值 1:1', '按量付费', '可开收据', '企业级对接'],
    consoleTitle: 'Cats AI Control Plane',
    consoleSubtitle: '全球模型流量实时调度',
    liveRouting: '实时路由',
    routingTitle: '自动故障切换',
    edgeTitle: '全球优化网络',
    edgeValue: '毫秒级响应',
    routes: [
      { name: 'Gemini Pro 5x', status: 'Ready' },
      { name: 'Claude Team 20x', status: 'Active' },
      { name: 'OpenAI Full Models', status: 'Failover' },
    ],
    metrics: [
      { value: '99.9%', label: 'SLA 可用性' },
      { value: '1M', label: '超长上下文' },
      { value: 'Auto', label: '宕机自动切换' },
    ],
    capabilityStrip: [
      { value: 'SLA', label: '可用性承诺' },
      { value: '5x/20x', label: 'Team 多渠道聚合' },
      { value: 'Scale', label: '高并发与负载均衡' },
      { value: 'Keys', label: '多用户 Key 权限' },
    ],
    platformEyebrow: 'Core capabilities',
    platformTitle: '为开发者打造，为规模而设计',
    platformLead: 'Cats AI 将账号池、模型路由、实时监控、权限分配和计费结算放在同一个控制平面里，既适合快速接入，也能承载企业规模化调用。',
    features: [
      { icon: 'globe', title: '全球覆盖', desc: '优化网络架构与区域入口，跨境调用保持稳定、低延迟和高可用。' },
      { icon: 'server', title: '自动负载均衡', desc: '多渠道账号池自动调度，请求高峰下也能平滑承载。' },
      { icon: 'chart', title: '实时使用监控', desc: '按用户、Key、模型和时间维度查看调用量、成本与趋势。' },
      { icon: 'key', title: '多用户 Key 管理', desc: '为团队成员分配独立 Key、额度和权限，适配企业协作。' },
      { icon: 'creditCard', title: '灵活付费', desc: '支持按量付费、套餐订阅、Alipay、信用卡、USDT 与企业收据。' },
      { icon: 'cpu', title: '企业级任务', desc: '可跑企业蒸馏、高并发专线和定制权限，需要联系客服开通。' },
    ],
    modelEyebrow: 'Models',
    modelTitle: '一个 API，多种选择',
    modelLead: '已支持 Gemini、Antigravity、OpenAI、GPT / Claude 全模型能力，保持完整输出，不做降智代理。',
    models: ['Gemini', 'Antigravity', 'OpenAI', 'GPT 全模型', 'Claude 全模型', '1M Context'],
    billingEyebrow: 'Billing',
    billingTitle: '充值 1:1，按量和订阅都能跑',
    billingLead: '余额、订单和使用记录实时可查。支持 Alipay、信用卡、USDT，可开收据并提供企业级对接流程。',
    paymentMethods: ['Alipay', 'Credit Card', 'USDT', 'Receipt', 'Subscription', 'Usage-based'],
    enterpriseEyebrow: 'Enterprise',
    enterpriseTitle: '为企业开发、规模化调用和蒸馏场景准备',
    enterpriseLead: '官方 Key 渠道、多渠道聚合、自动切换、权限控制、实时成本监控和高并发能力可以按需组合。',
    contactAction: '进入控制台',
    footer: `© ${currentYear} YUN AIS LTD. Cats AI 与大模型公司并无直接关联。`,
  },
  en: {
    navSubtitle: 'Enterprise AI API Console',
    navPlatform: 'Platform',
    navModels: 'Models',
    navBilling: 'Billing',
    themeTitle: 'Toggle theme',
    dashboard: 'Dashboard',
    login: 'Sign in',
    eyebrow: 'Official key channels · multi-model routing · SLA commitment',
    heroTitle: 'Cats AI turns global model capacity into one reliable gateway',
    heroLead: 'A unified AI API gateway for developers and enterprise teams. Connect Gemini, Antigravity, OpenAI, GPT and Claude model families, aggregate Pro 5x / 20x Team channels, and fail over automatically without waiting.',
    openConsole: 'Open console',
    startNow: 'Get started',
    docs: 'Integration docs',
    docsShort: 'Docs',
    trustItems: ['1:1 top-up', 'Usage billing', 'Receipts', 'Enterprise onboarding'],
    consoleTitle: 'Cats AI Control Plane',
    consoleSubtitle: 'Live orchestration for global AI traffic',
    liveRouting: 'Live routing',
    routingTitle: 'Automatic failover',
    edgeTitle: 'Optimized global network',
    edgeValue: 'Millisecond response',
    routes: [
      { name: 'Gemini Pro 5x', status: 'Ready' },
      { name: 'Claude Team 20x', status: 'Active' },
      { name: 'OpenAI Full Models', status: 'Failover' },
    ],
    metrics: [
      { value: '99.9%', label: 'SLA uptime' },
      { value: '1M', label: 'Long context' },
      { value: 'Auto', label: 'Failover' },
    ],
    capabilityStrip: [
      { value: 'SLA', label: 'Availability commitment' },
      { value: '5x/20x', label: 'Team channel aggregation' },
      { value: 'Scale', label: 'Concurrency and load balancing' },
      { value: 'Keys', label: 'Multi-user permissions' },
    ],
    platformEyebrow: 'Core capabilities',
    platformTitle: 'Built for developers, designed for scale',
    platformLead: 'Cats AI brings account pools, model routing, monitoring, permissions and billing into one control plane that works for fast integrations and enterprise-grade traffic.',
    features: [
      { icon: 'globe', title: 'Global coverage', desc: 'Optimized network paths and regional entry points keep cross-border calls stable and responsive.' },
      { icon: 'server', title: 'Automatic load balancing', desc: 'Requests are distributed across channel pools so traffic spikes remain smooth.' },
      { icon: 'chart', title: 'Real-time monitoring', desc: 'Track usage, cost and trends by user, key, model and time range.' },
      { icon: 'key', title: 'Multi-user key management', desc: 'Assign separate keys, quotas and permissions for team and enterprise workflows.' },
      { icon: 'creditCard', title: 'Flexible billing', desc: 'Supports usage billing, subscriptions, Alipay, credit card, USDT and receipts.' },
      { icon: 'cpu', title: 'Enterprise workloads', desc: 'Distillation, high concurrency lines and custom permissions can be enabled through support.' },
    ],
    modelEyebrow: 'Models',
    modelTitle: 'One API, many model choices',
    modelLead: 'Use Gemini, Antigravity, OpenAI, GPT and Claude model families while preserving full model capability.',
    models: ['Gemini', 'Antigravity', 'OpenAI', 'GPT full models', 'Claude full models', '1M Context'],
    billingEyebrow: 'Billing',
    billingTitle: '1:1 top-up with usage and subscription billing',
    billingLead: 'Balance, orders and usage records are visible in real time. Supports Alipay, credit card, USDT, receipts and enterprise integration.',
    paymentMethods: ['Alipay', 'Credit Card', 'USDT', 'Receipt', 'Subscription', 'Usage-based'],
    enterpriseEyebrow: 'Enterprise',
    enterpriseTitle: 'Ready for enterprise development, scale and distillation',
    enterpriseLead: 'Official key channels, channel aggregation, failover, permission control, cost monitoring and high concurrency can be combined as needed.',
    contactAction: 'Open console',
    footer: `© ${currentYear} YUN AIS LTD. Cats AI is not directly affiliated with model providers.`,
  },
}

const { locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

function normalizeSiteName(value?: string) {
  const name = value?.trim()
  return !name || name === 'Sub2API' || name === 'YUN AIS' ? 'Cats AI' : name
}

const copy = computed(() => String(locale.value).startsWith('zh') ? copies.zh : copies.en)
const displaySiteName = computed(() =>
  normalizeSiteName(appStore.cachedPublicSettings?.site_name || appStore.siteName)
)
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => isAdmin.value ? '/admin/dashboard' : '/dashboard')

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>

<style scoped>
.cats-home {
  --text: #111827;
  --muted: #667085;
  --soft: rgba(255, 255, 255, 0.68);
  --softer: rgba(255, 255, 255, 0.44);
  --line: rgba(255, 255, 255, 0.78);
  --accent: #007aff;
  --accent-soft: rgba(0, 122, 255, 0.12);
  --shadow: 0 26px 80px rgba(31, 45, 84, 0.12);
  min-height: 100dvh;
  color: var(--text);
  background:
    radial-gradient(circle at 18% 6%, rgba(0, 122, 255, 0.13), transparent 30%),
    radial-gradient(circle at 82% 12%, rgba(52, 199, 89, 0.1), transparent 28%),
    linear-gradient(180deg, #f7f9fd 0%, #eef3f9 58%, #f8fafc 100%);
}

:global(.dark) .cats-home {
  --text: #f8fafc;
  --muted: #aab6c8;
  --soft: rgba(15, 23, 42, 0.68);
  --softer: rgba(255, 255, 255, 0.07);
  --line: rgba(255, 255, 255, 0.12);
  --accent-soft: rgba(0, 122, 255, 0.18);
  --shadow: 0 28px 90px rgba(0, 0, 0, 0.32);
  background:
    radial-gradient(circle at 16% 8%, rgba(0, 122, 255, 0.2), transparent 30%),
    radial-gradient(circle at 84% 10%, rgba(52, 199, 89, 0.13), transparent 28%),
    linear-gradient(180deg, #070b14 0%, #0d1422 62%, #090d17 100%);
}

.cats-header {
  position: sticky;
  top: 0;
  z-index: 30;
  padding: 16px 24px 0;
}

.cats-nav,
main,
.cats-footer {
  width: min(1160px, calc(100% - 32px));
  margin: 0 auto;
}

.cats-nav {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 24px;
  min-height: 70px;
  border: 1px solid var(--line);
  border-radius: 28px;
  background: var(--soft);
  padding: 10px 12px;
  box-shadow: 0 18px 52px rgba(31, 45, 84, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(24px) saturate(170%);
  -webkit-backdrop-filter: blur(24px) saturate(170%);
}

.brand,
.nav-actions,
.hero-actions,
.trust-row,
.trust-row span,
.mini-title,
.live-pill,
.cats-footer,
.primary-cta,
.secondary-cta {
  display: flex;
  align-items: center;
}

.brand {
  min-width: 0;
  gap: 12px;
  color: inherit;
  text-decoration: none;
}

.brand-logo {
  display: grid;
  width: 46px;
  height: 46px;
  place-items: center;
  overflow: hidden;
  border-radius: 18px;
  color: #ffffff;
  font-size: 13px;
  font-weight: 950;
  background: linear-gradient(145deg, #101828, #344054);
  box-shadow: 0 14px 28px rgba(17, 24, 39, 0.2), inset 0 1px 0 rgba(255, 255, 255, 0.22);
}

.brand-logo img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.brand-text strong,
.brand-text small {
  display: block;
  white-space: nowrap;
}

.brand-text strong {
  font-size: 16px;
  font-weight: 950;
}

.brand-text small {
  margin-top: 2px;
  color: var(--muted);
  font-size: 12px;
  font-weight: 750;
}

.nav-links {
  display: flex;
  justify-content: center;
  gap: 6px;
}

.nav-links a,
.round-action,
.nav-cta,
.secondary-cta {
  color: var(--muted);
  font-size: 13px;
  font-weight: 850;
  text-decoration: none;
  border-radius: 999px;
}

.nav-links a {
  padding: 10px 12px;
}

.nav-actions {
  justify-content: flex-end;
  gap: 8px;
}

.round-action {
  display: grid;
  width: 42px;
  height: 42px;
  place-items: center;
  border: 1px solid rgba(152, 162, 179, 0.14);
  background: var(--softer);
  cursor: pointer;
}

.nav-cta {
  padding: 12px 16px;
  color: #ffffff;
  background: #111827;
  box-shadow: 0 12px 26px rgba(17, 24, 39, 0.16);
}

:global(.dark) .nav-cta {
  color: #07111f;
  background: #ffffff;
}

main {
  padding: 58px 0 46px;
}

.hero {
  display: grid;
  grid-template-columns: minmax(0, 0.92fr) minmax(390px, 1.08fr);
  gap: 48px;
  align-items: center;
  min-height: calc(100dvh - 178px);
}

.hero-copy {
  max-width: 670px;
}

.eyebrow {
  margin: 0 0 14px;
  color: var(--accent);
  font-size: 13px;
  font-weight: 950;
}

.hero h1 {
  margin: 0;
  max-width: 760px;
  font-size: 62px;
  line-height: 1.04;
  font-weight: 950;
  letter-spacing: 0;
}

.hero-lede,
.section-head p,
.split-panel p,
.billing-panel p,
.enterprise-callout p,
.feature-board p {
  color: var(--muted);
  line-height: 1.78;
}

.hero-lede {
  margin: 22px 0 0;
  font-size: 18px;
}

.hero-actions {
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 30px;
}

.primary-cta,
.secondary-cta {
  min-height: 48px;
  justify-content: center;
  gap: 8px;
  padding: 0 20px;
  text-decoration: none;
  transition: transform 180ms ease, box-shadow 180ms ease, background 180ms ease;
}

.primary-cta {
  color: #ffffff;
  font-size: 14px;
  font-weight: 950;
  border-radius: 999px;
  background: var(--accent);
  box-shadow: 0 18px 36px rgba(0, 122, 255, 0.26);
}

.secondary-cta {
  border: 1px solid rgba(152, 162, 179, 0.28);
  background: var(--soft);
}

.trust-row {
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 24px;
}

.trust-row span {
  gap: 6px;
  min-height: 34px;
  border: 1px solid rgba(152, 162, 179, 0.18);
  border-radius: 999px;
  padding: 0 12px;
  color: var(--muted);
  font-size: 13px;
  font-weight: 800;
  background: var(--softer);
}

.console-shell {
  position: relative;
  border: 1px solid var(--line);
  border-radius: 36px;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.8), rgba(255, 255, 255, 0.48)),
    var(--soft);
  box-shadow: var(--shadow), inset 0 1px 0 rgba(255, 255, 255, 0.86);
  padding: 18px;
  backdrop-filter: blur(28px) saturate(180%);
  -webkit-backdrop-filter: blur(28px) saturate(180%);
}

:global(.dark) .console-shell {
  background:
    linear-gradient(145deg, rgba(20, 31, 52, 0.86), rgba(7, 13, 25, 0.68)),
    var(--soft);
  box-shadow: var(--shadow), inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.console-chrome {
  display: flex;
  gap: 7px;
  padding: 2px 0 16px 4px;
}

.console-chrome span {
  width: 11px;
  height: 11px;
  border-radius: 999px;
  background: #ff5f57;
}

.console-chrome span:nth-child(2) {
  background: #ffbd2e;
}

.console-chrome span:nth-child(3) {
  background: #28c840;
}

.console-header {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  border-radius: 26px;
  padding: 18px;
  background: rgba(16, 24, 40, 0.94);
  color: #ffffff;
}

:global(.dark) .console-header {
  background: rgba(2, 6, 23, 0.82);
}

.console-header p,
.console-header strong {
  display: block;
  margin: 0;
}

.console-header p {
  color: rgba(255, 255, 255, 0.62);
  font-size: 12px;
  font-weight: 850;
}

.console-header strong {
  margin-top: 6px;
  font-size: 24px;
  line-height: 1.18;
}

.live-pill {
  align-self: flex-start;
  gap: 7px;
  border-radius: 999px;
  padding: 9px 11px;
  color: rgba(255, 255, 255, 0.86);
  font-size: 12px;
  font-weight: 850;
  background: rgba(255, 255, 255, 0.1);
  white-space: nowrap;
}

.live-pill i {
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #34c759;
  box-shadow: 0 0 0 6px rgba(52, 199, 89, 0.14);
}

.console-grid {
  display: grid;
  grid-template-columns: 1fr 0.82fr;
  gap: 14px;
  margin-top: 14px;
}

.routing-card,
.latency-card,
.signal-strip,
.feature-board article,
.split-panel,
.billing-panel,
.enterprise-callout {
  border: 1px solid var(--line);
  background: var(--soft);
  box-shadow: 0 18px 52px rgba(31, 45, 84, 0.08), inset 0 1px 0 rgba(255, 255, 255, 0.62);
  backdrop-filter: blur(22px) saturate(160%);
  -webkit-backdrop-filter: blur(22px) saturate(160%);
}

.routing-card,
.latency-card {
  border-radius: 26px;
  padding: 16px;
}

.mini-title {
  gap: 8px;
  color: var(--muted);
  font-size: 12px;
  font-weight: 900;
}

.route-row {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  margin-top: 10px;
  border-radius: 18px;
  padding: 13px;
  background: var(--softer);
}

.route-row span {
  font-size: 13px;
  font-weight: 850;
}

.route-row strong {
  color: #0b8f61;
  font-size: 12px;
}

.latency-card > strong {
  display: block;
  margin-top: 12px;
  font-size: 22px;
  line-height: 1.1;
}

.traffic-bars {
  display: flex;
  align-items: end;
  gap: 5px;
  height: 96px;
  margin-top: 18px;
}

.traffic-bars i {
  flex: 1;
  min-width: 4px;
  border-radius: 999px;
  background: var(--accent);
  opacity: 0.38;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  margin-top: 14px;
}

.metric-grid div {
  border-radius: 22px;
  padding: 18px;
  color: #ffffff;
  background: linear-gradient(145deg, #111827, #283549);
}

.metric-grid strong {
  display: block;
  font-size: 23px;
}

.metric-grid span {
  display: block;
  margin-top: 5px;
  color: rgba(255, 255, 255, 0.68);
  font-size: 12px;
  font-weight: 780;
}

.signal-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1px;
  overflow: hidden;
  margin-top: 30px;
  border-radius: 30px;
}

.signal-strip div {
  padding: 24px;
  background: rgba(255, 255, 255, 0.26);
}

:global(.dark) .signal-strip div {
  background: rgba(255, 255, 255, 0.04);
}

.signal-strip strong {
  display: block;
  font-size: 26px;
}

.signal-strip span {
  display: block;
  margin-top: 6px;
  color: var(--muted);
  font-size: 13px;
  font-weight: 820;
}

.section-head {
  max-width: 760px;
  margin-top: 78px;
}

.section-head h2,
.split-panel h2,
.billing-panel h2,
.enterprise-callout h2 {
  margin: 0;
  font-size: 38px;
  line-height: 1.1;
  font-weight: 950;
  letter-spacing: 0;
}

.feature-board {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  margin-top: 24px;
}

.feature-board article {
  min-height: 224px;
  border-radius: 30px;
  padding: 24px;
}

.feature-icon {
  display: grid;
  width: 52px;
  height: 52px;
  place-items: center;
  border-radius: 20px;
  color: var(--accent);
  background: var(--accent-soft);
}

.feature-board h3 {
  margin: 22px 0 0;
  font-size: 18px;
  font-weight: 950;
}

.feature-board p {
  margin: 8px 0 0;
  font-size: 14px;
}

.split-panel,
.billing-panel,
.enterprise-callout {
  display: grid;
  grid-template-columns: 0.9fr 1.1fr;
  gap: 28px;
  align-items: center;
  margin-top: 28px;
  border-radius: 34px;
  padding: 32px;
}

.model-cloud,
.billing-methods {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.model-cloud span,
.billing-methods span {
  border: 1px solid rgba(152, 162, 179, 0.16);
  border-radius: 999px;
  padding: 12px 16px;
  color: var(--text);
  font-size: 14px;
  font-weight: 900;
  background: var(--softer);
}

.billing-panel {
  background:
    linear-gradient(135deg, rgba(0, 122, 255, 0.1), rgba(255, 255, 255, 0.56)),
    var(--soft);
}

:global(.dark) .billing-panel {
  background:
    linear-gradient(135deg, rgba(0, 122, 255, 0.2), rgba(15, 23, 42, 0.54)),
    var(--soft);
}

.enterprise-callout {
  margin-bottom: 24px;
  background:
    linear-gradient(145deg, rgba(17, 24, 39, 0.94), rgba(31, 45, 84, 0.92));
  color: #ffffff;
}

.enterprise-callout .eyebrow,
.enterprise-callout p {
  color: rgba(255, 255, 255, 0.72);
}

.enterprise-callout .primary-cta {
  justify-self: end;
  background: #ffffff;
  color: #07111f;
  box-shadow: 0 18px 40px rgba(0, 0, 0, 0.16);
}

.cats-footer {
  justify-content: space-between;
  gap: 16px;
  padding: 0 0 34px;
  color: var(--muted);
  font-size: 13px;
  font-weight: 750;
}

.cats-footer a {
  color: inherit;
  text-decoration: none;
}

.primary-cta:hover,
.secondary-cta:hover,
.nav-cta:hover,
.round-action:hover {
  transform: translateY(-1px);
}

@media (max-width: 1020px) {
  .cats-nav {
    grid-template-columns: auto auto;
  }

  .nav-links {
    display: none;
  }

  .hero,
  .split-panel,
  .billing-panel,
  .enterprise-callout {
    grid-template-columns: 1fr;
  }

  .hero {
    min-height: auto;
  }

  .hero-copy {
    max-width: none;
  }

  .feature-board {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .model-cloud,
  .billing-methods {
    justify-content: flex-start;
  }

  .enterprise-callout .primary-cta {
    justify-self: start;
  }
}

@media (max-width: 680px) {
  .cats-header {
    padding: 10px 10px 0;
  }

  .cats-nav,
  main,
  .cats-footer {
    width: min(100% - 20px, 1160px);
  }

  .cats-nav {
    gap: 10px;
    border-radius: 24px;
  }

  .brand-text small,
  .round-action {
    display: none;
  }

  .nav-cta {
    padding: 11px 12px;
  }

  main {
    padding-top: 34px;
  }

  .hero {
    gap: 28px;
  }

  .hero h1 {
    font-size: 40px;
  }

  .hero-lede {
    font-size: 16px;
  }

  .console-shell,
  .split-panel,
  .billing-panel,
  .enterprise-callout {
    border-radius: 28px;
    padding: 20px;
  }

  .console-header,
  .console-grid,
  .metric-grid,
  .signal-strip,
  .feature-board {
    grid-template-columns: 1fr;
  }

  .console-header {
    display: grid;
  }

  .section-head {
    margin-top: 54px;
  }

  .section-head h2,
  .split-panel h2,
  .billing-panel h2,
  .enterprise-callout h2 {
    font-size: 30px;
  }

  .feature-board article {
    min-height: auto;
  }

  .cats-footer {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
