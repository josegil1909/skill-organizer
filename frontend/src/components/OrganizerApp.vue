<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';

// --- Domain Models ---
export type ItemType = 'skill' | 'mcp';
export type Origin = 'builtin' | 'installed' | 'custom' | string;
export type Provider =
  | 'global'
  | 'cursor'
  | 'claude'
  | 'kimi'
  | 'opencode'
  | 'hermes'
  | 'grok'
  | 'gemini'
  | 'pi'
  | string;

export interface Item {
  id: string;
  name: string;
  type: ItemType;
  provider: Provider;
  providers?: Provider[];
  instances?: Item[];
  origin?: Origin;
  category?: string;
  subCategory?: string;
  isClassified?: boolean;
  sourceUrl?: string;
  sourcePkg?: string;
  registryUrl?: string;
  sourcePath: string;
  description: string;
  command?: string;
  args?: string[];
  envKeys?: string[];
  url?: string;
  invocation: string;
  rawConfig?: string;
  family?: string;
  isParent?: boolean;
  childCount?: number;
}

export interface Stats {
  total: number;
  byType: Record<string, number>;
  byProvider: Record<string, number>;
  byOrigin?: Record<string, number>;
  classifiedCount?: number;
  unclassifiedCount?: number;
  byCategory?: Record<string, number>;
  bySubCategory?: Record<string, number>;
  byFamily?: Record<string, number>;
}

// --- Category Visual Config ---
interface CategoryStyle {
  label: string;
  badgeClass: string;
  cardBadgeClass: string;
  dotClass: string;
}

const CATEGORY_STYLES: Record<string, CategoryStyle> = {
  'bug bounty & security': {
    label: 'Bug Bounty & Security',
    badgeClass: 'border-rose-500/30 bg-rose-500/10 text-rose-300',
    cardBadgeClass: 'border-rose-500/30 bg-rose-500/10 text-rose-300',
    dotClass: 'bg-rose-400',
  },
  'gentle ai & sdd': {
    label: 'Gentle AI & SDD',
    badgeClass: 'border-violet-500/30 bg-violet-500/10 text-violet-300',
    cardBadgeClass: 'border-violet-500/30 bg-violet-500/10 text-violet-300',
    dotClass: 'bg-violet-400',
  },
  'frontend & ui/ux': {
    label: 'Frontend & UI/UX',
    badgeClass: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300',
    cardBadgeClass: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300',
    dotClass: 'bg-emerald-400',
  },
  'workflow & engineering': {
    label: 'Workflow & Engineering',
    badgeClass: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
    cardBadgeClass: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
    dotClass: 'bg-amber-400',
  },
  'ai core & meta': {
    label: 'AI Core & Meta',
    badgeClass: 'border-cyan-500/30 bg-cyan-500/10 text-cyan-300',
    cardBadgeClass: 'border-cyan-500/30 bg-cyan-500/10 text-cyan-300',
    dotClass: 'bg-cyan-400',
  },
  'cloud & integrations': {
    label: 'Cloud & Integrations',
    badgeClass: 'border-blue-500/30 bg-blue-500/10 text-blue-300',
    cardBadgeClass: 'border-blue-500/30 bg-blue-500/10 text-blue-300',
    dotClass: 'bg-blue-400',
  },
  'unclassified': {
    label: 'Unclassified',
    badgeClass: 'border-amber-500/40 bg-amber-950/30 text-amber-300',
    cardBadgeClass: 'border-amber-500/40 bg-amber-950/30 text-amber-300',
    dotClass: 'bg-amber-400',
  },
};

function getCategoryStyle(category?: string): CategoryStyle {
  const key = (category || 'unclassified').toLowerCase();
  if (CATEGORY_STYLES[key]) {
    return CATEGORY_STYLES[key];
  }
  return {
    label: category || 'Unclassified',
    badgeClass: 'border-zinc-500/30 bg-zinc-500/10 text-zinc-300',
    cardBadgeClass: 'border-zinc-500/30 bg-zinc-500/10 text-zinc-300',
    dotClass: 'bg-zinc-400',
  };
}

// --- Origin Visual Config ---
interface OriginStyle {
  label: string;
  badgeClass: string;
  cardBadgeClass: string;
  dotClass: string;
}

const ORIGIN_STYLES: Record<string, OriginStyle> = {
  custom: {
    label: 'Custom',
    badgeClass: 'border-purple-500/30 bg-purple-500/10 text-purple-300',
    cardBadgeClass: 'border-purple-500/30 bg-purple-500/10 text-purple-300',
    dotClass: 'bg-purple-400',
  },
  installed: {
    label: 'Installed',
    badgeClass: 'border-blue-500/30 bg-blue-500/10 text-blue-300',
    cardBadgeClass: 'border-blue-500/30 bg-blue-500/10 text-blue-300',
    dotClass: 'bg-blue-400',
  },
  builtin: {
    label: 'Built-in',
    badgeClass: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
    cardBadgeClass: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
    dotClass: 'bg-amber-400',
  },
};

function getOriginStyle(origin?: string): OriginStyle {
  const key = (origin || 'custom').toLowerCase();
  if (ORIGIN_STYLES[key]) {
    return ORIGIN_STYLES[key];
  }
  return {
    label: origin ? origin.charAt(0).toUpperCase() + origin.slice(1) : 'Custom',
    badgeClass: 'border-zinc-500/30 bg-zinc-500/10 text-zinc-300',
    cardBadgeClass: 'border-zinc-500/30 bg-zinc-500/10 text-zinc-300',
    dotClass: 'bg-zinc-400',
  };
}

// --- Provider Visual Config ---
interface ProviderStyle {
  label: string;
  badgeClass: string;
  cardBorderClass: string;
  dotClass: string;
}

const PROVIDER_STYLES: Record<string, ProviderStyle> = {
  cursor: {
    label: 'Cursor',
    badgeClass: 'border-sky-500/30 bg-sky-500/10 text-sky-300',
    cardBorderClass: 'hover:border-sky-500/40',
    dotClass: 'bg-sky-400',
  },
  claude: {
    label: 'Claude',
    badgeClass: 'border-amber-500/30 bg-amber-500/10 text-amber-300',
    cardBorderClass: 'hover:border-amber-500/40',
    dotClass: 'bg-amber-400',
  },
  kimi: {
    label: 'Kimi',
    badgeClass: 'border-purple-500/30 bg-purple-500/10 text-purple-300',
    cardBorderClass: 'hover:border-purple-500/40',
    dotClass: 'bg-purple-400',
  },
  opencode: {
    label: 'OpenCode',
    badgeClass: 'border-teal-500/30 bg-teal-500/10 text-teal-300',
    cardBorderClass: 'hover:border-teal-500/40',
    dotClass: 'bg-teal-400',
  },
  hermes: {
    label: 'Hermes',
    badgeClass: 'border-rose-500/30 bg-rose-500/10 text-rose-300',
    cardBorderClass: 'hover:border-rose-500/40',
    dotClass: 'bg-rose-400',
  },
  grok: {
    label: 'Grok',
    badgeClass: 'border-red-500/30 bg-red-500/10 text-red-300',
    cardBorderClass: 'hover:border-red-500/40',
    dotClass: 'bg-red-400',
  },
  gemini: {
    label: 'Gemini',
    badgeClass: 'border-indigo-500/30 bg-indigo-500/10 text-indigo-300',
    cardBorderClass: 'hover:border-indigo-500/40',
    dotClass: 'bg-indigo-400',
  },
  pi: {
    label: 'Pi',
    badgeClass: 'border-yellow-500/30 bg-yellow-500/10 text-yellow-300',
    cardBorderClass: 'hover:border-yellow-500/40',
    dotClass: 'bg-yellow-400',
  },
  global: {
    label: 'Global',
    badgeClass: 'border-zinc-500/30 bg-zinc-500/10 text-zinc-300',
    cardBorderClass: 'hover:border-zinc-500/40',
    dotClass: 'bg-zinc-400',
  },
};

function getProviderStyle(provider: string): ProviderStyle {
  const key = (provider || '').toLowerCase();
  if (PROVIDER_STYLES[key]) {
    return PROVIDER_STYLES[key];
  }
  return {
    label: provider ? provider.charAt(0).toUpperCase() + provider.slice(1) : 'Unknown',
    badgeClass: 'border-slate-500/30 bg-slate-500/10 text-slate-300',
    cardBorderClass: 'hover:border-slate-500/40',
    dotClass: 'bg-slate-400',
  };
}

// --- Sample / Mock Data for Offline Preview ---
const SAMPLE_ITEMS: Item[] = [
  {
    id: 'cursor:mcp:filesystem',
    name: 'filesystem',
    type: 'mcp',
    provider: 'cursor',
    origin: 'installed',
    category: 'Cloud & Integrations',
    subCategory: 'Developer Tools & Infra',
    isClassified: true,
    sourcePath: '/home/user/.cursor/mcp.json',
    description: 'Secure local filesystem MCP server enabling restricted read/write file access within designated workspace directories.',
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-filesystem', '/home/user/Projects'],
    envKeys: ['LOG_LEVEL', 'MAX_FILE_SIZE'],
    invocation: 'npx -y @modelcontextprotocol/server-filesystem /home/user/Projects',
    rawConfig: JSON.stringify(
      {
        command: 'npx',
        args: ['-y', '@modelcontextprotocol/server-filesystem', '/home/user/Projects'],
        env: { LOG_LEVEL: 'info' },
      },
      null,
      2
    ),
  },
  {
    id: 'opencode:mcp:github',
    name: 'github-mcp',
    type: 'mcp',
    provider: 'opencode',
    origin: 'installed',
    category: 'Cloud & Integrations',
    subCategory: 'Developer Tools & Infra',
    isClassified: true,
    sourcePath: '/home/user/.config/opencode/mcp.json',
    description: 'GitHub API integration for listing branches, managing pull requests, searching repositories, and reading issues.',
    command: 'docker',
    args: ['run', '-i', '--rm', 'mcp/github'],
    envKeys: ['GITHUB_PERSONAL_ACCESS_TOKEN'],
    invocation: 'docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN mcp/github',
    rawConfig: JSON.stringify(
      {
        command: 'docker',
        args: ['run', '-i', '--rm', 'mcp/github'],
        env: { GITHUB_PERSONAL_ACCESS_TOKEN: '***' },
      },
      null,
      2
    ),
  },
  {
    id: 'claude:mcp:postgres',
    name: 'postgres-inspector',
    type: 'mcp',
    provider: 'claude',
    origin: 'installed',
    category: 'Cloud & Integrations',
    subCategory: 'Developer Tools & Infra',
    isClassified: true,
    sourcePath: '/home/user/.config/Claude/claude_desktop_config.json',
    description: 'Direct read-only PostgreSQL schema inspection, table metadata analysis, and SQL query execution.',
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-postgres', 'postgresql://localhost/organizer_dev'],
    envKeys: ['POSTGRES_PASSWORD', 'PGSSLMODE'],
    invocation: 'npx -y @modelcontextprotocol/server-postgres postgresql://localhost/organizer_dev',
    rawConfig: JSON.stringify(
      {
        command: 'npx',
        args: ['-y', '@modelcontextprotocol/server-postgres', 'postgresql://localhost/organizer_dev'],
        env: { PGSSLMODE: 'disable' },
      },
      null,
      2
    ),
  },
  {
    id: 'hermes:mcp:brave-search',
    name: 'brave-search',
    type: 'mcp',
    provider: 'hermes',
    origin: 'installed',
    category: 'Cloud & Integrations',
    subCategory: 'Developer Tools & Infra',
    isClassified: true,
    sourcePath: '/home/user/.hermes/mcp.json',
    description: 'Web and real-time news search integration powered by the Brave Search REST API.',
    command: 'npx',
    args: ['-y', '@modelcontextprotocol/server-brave-search'],
    envKeys: ['BRAVE_API_KEY'],
    invocation: 'npx -y @modelcontextprotocol/server-brave-search',
    rawConfig: JSON.stringify(
      {
        command: 'npx',
        args: ['-y', '@modelcontextprotocol/server-brave-search'],
        env: { BRAVE_API_KEY: '***' },
      },
      null,
      2
    ),
  },
  {
    id: 'grok:mcp:fetch',
    name: 'fetch-converter',
    type: 'mcp',
    provider: 'grok',
    origin: 'installed',
    category: 'Cloud & Integrations',
    subCategory: 'Developer Tools & Infra',
    isClassified: true,
    sourcePath: '/home/user/.grok/mcp.json',
    description: 'High-performance web fetcher and HTML-to-Markdown converter for web scraping and LLM context ingestion.',
    command: 'uvx',
    args: ['mcp-server-fetch'],
    envKeys: [],
    invocation: 'uvx mcp-server-fetch',
    rawConfig: JSON.stringify(
      {
        command: 'uvx',
        args: ['mcp-server-fetch'],
      },
      null,
      2
    ),
  },
  {
    id: 'kimi:mcp:sqlite',
    name: 'sqlite-explorer',
    type: 'mcp',
    provider: 'kimi',
    origin: 'installed',
    category: 'Cloud & Integrations',
    subCategory: 'Developer Tools & Infra',
    isClassified: true,
    sourcePath: '/home/user/.kimi/mcp.json',
    description: 'Local SQLite database inspector for querying schemas, running parameter queries, and dumping tabular data.',
    command: 'uvx',
    args: ['mcp-server-sqlite', '--db', '/home/user/Projects/Organizer/backend/data.db'],
    envKeys: [],
    invocation: 'uvx mcp-server-sqlite --db /home/user/Projects/Organizer/backend/data.db',
    rawConfig: JSON.stringify(
      {
        command: 'uvx',
        args: ['mcp-server-sqlite', '--db', '/home/user/Projects/Organizer/backend/data.db'],
      },
      null,
      2
    ),
  },
  {
    id: 'global:skill:a11y-audit',
    name: 'a11y-debugging',
    type: 'skill',
    provider: 'global',
    origin: 'builtin',
    category: 'Frontend & UI/UX',
    subCategory: 'Web Standards & Testing',
    isClassified: true,
    sourcePath: '/home/user/.agents/skills/a11y-audit/SKILL.md',
    description: 'Automated accessibility audit based on web.dev guidelines. Verifies ARIA labels, contrast ratios, and keyboard navigation.',
    invocation: 'Audit accessibility for current view and components according to WCAG 2.1 AA',
    rawConfig: '---\nname: a11y-debugging\ndescription: Audit web views for accessibility compliance\n---\n\n# Accessibility Debugging\n\n1. Inspect semantic DOM tree\n2. Verify all interactive controls have accessible names\n3. Check color contrast ratio >= 4.5:1 for normal text\n4. Ensure full keyboard focusability and visible focus rings.',
  },
  {
    id: 'cursor:skill:modern-web',
    name: 'modern-web-guidance',
    type: 'skill',
    provider: 'cursor',
    origin: 'installed',
    category: 'Frontend & UI/UX',
    subCategory: 'Web Standards & Testing',
    isClassified: true,
    sourceUrl: 'https://github.com/GoogleChrome/modern-web-guidance.git',
    sourcePath: '/home/user/.cursor/skills/modern-web/SKILL.md',
    description: 'Search tool for modern web standards, Baseline browser compatibility, CSS container queries, and native dialog APIs.',
    invocation: 'Check modern web standards and baseline browser support for proposed UI features',
    rawConfig: '---\nname: modern-web-guidance\ndescription: Modern web standards search and recommendations\n---\n\n# Modern Web Guidance\n\nAdhere to Baseline widely available features. Prefer native <dialog> and Popover API over heavy external modal packages.',
  },
  {
    id: 'global:skill:security-audit',
    name: 'security-audit',
    type: 'skill',
    provider: 'global',
    origin: 'installed',
    category: 'Bug Bounty & Security',
    subCategory: 'Audit & Compliance',
    isClassified: true,
    sourceUrl: 'https://github.com/cloudflare/security-audit-skill.git',
    sourcePath: '/home/user/.agents/skills/security-audit/SKILL.md',
    description: 'Performs dependency vulnerability audits (npm audit, govulncheck) and inspects repository security baselines.',
    invocation: 'Audit dependency tree and source code for known security vulnerabilities and secrets leakage',
    rawConfig: '---\nname: security-audit\ndescription: Security vulnerability auditor\n---\n\n# Security Auditor\n\nRuns dependency checks, verifies that no raw secrets or tokens are committed, and checks TLS configurations.',
  },
  {
    id: 'pi:skill:code-refactor',
    name: 'clean-architecture-review',
    type: 'skill',
    provider: 'pi',
    origin: 'custom',
    category: 'Workflow & Engineering',
    subCategory: 'Code Quality & Testing',
    isClassified: true,
    sourcePath: '/home/user/.pi/agent/skills/clean-architecture/SKILL.md',
    description: 'Enforces hexagonal boundaries, container-presentational separation, and solid domain logic isolation.',
    invocation: 'Review architecture boundaries against hexagonal and clean architecture guidelines',
    rawConfig: '---\nname: clean-architecture-review\ndescription: Clean architecture review skill\n---\n\n# Clean Architecture Review\n\n- Ensure domain logic does not import HTTP or database adapters\n- Keep controllers thin and side-effect free\n- Validate interfaces at domain boundary',
  },
  {
    id: 'global:skill:vue',
    name: 'vue',
    type: 'skill',
    provider: 'global',
    origin: 'installed',
    category: 'Frontend & UI/UX',
    subCategory: 'Web Standards & Testing',
    isClassified: true,
    family: 'vue',
    isParent: true,
    childCount: 2,
    sourcePath: '/home/user/.agents/skills/vue/SKILL.md',
    description: 'Core Vue 3 architecture, Composition API reactivity patterns, and ecosystem standards.',
    invocation: 'Audit and guide Vue 3 application architecture and best practices',
    rawConfig: '---\nname: vue\ndescription: Core Vue 3 architecture guidelines\n---\n\n# Vue Core',
  },
  {
    id: 'global:skill:vue-best-practices',
    name: 'vue-best-practices',
    type: 'skill',
    provider: 'global',
    origin: 'installed',
    category: 'Frontend & UI/UX',
    subCategory: 'Web Standards & Testing',
    isClassified: true,
    family: 'vue',
    isParent: false,
    childCount: 0,
    sourcePath: '/home/user/.agents/skills/vue-best-practices/SKILL.md',
    description: 'Production standards for Vue 3 script setup, composable patterns, and performant computed caches.',
    invocation: 'Verify Vue 3 component patterns against official style guide',
    rawConfig: '---\nname: vue-best-practices\ndescription: Vue 3 production guidelines\n---\n\n# Vue Best Practices',
  },
  {
    id: 'global:skill:vue-pinia-best-practices',
    name: 'vue-pinia-best-practices',
    type: 'skill',
    provider: 'global',
    origin: 'installed',
    category: 'Frontend & UI/UX',
    subCategory: 'Web Standards & Testing',
    isClassified: true,
    family: 'vue',
    isParent: false,
    childCount: 0,
    sourcePath: '/home/user/.agents/skills/vue-pinia-best-practices/SKILL.md',
    description: 'Pinia state store architecture, typed actions, getter memoization, and SSR hydration safety.',
    invocation: 'Review Pinia store modularity and hydration safety in Vue 3',
    rawConfig: '---\nname: vue-pinia-best-practices\ndescription: Pinia store design patterns\n---\n\n# Pinia Best Practices',
  },
  {
    id: 'global:skill:hunt-xss',
    name: 'hunt-xss',
    type: 'skill',
    provider: 'global',
    origin: 'custom',
    category: 'Bug Bounty & Security',
    subCategory: 'Web Vulnerabilities',
    isClassified: true,
    family: 'hunt',
    isParent: false,
    childCount: 0,
    sourcePath: '/home/user/.agents/skills/hunt-xss/SKILL.md',
    description: 'Systematic Cross-Site Scripting (XSS) vulnerability detection and DOM sink taint analysis.',
    invocation: 'Hunt for reflected, stored, and DOM-based XSS vulnerabilities in target parameters',
    rawConfig: '---\nname: hunt-xss\ndescription: XSS vulnerability assessment methodology\n---\n\n# Hunt XSS',
  },
  {
    id: 'global:skill:hunt-sqli',
    name: 'hunt-sqli',
    type: 'skill',
    provider: 'global',
    origin: 'custom',
    category: 'Bug Bounty & Security',
    subCategory: 'Web Vulnerabilities',
    isClassified: true,
    family: 'hunt',
    isParent: false,
    childCount: 0,
    sourcePath: '/home/user/.agents/skills/hunt-sqli/SKILL.md',
    description: 'SQL injection payload construction, error analysis, and boolean/time-based blind exfiltration.',
    invocation: 'Analyze backend parameters for SQL injection and ORM bypass vulnerabilities',
    rawConfig: '---\nname: hunt-sqli\ndescription: SQL injection hunting guide\n---\n\n# Hunt SQLi',
  },
];

// --- Reactive State ---
const items = ref<Item[]>([]);
const stats = ref<Stats | null>(null);
const isLoading = ref<boolean>(true);
const isRefreshing = ref<boolean>(false);
const backendError = ref<string | null>(null);
const isConnected = ref<boolean>(false);
const isUsingSampleData = ref<boolean>(false);

// Filter & Search State
const searchQuery = ref<string>('');
const selectedType = ref<'all' | 'mcp' | 'skill'>('all');
const selectedProvider = ref<string>('all');
const selectedOrigin = ref<string>('all');
const selectedClassification = ref<'all' | 'classified' | 'unclassified'>('all');
const selectedCategory = ref<string>('all');
const selectedSubCategory = ref<string>('all');
const selectedFamily = ref<string>('all');
const isDeduplicated = ref<boolean>(true);

// Modal State
const selectedItem = ref<Item | null>(null);
const activeInstanceIndex = ref<number>(0);
const isModalOpen = ref<boolean>(false);
const activeModalTab = ref<'details' | 'raw'>('details');
const previouslyFocusedElement = ref<HTMLElement | null>(null);

// Search input DOM reference
const searchInputRef = ref<HTMLInputElement | null>(null);

// Copy state tracking
const copiedId = ref<string | null>(null);
let copyTimeout: ReturnType<typeof setTimeout> | null = null;

async function copyText(text: string, id: string, event?: Event) {
  if (event) {
    event.stopPropagation();
  }
  try {
    await navigator.clipboard.writeText(text);
    copiedId.value = id;
    if (copyTimeout) clearTimeout(copyTimeout);
    copyTimeout = setTimeout(() => {
      copiedId.value = null;
    }, 2000);
  } catch (err) {
    console.error('Failed to copy to clipboard', err);
  }
}

// --- API Service Calls ---
const API_BASE = 'http://localhost:8080';

async function fetchData(silent = false) {
  if (!silent) {
    isLoading.value = true;
  }
  backendError.value = null;

  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 4000);

    const [itemsRes, statsRes] = await Promise.all([
      fetch(`${API_BASE}/api/items`, { signal: controller.signal }),
      fetch(`${API_BASE}/api/stats`, { signal: controller.signal }),
    ]);

    clearTimeout(timeoutId);

    if (!itemsRes.ok || !statsRes.ok) {
      throw new Error(`API responded with status ${itemsRes.status}/${statsRes.status}`);
    }

    const fetchedItems: Item[] = await itemsRes.json();
    const fetchedStats: Stats = await statsRes.json();

    items.value = Array.isArray(fetchedItems) ? fetchedItems : [];
    stats.value = fetchedStats;
    isConnected.value = true;
    isUsingSampleData.value = false;
  } catch (err: unknown) {
    isConnected.value = false;
    const message = err instanceof Error ? err.message : 'Connection failed';
    backendError.value = `Backend unreachable at ${API_BASE} (${message}). Showing local sample catalog.`;

    // Fallback to sample data for smooth developer UX
    if (items.value.length === 0) {
      items.value = SAMPLE_ITEMS;
      stats.value = computeStatsFromItems(SAMPLE_ITEMS);
      isUsingSampleData.value = true;
    }
  } finally {
    isLoading.value = false;
  }
}

function computeStatsFromItems(itemList: Item[]): Stats {
  const byType: Record<string, number> = { mcp: 0, skill: 0 };
  const byProvider: Record<string, number> = {};
  const byOrigin: Record<string, number> = { custom: 0, installed: 0, builtin: 0 };
  const byCategory: Record<string, number> = {};
  const bySubCategory: Record<string, number> = {};
  const byFamily: Record<string, number> = {};
  let classifiedCount = 0;
  let unclassifiedCount = 0;

  for (const it of itemList) {
    byType[it.type] = (byType[it.type] || 0) + 1;
    byProvider[it.provider] = (byProvider[it.provider] || 0) + 1;
    const o = (it.origin || 'custom').toLowerCase();
    byOrigin[o] = (byOrigin[o] || 0) + 1;

    if (it.isClassified) {
      classifiedCount++;
    } else {
      unclassifiedCount++;
    }

    const cat = it.category || 'Unclassified';
    byCategory[cat] = (byCategory[cat] || 0) + 1;

    const sub = it.subCategory || 'Pending Triage';
    bySubCategory[sub] = (bySubCategory[sub] || 0) + 1;

    if (it.family) {
      const fam = it.family.toLowerCase();
      byFamily[fam] = (byFamily[fam] || 0) + 1;
    }
  }

  return {
    total: itemList.length,
    byType,
    byProvider,
    byOrigin,
    byCategory,
    bySubCategory,
    byFamily,
    classifiedCount,
    unclassifiedCount,
  };
}

async function handleRefresh() {
  if (isRefreshing.value) return;
  isRefreshing.value = true;

  try {
    const res = await fetch(`${API_BASE}/api/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });

    if (!res.ok) {
      throw new Error(`Refresh failed with status ${res.status}`);
    }

    await fetchData(true);
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : 'Refresh failed';
    backendError.value = `Unable to re-scan backend: ${msg}. Make sure 'go run main.go' is running.`;
    // If not connected, try re-fetching anyway
    await fetchData(true);
  } finally {
    isRefreshing.value = false;
  }
}

// --- Canonical Items (Aggregated across all AI clients) ---
const canonicalItems = computed<Item[]>(() => {
  const map = new Map<string, Item>();
  for (const item of items.value) {
    const key = `${item.type}:${item.name.toLowerCase()}`;
    if (!map.has(key)) {
      map.set(key, {
        ...item,
        providers: [item.provider],
        instances: [item],
      });
    } else {
      const existing = map.get(key)!;
      if (!existing.providers?.includes(item.provider)) {
        existing.providers = [...(existing.providers || []), item.provider];
      }
      existing.instances = [...(existing.instances || []), item];
      // Inherit taxonomy classification if another instance is classified
      if (!existing.isClassified && item.isClassified) {
        existing.isClassified = item.isClassified;
        existing.category = item.category;
        existing.subCategory = item.subCategory;
      }
      // Inherit family attributes if existing didn't have them
      if (!existing.family && item.family) {
        existing.family = item.family;
        existing.isParent = item.isParent;
        existing.childCount = item.childCount;
      }
      if (item.isParent) {
        existing.isParent = true;
      }
      if ((item.childCount || 0) > (existing.childCount || 0)) {
        existing.childCount = item.childCount;
      }
      // Inherit longer description if existing is minimal
      if ((!existing.description || existing.description.length < 15) && item.description) {
        existing.description = item.description;
      }
      // Inherit sourceUrl, sourcePkg, registryUrl if missing
      if (!existing.sourceUrl && item.sourceUrl) {
        existing.sourceUrl = item.sourceUrl;
      }
      if (!existing.sourcePkg && item.sourcePkg) {
        existing.sourcePkg = item.sourcePkg;
      }
      if (!existing.registryUrl && item.registryUrl) {
        existing.registryUrl = item.registryUrl;
      }
    }
  }
  return Array.from(map.values());
});

// --- Active Filters Status ---
const hasActiveFilters = computed<boolean>(() => {
  return Boolean(
    searchQuery.value.trim() !== '' ||
    selectedType.value !== 'all' ||
    selectedProvider.value !== 'all' ||
    selectedOrigin.value !== 'all' ||
    selectedClassification.value !== 'all' ||
    selectedCategory.value !== 'all' ||
    selectedSubCategory.value !== 'all' ||
    selectedFamily.value !== 'all'
  );
});

// --- Filtered / Display Items Computation ---
const displayItems = computed<Item[]>(() => {
  const query = searchQuery.value.trim().toLowerCase();
  const type = selectedType.value;
  const provider = selectedProvider.value.toLowerCase();
  const origin = selectedOrigin.value.toLowerCase();
  const classification = selectedClassification.value;
  const category = selectedCategory.value.toLowerCase();
  const subCategory = selectedSubCategory.value.toLowerCase();

  const sourceList = isDeduplicated.value ? canonicalItems.value : items.value;

  return sourceList.filter((item) => {
    // 1. Type filter
    if (type !== 'all' && item.type !== type) {
      return false;
    }

    // 2. Provider filter
    if (provider !== 'all') {
      if (item.providers && item.providers.length > 0) {
        if (!item.providers.some((p) => p.toLowerCase() === provider)) {
          return false;
        }
      } else if (item.provider.toLowerCase() !== provider) {
        return false;
      }
    }

    // 3. Origin filter
    if (origin !== 'all') {
      if (item.instances && item.instances.length > 0) {
        if (!item.instances.some((inst) => (inst.origin || 'custom').toLowerCase() === origin)) {
          return false;
        }
      } else if ((item.origin || 'custom').toLowerCase() !== origin) {
        return false;
      }
    }

    // 4. Classification status filter
    if (classification === 'classified' && !item.isClassified) {
      return false;
    }
    if (classification === 'unclassified' && item.isClassified) {
      return false;
    }

    // 5. Category filter
    if (category !== 'all' && (item.category || 'Unclassified').toLowerCase() !== category) {
      return false;
    }

    // 6. SubCategory filter
    if (subCategory !== 'all' && (item.subCategory || 'Pending Triage').toLowerCase() !== subCategory) {
      return false;
    }

    // 7. Family filter
    if (selectedFamily.value !== 'all' && (item.family || '').toLowerCase() !== selectedFamily.value.toLowerCase()) {
      return false;
    }

    // 8. Search query filter
    if (!query) {
      return true;
    }

    const matchesName = item.name.toLowerCase().includes(query);
    const matchesFamily = (item.family || '').toLowerCase().includes(query);
    const matchesCategory = (item.category || '').toLowerCase().includes(query);
    const matchesSubCategory = (item.subCategory || '').toLowerCase().includes(query);
    const matchesDesc = (item.description || '').toLowerCase().includes(query);
    const matchesProvider =
      item.providers?.some((p) => p.toLowerCase().includes(query)) ||
      item.provider.toLowerCase().includes(query);
    const matchesOrigin = (item.origin || 'custom').toLowerCase().includes(query);
    const matchesPath = (item.sourcePath || '').toLowerCase().includes(query);
    const matchesCommand = (item.command || '').toLowerCase().includes(query);
    const matchesInvocation = (item.invocation || '').toLowerCase().includes(query);
    const matchesSourceUrl = (item.sourceUrl || '').toLowerCase().includes(query);
    const matchesSourcePkg = (item.sourcePkg || '').toLowerCase().includes(query);
    const matchesRegistryUrl = (item.registryUrl || '').toLowerCase().includes(query);
    const matchesArgs = item.args?.some((arg) => arg.toLowerCase().includes(query)) ?? false;
    const matchesEnv = item.envKeys?.some((k) => k.toLowerCase().includes(query)) ?? false;

    // Check across instances if canonical
    const matchesInstances =
      item.instances?.some(
        (inst) =>
          (inst.sourcePath || '').toLowerCase().includes(query) ||
          (inst.command || '').toLowerCase().includes(query) ||
          (inst.invocation || '').toLowerCase().includes(query) ||
          (inst.sourceUrl || '').toLowerCase().includes(query) ||
          (inst.sourcePkg || '').toLowerCase().includes(query) ||
          (inst.registryUrl || '').toLowerCase().includes(query) ||
          inst.args?.some((arg) => arg.toLowerCase().includes(query)) ||
          inst.envKeys?.some((k) => k.toLowerCase().includes(query))
      ) ?? false;

    return (
      matchesName ||
      matchesFamily ||
      matchesCategory ||
      matchesSubCategory ||
      matchesDesc ||
      matchesProvider ||
      matchesOrigin ||
      matchesPath ||
      matchesCommand ||
      matchesInvocation ||
      matchesSourceUrl ||
      matchesSourcePkg ||
      matchesRegistryUrl ||
      matchesArgs ||
      matchesEnv ||
      matchesInstances
    );
  });
});

// Alias for backwards compatibility
const filteredItems = displayItems;

interface CatalogGroup {
  family: string;
  parent: Item | null;
  items: Item[];
}

const catalogGroups = computed<CatalogGroup[]>(() => {
  const list = displayItems.value;
  const grouped = new Map<string, Item[]>();
  const ungrouped: Item[] = [];

  for (const item of list) {
    const fam = (item.family || '').trim();
    if (!fam) {
      ungrouped.push(item);
      continue;
    }
    const key = fam.toLowerCase();
    const bucket = grouped.get(key) || [];
    bucket.push(item);
    grouped.set(key, bucket);
  }

  const sortPack = (items: Item[]) =>
    items.slice().sort((a, b) => {
      if (a.isParent && !b.isParent) return -1;
      if (!a.isParent && b.isParent) return 1;
      return a.name.localeCompare(b.name);
    });

  const groups: CatalogGroup[] = [...grouped.entries()]
    .sort((a, b) => b[1].length - a[1].length || a[0].localeCompare(b[0]))
    .map(([family, items]) => {
      const sorted = sortPack(items);
      return {
        family,
        parent: sorted.find((i) => i.isParent) || null,
        items: sorted,
      };
    });

  if (ungrouped.length > 0) {
    groups.push({ family: '', parent: null, items: ungrouped });
  }
  return groups;
});

// Currently selected instance inside the modal
const currentModalInstance = computed<Item>(() => {
  if (!selectedItem.value) return {} as Item;
  if (selectedItem.value.instances && selectedItem.value.instances.length > 0) {
    return selectedItem.value.instances[activeInstanceIndex.value] || selectedItem.value;
  }
  return selectedItem.value;
});

// Dynamic counts
const totalCount = computed(() => items.value.length);
const canonicalTotalCount = computed(() => canonicalItems.value.length);
const mcpCount = computed(() => (isDeduplicated.value ? canonicalItems.value : items.value).filter((i) => i.type === 'mcp').length);
const skillCount = computed(() => (isDeduplicated.value ? canonicalItems.value : items.value).filter((i) => i.type === 'skill').length);
const customCount = computed(() => (isDeduplicated.value ? canonicalItems.value : items.value).filter((i) => (i.origin || 'custom') === 'custom').length);
const installedCount = computed(() => (isDeduplicated.value ? canonicalItems.value : items.value).filter((i) => (i.origin || 'custom') === 'installed').length);
const builtinCount = computed(() => (isDeduplicated.value ? canonicalItems.value : items.value).filter((i) => (i.origin || 'custom') === 'builtin').length);

const classifiedCount = computed(() => (isDeduplicated.value ? canonicalItems.value : items.value).filter((i) => i.isClassified).length);
const unclassifiedCount = computed(() => (isDeduplicated.value ? canonicalItems.value : items.value).filter((i) => !i.isClassified).length);

const availableProviders = computed(() => {
  const counts: Record<string, number> = {};
  for (const item of items.value) {
    const p = item.provider.toLowerCase();
    counts[p] = (counts[p] || 0) + 1;
  }
  return Object.entries(counts).sort((a, b) => b[1] - a[1]);
});

const availableCategories = computed(() => {
  const counts: Record<string, number> = {};
  for (const item of items.value) {
    const cat = item.category || 'Unclassified';
    counts[cat] = (counts[cat] || 0) + 1;
  }
  return Object.entries(counts).sort((a, b) => b[1] - a[1]);
});

const availableSubCategories = computed(() => {
  const counts: Record<string, number> = {};
  for (const item of items.value) {
    if (selectedCategory.value !== 'all' && (item.category || 'Unclassified').toLowerCase() !== selectedCategory.value.toLowerCase()) {
      continue;
    }
    const sub = item.subCategory || 'Pending Triage';
    counts[sub] = (counts[sub] || 0) + 1;
  }
  return Object.entries(counts).sort((a, b) => b[1] - a[1]);
});

const availableFamilies = computed(() => {
  const counts: Record<string, number> = {};
  const source = isDeduplicated.value ? canonicalItems.value : items.value;
  for (const item of source) {
    if (item.family) {
      const fam = item.family.toLowerCase();
      counts[fam] = (counts[fam] || 0) + 1;
    }
  }
  return Object.entries(counts).sort((a, b) => b[1] - a[1]);
});

function setFilterFamily(familySlug: string, event?: Event) {
  if (event) {
    event.stopPropagation();
  }
  if (selectedFamily.value.toLowerCase() === familySlug.toLowerCase()) {
    selectedFamily.value = 'all';
  } else {
    selectedFamily.value = familySlug;
  }
}

// Modal Family Helpers
const currentItemFamily = computed(() => {
  if (!selectedItem.value) return '';
  return selectedItem.value.family || '';
});

const familyGroupItems = computed(() => {
  const fam = currentItemFamily.value.toLowerCase();
  if (!fam) return [];
  const source = isDeduplicated.value ? canonicalItems.value : items.value;
  return source.filter((i) => (i.family || '').toLowerCase() === fam);
});

const familyParentItem = computed<Item | null>(() => {
  const fam = currentItemFamily.value.toLowerCase();
  if (!fam) return null;
  return familyGroupItems.value.find((i) => i.isParent || i.name.toLowerCase() === fam) || null;
});

const familyChildItems = computed<Item[]>(() => {
  const fam = currentItemFamily.value.toLowerCase();
  if (!fam) return [];
  return familyGroupItems.value.filter((i) => !i.isParent && i.name.toLowerCase() !== fam);
});

const familySiblingItems = computed<Item[]>(() => {
  if (!selectedItem.value) return [];
  const curName = selectedItem.value.name.toLowerCase();
  return familyChildItems.value.filter((i) => i.name.toLowerCase() !== curName);
});

const isCurrentItemParent = computed(() => {
  if (!selectedItem.value) return false;
  const fam = currentItemFamily.value.toLowerCase();
  return Boolean(selectedItem.value.isParent || (fam && selectedItem.value.name.toLowerCase() === fam));
});

// --- Modal and Accessibility Controls ---
function openModal(item: Item) {
  previouslyFocusedElement.value = document.activeElement as HTMLElement;
  selectedItem.value = item;
  let defaultIdx = 0;
  if (selectedProvider.value !== 'all' && item.instances) {
    const foundIdx = item.instances.findIndex((i) => i.provider.toLowerCase() === selectedProvider.value.toLowerCase());
    if (foundIdx !== -1) {
      defaultIdx = foundIdx;
    }
  }
  activeInstanceIndex.value = defaultIdx;
  activeModalTab.value = 'details';
  isModalOpen.value = true;
  document.body.style.overflow = 'hidden';

  nextTick(() => {
    const modalContainer = document.getElementById('item-detail-dialog');
    modalContainer?.focus();
  });
}

function closeModal() {
  isModalOpen.value = false;
  selectedItem.value = null;
  document.body.style.overflow = '';

  nextTick(() => {
    if (previouslyFocusedElement.value && typeof previouslyFocusedElement.value.focus === 'function') {
      previouslyFocusedElement.value.focus();
    }
  });
}

// Global Keyboard Shortcuts
function handleGlobalKeydown(e: KeyboardEvent) {
  // Esc closes modal or clears search
  if (e.key === 'Escape') {
    if (isModalOpen.value) {
      closeModal();
      return;
    }
    if (searchQuery.value) {
      searchQuery.value = '';
      return;
    }
    if (document.activeElement === searchInputRef.value) {
      searchInputRef.value?.blur();
      return;
    }
  }

  // '/' or 'Cmd+K' / 'Ctrl+K' focuses search
  const isSearchFocused = document.activeElement === searchInputRef.value;
  const isInputActive =
    document.activeElement?.tagName === 'INPUT' ||
    document.activeElement?.tagName === 'TEXTAREA';

  if (!isModalOpen.value && !isInputActive) {
    if (e.key === '/' || ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k')) {
      e.preventDefault();
      searchInputRef.value?.focus();
      searchInputRef.value?.select();
    }
  } else if (!isModalOpen.value && isSearchFocused && (e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault();
    searchInputRef.value?.blur();
  }
}

function clearAllFilters() {
  searchQuery.value = '';
  selectedType.value = 'all';
  selectedProvider.value = 'all';
  selectedOrigin.value = 'all';
  selectedClassification.value = 'all';
  selectedCategory.value = 'all';
  selectedSubCategory.value = 'all';
  selectedFamily.value = 'all';
  searchInputRef.value?.focus();
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeydown);
  fetchData();
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeydown);
  document.body.style.overflow = '';
  if (copyTimeout) clearTimeout(copyTimeout);
});
</script>

<template>
  <div class="min-h-screen bg-[#0e0f12] text-[#f4f4f6] font-sans antialiased">
    <!-- Top Navigation / Application Header -->
    <header class="border-b border-[#272a34] bg-[#16181d]/80 backdrop-blur-md sticky top-0 z-30">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between gap-4">
        <!-- Logo / App Title -->
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-md bg-[#0e0f12] border border-[#272a34] flex items-center justify-center text-cyan-400">
            <!-- Terminal / Code SVG primitive -->
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 9l3 3-3 3m5 0h3M4 5h16a1 1 0 011 1v12a1 1 0 01-1 1H4a1 1 0 01-1-1V6a1 1 0 011-1z" />
            </svg>
          </div>
          <div>
            <div class="flex items-center gap-2">
              <h1 class="font-mono text-sm font-bold tracking-tight text-[#f4f4f6] uppercase">
                Organizer
              </h1>
              <span class="text-[10px] font-mono uppercase tracking-widest px-1.5 py-0.5 rounded-sm bg-[#0e0f12] border border-[#272a34] text-[#9ba1b0]">
                v1.0
              </span>
            </div>
            <p class="text-[11px] text-[#9ba1b0] hidden sm:block">
              AI Tools & Skills Unified Registry
            </p>
          </div>
        </div>

        <!-- Connection Indicator & Fast Actions -->
        <div class="flex items-center gap-3">
          <!-- Backend Status Pill -->
          <div
            class="flex items-center gap-2 px-2.5 py-1 rounded-md text-xs font-mono border"
            :class="isConnected ? 'bg-emerald-950/20 border-emerald-500/30 text-emerald-300' : 'bg-amber-950/20 border-amber-500/30 text-amber-300'"
          >
            <span
              class="w-2 h-2 rounded-full"
              :class="isConnected ? 'bg-emerald-400 animate-pulse' : 'bg-amber-400'"
              aria-hidden="true"
            />
            <span class="text-[11px]">
              {{ isConnected ? 'API Connected' : isUsingSampleData ? 'Preview Mode' : 'Offline' }}
            </span>
          </div>

          <!-- Re-scan Button -->
          <button
            type="button"
            @click="handleRefresh"
            :disabled="isRefreshing"
            class="flex items-center gap-2 px-3 py-1.5 rounded-md text-xs font-mono font-medium bg-[#1f2229] hover:bg-[#272a34] text-[#f4f4f6] border border-[#343846] transition-colors disabled:opacity-50 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-cyan-400"
            title="Re-scan local ecosystem for MCPs and Skills"
          >
            <!-- Refresh SVG icon -->
            <svg
              class="w-3.5 h-3.5"
              :class="{ 'animate-spin': isRefreshing }"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span>{{ isRefreshing ? 'Scanning...' : 'Re-scan' }}</span>
          </button>
        </div>
      </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
      <!-- Backend Offline Notice Banner -->
      <section
        v-if="backendError"
        class="border border-amber-500/30 bg-amber-950/10 rounded-lg p-4 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 text-xs"
        role="alert"
      >
        <div class="flex items-start gap-3">
          <div class="p-1 rounded bg-amber-500/10 text-amber-400 shrink-0 mt-0.5">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
          </div>
          <div>
            <p class="font-medium text-[#f4f4f6]">
              Backend scanner service is offline at <code class="font-mono text-amber-300">http://localhost:8080</code>
            </p>
            <p class="text-[#9ba1b0] mt-0.5">
              To connect live data: open a terminal, navigate to <code class="font-mono text-[#f4f4f6]">backend/</code>, and run <code class="font-mono text-cyan-300 bg-[#0e0f12] px-1 py-0.5 rounded">go run main.go</code>. Sample catalog is currently loaded for preview.
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="fetchData(false)"
          class="px-3 py-1.5 rounded-md font-mono bg-amber-500/20 hover:bg-amber-500/30 text-amber-200 border border-amber-500/30 whitespace-nowrap transition-colors"
        >
          Retry Connection
        </button>
      </section>

      <!-- SECTION A: Bento Stats Header -->
      <section aria-labelledby="bento-stats-title">
        <h2 id="bento-stats-title" class="sr-only">Registry Statistics</h2>
        
        <!-- Bento Grid Metric Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
          <!-- Metric 1: Total Artifacts -->
          <div class="bg-[#16181d] border border-[#272a34] rounded-lg p-5 flex flex-col justify-between">
            <div class="flex items-center justify-between text-[#9ba1b0]">
              <span class="text-xs font-mono uppercase tracking-wider font-semibold">Total Artifacts</span>
              <svg class="w-4 h-4 text-[#9ba1b0]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <div class="mt-4">
              <div class="text-3xl font-mono font-bold text-[#f4f4f6]">
                {{ isDeduplicated ? canonicalTotalCount : totalCount }}
              </div>
              <p class="text-xs text-[#9ba1b0] mt-1">
                {{ isDeduplicated ? `${totalCount} instances across ${availableProviders.length} clients` : 'Scanned across all AI client configs' }}
              </p>
            </div>
          </div>

          <!-- Metric 2: MCP Servers -->
          <div
            @click="selectedType = selectedType === 'mcp' ? 'all' : 'mcp'"
            class="bg-[#16181d] border rounded-lg p-5 flex flex-col justify-between cursor-pointer transition-all duration-150"
            :class="selectedType === 'mcp' ? 'border-cyan-500 bg-cyan-950/10' : 'border-[#272a34] hover:border-[#343846]'"
            role="button"
            tabindex="0"
            @keydown.enter="selectedType = selectedType === 'mcp' ? 'all' : 'mcp'"
            :aria-pressed="selectedType === 'mcp'"
          >
            <div class="flex items-center justify-between text-cyan-400">
              <span class="text-xs font-mono uppercase tracking-wider font-semibold">MCP Servers</span>
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
              </svg>
            </div>
            <div class="mt-4">
              <div class="text-3xl font-mono font-bold text-cyan-300">
                {{ mcpCount }}
              </div>
              <p class="text-xs text-[#9ba1b0] mt-1">
                Active protocol servers & transports
              </p>
            </div>
          </div>

          <!-- Metric 3: Skills -->
          <div
            @click="selectedType = selectedType === 'skill' ? 'all' : 'skill'"
            class="bg-[#16181d] border rounded-lg p-5 flex flex-col justify-between cursor-pointer transition-all duration-150"
            :class="selectedType === 'skill' ? 'border-emerald-500 bg-emerald-950/10' : 'border-[#272a34] hover:border-[#343846]'"
            role="button"
            tabindex="0"
            @keydown.enter="selectedType = selectedType === 'skill' ? 'all' : 'skill'"
            :aria-pressed="selectedType === 'skill'"
          >
            <div class="flex items-center justify-between text-emerald-400">
              <span class="text-xs font-mono uppercase tracking-wider font-semibold">Agent Skills</span>
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div class="mt-4">
              <div class="text-3xl font-mono font-bold text-emerald-300">
                {{ skillCount }}
              </div>
              <p class="text-xs text-[#9ba1b0] mt-1">
                Operational prompts & domain skills
              </p>
            </div>
          </div>

          <!-- Metric 4: Ecosystem Breakdown Summary -->
          <div class="bg-[#16181d] border border-[#272a34] rounded-lg p-5 flex flex-col justify-between">
            <div class="flex items-center justify-between text-[#9ba1b0]">
              <span class="text-xs font-mono uppercase tracking-wider font-semibold">Ecosystems</span>
              <svg class="w-4 h-4 text-[#9ba1b0]" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div class="mt-4">
              <div class="text-3xl font-mono font-bold text-[#f4f4f6]">
                {{ availableProviders.length }}
              </div>
              <p class="text-xs text-[#9ba1b0] mt-1">
                Detected provider integrations
              </p>
            </div>
          </div>

          <!-- Metric 5: Classification Status -->
          <div
            @click="selectedClassification = selectedClassification === 'unclassified' ? 'all' : 'unclassified'"
            class="bg-[#16181d] border rounded-lg p-5 flex flex-col justify-between cursor-pointer transition-all duration-150"
            :class="
              selectedClassification === 'unclassified'
                ? 'border-amber-500 bg-amber-950/20'
                : selectedClassification === 'classified'
                ? 'border-emerald-500 bg-emerald-950/20'
                : 'border-[#272a34] hover:border-[#343846]'
            "
            role="button"
            tabindex="0"
            @keydown.enter="selectedClassification = selectedClassification === 'unclassified' ? 'all' : 'unclassified'"
            :aria-pressed="selectedClassification !== 'all'"
            title="Click to filter unclassified items pending triage"
          >
            <div class="flex items-center justify-between text-[#9ba1b0]">
              <span class="text-xs font-mono uppercase tracking-wider font-semibold">Taxonomy</span>
              <svg class="w-4 h-4" :class="unclassifiedCount > 0 ? 'text-amber-400' : 'text-emerald-400'" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
              </svg>
            </div>
            <div class="mt-4">
              <div class="flex items-baseline gap-1.5">
                <span class="text-3xl font-mono font-bold text-emerald-300">{{ classifiedCount }}</span>
                <span class="text-xs font-mono text-[#656c7d]">/ {{ totalCount }}</span>
              </div>
              <div class="mt-2">
                <div
                  v-if="unclassifiedCount > 0"
                  class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-mono border border-amber-500/40 bg-amber-950/40 text-amber-300"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse" />
                  <span>{{ unclassifiedCount }} pending triage</span>
                </div>
                <div
                  v-else
                  class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-mono border border-emerald-500/40 bg-emerald-950/40 text-emerald-300"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                  <span>100% Classified</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Provider Breakdown Pills / Click-to-filter -->
        <div class="mt-4 bg-[#16181d] border border-[#272a34] rounded-lg p-3">
          <div class="flex flex-wrap items-center gap-2">
            <span class="text-[11px] font-mono text-[#9ba1b0] uppercase tracking-wider mr-1">
              Providers:
            </span>

            <!-- All Providers Chip -->
            <button
              type="button"
              @click="selectedProvider = 'all'"
              class="px-2.5 py-1 rounded-md text-xs font-mono border transition-colors flex items-center gap-1.5"
              :class="
                selectedProvider === 'all'
                  ? 'bg-[#272a34] border-[#474c5e] text-[#f4f4f6] font-semibold'
                  : 'bg-[#0e0f12] border-[#272a34] text-[#9ba1b0] hover:text-[#f4f4f6] hover:border-[#343846]'
              "
            >
              <span>All</span>
              <span class="text-[10px] px-1 py-0.2 rounded bg-[#16181d] border border-[#272a34] text-[#9ba1b0]">
                {{ totalCount }}
              </span>
            </button>

            <!-- Provider Chips -->
            <button
              v-for="[provKey, count] in availableProviders"
              :key="provKey"
              type="button"
              @click="selectedProvider = selectedProvider === provKey ? 'all' : provKey"
              class="px-2.5 py-1 rounded-md text-xs font-mono border transition-colors flex items-center gap-1.5"
              :class="
                selectedProvider === provKey
                  ? `${getProviderStyle(provKey).badgeClass} font-semibold ring-1 ring-white/20`
                  : 'bg-[#0e0f12] border-[#272a34] text-[#9ba1b0] hover:text-[#f4f4f6] hover:border-[#343846]'
              "
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :class="getProviderStyle(provKey).dotClass"
                aria-hidden="true"
              />
              <span>{{ getProviderStyle(provKey).label }}</span>
              <span class="text-[10px] px-1 py-0.2 rounded bg-[#16181d] border border-[#272a34] text-[#9ba1b0]">
                {{ count }}
              </span>
            </button>
          </div>
        </div>

      </section>


      <!-- SECTION B: Control Bar -->
      <section aria-labelledby="control-bar-title" class="space-y-4">
        <h2 id="control-bar-title" class="sr-only">Catalog Controls and Search</h2>

        <div class="space-y-3">
          <!-- Instant Search Input (Full Width & Prominent) -->
          <div class="relative w-full">
            <div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-[#9ba1b0]">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <input
              ref="searchInputRef"
              v-model="searchQuery"
              type="text"
              placeholder="Search by name, command, path, suite, category... (Press '/' to focus)"
              class="w-full bg-[#16181d] border border-[#272a34] focus:border-[#474c5e] text-[#f4f4f6] placeholder-[#656c7d] text-sm font-mono pl-10 pr-24 py-3 rounded-lg focus:outline-none transition-colors"
            />
            <div class="absolute inset-y-0 right-0 pr-3 flex items-center gap-1.5">
              <button
                v-if="searchQuery"
                type="button"
                @click="searchQuery = ''; searchInputRef?.focus()"
                class="p-1 text-[#9ba1b0] hover:text-[#f4f4f6] rounded"
                title="Clear search"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
              <kbd class="hidden sm:inline-block px-2 py-0.5 text-[11px] font-mono text-[#9ba1b0] bg-[#0e0f12] border border-[#272a34] rounded">
                /
              </kbd>
            </div>
          </div>

          <!-- Type, Status, Origin and Category Filter Toolbar -->
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex flex-wrap items-center gap-2">
            <!-- Classification Status Filter Tabs -->
            <div class="bg-[#16181d] border border-[#272a34] p-1 rounded-lg flex items-center gap-1">
              <button
                type="button"
                @click="selectedClassification = 'all'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors"
                :class="
                  selectedClassification === 'all'
                    ? 'bg-[#272a34] text-[#f4f4f6] font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                All Status
              </button>
              <button
                type="button"
                @click="selectedClassification = 'classified'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center gap-1.5"
                :class="
                  selectedClassification === 'classified'
                    ? 'bg-emerald-950/40 border border-emerald-500/40 text-emerald-300 font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400" aria-hidden="true" />
                Clasificados ({{ classifiedCount }})
              </button>
              <button
                type="button"
                @click="selectedClassification = 'unclassified'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center gap-1.5"
                :class="
                  selectedClassification === 'unclassified'
                    ? 'bg-amber-950/40 border border-amber-500/40 text-amber-300 font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :class="unclassifiedCount > 0 ? 'bg-amber-400 animate-pulse' : 'bg-zinc-500'"
                  aria-hidden="true"
                />
                Sin clasificar ({{ unclassifiedCount }})
              </button>
            </div>

            <!-- Type Filter Tabs -->
            <div class="bg-[#16181d] border border-[#272a34] p-1 rounded-lg flex items-center gap-1">
              <button
                type="button"
                @click="selectedType = 'all'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors"
                :class="
                  selectedType === 'all'
                    ? 'bg-[#272a34] text-[#f4f4f6] font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                All ({{ totalCount }})
              </button>
              <button
                type="button"
                @click="selectedType = 'mcp'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center gap-1.5"
                :class="
                  selectedType === 'mcp'
                    ? 'bg-cyan-950/40 border border-cyan-500/40 text-cyan-300 font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                <span class="w-1.5 h-1.5 rounded-full bg-cyan-400" aria-hidden="true" />
                MCPs ({{ mcpCount }})
              </button>
              <button
                type="button"
                @click="selectedType = 'skill'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center gap-1.5"
                :class="
                  selectedType === 'skill'
                    ? 'bg-emerald-950/40 border border-emerald-500/40 text-emerald-300 font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400" aria-hidden="true" />
                Skills ({{ skillCount }})
              </button>
            </div>

            <!-- Origin Filter Tabs -->
            <div class="bg-[#16181d] border border-[#272a34] p-1 rounded-lg flex items-center gap-1">
              <button
                type="button"
                @click="selectedOrigin = 'all'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors"
                :class="
                  selectedOrigin === 'all'
                    ? 'bg-[#272a34] text-[#f4f4f6] font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                All
              </button>
              <button
                type="button"
                @click="selectedOrigin = 'custom'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center gap-1.5"
                :class="
                  selectedOrigin === 'custom'
                    ? 'bg-purple-950/40 border border-purple-500/40 text-purple-300 font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                <span class="w-1.5 h-1.5 rounded-full bg-purple-400" aria-hidden="true" />
                My Skills
              </button>
              <button
                type="button"
                @click="selectedOrigin = 'installed'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center gap-1.5"
                :class="
                  selectedOrigin === 'installed'
                    ? 'bg-blue-950/40 border border-blue-500/40 text-blue-300 font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                <span class="w-1.5 h-1.5 rounded-full bg-blue-400" aria-hidden="true" />
                Installed
              </button>
              <button
                type="button"
                @click="selectedOrigin = 'builtin'"
                class="px-2.5 py-1.5 rounded-md text-xs font-mono transition-colors flex items-center gap-1.5"
                :class="
                  selectedOrigin === 'builtin'
                    ? 'bg-amber-950/40 border border-amber-500/40 text-amber-300 font-semibold'
                    : 'text-[#9ba1b0] hover:text-[#f4f4f6]'
                "
              >
                <span class="w-1.5 h-1.5 rounded-full bg-amber-400" aria-hidden="true" />
                Built-in
              </button>
            </div>

            <!-- Category & SubCategory Selects -->
            <div class="flex items-center gap-1.5 bg-[#16181d] border border-[#272a34] p-1 rounded-lg">
              <select
                v-model="selectedCategory"
                @change="selectedSubCategory = 'all'"
                class="bg-[#0e0f12] border border-[#272a34] rounded px-2 py-1 text-xs font-mono text-[#f4f4f6] focus:outline-none focus:border-cyan-400"
                aria-label="Filter by Category"
              >
                <option value="all">Category: All</option>
                <option v-for="[cat, count] in availableCategories" :key="cat" :value="cat">
                  {{ cat }} ({{ count }})
                </option>
              </select>

              <select
                v-if="selectedCategory !== 'all' && availableSubCategories.length > 0"
                v-model="selectedSubCategory"
                class="bg-[#0e0f12] border border-[#272a34] rounded px-2 py-1 text-xs font-mono text-[#f4f4f6] focus:outline-none focus:border-cyan-400"
                aria-label="Filter by SubCategory"
              >
                <option value="all">Sub: All</option>
                <option v-for="[sub, count] in availableSubCategories" :key="sub" :value="sub">
                  {{ sub }} ({{ count }})
                </option>
              </select>
            </div>

            <!-- Family Filter Select -->
            <div class="flex items-center gap-1.5 bg-[#16181d] border border-[#272a34] p-1 rounded-lg">
              <select
                v-model="selectedFamily"
                class="bg-[#0e0f12] border border-[#272a34] rounded px-2 py-1 text-xs font-mono text-[#f4f4f6] focus:outline-none focus:border-cyan-400"
                aria-label="Filter by Family"
              >
                <option value="all">Pack: All</option>
                <option v-for="[fam, count] in availableFamilies" :key="fam" :value="fam">
                  pack: {{ fam }} ({{ count }})
                </option>
              </select>
            </div>

            <!-- Active Family Quick-Filter Pill -->
            <div
              v-if="selectedFamily !== 'all'"
              class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-mono bg-cyan-950/40 border border-cyan-500/40 text-cyan-300"
            >
              <span class="text-[#656c7d]">Pack:</span>
              <span class="font-bold">{{ selectedFamily }}</span>
              <button
                type="button"
                @click="selectedFamily = 'all'"
                class="ml-1 p-0.5 text-cyan-400 hover:text-white rounded hover:bg-cyan-900/40"
                title="Clear family filter"
              >
                <svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- Deduplication Toggle Button -->
            <button
              type="button"
              @click="isDeduplicated = !isDeduplicated"
              class="px-2.5 py-1.5 rounded-lg text-xs font-mono transition-colors flex items-center gap-1.5 border"
              :class="
                isDeduplicated
                  ? 'bg-cyan-950/40 border-cyan-500/40 text-cyan-300 font-semibold'
                  : 'bg-[#16181d] border-[#272a34] text-[#9ba1b0] hover:text-[#f4f4f6] hover:border-[#343846]'
              "
              title="Group identical tools that exist across multiple AI clients"
            >
              <span
                class="w-1.5 h-1.5 rounded-full shrink-0"
                :class="isDeduplicated ? 'bg-cyan-400 animate-pulse' : 'bg-[#656c7d]'"
                aria-hidden="true"
              />
              <span>{{ isDeduplicated ? 'Deduplicado (ON)' : 'Todas las instancias' }}</span>
            </button>

            <!-- Clear filters button if any filter is active -->
            <button
              v-if="hasActiveFilters"
              type="button"
              @click="clearAllFilters"
              class="px-2.5 py-1.5 text-xs font-mono font-medium text-amber-300 hover:text-amber-200 bg-amber-950/40 hover:bg-amber-900/50 border border-amber-500/40 hover:border-amber-400 rounded-lg transition-all flex items-center gap-1.5 shadow-sm"
              title="Clear all filters and search"
            >
              <svg class="w-3.5 h-3.5 text-amber-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
              <span>Clear filters</span>
            </button>
          </div>
        </div>

        <!-- Subcategory Pills & Breadcrumb (when Category is chosen) -->
        <div
          v-if="selectedCategory !== 'all'"
          class="flex flex-wrap items-center gap-2 p-2.5 rounded-lg bg-[#14161b] border border-[#272a34] text-xs font-mono"
        >
          <!-- Category Breadcrumb -->
          <div class="flex items-center gap-1.5 pr-2.5 border-r border-[#272a34]">
            <span class="text-[#656c7d] uppercase tracking-wider text-[10px]">Category:</span>
            <span class="font-bold text-[#f4f4f6]">{{ selectedCategory }}</span>
            <button
              type="button"
              @click="selectedCategory = 'all'; selectedSubCategory = 'all'"
              class="p-0.5 text-[#9ba1b0] hover:text-white rounded hover:bg-[#272a34]"
              title="Clear category filter"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Subcategory Pills -->
          <span class="text-[10px] text-[#656c7d] uppercase tracking-wider">Subcategories:</span>
          <button
            type="button"
            @click="selectedSubCategory = 'all'"
            class="px-2.5 py-1 rounded-md text-xs font-mono border transition-colors"
            :class="
              selectedSubCategory === 'all'
                ? 'bg-[#272a34] text-[#f4f4f6] border-[#474c5e] font-semibold'
                : 'bg-[#0e0f12] text-[#9ba1b0] border-[#272a34] hover:text-[#f4f4f6] hover:border-[#343846]'
            "
          >
            All
          </button>
          <button
            v-for="[sub, count] in availableSubCategories"
            :key="sub"
            type="button"
            @click="selectedSubCategory = selectedSubCategory === sub ? 'all' : sub"
            class="px-2.5 py-1 rounded-md text-xs font-mono border transition-colors flex items-center gap-1.5"
            :class="
              selectedSubCategory === sub
                ? 'bg-cyan-950/40 text-cyan-300 border-cyan-500/40 font-semibold'
                : 'bg-[#0e0f12] text-[#9ba1b0] border-[#272a34] hover:text-[#f4f4f6] hover:border-[#343846]'
            "
          >
            <span>{{ sub }}</span>
            <span class="text-[10px] px-1 rounded bg-[#16181d] border border-[#272a34] text-[#656c7d]">
              {{ count }}
            </span>
          </button>
        </div>
      </div>

        <!-- Filter summary bar -->
        <div class="flex flex-wrap items-center justify-between gap-2 text-xs text-[#9ba1b0] font-mono px-1">
          <div class="flex flex-wrap items-center gap-2">
            <div>
              Showing <span class="text-[#f4f4f6] font-bold">{{ displayItems.length }}</span> {{ isDeduplicated ? 'unique tools' : 'instances' }} (of {{ totalCount }} total)
              <span v-if="searchQuery">
                for "<span class="text-[#f4f4f6]">{{ searchQuery }}</span>"
              </span>
            </div>

            <!-- Quick clear badge when filters are active -->
            <button
              v-if="hasActiveFilters"
              type="button"
              @click="clearAllFilters"
              class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] font-mono font-medium text-amber-300 bg-amber-950/40 border border-amber-500/40 hover:bg-amber-900/50 hover:border-amber-400 transition-colors cursor-pointer"
              title="Reset all filters"
            >
              <svg class="w-2.5 h-2.5 text-amber-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
              <span>Clear filters</span>
            </button>
          </div>
          <div class="text-[11px] text-[#656c7d]">
            Click a card to view details and configuration
          </div>
        </div>

      </section>

      <!-- SECTION C: Catalog Grid (Bento Layout) -->
      <section aria-labelledby="catalog-grid-title">
        <h2 id="catalog-grid-title" class="sr-only">Catalog Items</h2>

        <!-- Loading State Skeleton -->
        <div v-if="isLoading" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="i in 6"
            :key="i"
            class="bg-[#16181d] border border-[#272a34] rounded-lg p-5 animate-pulse space-y-4"
          >
            <div class="flex items-center justify-between">
              <div class="h-4 w-16 bg-[#272a34] rounded" />
              <div class="h-4 w-12 bg-[#272a34] rounded" />
            </div>
            <div class="h-5 w-3/4 bg-[#272a34] rounded" />
            <div class="h-10 w-full bg-[#272a34] rounded" />
            <div class="h-8 w-full bg-[#0e0f12] rounded" />
          </div>
        </div>

        <!-- Empty State -->
        <div
          v-else-if="displayItems.length === 0"
          class="bg-[#16181d] border border-[#272a34] rounded-lg p-12 text-center"
        >
          <div class="w-12 h-12 rounded-lg bg-[#0e0f12] border border-[#272a34] flex items-center justify-center mx-auto text-[#9ba1b0] mb-4">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
          <h3 class="text-sm font-mono font-bold text-[#f4f4f6]">
            No matching tools or skills found
          </h3>
          <p class="text-xs text-[#9ba1b0] max-w-md mx-auto mt-1 mb-6">
            We couldn't find any items matching your current search query or active filter settings.
          </p>
          <button
            type="button"
            @click="clearAllFilters"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg text-xs font-mono font-semibold bg-amber-950/40 hover:bg-amber-900/50 border border-amber-500/40 hover:border-amber-400 text-amber-300 transition-colors shadow-sm cursor-pointer"
          >
            <svg class="w-3.5 h-3.5 text-amber-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
            <span>Clear all filters and search</span>
          </button>
        </div>

        <!-- Grid of Bento Cards, grouped by install pack -->
        <div v-else class="space-y-8">
          <section
            v-for="group in catalogGroups"
            :key="group.family || '__ungrouped'"
            class="space-y-3"
          >
            <header
              v-if="group.family || catalogGroups.length > 1"
              class="flex flex-wrap items-center justify-between gap-2 px-1"
            >
              <div class="flex items-center gap-2 min-w-0">
                <h3 class="text-sm font-mono font-bold text-[#f4f4f6] truncate">
                  {{ group.parent?.name || group.family || 'Ungrouped' }}
                </h3>
                <span
                  v-if="group.family"
                  class="font-mono text-[10px] px-1.5 py-0.5 rounded-sm border border-cyan-500/30 bg-cyan-950/30 text-cyan-300"
                >
                  {{ group.items.length }} skills
                </span>
                <span
                  v-else
                  class="font-mono text-[10px] px-1.5 py-0.5 rounded-sm border border-[#272a34] text-[#9ba1b0]"
                >
                  {{ group.items.length }}
                </span>
              </div>
              <button
                v-if="group.family && selectedFamily === 'all'"
                type="button"
                class="text-[11px] font-mono text-cyan-300 hover:text-white border border-cyan-500/30 hover:border-cyan-400 rounded px-2 py-0.5"
                @click="setFilterFamily(group.family)"
              >
                View pack
              </button>
            </header>
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <article
            v-for="item in group.items"
            :key="item.id"
            @click="openModal(item)"
            @keydown.enter="openModal(item)"
            @keydown.space.prevent="openModal(item)"
            tabindex="0"
            role="button"
            :aria-label="`View details for ${item.name}`"
            class="bg-[#16181d] border border-[#272a34] hover:border-[#474c5e] rounded-lg p-5 flex flex-col justify-between transition-all duration-150 cursor-pointer group focus-visible:outline-none focus-visible:border-cyan-400"
            :class="getProviderStyle(item.provider).cardBorderClass"
          >
            <div>
              <!-- Top Row: Badges & Provider -->
              <div class="flex items-center justify-between gap-2 mb-3">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <!-- Type Badge -->
                  <span
                    v-if="item.type === 'mcp'"
                    class="font-mono text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-sm border border-cyan-500/30 bg-cyan-500/10 text-cyan-300"
                  >
                    MCP
                  </span>
                  <span
                    v-else
                    class="font-mono text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-sm border border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
                  >
                    SKILL
                  </span>

                  <!-- Provider Tag(s) -->
                  <template v-if="item.providers && item.providers.length > 1">
                    <span
                      v-for="prov in item.providers"
                      :key="prov"
                      class="font-mono text-[9px] px-1.5 py-0.5 rounded-sm border flex items-center gap-1"
                      :class="getProviderStyle(prov).badgeClass"
                      :title="`Available on ${getProviderStyle(prov).label}`"
                    >
                      <span class="w-1.5 h-1.5 rounded-full" :class="getProviderStyle(prov).dotClass" aria-hidden="true" />
                      {{ getProviderStyle(prov).label }}
                    </span>
                    <span class="font-mono text-[9px] px-1.5 py-0.5 rounded-sm bg-cyan-950/40 border border-cyan-500/40 text-cyan-300 font-bold" title="This tool is configured on multiple AI clients">
                      {{ item.providers.length }} clients
                    </span>
                  </template>
                  <span
                    v-else
                    class="font-mono text-[10px] px-2 py-0.5 rounded-sm border flex items-center gap-1"
                    :class="getProviderStyle(item.provider).badgeClass"
                  >
                    <span
                      class="w-1.5 h-1.5 rounded-full"
                      :class="getProviderStyle(item.provider).dotClass"
                      aria-hidden="true"
                    />
                    {{ getProviderStyle(item.provider).label }}
                  </span>


                  <!-- Origin Badge -->
                  <span
                    class="font-mono text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded-sm border flex items-center gap-1"
                    :class="getOriginStyle(item.origin).cardBadgeClass"
                  >
                    <span
                      class="w-1 h-1 rounded-full"
                      :class="getOriginStyle(item.origin).dotClass"
                      aria-hidden="true"
                    />
                    {{ (item.origin || 'custom') === 'builtin' ? 'BUILT-IN' : (item.origin || 'custom').toUpperCase() }}
                  </span>

                  <!-- skills.sh Registry Link Badge -->
                  <a
                    v-if="item.registryUrl"
                    :href="item.registryUrl"
                    target="_blank"
                    rel="noopener noreferrer"
                    @click.stop
                    class="font-mono text-[9px] font-semibold px-1.5 py-0.5 rounded-sm border border-violet-500/40 bg-violet-950/40 text-violet-300 hover:bg-violet-900/60 hover:border-violet-400/80 flex items-center gap-1 transition-colors shrink-0"
                    :title="`View on skills.sh: ${item.sourcePkg || item.name}`"
                  >
                    <span class="w-1 h-1 rounded-full bg-violet-400" aria-hidden="true" />
                    <span>skills.sh</span>
                    <svg class="w-2.5 h-2.5 text-violet-400 opacity-70" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                    </svg>
                  </a>

                  <!-- Category & SubCategory Tag (Only if classified) -->
                  <span
                    v-if="item.isClassified"
                    class="font-mono text-[9px] px-1.5 py-0.5 rounded-sm border flex items-center gap-1 truncate max-w-[180px]"
                    :class="getCategoryStyle(item.category).cardBadgeClass"
                    :title="`${item.category} / ${item.subCategory || 'General'}`"
                  >
                    <span
                      class="w-1 h-1 rounded-full shrink-0"
                      :class="getCategoryStyle(item.category).dotClass"
                      aria-hidden="true"
                    />
                    <span class="truncate">{{ item.category }}</span>
                  </span>

                  <!-- Unclassified Warning Badge (Only if NOT classified - single badge) -->
                  <span
                    v-else
                    class="font-mono text-[9px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded-sm border border-amber-500/40 bg-amber-950/40 text-amber-300 flex items-center gap-1 shrink-0"
                    title="Pending triage / suite assignment"
                  >
                    <span class="w-1 h-1 rounded-full bg-amber-400 animate-pulse" />
                    UNCLASSIFIED
                  </span>

                  <!-- Family Parent Badge -->
                  <button
                    v-if="item.isParent || (item.childCount && item.childCount > 0)"
                    type="button"
                    @click.stop="setFilterFamily(item.family || item.name, $event)"
                    class="font-mono text-[9px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-sm border border-cyan-500/40 bg-cyan-950/40 text-cyan-300 hover:bg-cyan-900/60 hover:border-cyan-400 transition-colors flex items-center gap-1 shrink-0"
                    :title="`Filter pack '${item.family || item.name}' (${item.childCount} sub-skills)`"
                  >
                    <svg class="w-2.5 h-2.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                    <span>{{ item.childCount }} sub-skills</span>
                  </button>

                  <!-- Family Child Badge -->
                  <button
                    v-else-if="item.family"
                    type="button"
                    @click.stop="setFilterFamily(item.family, $event)"
                    class="font-mono text-[9px] px-1.5 py-0.5 rounded-sm border border-[#343846] bg-[#1a1d24] text-[#9ba1b0] hover:text-[#f4f4f6] hover:border-[#474c5e] transition-colors flex items-center gap-1 shrink-0"
                    :title="`Filter by pack '${item.family}'`"
                  >
                    <span class="text-[#656c7d]">pack:</span>
                    <span class="text-[#f4f4f6] font-medium">{{ item.family }}</span>
                  </button>
                </div>


                <!-- Inspect Arrow Indicator -->
                <div class="text-[#656c7d] group-hover:text-[#f4f4f6] transition-colors">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                </div>
              </div>

              <!-- Item Name -->
              <h3 class="font-mono text-base font-bold text-[#f4f4f6] group-hover:text-white truncate mb-1">
                {{ item.name }}
              </h3>

              <!-- Source Package / Author Tag -->
              <div v-if="item.sourcePkg" class="font-mono text-[10px] text-violet-300/80 truncate mb-2 flex items-center gap-1">
                <span class="text-[#656c7d]">pkg:</span>
                <span class="hover:underline">{{ item.sourcePkg }}</span>
              </div>

              <!-- Description (Clamped) -->
              <p class="text-xs text-[#9ba1b0] line-clamp-2 leading-relaxed mb-4 min-h-[2.5rem]">
                {{ item.description || 'No description provided.' }}
              </p>
            </div>

            <div class="space-y-3">
              <!-- Quick Invocation Snippet with Copy Button -->
              <div
                class="bg-[#0e0f12] border border-[#272a34] rounded-md px-2.5 py-1.5 flex items-center justify-between gap-2 overflow-hidden"
              >
                <div class="font-mono text-[11px] text-[#9ba1b0] truncate select-all">
                  {{ item.invocation || item.command || 'No invocation snippet' }}
                </div>
                <button
                  type="button"
                  @click.stop="copyText(item.invocation || item.command || '', item.id, $event)"
                  class="p-1 rounded hover:bg-[#1f2229] text-[#9ba1b0] hover:text-[#f4f4f6] shrink-0 transition-colors"
                  :title="copiedId === item.id ? 'Copied to clipboard' : 'Copy invocation'"
                  :aria-label="copiedId === item.id ? 'Copied' : 'Copy invocation'"
                >
                  <!-- Checkmark when copied, copy icon otherwise -->
                  <svg
                    v-if="copiedId === item.id"
                    class="w-3.5 h-3.5 text-emerald-400"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  <svg
                    v-else
                    class="w-3.5 h-3.5"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                  </svg>
                </button>
              </div>

              <!-- Footer: File path / Command Summary -->
              <div class="pt-2 border-t border-[#272a34]/60 flex items-center justify-between text-[11px] font-mono text-[#656c7d]">
                <div class="flex items-center gap-1.5 truncate max-w-[85%]">
                  <!-- Folder icon -->
                  <svg class="w-3 h-3 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                  </svg>
                  <span v-if="item.instances && item.instances.length > 1" class="truncate text-cyan-400/90 font-medium">
                    {{ item.instances.length }} configured locations
                  </span>
                  <span v-else class="truncate">{{ item.sourcePath }}</span>
                </div>
                <span class="text-[10px] text-[#9ba1b0] uppercase tracking-wider">Inspect</span>
              </div>
            </div>
          </article>
          </div>
          </section>
        </div>
      </section>
    </main>

    <!-- SECTION D: Detail Modal / Drawer -->
    <div
      v-if="isModalOpen && selectedItem"
      id="modal-backdrop"
      @click="closeModal"
      class="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4 sm:p-6"
    >
      <div
        id="item-detail-dialog"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="'modal-title-' + selectedItem.id"
        tabindex="-1"
        @click.stop
        class="bg-[#16181d] border border-[#343846] rounded-lg max-w-3xl w-full max-h-[90vh] flex flex-col overflow-hidden shadow-2xl focus:outline-none"
      >
        <!-- Modal Top Bar -->
        <div class="px-6 py-4 border-b border-[#272a34] bg-[#14161b] flex items-center justify-between gap-4">
          <div class="flex flex-wrap items-center gap-2 sm:gap-3">
            <!-- Type badge -->
            <span
              v-if="currentModalInstance.type === 'mcp'"
              class="font-mono text-xs font-bold uppercase tracking-wider px-2.5 py-0.5 rounded-sm border border-cyan-500/30 bg-cyan-500/10 text-cyan-300"
            >
              MCP Server
            </span>
            <span
              v-else
              class="font-mono text-xs font-bold uppercase tracking-wider px-2.5 py-0.5 rounded-sm border border-emerald-500/30 bg-emerald-500/10 text-emerald-300"
            >
              Agent Skill
            </span>

            <!-- Provider Badge -->
            <span
              class="font-mono text-xs px-2.5 py-0.5 rounded-sm border flex items-center gap-1.5"
              :class="getProviderStyle(currentModalInstance.provider).badgeClass"
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :class="getProviderStyle(currentModalInstance.provider).dotClass"
                aria-hidden="true"
              />
              {{ getProviderStyle(currentModalInstance.provider).label }}
            </span>

            <!-- Multi-client badge -->
            <span
              v-if="selectedItem.instances && selectedItem.instances.length > 1"
              class="font-mono text-xs px-2 py-0.5 rounded-sm border border-cyan-500/40 bg-cyan-950/40 text-cyan-300 font-semibold"
            >
              {{ selectedItem.instances.length }} clients
            </span>

            <!-- Origin Badge -->
            <span
              class="font-mono text-xs px-2.5 py-0.5 rounded-sm border flex items-center gap-1.5"
              :class="getOriginStyle(currentModalInstance.origin).badgeClass"
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :class="getOriginStyle(currentModalInstance.origin).dotClass"
                aria-hidden="true"
              />
              {{ getOriginStyle(currentModalInstance.origin).label }}
            </span>

            <!-- Category Badge -->
            <span
              class="font-mono text-xs px-2.5 py-0.5 rounded-sm border flex items-center gap-1.5"
              :class="getCategoryStyle(currentModalInstance.category).badgeClass"
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :class="getCategoryStyle(currentModalInstance.category).dotClass"
                aria-hidden="true"
              />
              {{ currentModalInstance.category || 'Unclassified' }}
            </span>
          </div>

          <!-- Close Button -->
          <button
            type="button"
            @click="closeModal"
            class="p-1.5 rounded-md text-[#9ba1b0] hover:text-[#f4f4f6] hover:bg-[#272a34] transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-cyan-400"
            title="Close dialog (Esc)"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Modal Title & Tab Switcher -->
        <div class="px-6 pt-5 pb-3 border-b border-[#272a34] flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h3 :id="'modal-title-' + currentModalInstance.id" class="text-xl font-mono font-bold text-[#f4f4f6]">
              {{ selectedItem.name }}
            </h3>
            <p class="text-xs text-[#9ba1b0] mt-0.5">
              ID: <code class="font-mono text-[#f4f4f6]">{{ currentModalInstance.id }}</code>
            </p>
          </div>

          <!-- Inspector Tabs -->
          <div class="bg-[#0e0f12] border border-[#272a34] p-1 rounded-md flex items-center gap-1 self-start sm:self-auto">
            <button
              type="button"
              @click="activeModalTab = 'details'"
              class="px-3 py-1 rounded text-xs font-mono transition-colors"
              :class="activeModalTab === 'details' ? 'bg-[#272a34] text-[#f4f4f6] font-semibold' : 'text-[#9ba1b0] hover:text-[#f4f4f6]'"
            >
              Inspection
            </button>
            <button
              type="button"
              @click="activeModalTab = 'raw'"
              class="px-3 py-1 rounded text-xs font-mono transition-colors"
              :class="activeModalTab === 'raw' ? 'bg-[#272a34] text-[#f4f4f6] font-semibold' : 'text-[#9ba1b0] hover:text-[#f4f4f6]'"
            >
              Raw Config / JSON
            </button>
          </div>
        </div>

        <!-- Multi-Client Instance Switcher Bar (only when present in >1 provider) -->
        <div
          v-if="selectedItem.instances && selectedItem.instances.length > 1"
          class="px-6 py-2.5 bg-[#121419] border-b border-[#272a34] flex flex-wrap items-center gap-2"
        >
          <span class="text-xs font-mono text-[#9ba1b0] flex items-center gap-1.5 mr-1">
            <svg class="w-3.5 h-3.5 text-cyan-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            <span>Configured on {{ selectedItem.instances.length }} clients:</span>
          </span>
          <button
            v-for="(inst, idx) in selectedItem.instances"
            :key="inst.id"
            type="button"
            @click="activeInstanceIndex = idx"
            class="px-2.5 py-1 rounded-md text-xs font-mono border transition-all flex items-center gap-1.5"
            :class="
              activeInstanceIndex === idx
                ? 'bg-[#272a34] text-white border-cyan-500/60 shadow-sm font-semibold ring-1 ring-cyan-400/40'
                : 'bg-[#16181d] border-[#272a34] text-[#9ba1b0] hover:text-[#f4f4f6] hover:border-[#343846]'
            "
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="getProviderStyle(inst.provider).dotClass" />
            <span>{{ getProviderStyle(inst.provider).label }}</span>
            <span v-if="inst.origin" class="text-[9px] uppercase tracking-wider text-[#656c7d]">({{ inst.origin }})</span>
          </button>
        </div>

        <!-- Modal Scrollable Content -->
        <div class="p-6 overflow-y-auto space-y-6 flex-1">
          <!-- VIEW 1: Inspection Details -->
          <div v-if="activeModalTab === 'details'" class="space-y-6">
            <!-- Invocation Section -->
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-xs font-mono font-semibold uppercase tracking-wider text-[#9ba1b0]">
                  How to invoke / execute
                </span>
                <button
                  type="button"
                  @click="copyText(currentModalInstance.invocation || currentModalInstance.command || '', 'modal-invoc')"
                  class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs font-mono bg-[#272a34] hover:bg-[#343846] text-[#f4f4f6] transition-colors"
                >
                  <svg
                    v-if="copiedId === 'modal-invoc'"
                    class="w-3.5 h-3.5 text-emerald-400"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  <svg
                    v-else
                    class="w-3.5 h-3.5"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    viewBox="0 0 24 24"
                    aria-hidden="true"
                  >
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
                  </svg>
                  <span>{{ copiedId === 'modal-invoc' ? 'Copied' : 'Copy Invocation' }}</span>
                </button>
              </div>
              <div class="bg-[#0e0f12] border border-[#272a34] rounded-lg p-3 font-mono text-xs text-[#f4f4f6] whitespace-pre-wrap break-all select-all">
                {{ currentModalInstance.invocation || currentModalInstance.command || 'No invocation defined' }}
              </div>
            </div>

            <!-- Description -->
            <div class="space-y-2">
              <span class="text-xs font-mono font-semibold uppercase tracking-wider text-[#9ba1b0]">
                Description
              </span>
              <div class="bg-[#14161b] border border-[#272a34] rounded-lg p-4 text-xs text-[#f4f4f6] leading-relaxed">
                {{ currentModalInstance.description || selectedItem.description || 'No description available for this item.' }}
              </div>
            </div>

            <!-- Family & Sub-skills Section -->
            <div
              v-if="currentItemFamily"
              class="bg-[#14161b] border border-[#272a34] rounded-lg p-4 space-y-3"
            >
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-[#272a34] pb-2.5">
                <div class="flex items-center gap-2">
                  <div class="w-5 h-5 rounded bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
                    <svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                  </div>
                  <span class="text-xs font-mono font-bold uppercase tracking-wider text-[#f4f4f6]">
                    Family: {{ currentItemFamily }}
                  </span>
                  <span
                    v-if="isCurrentItemParent"
                    class="font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-sm bg-cyan-950/40 border border-cyan-500/40 text-cyan-300 font-semibold"
                  >
                    Parent Skill ({{ familyChildItems.length }} sub-skills)
                  </span>
                  <span
                    v-else
                    class="font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-sm bg-[#1f2229] border border-[#343846] text-[#9ba1b0]"
                  >
                    Sub-Skill
                  </span>
                </div>

                <!-- Quick filter button -->
                <button
                  type="button"
                  @click="selectedFamily = currentItemFamily; closeModal()"
                  class="text-[11px] font-mono text-cyan-400 hover:text-cyan-300 hover:underline flex items-center gap-1 self-start sm:self-auto"
                >
                  <span>Filter catalog by this family</span>
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
                  </svg>
                </button>
              </div>

              <!-- If Parent Skill: list child skills -->
              <div v-if="isCurrentItemParent" class="space-y-2">
                <div class="text-[11px] font-mono text-[#9ba1b0]">
                  Sub-skills in this family ({{ familyChildItems.length }}):
                </div>
                <div v-if="familyChildItems.length > 0" class="flex flex-wrap gap-2 max-h-48 overflow-y-auto p-1">
                  <button
                    v-for="child in familyChildItems"
                    :key="child.id"
                    type="button"
                    @click="openModal(child)"
                    class="px-2.5 py-1.5 rounded-md text-xs font-mono bg-[#0e0f12] border border-[#272a34] hover:border-cyan-500/50 text-[#f4f4f6] hover:text-cyan-300 transition-all flex items-center gap-1.5"
                    :title="`Open ${child.name}`"
                  >
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                    <span>{{ child.name }}</span>
                  </button>
                </div>
                <div v-else class="text-xs font-mono text-[#656c7d] italic">
                  No sub-skills registered yet in this family.
                </div>
              </div>

              <!-- If Child Skill: show parent skill (if exists) and sibling skills -->
              <div v-else class="space-y-3">
                <!-- Parent Skill row -->
                <div v-if="familyParentItem" class="space-y-1.5">
                  <div class="text-[11px] font-mono text-[#9ba1b0]">
                    Parent Skill:
                  </div>
                  <button
                    type="button"
                    @click="openModal(familyParentItem)"
                    class="px-3 py-1.5 rounded-md text-xs font-mono bg-[#0e0f12] border border-cyan-500/40 hover:border-cyan-400 text-cyan-300 transition-all flex items-center gap-2"
                    :title="`Open parent skill: ${familyParentItem.name}`"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 10l7-7m0 0l7 7m-7-7v18" />
                    </svg>
                    <span class="font-bold">{{ familyParentItem.name }}</span>
                    <span class="text-[10px] text-[#9ba1b0]">({{ familyParentItem.childCount }} sub-skills)</span>
                  </button>
                </div>

                <!-- Sibling Skills row -->
                <div v-if="familySiblingItems.length > 0" class="space-y-1.5">
                  <div class="text-[11px] font-mono text-[#9ba1b0]">
                    Sibling Skills in '{{ currentItemFamily }}' ({{ familySiblingItems.length }}):
                  </div>
                  <div class="flex flex-wrap gap-2 max-h-40 overflow-y-auto p-2 bg-[#0e0f12] rounded-md border border-[#272a34]">
                    <button
                      v-for="sibling in familySiblingItems"
                      :key="sibling.id"
                      type="button"
                      @click="openModal(sibling)"
                      class="px-2.5 py-1 rounded text-xs font-mono bg-[#16181d] border border-[#272a34] hover:border-[#474c5e] text-[#9ba1b0] hover:text-[#f4f4f6] transition-colors flex items-center gap-1.5"
                      :title="`Switch to ${sibling.name}`"
                    >
                      <span class="w-1.5 h-1.5 rounded-full bg-[#656c7d]" />
                      <span>{{ sibling.name }}</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Metadata Bento Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <!-- Taxonomy Classification Card -->
              <div class="bg-[#14161b] border border-[#272a34] rounded-lg p-3 space-y-2 sm:col-span-2">
                <div class="flex items-center justify-between text-[#9ba1b0]">
                  <span class="text-[11px] font-mono uppercase tracking-wider">Taxonomy Classification</span>
                  <span
                    class="px-2 py-0.5 rounded text-[10px] font-mono font-bold border flex items-center gap-1.5"
                    :class="
                      currentModalInstance.isClassified
                        ? 'bg-emerald-950/20 border-emerald-500/30 text-emerald-300'
                        : 'bg-amber-950/20 border-amber-500/30 text-amber-300'
                    "
                  >
                    <span
                      class="w-1.5 h-1.5 rounded-full"
                      :class="currentModalInstance.isClassified ? 'bg-emerald-400' : 'bg-amber-400 animate-pulse'"
                      aria-hidden="true"
                    />
                    {{ currentModalInstance.isClassified ? 'CLASSIFIED' : 'PENDING TRIAGE' }}
                  </span>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <div class="flex items-center gap-1.5 text-xs font-mono text-[#f4f4f6]">
                    <span class="text-[#9ba1b0]">Category:</span>
                    <span
                      class="px-2 py-0.5 rounded border text-[11px] font-medium flex items-center gap-1"
                      :class="getCategoryStyle(currentModalInstance.category).badgeClass"
                    >
                      <span class="w-1.5 h-1.5 rounded-full" :class="getCategoryStyle(currentModalInstance.category).dotClass" />
                      {{ currentModalInstance.category || 'Unclassified' }}
                    </span>
                  </div>
                  <span class="text-[#656c7d]">/</span>
                  <div class="flex items-center gap-1.5 text-xs font-mono text-[#f4f4f6]">
                    <span class="text-[#9ba1b0]">SubCategory:</span>
                    <span class="px-2 py-0.5 rounded bg-[#0e0f12] border border-[#272a34] text-[11px] text-[#f4f4f6]">
                      {{ currentModalInstance.subCategory || 'Pending Triage' }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- Source Path Card -->
              <div class="bg-[#14161b] border border-[#272a34] rounded-lg p-3 space-y-1">
                <div class="flex items-center justify-between text-[#9ba1b0]">
                  <span class="text-[11px] font-mono uppercase tracking-wider">Source Config Path</span>
                  <button
                    type="button"
                    @click="copyText(currentModalInstance.sourcePath, 'modal-path')"
                    class="text-[10px] font-mono text-cyan-400 hover:underline"
                  >
                    {{ copiedId === 'modal-path' ? 'Copied' : 'Copy' }}
                  </button>
                </div>
                <div class="font-mono text-xs text-[#f4f4f6] break-all select-all">
                  {{ currentModalInstance.sourcePath }}
                </div>
              </div>

              <!-- Provider Card -->
              <div class="bg-[#14161b] border border-[#272a34] rounded-lg p-3 space-y-1">
                <span class="text-[11px] font-mono uppercase tracking-wider text-[#9ba1b0]">
                  AI Ecosystem
                </span>
                <div class="flex items-center gap-2 text-xs font-mono text-[#f4f4f6]">
                  <span
                    class="w-2 h-2 rounded-full"
                    :class="getProviderStyle(currentModalInstance.provider).dotClass"
                    aria-hidden="true"
                  />
                  <span>{{ getProviderStyle(currentModalInstance.provider).label }}</span>
                </div>
              </div>

              <!-- Origin Card -->
              <div class="bg-[#14161b] border border-[#272a34] rounded-lg p-3 space-y-1">
                <span class="text-[11px] font-mono uppercase tracking-wider text-[#9ba1b0]">
                  Origin
                </span>
                <div class="flex items-center gap-2 text-xs font-mono text-[#f4f4f6]">
                  <span
                    class="px-2 py-0.5 rounded-sm border font-bold text-[10px] uppercase"
                    :class="getOriginStyle(currentModalInstance.origin).badgeClass"
                  >
                    {{ (currentModalInstance.origin || 'custom') === 'builtin' ? 'BUILT-IN' : (currentModalInstance.origin || 'custom').toUpperCase() }}
                  </span>
                  <span class="text-xs text-[#9ba1b0]">
                    {{ (currentModalInstance.origin || 'custom') === 'builtin' ? 'Built-in' : (currentModalInstance.origin || 'custom') === 'installed' ? 'Installed' : 'My Skills (Custom)' }}
                  </span>
                </div>
              </div>

              <!-- Public Registry Card (skills.sh) -->
              <div
                v-if="currentModalInstance.registryUrl"
                class="bg-[#14161b] border border-violet-500/40 rounded-lg p-3 space-y-2 sm:col-span-2"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <span class="w-2 h-2 rounded-full bg-violet-400" />
                    <span class="text-[11px] font-mono font-bold uppercase tracking-wider text-violet-300">
                      Public Registry (skills.sh)
                    </span>
                    <span
                      v-if="currentModalInstance.sourcePkg"
                      class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-violet-950/60 border border-violet-500/40 text-violet-200"
                    >
                      {{ currentModalInstance.sourcePkg }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2">
                    <button
                      type="button"
                      @click="copyText(currentModalInstance.registryUrl, 'modal-registry-url')"
                      class="text-[10px] font-mono text-violet-400 hover:underline"
                    >
                      {{ copiedId === 'modal-registry-url' ? 'Copied' : 'Copy URL' }}
                    </button>
                    <a
                      :href="currentModalInstance.registryUrl"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="text-[10px] font-mono text-violet-300 hover:text-white hover:underline inline-flex items-center gap-1 bg-violet-900/40 border border-violet-500/40 px-2 py-0.5 rounded-sm"
                    >
                      <span>Open on skills.sh</span>
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                      </svg>
                    </a>
                  </div>
                </div>
                <div class="font-mono text-xs text-violet-200 break-all select-all">
                  <a
                    :href="currentModalInstance.registryUrl"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="hover:underline text-violet-300"
                  >
                    {{ currentModalInstance.registryUrl }}
                  </a>
                </div>
              </div>

              <!-- Repository / Source URL Card (if present) -->
              <div
                v-if="currentModalInstance.sourceUrl"
                class="bg-[#14161b] border border-[#272a34] rounded-lg p-3 space-y-1 sm:col-span-2"
              >
                <div class="flex items-center justify-between text-[#9ba1b0]">
                  <span class="text-[11px] font-mono uppercase tracking-wider">Repository / Source URL</span>
                  <div class="flex items-center gap-2">
                    <button
                      type="button"
                      @click="copyText(currentModalInstance.sourceUrl, 'modal-source-url')"
                      class="text-[10px] font-mono text-cyan-400 hover:underline"
                    >
                      {{ copiedId === 'modal-source-url' ? 'Copied' : 'Copy URL' }}
                    </button>
                    <a
                      :href="currentModalInstance.sourceUrl"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="text-[10px] font-mono text-cyan-400 hover:underline inline-flex items-center gap-1"
                    >
                      <span>Open Link</span>
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                      </svg>
                    </a>
                  </div>
                </div>
                <div class="font-mono text-xs text-cyan-300 break-all select-all">
                  <a
                    :href="currentModalInstance.sourceUrl"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="hover:underline"
                  >
                    {{ currentModalInstance.sourceUrl }}
                  </a>
                </div>
              </div>
            </div>

            <!-- MCP SPECIFIC SECTION -->
            <div v-if="currentModalInstance.type === 'mcp'" class="space-y-4 pt-2 border-t border-[#272a34]">
              <h4 class="text-xs font-mono font-bold uppercase tracking-wider text-cyan-400">
                MCP Protocol Configuration
              </h4>

              <!-- Command & Arguments -->
              <div class="space-y-2">
                <span class="text-[11px] font-mono text-[#9ba1b0] uppercase tracking-wider">
                  Command & Arguments
                </span>
                <div class="bg-[#0e0f12] border border-[#272a34] rounded-lg p-3 space-y-2 font-mono text-xs">
                  <div class="flex items-start gap-2">
                    <span class="text-[#656c7d] select-none">$</span>
                    <span class="text-cyan-300 font-semibold">{{ currentModalInstance.command || 'N/A' }}</span>
                    <span v-if="currentModalInstance.args && currentModalInstance.args.length" class="text-[#f4f4f6]">
                      {{ currentModalInstance.args.join(' ') }}
                    </span>
                  </div>
                </div>
              </div>

              <!-- URL / Transport if available -->
              <div v-if="currentModalInstance.url" class="space-y-2">
                <span class="text-[11px] font-mono text-[#9ba1b0] uppercase tracking-wider">
                  Transport URL
                </span>
                <div class="bg-[#0e0f12] border border-[#272a34] rounded-lg p-3 font-mono text-xs text-[#f4f4f6]">
                  {{ currentModalInstance.url }}
                </div>
              </div>

              <!-- Environment Variable Keys (With Security Notice) -->
              <div class="space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-mono text-[#9ba1b0] uppercase tracking-wider">
                    Environment Variables Required
                  </span>
                  <span class="text-[11px] font-mono text-[#656c7d]">
                    {{ currentModalInstance.envKeys?.length || 0 }} keys
                  </span>
                </div>

                <div v-if="currentModalInstance.envKeys && currentModalInstance.envKeys.length > 0" class="flex flex-wrap gap-2">
                  <span
                    v-for="key in currentModalInstance.envKeys"
                    :key="key"
                    class="px-2.5 py-1 rounded bg-[#0e0f12] border border-[#272a34] font-mono text-xs text-cyan-300"
                  >
                    {{ key }}
                  </span>
                </div>
                <div v-else class="text-xs font-mono text-[#656c7d] italic">
                  No environment variables declared.
                </div>

                <!-- Security notice -->
                <div class="flex items-start gap-2 p-2.5 rounded bg-[#0e0f12] border border-[#272a34] text-[11px] text-[#9ba1b0]">
                  <svg class="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                  </svg>
                  <span>
                    <strong class="text-[#f4f4f6] font-medium">Security Protected:</strong> Secret values are permanently omitted by the backend scanner. Only environment variable keys are cataloged.
                  </span>
                </div>
              </div>
            </div>

            <!-- SKILL SPECIFIC SECTION -->
            <div v-if="currentModalInstance.type === 'skill'" class="space-y-4 pt-2 border-t border-[#272a34]">
              <h4 class="text-xs font-mono font-bold uppercase tracking-wider text-emerald-400">
                Skill Specification & Content
              </h4>

              <div class="space-y-2">
                <span class="text-[11px] font-mono text-[#9ba1b0] uppercase tracking-wider">
                  SKILL.md Definition
                </span>
                <div class="bg-[#0e0f12] border border-[#272a34] rounded-lg p-4 font-mono text-xs text-[#f4f4f6] whitespace-pre-wrap overflow-x-auto leading-relaxed max-h-64">
                  {{ currentModalInstance.rawConfig || currentModalInstance.description }}
                </div>
              </div>
            </div>
          </div>

          <!-- VIEW 2: Raw JSON Config -->
          <div v-else class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-xs font-mono font-semibold uppercase tracking-wider text-[#9ba1b0]">
                Raw Item Payload ({{ getProviderStyle(currentModalInstance.provider).label }})
              </span>
              <button
                type="button"
                @click="copyText(JSON.stringify(currentModalInstance, null, 2), 'modal-raw')"
                class="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs font-mono bg-[#272a34] hover:bg-[#343846] text-[#f4f4f6] transition-colors"
              >
                <span>{{ copiedId === 'modal-raw' ? 'Copied' : 'Copy JSON' }}</span>
              </button>
            </div>
            <pre class="bg-[#0e0f12] border border-[#272a34] rounded-lg p-4 font-mono text-xs text-[#f4f4f6] overflow-x-auto leading-relaxed max-h-[50vh] select-all">{{ JSON.stringify(currentModalInstance, null, 2) }}</pre>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="px-6 py-3 border-t border-[#272a34] bg-[#14161b] flex items-center justify-between text-xs font-mono text-[#9ba1b0]">
          <span>Press <kbd class="px-1 py-0.5 rounded bg-[#0e0f12] border border-[#272a34] text-[#f4f4f6]">Esc</kbd> to close</span>
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-1.5 rounded-md bg-[#272a34] hover:bg-[#343846] text-[#f4f4f6] transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
