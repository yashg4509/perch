export const mockNodes = [
  {
    id: 'github',
    provider: 'github',
    label: 'repo',
    url: 'org/my-app',
    status: 'source',
    meta: [
      { key: 'branch', value: 'main' },
      { key: 'commit', value: 'a3f9c12' },
      { key: 'CI', value: 'passing', semantic: 'success' },
    ],
    tabs: ['PRs', 'actions', 'commits'],
  },
  {
    id: 'vercel',
    provider: 'vercel',
    label: 'web',
    url: 'my-app.vercel.app',
    status: 'healthy',
    meta: [
      { key: 'deploy', value: 'dpl_xK92m...' },
      { key: 'region', value: 'iad1' },
      { key: 'last deploy', value: '2h ago' },
    ],
    tabs: ['logs', 'env', 'deploy'],
  },
  {
    id: 'render',
    provider: 'render',
    label: 'api',
    url: 'api.my-app.onrender.com',
    status: 'degraded',
    meta: [
      { key: 'instance', value: 'starter' },
      { key: 'region', value: 'oregon' },
      { key: 'uptime', value: '94.1%', semantic: 'warning' },
    ],
    tabs: ['logs', 'env', 'metrics'],
  },
  {
    id: 'supabase',
    provider: 'supabase',
    label: 'db',
    url: 'abcxyz.supabase.co',
    status: 'healthy',
    meta: [
      { key: 'plan', value: 'pro' },
      { key: 'region', value: 'us-east-1' },
      { key: 'db size', value: '1.2 GB' },
    ],
    tabs: ['logs', 'tables', 'metrics'],
  },
]

export const mockEdges = [
  { id: 'e1', source: 'github', target: 'vercel', animated: true },
  { id: 'e2', source: 'github', target: 'render', animated: true },
  { id: 'e3', source: 'vercel', target: 'supabase', animated: true },
  { id: 'e4', source: 'render', target: 'supabase', animated: true },
]

export const mockDeployments = {
  vercel: [
    {
      id: 'E3QPGJfsv',
      env: 'Production',
      current: true,
      branch: 'main',
      commit: 'da26937 docs(demo): langchain integration',
      date: 'Apr 10',
      duration: '31s',
    },
    {
      id: 'C4XLAWPTd',
      env: 'Preview',
      current: false,
      branch: 'mcp-server',
      commit: '9f8ebf5 fix(mcp): upstream base timeouts',
      date: 'Apr 11',
      duration: '29s',
    },
    {
      id: 'J57BKWeLv',
      env: 'Production',
      current: false,
      branch: 'main',
      commit: 'f6a4622 Saved progress at end of loop',
      date: 'Apr 9',
      duration: '33s',
    },
  ],
  github: [
    {
      id: 'run_4821',
      env: 'CI',
      current: true,
      branch: 'main',
      commit: 'da26937 docs(demo): langchain integration',
      date: 'Apr 10',
      duration: '1m 12s',
    },
    {
      id: 'run_4820',
      env: 'CI',
      current: false,
      branch: 'mcp-server',
      commit: '0b9c82b Add image for MCP server',
      date: 'Apr 12',
      duration: '58s',
    },
  ],
  render: [
    {
      id: 'dep-abc123',
      env: 'Production',
      current: true,
      branch: 'main',
      commit: 'f6a4622 Saved progress at end of loop',
      date: 'Apr 9',
      duration: '45s',
    },
    {
      id: 'dep-abc122',
      env: 'Production',
      current: false,
      branch: 'main',
      commit: 'be599bc Saved progress at end of loop',
      date: 'Apr 9',
      duration: '41s',
    },
  ],
  supabase: [
    {
      id: 'migration_012',
      env: 'Production',
      current: true,
      branch: 'main',
      commit: 'add_user_sessions table',
      date: 'Apr 10',
      duration: '4s',
    },
    {
      id: 'migration_011',
      env: 'Production',
      current: false,
      branch: 'main',
      commit: 'add indexes on events',
      date: 'Apr 8',
      duration: '2s',
    },
  ],
}

export const mockLogs = {
  vercel: [
    { ts: '18:21:01', level: 'info', msg: 'server started on :3000' },
    { ts: '18:21:03', level: 'info', msg: 'connected to db' },
    { ts: '18:21:44', level: 'info', msg: 'GET /health 200 4ms' },
    { ts: '18:23:15', level: 'warn', msg: 'upstream timeout on /ai 504' },
    { ts: '18:23:16', level: 'error', msg: 'retrying... attempt 1/3' },
  ],
  render: [
    { ts: '18:20:11', level: 'info', msg: 'worker started' },
    { ts: '18:22:30', level: 'warn', msg: 'memory usage at 87%' },
    { ts: '18:23:01', level: 'error', msg: 'health check failed: connection refused' },
  ],
  github: [
    { ts: '18:10:00', level: 'info', msg: 'workflow triggered: push to main' },
    { ts: '18:10:45', level: 'info', msg: 'all checks passed' },
  ],
  supabase: [
    { ts: '18:00:00', level: 'info', msg: 'realtime connected' },
    { ts: '18:15:22', level: 'info', msg: 'query executed in 3ms' },
  ],
}
