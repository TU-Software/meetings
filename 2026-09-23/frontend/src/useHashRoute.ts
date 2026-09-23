import { useCallback, useEffect, useState } from 'react'

function getHash() {
  return window.location.hash.slice(1) || '/'
}

export function useHashRoute() {
  const [route, setRoute] = useState(getHash)

  useEffect(() => {
    const handler = () => setRoute(getHash())
    window.addEventListener('hashchange', handler)
    return () => window.removeEventListener('hashchange', handler)
  }, [])

  const navigate = useCallback((path: string) => {
    window.location.hash = path
  }, [])

  return { route, navigate }
}
