# Open questions

Decisions that are not made yet. Each is an invitation: if you have a view, open an
issue, say what you propose and why, and the maintainers will discuss it in the open.

## 1. The name

"Herald" is a working name. The design document has said so from the start
("placeholder, rename freely").

Two names are in play. The **project's** name is on the repository, the command and the
documents. The **officer's** default name is what it calls itself to customers:
"Hi, I'm Herald, the assistant for Bloom Café." Each business can already set its own
officer name in its configuration, so the question is only about the defaults.

What a good name has to do:

- carry no gender, so that it fits any business and any customer;
- be plain to the owner of a café, a clinic or a church office;
- be short enough to say;
- not already belong to a product in this field.

Another route is to give the officer no personal name by default ("Hi, I'm the
assistant for Bloom Café") and let each business choose.

To propose a name, open an issue with the name, your reason, and what you found when
you searched for other software that uses it.

## 2. Sign-in for the console

The console has no sign-in ([SECURITY-MODEL.md](SECURITY-MODEL.md)). What is the
simplest sound design that keeps the single binary and zero dependencies: one shared
password from the environment, per-user accounts, or leaving it to the proxy?

## 3. Where the knowledge index lives

Today it is in memory. The design names `sqlite-vec` as the way to keep it on disk
without a separate database service. Is that still the right choice, given the wish to
keep dependencies at zero?

## 4. The default local embedder

Which small embedding model that runs on a CPU should ship as the default, so that
good retrieval needs no API key?

## 5. Smarter matching of facts

Price, hours and policy questions are recognised by plain words inside the question
([CONFIGURATION.md](CONFIGURATION.md)). How should that improve without giving up
what makes it trustworthy: exact answers that a model never paraphrases?

## 6. The console's shape

Server-rendered pages, as now, or a small single-page application? And a per-field
editor for the profile in place of raw JSON.

## 7. Multi-step tasks

Booking and lead capture need more than one turn. What shape should that take, and
when is the first real case for it?

The last four come from the "Open decisions" and "Feature cut" sections of
[HERALD-V1-DESIGN.md](HERALD-V1-DESIGN.md).
