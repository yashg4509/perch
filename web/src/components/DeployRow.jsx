export function DeployRow({ deployment }) {
  const d = deployment
  return (
    <div className="border-b border-gray-200 px-3 py-2.5 last:border-b-0">
      <div className="flex items-start justify-between gap-2">
        <span className="font-mono text-[12px] text-gray-900">{d.id}</span>
        <div className="flex shrink-0 items-center gap-2">
          {d.current && (
            <span className="rounded bg-blue-50 px-1.5 py-0.5 text-[9px] font-medium text-blue-600">current</span>
          )}
          <span className="text-xs text-gray-500">{d.date}</span>
        </div>
      </div>
      <div className="mt-1 flex items-center justify-between gap-2 text-xs text-gray-500">
        <span>
          {d.env} · <span className="font-mono">⎇</span> {d.branch}
        </span>
        <span className="shrink-0 text-gray-400">{d.duration}</span>
      </div>
      <div className="mt-1.5 truncate font-mono text-xs text-gray-700">
        <span className="mr-1 text-gray-400">◎</span>
        {d.commit}
      </div>
    </div>
  )
}
