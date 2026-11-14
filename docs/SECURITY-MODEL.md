# Security model

What Herald protects, what it does not protect yet, and what to do about it. This is
written plainly so that nobody deploys Herald believing it does more than it does.

To report a vulnerability that is not listed here, follow [SECURITY.md](../SECURITY.md).

## What Herald is built to protect

**The truth of what a business tells its customers.** That is the design's centre:

- Prices, hours and policies are looked up from the business's own profile and given
  exactly. A language model never paraphrases them.
- Any other answer must be backed by a knowledge passage that scores above the
  grounding threshold. If none does, the question is handed to a person.
- When a language model drafts an answer, it is told to use only the retrieved
  passages and to offer a person otherwise.
- Tests hold these behaviours in place: an unlisted price is never invented, an
  unsupported question is handed off, and an exact fact is answered without calling
  the model at all.

Grounding reduces wrong answers. It does not make them impossible. A model can still
misread a passage, which is why the threshold is conservative and handoff is the
default when in doubt.

## Known limits today

| Limit | What it means | What to do |
|---|---|---|
| **The console has no sign-in** | Anyone who can reach `/admin`, `/admin/config` or `/admin/knowledge` can change the profile and the knowledge, and so what customers are told | Bind Herald to `127.0.0.1`, expose only the public routes through a proxy, and reach the console over an SSH tunnel. See [DEPLOYMENT.md](DEPLOYMENT.md) |
| **By default the server listens on every interface** | `herald serve` without `-addr` listens on port 8080 on all network interfaces, console included | Always pass `-addr 127.0.0.1:8080` |
| **The chat API accepts any origin** | Any website can call `/api/chat` on your Herald | Acceptable for a public chat; add rate limiting at the proxy |
| **No rate limit** | A script can send many questions; with an LLM key set, each grounded one costs a model call | Limit requests at the proxy |
| **The input screen is a simple filter** | It declines a short list of known override phrases. In the code's own words it is "a seam, not a security boundary" | Do not rely on it. The grounding gate is what keeps answers honest |
| **Knowledge is trusted** | Whatever is in the knowledge base can be quoted to customers, and is given to the model as context | Only add text you would say to a customer yourself. This is one more reason to keep the console closed |
| **The escalation contact is public** | It is included in every handoff reply | Use a contact you are content to publish |
| **Keys can be written into the configuration** | The `APIKey` and `api_key` fields exist | Leave them empty and use environment variables |
| **No HTTPS in the binary** | Herald speaks plain HTTP | Terminate HTTPS at the proxy |
| **No conversation log** | Herald does not record questions or replies, so there is no audit trail either | If you need one, log at the proxy, and tell your visitors |

## What Herald does not do

It does not store customer conversations, set cookies, or track visitors. What it
sends to third parties, and when, is in [DATA.md](DATA.md).

## Helping close these

Sign-in for the console, an origin allow-list, and rate limiting are open areas, and
contributions are welcome. Start with an issue so the approach can be agreed first.
