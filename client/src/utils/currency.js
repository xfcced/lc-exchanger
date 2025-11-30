export function normalizeRates(payload, currencyOptions) {
  if (!payload || typeof payload !== 'object') return null
  const containers = [payload?.rates, payload?.usd, payload?.data?.rates, payload?.data?.usd, payload].filter(Boolean)

  for (const container of containers) {
    const extracted = {}
    // USD is always 1 for conversion
    for (const { code } of currencyOptions) {
      if (code === 'USD') {
        extracted[code] = 1
      } else {
        const value = findRate(container, code)
        if (typeof value === 'number' && Number.isFinite(value)) {
          extracted[code] = value
        }
      }
    }
    if (Object.keys(extracted).length === currencyOptions.length) {
      return {
        rates: extracted,
        updatedAt: payload.updatedAt ?? payload.timestamp ?? payload.lastUpdated ?? null,
      }
    }
  }
  return null
}

export function findRate(container, code) {
  const lower = code.toLowerCase()
  const candidates = [
    container?.[code],
    container?.[code.toUpperCase()],
    container?.[lower],
    container?.[`usdTo${code}`],
    container?.[`USDTo${code}`],
    container?.[`usd_${lower}`],
    container?.[`usd-${lower}`],
    container?.[`USD_${code}`],
    container?.[`USD-${code}`],
  ]
  for (const value of candidates) {
    const numeric = Number(value)
    if (!Number.isNaN(numeric)) {
      return numeric
    }
  }
  return undefined
}

export function formatAmount(value) {
  if (!Number.isFinite(value)) return ''
  const rounded = Math.round(value * 10000) / 10000
  return rounded.toString()
}

export function formatUsdDisplay(code, usdRates) {
  const rate = usdRates[code]
  if (code === 'USD') return '1 USD'
  return rate ? `${formatAmount(rate)} ${code}` : 'N/A'
}

export function formatUsdInverse(code, usdRates) {
  const rate = usdRates[code]
  if (code === 'USD') return '1 USD'
  if (!rate) return 'N/A'
  return `${formatAmount(1 / rate)} USD`
}

export function formatTimestamp(value) {
  const date = typeof value === 'number' ? new Date(value) : new Date(String(value))
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString()
}
