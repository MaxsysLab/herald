# Configuration

One business is one JSON file. It holds the business's facts, its knowledge, which
providers to use, and how strict to be. Herald writes it with permissions `0600`.

You can edit it in the console (`/admin`, "Configuration") or in a text editor. Saving
from the console takes effect at once. After editing the file by hand, restart Herald.

## The file

This is the sample Herald creates, shortened. Key names are shown exactly as the
program reads and writes them today: the `profile` and `llm` sections use capitalized
keys, and `knowledge` and `embed` use lowercase ones.

```json
{
  "name": "sample",
  "profile": {
    "Name": "Bloom Café",
    "Tagline": "Neighborhood coffee, done right",
    "Industry": "coffee shop",
    "Description": "",
    "OfficerName": "Herald",
    "BrandVoice": "warm, concise, welcoming",
    "Services": [
      { "Name": "Cappuccino", "Description": "", "Price": { "Amount": "4.50", "Currency": "USD" }, "Aliases": ["cap"] },
      { "Name": "Catering", "Description": "Event catering", "Price": null, "Aliases": null }
    ],
    "Hours": [
      { "Day": "Mon-Fri", "Open": "07:00", "Close": "18:00", "Note": "" }
    ],
    "Policies": [
      { "Topic": "refund", "Text": "We happily remake or refund any drink you're not satisfied with, no receipt needed." }
    ],
    "Contacts": [
      { "Kind": "escalation", "Value": "manager@bloomcafe.example", "Label": "Café manager" }
    ]
  },
  "knowledge": [
    { "id": "wifi", "source": "kb:amenities.md", "text": "free guest wifi is available; the network is BloomGuest and the password is on your receipt" }
  ],
  "llm": { "Provider": "groq", "BaseURL": "", "Model": "", "APIKey": "" },
  "embed": { "provider": "fake", "base_url": "", "model": "", "api_key": "" },
  "min_score": 0.25
}
```

## `profile`: the facts of record

| Field | What it is |
|---|---|
| `Name` | The business's name. Used in the greeting, in handoffs and in refusals |
| `OfficerName` | The name the officer introduces itself by. Empty means "Herald" |
| `BrandVoice` | Tone guidance, given to the LLM when it drafts an answer |
| `Tagline`, `Industry`, `Description` | Kept with the profile. Not used in answers yet |
| `Services` | What the business sells. `Price.Amount` is an exact string, never a rounded number. A service with `"Price": null` means "not listed": a price question about it goes to a person. `Aliases` are other names customers use |
| `Hours` | Rows of opening hours: `Day`, `Open`, `Close`, and an optional `Note` such as "by appointment" |
| `Policies` | A `Topic` word and the policy `Text`. The text is given word for word |
| `Contacts` | Ways to reach the business: `Kind` is `phone`, `email`, `address` or `escalation` |

**The escalation contact** is where a customer is sent when the officer cannot answer.
Herald uses the first contact of kind `escalation`; if there is none, the first phone;
if none, the first email. This value is included in chat replies, so treat it as public.

## `knowledge`: what the officer may quote

Each passage has a `text`, a `source` label that is shown as the citation, and an
optional `id`. Keep passages short and about one thing: one FAQ answer, one paragraph
of a policy. The console's "Add knowledge" box appends to this list.

## `llm`: natural drafting (optional)

| `Provider` | Default endpoint | Default model | Key from |
|---|---|---|---|
| `groq` (default) | `https://api.groq.com/openai/v1` | `llama-3.3-70b-versatile` | `GROQ_API_KEY` |
| `openai` | `https://api.openai.com/v1` | `gpt-4o-mini` | `OPENAI_API_KEY` |
| `grok` | `https://api.x.ai/v1` | `grok-2-latest` | `XAI_API_KEY` |
| `ollama` | `http://localhost:11434/v1` | `llama3.1` | none |
| `custom` | set `BaseURL` | set `Model` | `APIKey` |

Set `BaseURL` or `Model` to override a default. Any endpoint that speaks the OpenAI
chat-completions format works as `custom`. Claude is not wired yet.

When a keyed provider has no key, Herald does not fail: it answers by quoting the
best-matching passage word for word.

**Keys belong in the environment, not in this file.** The `APIKey` field exists, but
leave it empty and set the provider's environment variable instead, so that a copy of
the configuration never carries a secret.

## `embed`: how knowledge is matched

| `provider` | Default model | Key from |
|---|---|---|
| `fake` (default) | none: an offline word-matching stand-in | none |
| `openai` | `text-embedding-3-small` | `OPENAI_API_KEY` |
| `voyage` | `voyage-3-lite` | `VOYAGE_API_KEY` |
| `custom` | set `base_url` and `model` | `api_key` |

The default needs no key and nothing leaves the machine, but it matches on exact words
only, so it hands off more often than it needs to. For a live site use a real embedder.

## `min_score`: how strict to be

A knowledge passage is used only if its match score is at least `min_score`. Below it,
the question goes to a person. `0` means the built-in default of `0.35`. The sample
uses `0.25` because the offline stand-in scores low.

Raise it and the officer hands off more. Lower it and the officer answers more, with
more risk of an answer that does not fit the question. Tune it by asking real customer
questions in the console and watching the score beside each citation.

## How a question is answered

In this order:

1. **Screen.** An empty message, or one that looks like an attempt to override the
   officer's instructions, is declined.
2. **Price.** If the question contains a price word (price, cost, how much, charge,
   fee, rate), Herald looks for a service name or alias in it. A listed price is given
   exactly. Anything else goes to a person, with the reason `price_not_listed`.
3. **Hours.** If the question contains hour, open, close or "when are you", the hours
   are given exactly as configured.
4. **Policy.** If the question contains a policy's topic word, that policy's text is
   given word for word.
5. **Knowledge.** Otherwise Herald retrieves the four best-matching passages. If the
   best one clears `min_score`, the answer is drafted from the passages that clear it
   (or the best passage is quoted, when no LLM is set). If none clears it, the question
   goes to a person, with the reason `ungrounded`.

Steps 2 to 4 match plain words inside the question today. That is simple and
predictable, and it also means a word such as "celebrate" contains "rate". Smarter
matching is an open area for contributors.
