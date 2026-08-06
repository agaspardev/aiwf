const windows = new Map()

function pruneWindows() {
  const now = Date.now()
  for (const [key, entries] of windows) {
    const active = entries.filter(t => now - t < 60000)
    if (active.length === 0) {
      windows.delete(key)
    } else {
      windows.set(key, active)
    }
  }
}

module.exports = {
  name: "rate-limiter",
  priority: 50,
  onRequest: async (ctx) => {
    const now = Date.now()
    const maxReqs = 60
    const maxPerProv = 200

    pruneWindows()

    const modelKey = `model:${ctx.model}`
    const provKey = `provider:${ctx.provider}`

    const modelHits = (windows.get(modelKey) || []).filter(t => now - t < 60000)
    const provHits = (windows.get(provKey) || []).filter(t => now - t < 60000)

    if (modelHits.length >= maxReqs) {
      console.warn(`[rate-limiter] Blocked ${ctx.requestId}: model ${ctx.model} exceeded ${maxReqs} req/min`)
      return { blocked: true, response: { status: 429, body: { error: "rate_limit_exceeded", message: `Model ${ctx.model} rate limit exceeded. Try again later.` } } }
    }

    if (provHits.length >= maxPerProv) {
      console.warn(`[rate-limiter] Blocked ${ctx.requestId}: provider ${ctx.provider} exceeded ${maxPerProv} req/min`)
      return { blocked: true, response: { status: 429, body: { error: "provider_rate_limit", message: `Provider ${ctx.provider} rate limit exceeded. Try again later.` } } }
    }

    windows.set(modelKey, [...modelHits, now])
    windows.set(provKey, [...provHits, now])
  }
}
