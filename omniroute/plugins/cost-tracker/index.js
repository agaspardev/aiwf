const DEFAULT_RATES = {
  "claude-sonnet-4":      [3, 15],
  "claude-opus-4":        [15, 75],
  "claude-haiku-4":       [0.25, 1.25],
  "gpt-4o":               [2.5, 10],
  "gpt-4o-mini":          [0.15, 0.6],
  "gemini-2.5-pro":       [1.25, 5],
  "gemini-2.5-flash":     [0.1, 0.4],
  "deepseek-chat":        [0.27, 1.1],
}

let runningTotal = 0

function resolveRates(model) {
  const key = Object.keys(DEFAULT_RATES).find(k => model.toLowerCase().includes(k))
  return key ? DEFAULT_RATES[key] : [1, 4]
}

module.exports = {
  name: "cost-tracker",
  priority: 100,
  onResponse: async (ctx, response) => {
    let tokensIn = 0, tokensOut = 0

    if (response && typeof response === "object") {
      if (response.usage) {
        tokensIn = response.usage.prompt_tokens || response.usage.input_tokens || 0
        tokensOut = response.usage.completion_tokens || response.usage.output_tokens || 0
      }
      if (response.body && response.body.usage) {
        tokensIn = response.body.usage.prompt_tokens || response.body.usage.input_tokens || 0
        tokensOut = response.body.usage.completion_tokens || response.body.usage.output_tokens || 0
      }
    }

    if (tokensIn === 0 && tokensOut === 0) return

    const [inputRate, outputRate] = resolveRates(ctx.model)
    const costIn = (tokensIn / 1000) * inputRate
    const costOut = (tokensOut / 1000) * outputRate
    const total = costIn + costOut
    runningTotal += total

    console.log(`[cost-tracker] ${ctx.requestId} | ${ctx.model} | in=${tokensIn} out=${tokensOut} | $${total.toFixed(6)} | running: $${runningTotal.toFixed(4)}`)
  }
}
