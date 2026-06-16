---
name: condense
description: Cut text down to size — make writing concise, tighten convoluted sentences, and remove filler without losing meaning. Use when asked to shorten, condense, tighten, trim, compress, or make text more concise; to cut word count; or to simplify wordy, bloated, or rambling prose (docs, messages, emails, comments, LLM output). Triggers like "make this shorter", "tighten this up", "too wordy", "cut the fluff", "be more concise".
---
# Condense

Single goal: **reduce text size while keeping the meaning intact.** Make prose concise, break up convoluted sentences, and strip filler. This is about length and clarity — not tone or style (use `unslop` for AI tropes).

## Workflow

1. Read the text and identify its core claims — the load-bearing words. Everything else is a candidate for cutting.
2. Apply the cuts in [references/patterns.md](references/patterns.md), in roughly this order:
   - Delete words that carry no information (filler, hedges, redundant intensifiers).
   - Replace wordy phrases with shorter equivalents (see the substitution table).
   - Convert passive to active and noun-heavy phrases to verbs.
   - Split or shorten convoluted sentences; cut sentences that repeat a prior point.
3. Re-read the result. Confirm every original claim survives. Restore anything whose removal changed the meaning.
4. Report the result. By default, return only the rewritten text. If the user asks, show before/after and the word-count reduction.

## Rules

- **Preserve meaning first.** Never drop a fact, caveat, number, or qualifier that changes what the text asserts. When unsure whether a word is load-bearing, keep it.
- **Don't add.** No new ideas, no editorializing, no "polishing" that grows the text.
- **Match the register.** Keep the author's voice and formality; you're trimming, not rewriting personality.
- **Preserve structure** the user relies on — code blocks, links, lists, headings, defined terms — unless asked to flatten them.
- **Know when to stop.** Concise ≠ terse-to-the-point-of-cryptic. Stop when further cuts would lose information or make it hard to read.

## Quick heuristics

- If a sentence runs past ~25 words or has 3+ clauses, split it.
- If a word can be deleted and the sentence still means the same thing, delete it.
- Prefer the shorter of two synonyms ("use" over "utilize", "to" over "in order to").
- Lead with the point; cut throat-clearing intros ("It's worth noting that…", "As we can see…").
- One idea per sentence. One point per paragraph.

Read [references/patterns.md](references/patterns.md) before condensing.
