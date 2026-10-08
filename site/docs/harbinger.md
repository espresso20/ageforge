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

<figure class="screen" data-screen="harbinger"><figcaption>The Harbinger panel in the Iron Age: the warning of the Desert Prophet, its severity, and what the next level of Appease and of Brace would cost.</figcaption></figure>

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

Without any Appease that works out to **low** at high faith (60%), **medium** at mid faith (75%), and **high** at low faith (90%). For a real doom, one level of Appease brings any band down to low. See [Faith and the odds](faith.md#faith-threshold-bands). The severity is live: if your faith strength changes band, or you Appease, it moves.

The Last Passage keeps its own scale: low under 14%, medium from 14% to under 17%, high at 17% or more.

The `catastrophe` command and the Epoch panel repeat the warning in the same words, for example "The Oracle warns of doom before this age is out: medium risk of catastrophe (no figures this early), faith strength 44% (devotion 2.0x, 100% of your faith kept)." In the Cosmic Era they show the doom's warning while its harbinger speaks and, on a line of its own, the Last Passage at its own odds.

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
| 2 | the same again as level 1 | ×0.36 (×0.6 again) |

- **15% of the age** is three quarters of what the shortest warning makes. A harbinger comes 20% to 60% of an age's target length before its doom (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)), and an advance can cut that short. So level 1 is priced on faith you can gather while the warning lasts, not on faith you had to save before it. Level 2 costs the same again, so both levels together cost 30% of what the age makes, about what an average warning brings in: the second level takes a warning of ordinary length or faith you kept in storage.
- **What an age makes** is what a player who invests moderately in faith makes in it: five fully staffed copies of every faith building so far, the faith of every wonder already built, the flat faith of the techs, all multiplied by the production bonuses a typical player holds by then: the all-production pool of milestones and wonders (past +200% from the Digital Age on, where a point counts a quarter) and the techs' own bonus on top. Culture is counted the same way.
- The price belongs to the thread. It is set from the age the harbinger arrives in and does not change when the thread passes to the next age's figure. It is the same on known ground: a mastered age makes more per tick for a shorter warning.
- Culture is only charged if you already had it when the era began. It unlocks in the Classical Age, so Iron Era threads cost faith only.
- The Last Passage's thread is priced by the same rule on a longer warning. It comes as you enter the Cosmic Era and lasts until you prestige, so its shortest warning counts as two thirds of the Interstellar Age, not a fifth of it: level 1 is half of what that age makes, a little over three times a doom foretold there, and level 2 costs the same again (see [What it costs](#what-it-costs-by-epoch)).
- Two levels at most.
- Appease always lowers the chance a real doom strikes. Against a false prophet it only lowers the claim, since nothing is coming.
- **Not after Invite.** Once you invite the catastrophe, Appease is refused: it will come whatever you offer.

Example: at mid faith a fated doom strikes 75% of the time. One level of Appease takes it to 45%, two levels to 27%.

Spending faith lowers the share of your faith you have kept, and your [faith strength](faith.md#faith-threshold-bands) with it, and that can drop you into a worse faith band for the strike. Appease still comes out ahead. In the worst case one level leaves the odds at 0.9× of where they started (from the 60% band to the 90% band, then ×0.6 is 54%), so every level lowers the real chance. A typical town is in the bottom band before and after it pays, so for it Appease is the whole 40% off.

### Brace: soften an Endure

Brace spends the materials your age makes. Level 1 costs a third of what a moderate economy makes of each of them during the doom's shortest warning: a fifth of the age the harbinger arrives in, so a fifteenth of what that age makes. That is 27 minutes of income in the Iron Age and about four hours in the Space Age: the effort grows with the age, as the warning does.

- **The materials** are the era's core resources (the ones its advances ask for, that you already had when the era began, with faith and culture left out) that the buildings of a moderate economy make by then. The Modern Age buys its data at the market, so a harbinger who arrives there asks for gold and electricity only. Nothing but a wonder makes crypto, so no Neon Era harbinger asks for it.
- **A moderate economy** is five fully staffed copies of every producer so far, with the bonuses a typical player holds by then: the same measure Appease uses for faith.
- The price belongs to the thread, like Appease's: it is set from the age the harbinger arrives in and keeps that price through every age the thread lives in.
- Level 2 costs the same again as level 1.

The Cosmic Era's two threads are priced by the same rule and cost more of the warning: the Reality Tear's five sixths of its shortest warning, the Last Passage's half of its much longer one (see [What it costs](#what-it-costs-by-epoch)).

| Level | Cost | If you Endure: buildings destroyed | If you Endure: stored resources kept |
|-------|------|------------------------------------|--------------------------------------|
| 0 (unbraced) | none | 20% | 15% |
| 1 | a third of what the shortest warning makes of each material | 15% | 30% |
| 2 | the same again | 10% | 45% |

Buildings destroyed are counted from your buildings other than wonders and storage, rounded down, with at least 1 if you have any. Wonders and storage are never destroyed. The rest of Endure (25% of workers lost, the 562-tick reconstruction debuff, −10 morale) is the same at every Brace level. See [Endure](catastrophe.md#endure).

**Brace and your garrison.** The table is Brace alone. If you have soldiers, Brace applies first and then your garrison blunts its share of what Brace leaves: fewer buildings fall and more stock is kept. The garrison's share is measured against the raid threat of the age the doom strikes in. The panel measures it against the age you are in now; if the doom strikes after you advance, the threat there is double and your garrison blunts less than the preview shows. Brace and garrison together can cut the unbraced loss by at most **60%**: at least 8% of buildings fall and at most 66% of stock is kept. That cap only bites at level 2. For example, level 1 with a garrison that blunts 20% of a raid means 12% of buildings fall and 44% of stock is kept; level 2 with a strong garrison stops at 8% and 66%. See [Your garrison](catastrophe.md#your-garrison) and [Defense: what your army blunts](military.md#defense-what-your-army-blunts).

The panel's Brace preview ("If it comes and you Endure: N% of buildings fall, N% of stock is kept", and the same for the next level) already counts your garrison, so it shows what you would face in your current age. A line under it says so: "Your garrison is counted: it blunts about N% of what Brace leaves." With no soldiers it reads "No garrison counted: soldiers would soften an Endure further (Army panel)." When the 60% cap cuts in, a further line says so.

Things to know:

- Two levels at most.
- Brace only matters if the catastrophe comes **and** you choose Endure. If no catastrophe comes, or you Succumb, the resources are simply spent.
- The Brace is attached to the pending catastrophe. If you close the choice with Esc and Endure later, or save and load in between, it still applies.
- Brace is allowed after Invite.
- **Against the Last Passage, Brace protects points, not your civilization, and it takes real effort.** An Endure at the Last Passage keeps 50% of the run's prestige points unbraced, 70% at level 1 and 85% at level 2. The building and resource numbers above don't apply there, and your garrison doesn't count. It costs more than a doom's: each level costs a third of what the Interstellar Age makes of dark matter and titanium at a moderate economy, about 22 hours of income. See [Brace or Appease against the Last Passage](#brace-or-appease-against-the-last-passage) and [The Last Passage](prestige.md#the-last-passage). Against the Cosmic Era's fated doom, the Reality Tear, Brace works as above, garrison included, at half the Last Passage's price: a sixth of what the age makes, about eleven hours of income a level.

### What it costs, by epoch

Level 1 prices. Level 2 of either answer costs the same again.

Both answers are priced by the age the harbinger arrives in, and keep that price for the whole thread. **Brace** is a third of what the shortest warning makes of each material (more in the Cosmic Era); the hours are how long a moderate economy takes to make it. **Appease** is three quarters of what the shortest warning makes of faith and culture. One tech cuts Appease: with **Void Contemplation** (Interstellar Age) every Appease level costs 20% less, on every thread, the Last Passage's included, and the Harbinger panel shows the price you would pay.

| Era | Harbinger arrives in | Brace (level 1) | Hours of income | Appease (level 1) | The warning lasts |
|-----|----------------------|-----------------|-----------------|-------------------|-------------------|
| Stone Era | any age | refused: nothing can strike there | | refused | |
| Iron Era | Iron Age | 37K stone, 170K iron, 140K gold | 27 min | 1.6K faith | 1.3 h to 3.9 h |
| | Classical Age | 51K stone, 760K iron, 920K gold | 37 min | 2.7K faith | 1.8 h to 5.5 h |
| | Medieval Age | 66K stone, 3.7M iron, 4.9M gold | 47 min | 5K faith | 2.3 h to 7.0 h |
| Steel Era | Renaissance Age | 23M gold, 4.5M steel | 1.1 h | 11K faith, 96K culture | 3.1 h to 9.4 h |
| | Colonial Age | 510M gold, 69M steel | 1.2 h | 32K faith, 240K culture | 3.6 h to 10.9 h |
| | Industrial Age | 3.8B gold, 2.8B steel | 1.4 h | 110K faith, 1.1M culture | 4.2 h to 12.5 h |
| Electric Era | Victorian Age | 13B steel, 1.3B oil, 1.7M electricity | 1.6 h | 210K faith, 2.5M culture | 4.7 h to 14.0 h |
| | Electric Age | 58B steel, 5.2B oil, 27B electricity | 1.7 h | 500K faith, 6.3M culture | 5.2 h to 15.6 h |
| | Atomic Age | 240B steel, 7B oil, 180B electricity | 2.2 h | 1.2M faith, 17M culture | 6.2 h to 18.7 h |
| Digital Era | Modern Age | 1.7T gold, 1.3T electricity | 2.1 h | 2.6M faith, 36M culture | 6.2 h to 18.7 h |
| | Information Age | 40T gold, 17T electricity, 600B data | 2.5 h | 6.2M faith, 92M culture | 7.3 h to 21.8 h |
| | Digital Age | 49T gold, 87T electricity, 4.3T data | 2.8 h | 16M faith, 230M culture | 8.3 h to 25.0 h |
| Neon Era | Cyberpunk Age | 600T electricity, 21T data | 3.2 h | 34M faith, 510M culture | 9.4 h to 28.1 h |
| | Fusion Age | 1.6Q electricity, 24T data | 3.6 h | 76M faith, 1.2B culture | 10.4 h to 31.2 h |
| | Space Age | 4.8Q electricity, 28T data | 3.9 h | 170M faith, 2.7B culture | 11.4 h to 34.3 h |
| Cosmic Era: the Reality Tear | Interstellar Age | 14Q titanium, 11Q dark matter | 11.4 h | 380M faith, 5.8B culture | 12.5 h to 37.4 h |
| | Galactic Age | 14Q titanium, 240Q dark matter | 10.8 h | 780M faith, 13B culture | 12.5 h to 37.4 h |
| | Quantum Age | 15Q titanium, 250Q dark matter | 10.9 h | 1.7B faith, 27B culture | 12.5 h to 37.4 h |
| | Transcendent Age | 16Q titanium, 270Q dark matter | 10.7 h | 1.8B faith, 29B culture | 12.5 h to 37.4 h |
| Cosmic Era: the Last Passage | Interstellar Age, as you enter the era | 27Q titanium, 21Q dark matter | 21.7 h | 1.3B faith, 20B culture | until you prestige |

The warning times are for a first run. On known ground an age and its warning are both shorter by the age's [Era Mastery](prestige.md#era-mastery) speed, and the age makes that much more per tick, so the price is the same.

The Cosmic Era's passage is prestige, which you may take in any of its ages. The Last Passage's thread runs from your arrival in the era until you prestige, and nothing about its timing is hidden: you pick the moment. So where a doom's shortest warning is a fifth of an age, the Last Passage's is counted as two thirds of the age its harbinger arrives in, the Interstellar Age (41.6 h of its 62.4 h on a first run): longer than any doom's. Both of its answers are priced on that warning, at what a moderate economy (five staffed copies of every producer) makes in it:

- **Appease** level 1 is three quarters of what the warning makes in faith and culture, half of what the whole age does: 1.3B faith and 20B culture, a little over three times a doom foretold in the same age, and about 33 hours of income. Level 2 costs the same again, so both levels together cost what the whole Interstellar Age makes: more than the warning brings in, so the second takes playing the age out, or a stock you brought with you.
- **Brace** level 1 is half of what the warning makes in dark matter and titanium, a third of what the whole age does: 21Q dark matter and 27Q titanium, about 22 hours of income. Level 2 costs the same again, so both levels together cost what the warning makes, two thirds of the Interstellar Age. The Galactic Age makes about 21 times the dark matter but no more titanium, so later in the era the titanium is the real price. Each level fits the storage you can build in the Interstellar Age; if yours can't hold the price yet, the refusal says how much you need.

If you prestige as soon as you arrive, only what you already hold can pay. A thread that begins in a later age (only a save from an older version that was already deeper in the era) is priced on that age: 2.6B faith, 42B culture, 470Q dark matter and 28Q titanium in the Galactic Age, and 5.4B faith, 89B culture, 490Q dark matter and 29Q titanium in the Quantum Age.

The Reality Tear's thread is priced the same way on a doom's warning, which is at least a fifth of the age. Its Brace level 1 is five sixths of what a moderate economy makes of dark matter and titanium in that fifth, a sixth of what the whole age does and half the Last Passage's: 11Q dark matter and 14Q titanium if its harbinger arrives in the Interstellar Age, about eleven hours of income. Its Appease is a doom's: 380M faith and 5B culture there, about nine and a half hours. Level 2 of each costs the same again. Each thread keeps its own levels, so answering one doesn't answer the other.

#### Brace or Appease against the Last Passage

For a player who would Endure, the two answers work on the same loss from opposite ends: Appease makes the Last Passage less likely, Brace makes it cost less. What you can expect to lose is the chance it comes times the share of the run's points an Endure gives up. At mid faith, in percent of the run's prestige points:

| | Unbraced (Endure keeps 50%) | Brace 1 (keeps 70%) | Brace 2 (keeps 85%) |
|---|---|---|---|
| No Appease (15% chance) | 7.5% | 4.5% | 2.25% |
| Appease 1 (9%) | 4.5% | 2.7% | 1.35% |
| Appease 2 (5.4%) | 2.7% | 1.62% | 0.81% |

In the bottom band, where a typical town sits, the chance is 18% and every figure is a fifth higher; in the top band, a fifth lower. A run that prestiges in the Interstellar Age is worth 1,092 points, so 3% of it is 33 points; from the Galactic Age it is 1,821 and 55.

What each step costs, for the thread a run meets (foretold in the Interstellar Age), in hours of a moderate economy's income:

| Step | Price | Hours of income | Saves, with none of the other answer | Saves, with the other answer at level 1 |
|------|-------|-----------------|--------------------------------------|-----------------------------------------|
| Brace 1 | 21Q dark matter, 27Q titanium | 22 | 3.0% of the run's points | 1.8% |
| Appease 1 | 1.3B faith, 20B culture | 33 | 3.0% | 1.8% |
| Brace 2 | the same again | 22 more | 2.25% | 1.35% |
| Appease 2 | the same again | 33 more | 1.8% | 1.08% |

- The first level of each saves the same share. Brace level 1 is the cheaper: per hour of income, Appease level 1 saves about two thirds of what Brace level 1 does.
- The two are paid in different resources, so you gather both at once. Holding both first levels leaves 2.7% at risk, for 22 hours of dark matter and titanium and 33 of faith and culture.
- Once you hold Brace level 1, Brace level 2 is the next best step for an hour of income: it saves 2.25% for another 22 hours, where Appease level 1 saves 1.8% for 33. They are still paid in different resources.
- A second level costs what the first did and saves a little less: Brace level 2 saves three quarters of what level 1 did, Appease level 2 three fifths.
- Appease is paid in faith, and the share of your faith you have kept is half of your [faith strength](faith.md#faith-threshold-bands). A typical town is in the bottom band before and after it pays, so the tables above hold for it as they stand (at the bottom band's 18%: every figure a fifth higher). A town that has worked its way into the top band can pay itself out of it: one with three and a half times the moderate set of faith buildings holds about 16B faith by the end of the Interstellar Age, and paying 1.3B takes its faith strength from 78% to about 71%, the middle band. The roll's chance then goes up a step (12% to 15%), which takes back part of what Appease bought, never all of it.
- All of it is spent whether or not the Last Passage comes, and none of it helps if you mean to Succumb for the [Cosmic Legacy](prestige.md#cosmic-legacy): Invite instead.

#### Can you afford it?

Against a doom: Brace, nearly always. Appease level 1, in about three warnings out of four. Level 2 was affordable in about half while it cost double; it costs the same again now. The table below was measured before an ordinary doom's Brace was priced on the warning and before second levels came down, and the Cosmic Era's two threads have not been measured at their prices either (see the notes under the table).

A doom's Brace is priced in materials the age makes, a third of the shortest warning's worth: a player who keeps their producers running holds it, or gathers it well inside the warning. Appease is priced in faith (plus culture from the Steel Era on), and faith can't be bought at the market (culture can, gold → culture). Level 1 costs less than the shortest warning makes at a moderate faith economy, so a player who keeps faith buildings staffed can pay it from what they hold or gather it before the doom's moment. Level 2 takes a long warning, or faith and culture kept in storage.

In the smoke-test bot's runs (24 seeds from a new game to a Quantum Age prestige: a first run on the one-week curve, so every age at 1x), the bot saved toward both answers and bought each level as soon as it could, keeping what its next advance needed. Each real doom's thread, from the harbinger's arrival until the doom struck or passed, went like this (false prophets left out):

| Thread | Threads seen | Thread length (median) | Appease level 1 affordable | Appease level 2 affordable | Brace level 1 affordable |
|--------|--------------|------------------------|----------------------------|----------------------------|--------------------------|
| Iron Era | 7 | 2.0 h | 7 of 7, on arrival | 6 of 7 | 6 of 7: 3 on arrival |
| Steel Era | 6 | 2.6 h | 5 of 6, on arrival | 4 of 6 | 6 of 6: 5 on arrival |
| Electric Era | 8 | 5.7 h | 6 of 8: 2 on arrival, the rest after 1.9 h to 12.9 h | 2 of 8 | 8 of 8: 7 on arrival |
| Digital Era | 8 | 12.0 h | 5 of 8, after 2.7 h to 11.0 h | 3 of 8 | 7 of 8: 6 on arrival |
| Neon Era | 6 | 19.5 h | 6 of 6: 4 on arrival, the others within 8.3 h | 4 of 6 | 6 of 6, on arrival |
| Cosmic Era: the Reality Tear, at its old Brace price of 1.56T dark matter and 75.6B titanium, and with Appease level 2 at double | 6 | 5.4 h | 3 of 6: 2 on arrival, one after 13.3 h | 2 of 6 | 3 of 6, on arrival |
| All dooms | 41 | | **32 of 41** (20 on arrival) | **21 of 41** | **36 of 41** |
| Cosmic Era: the Last Passage, at its old prices: Appease 3.1B faith and 48B culture, Brace 1.56T dark matter and 75.6B titanium (to a Quantum Age prestige) | 24 runs | 5.9 days | 2 of 24 | 2 of 24 | 24 of 24, on arrival |

A doom is fated in only 27% of eras, so these are 41 dooms across 24 runs: read the shares, not the exact counts. Before Appease was priced on the warning, the same bot on the same seeds could afford level 1 in 15 of 46 warnings and level 2 in 6, nearly always from faith it already held, against Brace in 41.

Level 1 stayed out of reach in 9 warnings. Four of them gave no time to work with: the harbinger came at the very advance its doom struck on, which leaves only what you already hold (see [No Outrunning a Doom](#no-outrunning-a-doom)). In the other five the bot, which keeps faith only for age requirements and the Sistine Chapel, ended the warning with 30% to 90% of the price in whichever of faith and culture it was shorter of.

The Last Passage's row was measured at its old prices. Appease cost 3.1B faith and 48B culture: the bot had the faith in most runs and almost never the culture, so it could pay in 2 runs of 24. Brace cost the era's 1.56T dark matter and 75.6B titanium, which the bot held on arrival every time. Neither is the price now. Appease level 1 is 1.3B faith and 20B culture, under half of what it was. Brace level 1 is 21Q dark matter and 27Q titanium, about 13,000 times the dark matter and 330,000 times the titanium. The new prices have not been run through the bot yet, so there is no count for them: what they rest on is the rule above. A player who keeps about five copies of each producer running gathers Brace level 1 in about 22 hours of the Interstellar Age and Appease level 1 in about 33. One who has let culture lie, or who spends titanium as fast as it comes, does not, and has to stay longer or go without.

The Reality Tear's row is from the same runs, when its Brace was the era's price and its second Appease level cost double. Its Brace level 1 is now 11Q dark matter and 14Q titanium, about eleven hours of a moderate economy's income, and has no count yet either.

All of these runs were played while the faith bands read storage, so every doom in them rolled at the bottom band's 90% (see [Why it isn't a share of storage](faith.md#faith-threshold-bands)). That is still what a typical town rolls at, and the bot is one.

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

A doom's harbinger gives you a fifth to three fifths of an age. Appease used to cost a quarter of what the whole era makes in faith, which no warning is long enough to earn: it could only be paid from faith saved before anyone had warned you, and in test games it was affordable in a third of the warnings (15 of 46). Now level 1 is 15% of what the age makes, three quarters of the shortest warning, so the warning itself can pay for it, and level 2 is the stretch.

The Last Passage's thread kept an era price for longer, because it lasts days, not hours: a quarter of what the Interstellar, Galactic and Quantum Ages make together, 3.1B faith and 48B culture. That is 1.8 times the culture the whole Interstellar Age makes, so the price of lowering the odds on the roll that ends your run was most of the era, and in test games it could be paid in 2 runs of 24. It now follows the rule every other thread does, with two thirds of the arrival age as its warning: dearer than any doom's Appease, since it guards the end of the run, but something the age it arrives in can pay for.

The price is fixed when the harbinger arrives, so it doesn't climb if the thread carries into the next age, and it comes from what the age makes, not from your storage: a harbinger who comes while your stores are small is no cheaper to answer than the age warrants.

Brace is priced the same way. It used to be priced by the era: 12% of the most the era's advances ask of each resource, the same in every age of it. Requirements grow far slower than income, so the price collapsed as the ages passed. The Electric Era's (56.4M steel, 924K oil, 3.96M electricity) took over three hours of a moderate Victorian Age's income, seven seconds of the Electric Age's and two of the Atomic Age's, and the Digital Era's asked the Modern Age for 19.2B data, which it makes three of a tick. Now every doom's Brace is a third of what its shortest warning makes, in materials the age makes, so it takes the same share of the warning in every age. The Iron Age's price is about what it was (24 minutes of income then, 27 now), with more of it in gold and iron and less in stone.

The Cosmic Era's threads got there first. Both once shared the era's price, 12% of the most its advances ask of the resources you have held since the Interstellar Age (1.56T dark matter and 75.6B titanium), and the smoke-test bot could brace against the Last Passage on arrival in 24 runs of 24. Next to a Brace that close to free, Appease bought little: with Brace level 2 held, Appease level 1 took the expected loss from 2.25% of the run's points to 1.35%, for two days of culture.

Second levels cost the same again as the first, in every era. They used to cost double, and a second level bought less for more: per hour of income the Last Passage's Brace level 2 saved three eighths of what level 1 did, and any thread's Appease level 2 three tenths. At the same price they save three quarters and three fifths, and a doom's Brace level 2 softens an Endure by as much as level 1 for the same price. The Cosmic Era's threads changed first; every other era's followed.

**Brace or Appease against a doom.** Appease level 1 takes two fifths off the chance the doom strikes. Brace level 1 softens the Endure if it does: a quarter of the buildings that would fall stand, and you keep 30% of your stock in place of 15%. Appease is the dearer (three quarters of the shortest warning in faith and culture, against a third of it in materials), and per hour of income it saves about seven tenths of what Brace does. They are paid in different resources, so a town that keeps faith buildings and producers both running can hold both.

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

While the Reality Tear's harbinger speaks, the Last Passage's thread waits behind it with its answers intact. Appease, Brace and Invite go to the thread that is speaking, and the panel says "The Last Passage still waits at your next prestige. Your answers to it stand." When the doom resolves (struck or spared), the Last Passage's thread takes up the warning again in the current age's voice, with a log line if the age moved on while it waited. The `catastrophe` command and the Epoch panel show both: the doom's warning, for example "Your future self warns of doom before this age is out: 90% catastrophe chance (high), faith strength 18% (devotion 0.8x, 100% of your faith kept).", and the Last Passage at its own odds.

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

- **Keep some faith in hand.** A harbinger gives you 20% to 60% of an age's time before its doom strikes, and sometimes less. Faith you already hold pays for Appease level 1 the moment the harbinger comes, and is what puts level 2 within reach. It does not buy better odds by itself: the strike chance (90%, 75% or 60%) follows your [faith strength](faith.md#faith-threshold-bands), which is about how much your town has put into faith buildings, and a typical town rolls at 90%.
- **Answer before the moment.** The price doesn't change once the harbinger has come, but the thread lasts only until the doom's moment, and advancing out of the era, or out of the age a figure named, brings the strike to that advance (in the Cosmic Era, so does prestige).
- **Want to keep your run?** Appease is the direct answer. One level cuts the chance it strikes by 40%, and it can't hurt the odds.
- **Worried but short on faith?** Brace instead. It doesn't lower the odds, but it makes Endure much cheaper if the catastrophe comes, and it is paid in materials your producers are making anyway: a third of the shortest warning's worth.
- **Early warnings deserve a little doubt.** Before the Industrial Age a warning may be false: in the Stone Era it always is, and in the Iron and Steel Eras now and then. Your own faith strength tells you the real band (see [Faith and the odds](faith.md#faith-threshold-bands)).
- **Prestiging from the Cosmic Era?** An open Reality Tear settles first, so answer its harbinger if one is speaking. The Last Passage then costs you points, not buildings. On its thread, Appease to lower the odds, Brace to keep more points on an Endure, Invite if you want the Cosmic Legacy. Both answers cost hours of income, about 22 for each level of Brace (dark matter and titanium) and 33 for each level of Appease (faith and culture): keep your faith and culture buildings going from the moment you enter the era, and set titanium aside. Per hour of income Brace is the better value at both levels, but the two are paid in different resources, and faith you keep also holds up your faith strength (see [Brace or Appease against the Last Passage](#brace-or-appease-against-the-last-passage)).
- **Hunting legacy bonuses?** Invite, then Succumb when it strikes. Skip Appease; it is refused after Invite anyway. Brace only if you think you might change your mind and Endure. You can't choose where a doom is fated, so collecting every era's legacy takes several runs.
