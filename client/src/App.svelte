<script>
  import { onMount } from 'svelte'
  import { normalizeRates, formatAmount, formatUsdDisplay, formatUsdInverse, formatTimestamp } from './utils/currency.js'

  const currencyOptions = [
    { code: 'CNY', name: 'Chinese Yuan' },
    { code: 'SEK', name: 'Swedish Krona' },
    { code: 'EUR', name: 'Euro' },
    { code: 'USD', name: 'US Dollar' },
  ]

  const defaultAmount = '100'

  let usdRates = {}
  let amounts = {
    CNY: defaultAmount,
    SEK: '',
    EUR: '',
    USD: '',
  }
  let activeCurrency = 'CNY'
  let loading = true
  let errorMessage = ''
  let lastUpdated = ''

  $: hasRates = Object.keys(usdRates).length === currencyOptions.length
  $: lastUpdatedLabel = lastUpdated ? formatTimestamp(lastUpdated) : ''

  onMount(() => {
    fetchRates()
  })

  async function fetchRates() {
    loading = true
    errorMessage = ''
    try {
      const response = await fetch('/api/all-rate', { headers: { accept: 'application/json' } })
      if (!response.ok) {
        throw new Error(`Failed to load rates (${response.status})`)
      }
      const payload = await response.json()
      const normalized = normalizeRates(payload, currencyOptions)
      if (!normalized) {
        throw new Error('The API response is missing USD -> CNY/SEK/EUR rates.')
      }
      usdRates = normalized.rates
      lastUpdated = normalized.updatedAt ?? ''

      if (!amounts[activeCurrency]) {
        amounts = { ...amounts, [activeCurrency]: defaultAmount }
      }
      convertFrom(activeCurrency, amounts[activeCurrency] || defaultAmount)
    } catch (error) {
      errorMessage = error?.message ?? 'Unable to load exchange rates.'
    } finally {
      loading = false
    }
  }

  function convertFrom(code, rawValue) {
    activeCurrency = code
    const value = rawValue ?? ''
    const next = { ...amounts, [code]: value }
    const cleaned = value.replace(/,/g, '.').trim()
    const numericValue = parseFloat(cleaned)

    if (!cleaned || Number.isNaN(numericValue) || !usdRates[code]) {
      currencyOptions.forEach(({ code: other }) => {
        if (other !== code) {
          next[other] = ''
        }
      })
      amounts = next
      return
    }

    const usdAmount = numericValue / usdRates[code]
    currencyOptions.forEach(({ code: other }) => {
      if (other === code) return
      const usdToOther = usdRates[other]
      if (!usdToOther) {
        next[other] = ''
        return
      }
      next[other] = formatAmount(usdAmount * usdToOther)
    })
    amounts = next
  }
</script>

<main class="app-shell">
  <section class="converter-card">
    {#if errorMessage}
      <div class="state state-error" role="alert">
        <p>{errorMessage}</p>
        <button on:click={fetchRates} class="secondary-btn">Try again</button>
      </div>
    {:else if !hasRates && loading}
      <div class="state state-loading" role="status">
        <p>Loading the latest exchange rates...</p>
      </div>
    {:else}
      <div class="inputs-grid">
        {#each currencyOptions as currency}
          <label class={`currency-field ${activeCurrency === currency.code ? 'active' : ''}`}>
            <div class="field-head">
              <div>
                <span class="currency-code">{currency.code}</span>
                <span class="currency-name">{currency.name}</span>
              </div>
              <span class="helper">1 {currency.code} ~ {formatUsdInverse(currency.code, usdRates)}</span>
            </div>
            <input
              type="text"
              inputmode="decimal"
              placeholder="0.00"
              value={amounts[currency.code] ?? ''}
              on:input={(event) => convertFrom(currency.code, event.currentTarget.value)}
              on:focus={() => (activeCurrency = currency.code)}
              disabled={!hasRates}
            />
          </label>
        {/each}
      </div>

      <div class="rates-grid">
        {#each currencyOptions as currency}
          <div class="rate-card">
            <p class="label">1 USD</p>
            <p class="value">{formatUsdDisplay(currency.code, usdRates)}</p>
            <p class="note">{currency.name}</p>
          </div>
        {/each}
      </div>

      <div class="meta">
        <span class="meta-pill">Active input: {activeCurrency}</span>
        {#if loading}
          <span class="meta-pill subtle">Updating rates...</span>
        {:else if lastUpdatedLabel}
          <span class="meta-pill subtle">Synced {lastUpdatedLabel}</span>
        {/if}
      </div>
    {/if}
  </section>
</main>
