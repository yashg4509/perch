const variants = {
  vercel: {
    className: 'bg-black text-white',
    label: '▲',
  },
  supabase: {
    className: 'bg-[#3ecf8e] text-emerald-900',
    label: 'S',
  },
  github: {
    className: 'bg-[#24292e] text-white',
    label: 'G',
  },
  render: {
    className: 'bg-[#46e3b7] text-emerald-900',
    label: 'R',
  },
  custom: {
    className: 'bg-gray-200 text-gray-600',
    label: '⌘',
  },
}

export function ProviderBadge({ provider }) {
  const v = variants[provider] ?? variants.custom
  return (
    <span
      className={`inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-md text-[11px] font-semibold leading-none ${v.className}`}
      aria-hidden
    >
      {v.label}
    </span>
  )
}
