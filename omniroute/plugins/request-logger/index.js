const startTimes = new Map()

module.exports = {
  name: "request-logger",
  priority: 200,
  onRequest: async (ctx) => {
    startTimes.set(ctx.requestId, Date.now())
    const bodyInfo = ctx.body && typeof ctx.body === "object" && ctx.body.messages
      ? `${ctx.body.messages.length} messages`
      : "no messages field"
    console.log(`[request-logger] >>> ${ctx.requestId} | ${ctx.model}@${ctx.provider} | ${bodyInfo}`)
  },
  onResponse: async (ctx, response) => {
    const start = startTimes.get(ctx.requestId)
    const elapsed = start ? Date.now() - start : -1
    startTimes.delete(ctx.requestId)

    let status = "ok"
    let tokensOut = 0
    if (response && typeof response === "object") {
      if (response.body && response.body.usage) {
        tokensOut = response.body.usage.completion_tokens || response.body.usage.output_tokens || 0
      }
      if (response.status && response.status >= 400) {
        status = `error_${response.status}`
      }
    }

    console.log(`[request-logger] <<< ${ctx.requestId} | ${elapsed}ms | ${tokensOut} out tokens | ${status}`)
  },
  onError: async (ctx, error) => {
    const start = startTimes.get(ctx.requestId)
    const elapsed = start ? Date.now() - start : -1
    startTimes.delete(ctx.requestId)
    console.error(`[request-logger] XXX ${ctx.requestId} | ${elapsed}ms | ERROR: ${error.message || error}`)
  }
}
