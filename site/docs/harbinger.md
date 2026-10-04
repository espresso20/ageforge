# The Harbinger

Harbingers are figures who warn you that a [catastrophe](catastrophe.md) is coming, and tell you how worried to be. From the Iron Era on, a harbinger comes only when a doom is fated to strike in your era, some while before it does. In the Cosmic Era, the last one, a second thread also warns of your next prestige: the [Last Passage](#the-last-passage). You can pay to lower the odds, pay to soften the blow, or invite the catastrophe on purpose.

A harbinger never blocks anything and never expires. You can ignore it completely and the game plays on as normal.

---

## When Harbingers Come

### A doom fated in secret

When you enter an era from the Iron Era on, the Cosmic Era included, a hidden roll decides whether a doom is fated there: 27% of the time, one is. Nothing is ever fated in the Stone Era.

A fated doom strikes at a random moment anywhere in the era: in any of its ages, early or late, mid-age included. The moment is drawn across the era's expected length, counted from when you entered it. The expected length is the target times of the era's ages added up (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)):

| Era | Expected length at 1x | Figures, in order | Can a doom be fated? |
|-----|-----------------------|-------------------|----------------------|
| Stone Era | 4.9 h | the Wild Man, the Hermit, the Soothsayer | No. Its harbingers are always [false prophets](#false-prophets). |
| Iron Era | 27.3 h | the Desert Prophet, the Oracle, the Town Crier | Yes, 27% of the time |
| Steel Era | 54.6 h | the Court Astrologer, the Pamphleteer, the Newsboy | Yes, 27% of the time |
| Electric Era | 80.6 h | the Doomsayer, the Telegraph, the Civil Defense Broadcast | Yes, 27% of the time |
| Digital Era | 109.2 h | the Evening News, the Chain Email, the Viral Video | Yes, 27% of the time |
| Neon Era | 156 h | the Ghost in the Net, the Reactor Warden, the Deep Space Monitor | Yes, 27% of the time |
| Cosmic Era | 249.6 h | the Distress Beacon, the Elder Relay, your future self, your unmade self | Yes, 27% of the time: the Reality Tear. A second thread warns of [the Last Passage](#the-last-passage). |

The fate is saved with your game, so reloading can't re-roll it.

Until a harbinger arrives, nothing in the game tells a fated era from a quiet one: not the panels, the Epoch panel, the `catastrophe` command, the status bar, the log or the map. A quiet era is safe, for now.

Every window on this page (the era's length, the harbinger's lead, what a figure means by "this age") is measured against the ages' target times, so the hours here follow the target table. They are the hours at 1x, on the frontier. On known ground each age's target is first divided by its [Era Mastery](prestige.md#era-mastery) speed, so a mastered era is shorter and its doom and its warning both fall inside it: at 4x every window here is a quarter as long. Mastery is fixed for a run, so a fate rolled when you enter an era never shifts.

### When the harbinger arrives

A harbinger comes only when a doom is fated, or as a false prophet (see [False prophets](#false-prophets)). It arrives a random lead before the strike: 20% to 60% of the target time of the age you are in, drawn once per fate.

| Era | Lead in its first age | Lead in its last age |
|-----|-----------------------|----------------------|
| Iron Era | 1.3 to 3.9 hours (Iron Age) | 2.3 to 7 hours (Medieval Age) |
| Steel Era | 3.1 to 9.4 hours (Renaissance Age) | 4.2 to 12.5 hours (Industrial Age) |
| Electric Era | 4.7 to 14 hours (Victorian Age) | 6.2 to 18.7 hours (Atomic Age) |
| Digital Era | 6.2 to 18.7 hours (Modern Age) | 8.3 to 25 hours (Digital Age) |
| Neon Era | 9.4 to 28.1 hours (Cyberpunk Age) | 11.4 to 34.3 hours (Space Age) |
| Cosmic Era | 12.5 to 37.4 hours (Interstellar Age) | 12.5 to 37.4 hours (Transcendent Age) |

Rules:

- **The lead follows the age you are in.** Advancing into a longer age lengthens it, so a harbinger can arrive the moment you advance.
- **Never before the era began.** If the lead would reach back past the moment you entered the era, the harbinger comes as soon as you enter it.
- **Always some warning.** A doom always gets at least the shortest lead, 20% of the current age's target. If it was fated for the era's first moments, before any lead could reach back, the strike is held until that long after the harbinger arrives.
- **One per era per run.** Once an era's doom has struck, passed you by, or been exposed as a false prophet's invention, no other harbinger comes in that era. Succumb and prestige start a new run, and each era rolls its fate again.
- **Offline too.** While you are away, the harbinger arrives and the doom strikes at their own moments.
- **The Cosmic Era also warns of prestige.** Beside its fated doom, a second thread warns of the Last Passage from the moment you enter the era until you prestige. See [The Last Passage](#the-last-passage).

### The figures

A harbinger and the figures who take up its warning make one **harbinger thread**, which lasts until the doom resolves. The speaker changes with the age: each time you advance while the thread lasts, the new age's figure takes up the warning, with its own word on when (see [Reading the Warning](#reading-the-warning)). Levels you bought and an Invite carry over, and the price is the same in every age of the era.

When a harbinger arrives, or a new figure takes up the warning, you get:

- a log entry naming the figure and what it warns of, for example "⚑ The Oracle has come, warning of impending doom before this age is out. Type 'harbinger' to answer.", followed by its arrival and warning lines,
- a toast announcing the figure (the first one arriving, or a later one taking up the warning),
- a **⚑ Harbinger** badge in the status bar for as long as the thread lasts.

Type `harbinger` (or `harb`) to open the Harbinger panel. It shows the current figure and, after a handoff, who it took up the warning from.

A harbinger never names an era you haven't reached. It warns of **impending doom** in the era you are in (in the Cosmic Era, also of the Last Passage), and the only era it ever names is your current one.

---

## The Roster

Every age has a figure written for it, and all 22 appear in play: three for each era from the Stone Era to the Neon Era, and four for the Cosmic Era, who warn of its doom and of the Last Passage.

### Who they are

| Age | Harbinger | Odds | Timing | False-prophet chance | Who they are |
|-----|-----------|------|--------|----------------------|--------------|
| Primitive | the Wild Man | Vague | None | 8/64 (12.5%) | A man who lives past the last fire walks in from the wilderness, gray with ash, to say what he has seen. |
| Stone | the Hermit | Vague | None | 7/64 (10.9%) | Comes down from the high caves once in a generation, and never with good news. |
| Bronze | the Soothsayer | Vague | None | 6/64 (9.4%) | Reads the future in knucklebones, sparrows and goat livers, and wants paying before and after. |
| Iron | the Desert Prophet | Vague | None | 5/64 (7.8%) | Walks in from the dry country with sand in his beard and one message for the city. |
| Classical | the Oracle | Vague | Age or era | 4/64 (6.3%) | Speaks from the smoke over the cleft rock, through priests who charge by the question. |
| Medieval | the Town Crier | Vague | Age or era | 3/64 (4.7%) | Rings his bell at the market cross and reads out doom in the voice he uses for tolls. |
| Renaissance | the Court Astrologer | Vague | Age or era | 2/64 (3.1%) | Casts the prince's horoscope, and lately the prince's horoscope has been bad for everyone. |
| Colonial | the Pamphleteer | Vague | Age or era | 1/64 (1.6%) | Prints doom on cheap paper and sells it outside the coffee house for a penny. |
| Industrial | the Newsboy | Numeric | Age or era | 0 | Shouts the late edition from the corner, and the late edition has the odds printed on it. |
| Victorian | the Doomsayer | Numeric | Age or era | 0 | Stands on a soap crate by the park railings with a sandwich board and a table of figures. |
| Electric | the Telegraph | Numeric | Age or era | 0 | Chatters all night at the post office with dispatches from stations that have stopped answering. |
| Atomic | the Civil Defense Broadcast | Numeric | Age or era | 0 | Three long notes on every wireless, then a calm voice reading the odds from a card. |
| Modern | the Evening News | Numeric | Age or era | 0 | Leads with it at six, with a graphic, an expert and an anchor trying not to look worried. |
| Information | the Chain Email | Numeric | Age or era | 0 | Forward this to ten people or it happens to you. It has a spreadsheet attached. |
| Digital | the Viral Video | Numeric | Age or era | 0 | Shaky, portrait, a million views by lunch, and the odds on a whiteboard at the end. |
| Cyberpunk | the Ghost in the Net | Numeric | Age or era | 0 | A dead corporate AI that leaks internal risk memos through the net, glitching on every third word. |
| Fusion | the Reactor Warden | Numeric | Age or era | 0 | The plant's safety intelligence, which has never before spoken outside a scheduled drill. |
| Space | the Deep Space Monitor | Numeric | Age or era | 0 | A station behind the moon that has watched one patch of sky for forty years, and has just marked a packet urgent. |
| Interstellar | the Distress Beacon | Numeric | Age or era | 0 | Still looping from a colony that went silent eighty years ago, and the loop has changed. |
| Galactic | the Elder Relay | Numeric | Age or era | 0 | An alien relay older than the species that found it, speaking in geometry for the first time in an age. |
| Quantum | your future self | Numeric | Age or era | 0 | A message in your handwriting, stamped nine years from now, that knows your passcode. |
| Transcendent | your unmade self | Numeric | Age or era | 0 | A version of you from a branch that ended, come to see whether this one ends the same way. |

**Odds:** Vague figures give a severity only; Numeric figures also publish the chance. **Timing:** None means no word of when; Age or era means "before this age is out" or "before the era ends" (see [Reading the Warning](#reading-the-warning)). The Cosmic Era's figures also warn of the Last Passage, which needs no timing: it comes at your next prestige.

Only the figure of an era's first age rolls its false-prophet chance, and only when nothing is fated there: the Wild Man for the Stone Era, the Desert Prophet for the Iron Era and the Court Astrologer for the Steel Era. The other ages' chances only come into play when a save from an older version loads partway through an era. See [False prophets](#false-prophets).

### What the actions are called

Each figure names the three actions in its own terms. The effect is the same whoever is speaking.

| Age | Appease | Brace | Invite |
|-----|---------|-------|--------|
| Primitive | Leave offerings at the stones | Dig in and hoard | Howl with the Wild Man |
| Stone | Carve his sign on the cave wall | Wall up the cave mouth | Climb the high rock and shout back |
| Bronze | Sacrifice a goat at the altar | Seal the grain in jars | Pour the omen-wine on the ground |
| Iron | Fast and wear sackcloth | Raise the walls a course higher | Curse the city alongside him |
| Classical | Send rich gifts to the shrine | Provision the citadel | Ask the Oracle for the worst |
| Medieval | Pay the monks to pray | Shore up the walls | Ring the bells backwards |
| Renaissance | Commission a votive altarpiece | Lay in stores behind the bastions | Have the astrologer cast for ruin |
| Colonial | Proclaim a day of fasting and prayer | Stockpile powder and flour | Pay her to print the date |
| Industrial | Sponsor a revival meeting | Reinforce the mills | Buy every copy and print the rest yourself |
| Victorian | Hold a national day of prayer | Lay in tinned goods and sandbags | Climb onto the soapbox with him |
| Electric | Fund a mass revival tour | Wire the city for emergency power | Wire back SEND IT |
| Atomic | Fund the early-warning network | Stock the fallout shelters | Stand on the roof and wave at the sky |
| Modern | Run a national unity telethon | Harden the power grid | Go on air and dare it |
| Information | Forward it to everyone you know | Back up everything to tape | Reply all and say bring it on |
| Digital | Fund a feel-good hashtag campaign | Buy up the bottled water | Duet the video, smiling |
| Cyberpunk | Bribe the oracle AI | Air-gap the grid | Join the accelerationist collective |
| Fusion | Hold a vigil in the containment hall | Shunt power to the containment fields | Tell the Warden to stop holding back |
| Space | Broadcast a peace hymn into the dark | Harden the orbital shields | Point every dish at it and wave |
| Interstellar | Hold a vigil for the lost colony | Pull the outposts back behind the shield | Answer the beacon and ask for it |
| Galactic | Answer the relay in the old tongue | Fold the fleet into the nebula | Tell the relay to send them |
| Quantum | Keep the promises you made yourself | Branch the timeline and hedge | Tell your future self to go ahead |
| Transcendent | Mourn the branches that ended | Anchor yourself to this reality | Let the other you in |

---

## Reading the Warning

The figure's words tell you how worried to be, and from the Classical Age on, roughly when. What you learn depends on who is speaking now.

### When

| Who is speaking | What they say about when |
|-----------------|--------------------------|
| Before the Classical Age: the Wild Man, the Hermit, the Soothsayer, the Desert Prophet | Nothing. They warn of impending doom, and the panel says so: "The Desert Prophet gives no word of when." |
| From the Classical Age on: the Oracle and every figure after her | **before this age is out**, or **before the Iron Era ends** (naming the era you are in) |

A figure who tells time says **before this age is out** when the doom falls in the current age on the era's expected schedule: each of its ages at its target time, counted from when you entered the era. In the era's last age it always says so, since the doom can't outlast the era. Otherwise it says the doom comes before the era ends.

For a real doom the words always hold. Its moment is fixed, and if you would advance past it first, it strikes at that advance instead (see [No Outrunning a Doom](#no-outrunning-a-doom)). Each new figure gives its own word, so the warning can narrow as the thread passes from age to age.

### How likely

**Before the Industrial Age**, the panel shows only the figure's words and a severity: **low**, **medium** or **high risk**.

**From the Industrial Age on**, the panel also prints the odds as a percentage. The number already includes any Appease you have bought. In the Steel Era this means a thread that starts vague turns numeric when the Newsboy takes it up.

For an era's doom, the severity follows the chance it strikes:

| Chance it strikes | Severity shown |
|-------------------|----------------|
| under 70% | low risk |
| 70% to under 85% | medium risk |
| 85% or more | high risk |

Without any Appease that works out to **low** at high faith (60%), **medium** at mid faith or with no faith storage yet (75%), and **high** at low faith (90%). For a real doom, one level of Appease brings any band down to low. See [Faith and the odds](faith.md#faith-threshold-bands). The severity is live: if your faith fill changes band, or you Appease, it moves.

The Last Passage keeps its own scale: low under 14%, medium from 14% to under 17%, high at 17% or more.

The `catastrophe` command and the Epoch panel repeat the warning in the same words, for example "The Oracle warns of doom before this age is out: medium risk of catastrophe (no figures this early), faith 40% full." In the Cosmic Era they show the doom's warning while its harbinger speaks and, on a line of its own, the Last Passage at its own odds.

### False prophets

A warning may be a lie. False prophets come only before the Industrial Age, and only to an era where nothing is fated. When you enter such an era, the false-prophet chance of its first figure decides whether one will come:

| Era | Chance of a false prophet |
|-----|---------------------------|
| Stone Era (the Wild Man's chance) | 8/64: 12.5% of runs |
| Iron Era (the Desert Prophet's chance) | 5/64 of the Iron Eras with nothing fated: about 5.7% of all Iron Eras |
| Steel Era (the Court Astrologer's chance) | 2/64 of the Steel Eras with nothing fated: about 2.3% of all Steel Eras |
| Electric, Digital, Neon and Cosmic Eras | none |

A false prophet picks its foretold moment and its lead exactly like a real doom, and arrives the same way. It claims **medium** or **high** risk (picked at random), and every figure in its thread repeats the claim, with the same warning lines a real harbinger would use. The `catastrophe` command and the Epoch panel repeat it too, so nothing the game shows tells a false prophet apart from a real one.

The claim is kept as a fixed multiple of what a real doom's chance would be. Appease, a change in faith, or an Invite move it exactly as they would move a real warning. It never drops below low risk. A false Steel Era thread that reaches the Newsboy prints the claimed figure.

At the foretold moment nothing happens. The lie comes out once the foretold window has passed:

- if the figure speaking said **before this age is out**, when you advance out of that age;
- otherwise, when you advance out of the era. That includes every figure who gives no word of when.

The log then says so, for example "⚑ This age ends without the doom the Oracle foretold. The warning had been invented from the start." or "⚑ The Iron Era ends without the doom the Desert Prophet foretold. The warning had been invented from the start." The verdict is **Discredited**.

Appease and Brace bought against a false prophet buy nothing, since no doom is coming, unless you [Invite](#invite-choose-the-catastrophe) it.

**The Stone Era.** Nothing can strike in the Stone Era, so every harbinger there (the Wild Man, the Hermit or the Soothsayer) is a false prophet, and all three answers are refused: "Unavailable: no catastrophe can strike in the Stone Era." It is a harmless first taste of harbingers.

---

## Answering the Harbinger

While a harbinger is present you can Appease, Brace or Invite, in any age of the era. Your answers belong to the doom, not to the figure: levels you buy and an Invite carry over when the next figure takes up the warning. Nothing expires, but a real doom's thread lasts only until the strike: usually its lead, 20% to 60% of an age, and less if an advance brings the strike forward (see [No Outrunning a Doom](#no-outrunning-a-doom)). In the Cosmic Era your answers go to the thread that is speaking (see [The Last Passage](#the-last-passage)).

A thread's price is set when its harbinger arrives and stays the same for as long as the thread lasts, whoever is speaking. If your storage can't hold the price yet, the refusal tells you how much storage you need.

### Appease: lower the odds

| Level | Cost | Chance the doom strikes |
|-------|------|-------------------------|
| 0 | none | unchanged |
| 1 | 15% of what the age the harbinger arrives in makes, in faith (and the same in culture, from the Steel Era on) | ×0.6 |
| 2 | double level 1 | ×0.36 (×0.6 again) |

- **15% of the age** is three quarters of what the shortest warning makes. A harbinger comes 20% to 60% of an age's target length before its doom (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)), and an advance can cut that short. So level 1 is priced on faith you can gather while the warning lasts, not on faith you had to save before it. Level 2 costs double, and both levels together cost 45% of what the age makes, more than an average warning brings in: the second level takes a long warning or faith you kept in storage.
- **What an age makes** is what a player who invests moderately in faith makes in it: five fully staffed copies of every faith building so far, the faith of every wonder already built, the flat faith of the techs, all multiplied by the production bonus of the techs and wonders you hold by then (it reaches the ×3 cap in the Electric Age). Culture is counted the same way.
- The price belongs to the thread. It is set from the age the harbinger arrives in and does not change when the thread passes to the next age's figure. It is the same on known ground: a mastered age makes more per tick for a shorter warning.
- Culture is only charged if you already had it when the era began. It unlocks in the Classical Age, so Iron Era threads cost faith only.
- The Last Passage's thread is priced differently, because it lasts from your arrival in the Cosmic Era until you prestige: a quarter of what the era's ages make (see [What it costs](#what-it-costs-by-epoch)).
- Two levels at most.
- Appease always lowers the chance a real doom strikes. Against a false prophet it only lowers the claim, since nothing is coming.
- **Not after Invite.** Once you invite the catastrophe, Appease is refused: it will come whatever you offer.

Example: at mid faith a fated doom strikes 75% of the time. One level of Appease takes it to 45%, two levels to 27%.

Spending faith lowers your faith fill, and that can drop you into a worse faith band for the strike. Appease still comes out ahead. In the worst case one level leaves the odds at 0.9× of where they started (from the 60% band to the 90% band, then ×0.6 is 54%), so every level lowers the real chance.

### Brace: soften an Endure

Brace spends the resources the era asks of you. For each resource, the price is based on the most the era asks of it across its advances (into its later ages and into the next era). Only resources you already had when the era began count, and faith and culture are left out.

| Level | Cost | If you Endure: buildings destroyed | If you Endure: stored resources kept |
|-------|------|------------------------------------|--------------------------------------|
| 0 (unbraced) | none | 20% | 15% |
| 1 | 12% of the most the era asks of each of those resources | 15% | 30% |
| 2 | 24% of the same | 10% | 45% |

Buildings destroyed are counted from your buildings other than wonders and storage, rounded down, with at least 1 if you have any. Wonders and storage are never destroyed. The rest of Endure (25% of workers lost, the 562-tick reconstruction debuff, −10 morale) is the same at every Brace level. See [Endure](catastrophe.md#endure).

**Brace and your garrison.** The table is Brace alone. If you have soldiers, Brace applies first and then your garrison blunts its share of what Brace leaves: fewer buildings fall and more stock is kept. The garrison's share is measured against the raid threat of the age the doom strikes in. The panel measures it against the age you are in now; if the doom strikes after you advance, the threat there is double and your garrison blunts less than the preview shows. Brace and garrison together can cut the unbraced loss by at most **60%**: at least 8% of buildings fall and at most 66% of stock is kept. That cap only bites at level 2. For example, level 1 with a garrison that blunts 20% of a raid means 12% of buildings fall and 44% of stock is kept; level 2 with a strong garrison stops at 8% and 66%. See [Your garrison](catastrophe.md#your-garrison) and [Defense: what your army blunts](military.md#defense-what-your-army-blunts).

The panel's Brace preview ("If it comes and you Endure: N% of buildings fall, N% of stock is kept", and the same for the next level) already counts your garrison, so it shows what you would face in your current age. A line under it says so: "Your garrison is counted: it blunts about N% of what Brace leaves." With no soldiers it reads "No garrison counted: soldiers would soften an Endure further (Army panel)." When the 60% cap cuts in, a further line says so.

Things to know:

- Two levels at most.
- Brace only matters if the catastrophe comes **and** you choose Endure. If no catastrophe comes, or you Succumb, the resources are simply spent.
- The Brace is attached to the pending catastrophe. If you close the choice with Esc and Endure later, or save and load in between, it still applies.
- Brace is allowed after Invite.
- **Against the Last Passage, Brace protects points, not your civilization.** An Endure at the Last Passage keeps 50% of the run's prestige points unbraced, 70% at level 1 and 85% at level 2. The building and resource numbers above don't apply there, and your garrison doesn't count. See [The Last Passage](prestige.md#the-last-passage). Against the Cosmic Era's fated doom, the Reality Tear, Brace works as above, garrison included.

### What it costs, by epoch

Level 1 prices. Level 2 costs double.

**Brace** is priced by the era:

| Era | Brace (level 1) |
|-----|-----------------|
| Stone Era | refused: nothing can strike there |
| Iron Era | 26.4K knowledge, 26.4K stone, 6.36K iron, 21.6K gold |
| Steel Era | 3.6M knowledge, 1.8M gold, 288K steel |
| Electric Era | 56.4M steel, 924K oil, 3.96M electricity |
| Digital Era | 156M gold, 117.6B electricity, 19.2B data |
| Neon Era | 288B electricity, 46.8B data, 3B crypto |
| Cosmic Era | 1.56T dark matter, 75.6B titanium |

**Appease** is priced by the age the harbinger arrives in, and keeps that price for the whole thread:

| Era | Harbinger arrives in | Appease (level 1) | The warning lasts |
|-----|----------------------|-------------------|-------------------|
| Stone Era | any age | refused: nothing can strike there | |
| Iron Era | Iron Age | 1.4K faith | 1.3 h to 3.9 h |
| | Classical Age | 2.3K faith | 1.8 h to 5.5 h |
| | Medieval Age | 4.9K faith | 2.3 h to 7.0 h |
| Steel Era | Renaissance Age | 9.2K faith, 95K culture | 3.1 h to 9.4 h |
| | Colonial Age | 26K faith, 230K culture | 3.6 h to 10.9 h |
| | Industrial Age | 79K faith, 870K culture | 4.2 h to 12.5 h |
| Electric Era | Victorian Age | 210K faith, 2.6M culture | 4.7 h to 14.0 h |
| | Electric Age | 490K faith, 6.7M culture | 5.2 h to 15.6 h |
| | Atomic Age | 1.2M faith, 16M culture | 6.2 h to 18.7 h |
| Digital Era | Modern Age | 2.2M faith, 31M culture | 6.2 h to 18.7 h |
| | Information Age | 5M faith, 74M culture | 7.3 h to 21.8 h |
| | Digital Age | 12M faith, 170M culture | 8.3 h to 25.0 h |
| Neon Era | Cyberpunk Age | 25M faith, 380M culture | 9.4 h to 28.1 h |
| | Fusion Age | 56M faith, 850M culture | 10.4 h to 31.2 h |
| | Space Age | 130M faith, 1.9B culture | 11.4 h to 34.3 h |
| Cosmic Era: the Reality Tear | Interstellar Age | 270M faith, 4.1B culture | 12.5 h to 37.4 h |
| | Galactic Age | 540M faith, 8.1B culture | 12.5 h to 37.4 h |
| | Quantum or Transcendent Age | 1.1B faith, 17B culture | 12.5 h to 37.4 h |
| Cosmic Era: the Last Passage | on arrival in the era | 3.1B faith, 48B culture | until you prestige |

The warning times are for a first run. On known ground an age and its warning are both shorter by the age's [Era Mastery](prestige.md#era-mastery) speed, and the age makes that much more per tick, so the price is the same.

The Cosmic Era's passage is prestige, which you may take in any of its ages. The Last Passage's thread runs from your arrival in the era until you prestige, so its Appease is priced on the era, not on a warning: a quarter of what the Interstellar, Galactic and Quantum Ages make. If you prestige soon after you arrive, it is out of reach. In the smoke-test bot's runs, level 1 on the Last Passage's thread came about 4 days into the era and level 2 about 7 (see [Can you afford it?](#can-you-afford-it)). Brace is based on the most the Cosmic Era's own advances ask of each resource you have held since the Interstellar Age: 13T dark matter (for the Quantum Age) and 630B titanium (for the Galactic Age). Antimatter and quantum flux arrive later, so they are left out.

The two Cosmic Era threads, the Reality Tear's and the Last Passage's, share the era's Brace price; each has its own Appease price, as above. Each keeps its own levels, so answering one doesn't answer the other.

#### Can you afford it?

Brace, nearly always. Appease, only if you banked faith before the harbinger came.

Brace is priced in the construction resources the era's advances already ask for, so you usually hold enough the moment the harbinger arrives. Appease is priced in faith (plus culture from the Steel Era on), a quarter of what a moderate faith economy makes over the whole era. But a doom's thread lasts only its lead, 20% to 60% of an age, and faith can't be bought at the market (culture can, gold → culture). A thread is too short to earn the price, so Appease has to come out of faith you already kept in storage. That faith also lowers the strike chance on its own.

In the smoke-test bot's runs (24 seeds from a new game to a Quantum Age prestige: a first run on the one-week curve, so every age at 1x), the bot saved toward both answers and bought each level as soon as it could, keeping what its next advance needed. Each real doom's thread, from the harbinger's arrival until the doom struck or passed, went like this (false prophets left out):

| Thread | Threads seen | Thread length (median) | Appease level 1 affordable | Appease level 2 affordable | Brace level 1 affordable |
|--------|--------------|------------------------|----------------------------|----------------------------|--------------------------|
| Iron Era | 4 | 3.4 h | 3 of 4, on arrival | 1 of 4 | 4 of 4, on arrival |
| Steel Era | 3 | 6.8 h | 2 of 3, after 19 minutes and 3.2 h | none | 3 of 3, within 35 minutes |
| Electric Era | 5 | 11.7 h | 2 of 5, after 3.7 h and 12.7 h | 1 of 5 | 4 of 5, on arrival |
| Digital Era | 7 | 12.7 h | 1 of 7 | none | 7 of 7: 4 on arrival, the rest within 5.2 h |
| Neon Era | 7 | 14.1 h | 1 of 7 | 1 of 7 | 7 of 7, on arrival (one after 6.1 h) |
| Cosmic Era: the Reality Tear | 4 | 13.4 h | 1 of 4 | none | 4 of 4, on arrival |
| Cosmic Era: the Last Passage (to a Quantum Age prestige) | 24 | 8.2 days | 24 of 24, after about 4 days | 24 of 24, after about 7 days | 24 of 24, on arrival |

A doom is fated in only 27% of eras, so these are 30 dooms across 24 runs: read the shares, not the exact counts. The bot keeps faith only for age requirements and the Sistine Chapel. When a harbinger came it usually held well under half the Appease price in faith (in the Neon Era, about 2%), and twice it had the faith but not the culture. Only in the Iron Era, the cheapest, did the faith it already held cover level 1 in most threads. The Last Passage's thread runs from your arrival in the Cosmic Era until you prestige, which is long enough for faith buildings to pay for both levels.

Faith you spend on Appease is faith the next age requirement can't count: keep what the next age asks for, and what the Sistine Chapel still needs, before you appease.

### Invite: choose the catastrophe

- Free.
- Guarantees the strike. It still comes at its fated moment, or at an advance that would outrun it.
- **Can't be undone.** It stays with the thread when the next figure takes over.
- After inviting, Appease is refused. Brace is still allowed.
- Only while a harbinger is present, so at most once per era (in the Cosmic Era, once per thread). In the Stone Era it is refused like the other answers.
- **Inviting a false prophet** makes its invented doom real: it strikes at the foretold moment (at once, if that moment has passed), the verdict is Fulfilled, and the log notes that the warning had been invented.

Invite is for players who want to Succumb on purpose, to collect an era's legacy bonus and Ancient Knowledge. It is the only way to choose a catastrophe; there is no command to trigger one directly. You can't choose where a doom is fated, only answer one when its harbinger comes. See [Succumb](catastrophe.md#succumb).

In the Cosmic Era, Invite answers the thread that is speaking. Invite the Reality Tear's harbinger and that doom is certain, like any era's. Invite the Last Passage's thread and your next prestige brings the Last Passage: it is how you choose the [Cosmic Legacy](prestige.md#cosmic-legacy) on purpose.

### Why Appease is priced on the warning

A doom's harbinger gives you a fifth to three fifths of an age. Appease used to cost a quarter of what the whole era makes in faith, which no warning is long enough to earn: it could only be paid from faith saved before anyone had warned you, and in test games it was affordable in a third of the warnings. Now level 1 is 15% of what the age makes, three quarters of the shortest warning, so the warning itself can pay for it, and level 2 is the stretch.

The price is fixed when the harbinger arrives, so it doesn't climb if the thread carries into the next age, and it comes from what the age makes, not from your storage: a harbinger who comes while your stores are small is no cheaper to answer than the age warrants. Brace is priced by the era, from what its advances ask of you, and is the same in every age of it.

---

## No Outrunning a Doom

Advancing faster doesn't dodge a fated doom:

- **Leaving the era.** If you advance out of the era before the doom's moment, the strike rolls at that advance, before it goes through. If it hits, the advance waits behind the pending catastrophe. If it misses, you are spared and the advance goes through.
- **Leaving the age it falls in.** Once the figure speaking has said **before this age is out**, any advance out of that age works the same way.
- **No harbinger yet.** If you reach the era's last advance and no harbinger has come, it comes at that advance, and the advance waits for one more try: "The Town Crier stands in your way, warning of impending doom before this age is out. Type 'harbinger' to answer, or advance again to meet it." Your next advance brings the strike, or reveals a false prophet. In the Stone Era, where nothing can strike, a false prophet who hasn't come by the era's end never comes, and the advance goes through.
- **The build plan** follows the same rules. A `plan advance` waits behind a pending catastrophe, and if a harbinger comes at the plan's advance, the plan tries again on the next tick.
- **Offline** changes nothing: the harbinger arrives and the doom strikes at their own moments while you are away. A doom that strikes waits, pending, until you come back, and the log shows the warning and the strike.
- **Prestige before the Cosmic Era** (it opens at the Medieval Age) ends the run: a doom that hasn't struck there is gone with it. An early taste in the Medieval Age escapes an Iron Era doom this way, at the price of the run.
- **Prestige from the Cosmic Era** is the era's passage, so an open doom settles before the Last Passage rolls. If its harbinger hasn't come, it comes at the prestige, and the prestige waits for one more try: "... Type 'harbinger' to answer, or confirm prestige again to meet it." Then the strike rolls first. A hit holds the prestige behind the pending Reality Tear: Endure it, then prestige again, which rolls the Last Passage (or Succumb, which resets the run with no prestige). A miss lets the same confirm go on to the Last Passage roll.

---

## When the Doom Resolves

The thread is settled when its doom resolves: at the fated moment, at an advance (in the Cosmic Era, a prestige) that would outrun it, or, for a false prophet, when its foretold window passes. The last figure speaks the verdict, and the log records one of four:

| Verdict | What happened |
|---------|---------------|
| **Fulfilled** | You invited the catastrophe, and it came. An invited false prophet ends this way too; the log notes that its warning had been invented. |
| **Vindicated** | It warned you, you didn't invite it, and the catastrophe came. |
| **Spared** | A real warning, and the strike missed. |
| **Discredited** | A false prophet, revealed when its foretold window passed. |

A miss reads "⚑ The doom the Oracle foretold passed you by. The warning was real, and you were spared." After a strike or a miss nothing more strikes in that era: the Epoch panel shows "Catastrophe: spared (nothing more will strike this era)" after a miss, and the `catastrophe` command and the Harbinger panel say nothing more will strike before the era ends.

If the catastrophe came and you had braced, the log also says what Endure will cost you at that Brace level. Then the thread ends and the catastrophe choice works as usual (see [Catastrophe System](catastrophe.md)).

The Epoch panel (`epoch`) shows the current thread's status and, for past threads this run, the whole chain of figures and the verdict, for example "The Oracle, The Town Crier → doom in the Iron Era: Spared (settled as you advanced)".

### The Last Passage

The Cosmic Era runs two threads. Its fated doom, the Reality Tear, gets a harbinger like any era's. The Last Passage's thread starts when you enter the era and is settled when you confirm prestige, not at an age advance. The panel and the log both say its figure warns of the Last Passage: the end of this civilization at your next prestige.

While the Reality Tear's harbinger speaks, the Last Passage's thread waits behind it with its answers intact. Appease, Brace and Invite go to the thread that is speaking, and the panel says "The Last Passage still waits at your next prestige. Your answers to it stand." When the doom resolves (struck or spared), the Last Passage's thread takes up the warning again in the current age's voice, with a log line if the age moved on while it waited. The `catastrophe` command and the Epoch panel show both: the doom's warning, for example "Your future self warns of doom before this age is out: 90% catastrophe chance (high), faith 0% full.", and the Last Passage at its own odds.

Confirming prestige settles an open doom first (see [No Outrunning a Doom](#no-outrunning-a-doom)). Then the Last Passage rolls once, with the odds for your faith band (18%, 15% or 12%), times 0.6 per level of Appease on its own thread, or certain if you invited it. If nothing comes, the verdict is Spared and prestige completes. If it comes, prestige waits for you to Endure or Succumb. Endure keeps part of the run's points and Succumb grants the Cosmic Legacy. See [The Last Passage](prestige.md#the-last-passage).

If both are ever pending at once, the Reality Tear is answered first: the choice window shows it first, and the Last Passage's Endure and Succumb are refused until it is ("The Reality Tear came first. Answer it before the Last Passage.").

---

## Saving, Succumb and Prestige

- Your current era's fate is saved with your game, so reloading can't re-roll it. So is the thread: the figures who have spoken, what the current one said, your Appease and Brace levels, whether you invited the catastrophe, and in the Cosmic Era a Last Passage thread waiting behind the doom's. A Brace attached to a pending catastrophe is saved too.
- **Succumb** clears the live thread, the era's fate, the invite and any Brace. Past verdicts are kept, like the epoch event history. The new run's Stone Era rolls its fate on the first tick.
- **Prestige** clears all of it, past verdicts included. Before the Cosmic Era a doom that hasn't struck ends with the run; in the Cosmic Era it settles before the Last Passage rolls.
- A pending Last Passage is saved with your game, and the choice is still waiting when you load.

---

## Commands and Keys

| Command | What it does |
|---------|--------------|
| `harbinger` (or `harb`) | Open the Harbinger panel. It is also listed under Panels in the sidebar. |
| `harbinger appease` | Buy the next Appease level without opening the panel |
| `harbinger brace` | Buy the next Brace level without opening the panel |
| `harbinger invite` | Invite the catastrophe without opening the panel. Can't be undone. |

In the panel:

| Key | Action |
|-----|--------|
| **A** | Appease |
| **B** | Brace |
| **I** | Invite. Press **I** twice to confirm, since it can't be undone. |
| **Esc** | Close the panel |

With no harbinger present, the panel says so, explains that a harbinger comes only when doom is on its way, some while before it strikes, and shows the outlook in plain words: "No harbinger has come: the Iron Era is quiet, for now. A doom is always foretold before it strikes." In the Stone Era it says no catastrophe can strike there. Once the era's doom has struck or passed you by, it says nothing more will strike before the era ends. In the Cosmic Era it shows the risk of the Last Passage at your next prestige.

---

## Strategy

- **Bank faith before you need it.** A harbinger gives you 20% to 60% of an age's time before its doom strikes. Faith already in storage lowers the strike chance by itself and pays for Appease the moment the harbinger comes.
- **Answer before the moment.** The price doesn't change once the harbinger has come, but the thread lasts only until the doom's moment, and advancing out of the era, or out of the age a figure named, brings the strike to that advance (in the Cosmic Era, so does prestige).
- **Want to keep your run?** Appease is the direct answer. One level cuts the chance it strikes by 40%, and it can't hurt the odds.
- **Worried but short on faith?** Brace instead. It doesn't lower the odds, but it makes Endure much cheaper if the catastrophe comes.
- **Early warnings deserve a little doubt.** Before the Industrial Age a warning may be false: in the Stone Era it always is, and in the Iron and Steel Eras now and then. Your own faith fill tells you the real band (see [Faith and the odds](faith.md#faith-threshold-bands)).
- **Prestiging from the Cosmic Era?** An open Reality Tear settles first, so answer its harbinger if one is speaking. The Last Passage then costs you points, not buildings. On its thread, Appease to lower the odds, Brace to keep more points on an Endure, Invite if you want the Cosmic Legacy.
- **Hunting legacy bonuses?** Invite, then Succumb when it strikes. Skip Appease; it is refused after Invite anyway. Brace only if you think you might change your mind and Endure. You can't choose where a doom is fated, so collecting every era's legacy takes several runs.
