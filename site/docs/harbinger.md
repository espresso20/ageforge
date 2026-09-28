# The Harbinger

Harbingers are figures who walk with you through an epoch whose passage can bring a [catastrophe](catastrophe.md), and tell you how worried to be. For most epochs the passage is the crossing into the next epoch. For the Cosmic Era, the last one, it is your next prestige: the [Last Passage](#the-last-passage). You can pay to lower the odds, pay to soften the blow, or invite the catastrophe on purpose.

A harbinger never blocks anything and never expires. You can ignore it completely and the game plays on as normal.

---

## When Harbingers Come

Each epoch whose passage can bring a catastrophe gets one **harbinger thread**. The thread starts when you enter the epoch's **first age** and lasts until the passage settles it: crossing into the next epoch, or in the Cosmic Era, confirming prestige.

The speaker changes with the age. Each time you advance within the epoch, that age's figure takes up the warning:

| Epoch | Figures, in order | Warns about the passage into |
|-------|-------------------|------------------------------|
| Stone Era | the Wild Man, the Hermit, the Soothsayer | Iron Era |
| Iron Era | the Desert Prophet, the Oracle, the Town Crier | Steel Era |
| Steel Era | the Court Astrologer, the Pamphleteer, the Newsboy | Electric Era |
| Electric Era | the Doomsayer, the Telegraph, the Civil Defence Broadcast | Digital Era |
| Digital Era | the Evening News, the Chain Email, the Viral Video | Neon Era |
| Neon Era | the Ghost in the Net, the Reactor Warden, the Deep Space Monitor | Cosmic Era |
| Cosmic Era | the Distress Beacon, the Elder Relay, your future self, your unmade self | the Last Passage (your next prestige) |

Rules:

- **One thread per epoch per run.** Succumb and prestige start a new run, so the threads start again.
- **A new run starts with one.** In a new game, and after Succumb or prestige, the Stone Era thread starts on the first tick. The Wild Man greets every new game.
- **Loading a save counts.** If a save sits in any age of a qualifying epoch without a thread, one starts there when the save loads, with that age's figure.
- **Only when a catastrophe is possible.** If the epoch's transition can't bring one (it already rolled this run, for example), no thread starts.
- **The Cosmic Era warns of prestige.** It is the last epoch, so its passage is the end of the run. See [The Last Passage](#the-last-passage).

When a thread starts, or a new figure takes it up, you get:

- a log entry naming the figure and the epoch it warns about, followed by its arrival and warning lines,
- a toast ("... has come" for the first figure, "... takes up the warning" for the rest),
- a **⚑ HARBINGER** badge in the status bar for as long as the thread lasts.

Type `harbinger` (or `harb`) to open the Harbinger panel. It shows the current figure and, after a handoff, who it took up the warning from.

---

## The Roster

Every age has a figure written for it, and all 22 appear in play: three for each epoch from the Stone Era to the Neon Era, and four for the Cosmic Era, whose thread warns of the Last Passage.

### Who they are

| Age | Harbinger | Appears? | Forecast | False-prophet chance if it starts the thread | Who they are |
|-----|-----------|----------|----------|----------------------------------------------|--------------|
| Primitive | the Wild Man | Yes | Vague | 8/64 (12.5%) | A man who lives past the last fire walks in from the wilderness, grey with ash, to say what he has seen. |
| Stone | the Hermit | Yes | Vague | 7/64 (10.9%) | Comes down from the high caves once in a generation, and never with good news. |
| Bronze | the Soothsayer | Yes | Vague | 6/64 (9.4%) | Reads the future in knucklebones, sparrows and goat livers, and wants paying before and after. |
| Iron | the Desert Prophet | Yes | Vague | 5/64 (7.8%) | Walks in from the dry country with sand in his beard and one message for the city. |
| Classical | the Oracle | Yes | Vague | 4/64 (6.3%) | Speaks from the smoke over the cleft rock, through priests who charge by the question. |
| Medieval | the Town Crier | Yes | Vague | 3/64 (4.7%) | Rings his bell at the market cross and reads out doom in the voice he uses for tolls. |
| Renaissance | the Court Astrologer | Yes | Vague | 2/64 (3.1%) | Casts the prince's horoscope, and lately the prince's horoscope has been bad for everyone. |
| Colonial | the Pamphleteer | Yes | Vague | 1/64 (1.6%) | Prints doom on cheap paper and sells it outside the coffee house for a penny. |
| Industrial | the Newsboy | Yes | Numeric | 0 | Shouts the late edition from the corner, and the late edition has the odds printed on it. |
| Victorian | the Doomsayer | Yes | Numeric | 0 | Stands on a soap crate by the park railings with a sandwich board and a table of figures. |
| Electric | the Telegraph | Yes | Numeric | 0 | Chatters all night at the post office with dispatches from stations that have stopped answering. |
| Atomic | the Civil Defence Broadcast | Yes | Numeric | 0 | Three long notes on every wireless, then a calm voice reading the odds from a card. |
| Modern | the Evening News | Yes | Numeric | 0 | Leads with it at six, with a graphic, an expert and an anchor trying not to look worried. |
| Information | the Chain Email | Yes | Numeric | 0 | Forward this to ten people or it happens to you. It has a spreadsheet attached. |
| Digital | the Viral Video | Yes | Numeric | 0 | Shaky, portrait, a million views by lunch, and the odds on a whiteboard at the end. |
| Cyberpunk | the Ghost in the Net | Yes | Numeric | 0 | A dead corporate AI that leaks internal risk memos through the net, glitching on every third word. |
| Fusion | the Reactor Warden | Yes | Numeric | 0 | The plant's safety intelligence, which has never before spoken outside a scheduled drill. |
| Space | the Deep Space Monitor | Yes | Numeric | 0 | A station behind the moon that has watched one patch of sky for forty years, and has just marked a packet urgent. |
| Interstellar | the Distress Beacon | Yes | Numeric | 0 | Still looping from a colony that went silent eighty years ago, and the loop has changed. |
| Galactic | the Elder Relay | Yes | Numeric | 0 | An alien relay older than the species that found it, speaking in geometry for the first time in an age. |
| Quantum | your future self | Yes | Numeric | 0 | A message in your handwriting, stamped nine years from now, that knows your passcode. |
| Transcendent | your unmade self | Yes | Numeric | 0 | A version of you from a branch that ended, come to see whether this one ends the same way. |

The false-prophet chance only counts for the figure who starts the thread. Normally that is the epoch's first age; the other chances only matter when a save loads mid-epoch without a thread. See [False prophets](#false-prophets).

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

The figure's words tell you how bad the risk is. How much detail you get depends on who is speaking now.

**Before the Industrial Age**, the panel shows only the figure's words and a vague severity: **low**, **medium** or **high risk**.

**From the Industrial Age on**, the panel also prints the odds, for example `Odds published: 9%`. The number already includes any Appease you have bought. In the Steel Era this means the thread starts vague and becomes numeric when the Newsboy takes it up.

The severity follows the catastrophe chance for the transition:

| Chance | Severity shown |
|--------|----------------|
| under 14% | low risk |
| 14% to under 17% | medium risk |
| 17% or more | high risk |

Without any Appease that works out to **low** at high faith (12%), **medium** at mid faith (15%) and **high** at low faith (18%). See [Faith and the Odds](catastrophe.md#faith-and-the-odds). The severity is live: if your faith fill changes band, or you Appease, it moves.

### False prophets

A thread may be a lie. It rolls once, when its first figure arrives, using that figure's chance from the roster. For threads that start normally, in the epoch's first age:

| Thread | Chance it is false |
|--------|--------------------|
| Stone Era (starts with the Wild Man) | 8/64, 12.5% |
| Iron Era (starts with the Desert Prophet) | 5/64, about 7.8% |
| Steel Era (starts with the Court Astrologer) | 2/64, about 3.1% |
| Electric, Digital, Neon and Cosmic Eras | none |

A false thread claims **medium** or **high** risk (picked at random when it starts), whatever the real odds are. Every figure in that epoch repeats the same false claim, using the same warning lines a real harbinger would. The `catastrophe` command and the Epoch tab repeat the warning too, so you can't tell a false thread apart from the screen.

The claim is kept as a fixed multiple of the real chance. Appease, a change in faith, or an Invite move it exactly as they would move a real warning. It never drops below low risk. A false Steel Era thread that reaches the Newsboy prints the claimed figure, not the real one.

The truth comes out at the transition. Appease still lowers the real odds when the thread is false, and Brace still works if the catastrophe happens to come anyway.

---

## Answering the Harbinger

While a thread lasts you can Appease, Brace or Invite, in any age of the epoch. Your answers belong to the passage, not to the figure: levels you buy and an Invite carry over when the next figure takes up the warning. Nothing expires.

The price is set by the passage, so it is the same in every age of the epoch. If your storage can't hold the price yet, the refusal tells you how much storage you need.

### Appease: lower the odds

| Level | Cost | Real catastrophe chance |
|-------|------|-------------------------|
| 0 | none | unchanged |
| 1 | a quarter of the thread's faith income, in faith (and a quarter of its culture income, in culture, from the Steel Era on) | ×0.6 |
| 2 | double level 1 | ×0.36 (×0.6 again) |

- The **thread's income** is what a player who invests moderately in faith makes over the epoch's ages, each at its target length (see [How Long Each Age Takes](ages.md#how-long-each-age-takes)): five fully staffed copies of every faith building so far, the faith of every wonder already built, the flat faith of the techs, all multiplied by the production bonus of the techs and wonders you hold by then (it reaches the ×3 cap in the Electric Age). Culture is counted the same way. So level 1 comes partway through the epoch, and level 2 on top of it before the passage, if you keep a few faith buildings staffed.
- Culture is only charged if you already had it when the epoch began. It unlocks in the Classical Age, so Stone and Iron Era threads cost faith only.
- Two levels at most.
- Appease always changes the real odds, including with a false prophet.
- **Not after Invite.** Once you invite the catastrophe, Appease is refused: it will come whatever you offer.

Example: at mid faith the base chance is 15%. One level of Appease takes it to 9%, two levels to 5.4%.

Spending faith lowers your faith fill, and that can drop you into a worse faith band for the roll. Appease still comes out ahead. In the worst case one level leaves the odds at 0.9× of where they started, so every level lowers the real chance.

### Brace: soften an Endure

Brace spends the resources the epoch still asks of you. For each resource, the price is based on the most the epoch asks of it across its remaining advances (into its later ages and into the next epoch). Only resources you already had when the epoch began count, and faith and culture are left out.

| Level | Cost | If you Endure: buildings destroyed | If you Endure: stored resources kept |
|-------|------|------------------------------------|--------------------------------------|
| 0 (unbraced) | none | 20% | 15% |
| 1 | 12% of the most the epoch asks of each of those resources | 15% | 30% |
| 2 | 24% of the same | 10% | 45% |

Buildings destroyed are counted from your non-wonder buildings, rounded down, with at least 1 if you have any. Wonders are never destroyed. The rest of Endure (25% of workers lost, the 216-tick reconstruction debuff, −10 morale) is the same at every Brace level. See [Endure](catastrophe.md#endure).

Things to know:

- Two levels at most.
- Brace only matters if the catastrophe comes **and** you choose Endure. If no catastrophe comes, or you Succumb, the resources are simply spent.
- The Brace is attached to the pending catastrophe. If you close the choice with Esc and Endure later, or save and load in between, it still applies.
- Brace is allowed after Invite.
- **In the Cosmic Era, Brace protects points, not your civilization.** An Endure at the Last Passage keeps 50% of the run's prestige points unbraced, 70% at level 1 and 85% at level 2. The building and resource numbers above don't apply there. See [The Last Passage](prestige.md#the-last-passage).

### What it costs, by epoch

Level 1 prices. Level 2 costs double.

| Thread | Appease (level 1) | Brace (level 1) |
|--------|-------------------|-----------------|
| Stone Era | 59 faith | 9,600 food, 4,800 wood, 2,400 knowledge |
| Iron Era | 5,400 faith | 26,400 knowledge, 26,400 stone, 6,360 iron, 21,600 gold |
| Steel Era | 74K faith, 770K culture | 3.6M knowledge, 1.8M gold, 288K steel |
| Electric Era | 1.2M faith, 16M culture | 56.4M steel, 924K oil, 3.96M electricity |
| Digital Era | 12M faith, 180M culture | 156M gold, 117.6B electricity, 19.2B data |
| Neon Era | 130M faith, 2B culture | 288B electricity, 46.8B data, 3B crypto |
| Cosmic Era | 1.2B faith, 19B culture | 1.56T dark matter, 75.6B titanium |

Appease used to cost 15% of the passage storage (12,000 faith in the Stone Era, 147B in the Digital Era), which faith could almost never reach once ages became short.

The Cosmic Era's passage is prestige, which you may take in any of its ages. Its Appease counts the Interstellar, Galactic and Quantum Ages, so if you prestige as soon as you arrive, Appease is out of reach; stay a day or two and it isn't. Its Brace is based on the most the Cosmic Era's own advances ask of each resource you have held since the Interstellar Age: 13T dark matter (for the Quantum Age) and 630B titanium (for the Galactic Age). Antimatter and quantum flux arrive later, so they are left out.

#### Can you afford it?

Yes, both, if you invest a little in faith. Faith can't be bought at the market (culture can, gold → culture), so Appease is priced off what faith buildings actually make. In the smoke-test bot's runs (five seeds, the bot building toward Appease and buying each level as soon as it could), the median thread went like this:

| Thread | Thread length | Level 1 affordable | Level 2 affordable (after level 1) |
|--------|---------------|--------------------|------------------------------------|
| Stone Era | 3.2 h | after 58 min (30%) | after 1.6 h (50%) |
| Iron Era | 9.2 h | after 2.6 h (28%) | after 5.9 h (64%) |
| Steel Era | 23 h | after 8 h (34%) | after 19 h (80%) |
| Electric Era | 1.1 days | after 11 h (44%) | after 22 h (85%) |
| Digital Era | 1.8 days | after 9 h (21%) | after 23 h (56%) |
| Neon Era | 3.0 days | after 7 h (9%) | after 1 day (35%) |
| Cosmic Era (to a Quantum Age prestige) | 3.3 days | after 11 h (14%) | after 1.8 days (57%) |

From the Digital Era on, faith saved in earlier eras carries over (faith is never reduced at an advance), so Appease comes earlier in the thread. Brace is priced in construction resources and became affordable within 0 to 6 hours of each thread starting.

Faith you spend on Appease is faith the gate can't use: keep what the next age asks for, and what the Sistine Chapel still needs, before you appease.

### Invite: choose the catastrophe

- Free.
- Guarantees the catastrophe at this transition.
- **Can't be undone.** It stays with the thread when the next figure takes over.
- After inviting, Appease is refused. Brace is still allowed.

Invite is for players who want to Succumb on purpose, to collect an epoch's legacy bonus and Ancient Knowledge. It is the only way to choose a catastrophe; there is no command to trigger one directly. See [Succumb](catastrophe.md#succumb).

In the Cosmic Era, Invite means your next prestige brings the Last Passage. It is how you choose the [Cosmic Legacy](prestige.md#cosmic-legacy) on purpose.

### Why the price is tied to the passage

The price comes from what the passage asks, not from your current storage. Your storage is smallest in the epoch's first age, so a price based on it would make paying early a discount. Tying it to the passage keeps it the same in every age of the epoch, whoever is speaking.

---

## At the Transition

When you cross into the next epoch the thread is settled, in the voice of the last figure, and the log records one of four verdicts:

| Verdict | What happened |
|---------|---------------|
| **Fulfilled** | You invited the catastrophe, and it came. |
| **Vindicated** | It warned you, you didn't invite it, and the catastrophe came. This also covers a false thread whose warning came true by chance; the log notes that the warning had been invented. |
| **Spared** | A real warning, and no catastrophe came. |
| **Discredited** | A false prophet, and no catastrophe came. |

If the catastrophe came and you had braced, the log also says what Endure will cost you at that Brace level. Then the thread ends and the catastrophe choice works as usual (see [Catastrophe System](catastrophe.md)). If the new epoch can also bring a catastrophe, its own thread starts with its first figure.

The Epoch tab (`epoch`) shows the current thread's status and, for past threads this run, the whole chain of figures and the verdict.

### The Last Passage

The Cosmic Era's thread is settled when you confirm prestige, not at an age advance. The panel names its passage: "Warning of the Last Passage: the end of this civilization, when you next prestige." When the thread starts, the log reads `⚑ <Figure> has come, warning of the Last Passage.`

Confirming prestige rolls once, with the usual odds for the faith band (18%, 15% or 12%), times 0.6 per level of Appease, or certain if you invited it. If nothing comes, the verdict is Spared ("The Last Passage opens, and nothing comes through it. ...") and prestige completes. If it comes, prestige waits for you to Endure or Succumb. Endure keeps part of the run's points and Succumb grants the Cosmic Legacy. See [The Last Passage](prestige.md#the-last-passage).

---

## Saving, Succumb and Prestige

- The thread is saved with your game: the figures who have spoken, what the current one said, your Appease and Brace levels, and whether you invited the catastrophe. A Brace attached to a pending catastrophe is saved too.
- **Succumb** clears the live thread, the invite and any Brace. Past verdicts are kept, like the epoch event history. The new run's Stone Era thread starts on the first tick.
- **Prestige** clears all of it, past verdicts included, and the Wild Man greets the new run.
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

With no harbinger present, the panel says so and describes the outlook for your next passage in plain words.

---

## Strategy

- **No hurry.** The price is the same in every age of the epoch, so pay whenever your storage (and, for Appease, your faith) can cover it. The last age of the epoch is fine.
- **Want to keep your run?** Appease is the direct answer. One level cuts the risk by 40%, and it can't hurt the odds.
- **Worried but short on faith?** Brace instead. It doesn't lower the odds, but it makes Endure much cheaper if the catastrophe comes.
- **Early warnings deserve a little doubt.** A Stone, Iron or Steel Era warning of high risk may be false. Your own faith fill tells you the real band (see [Faith and the Odds](catastrophe.md#faith-and-the-odds)).
- **Prestiging from the Cosmic Era?** The Last Passage costs you points, not buildings. Appease to lower the odds, Brace to keep more points on an Endure, Invite if you want the Cosmic Legacy.
- **Hunting legacy bonuses?** Invite, then Succumb at the transition. Skip Appease; it is refused after Invite anyway. Brace only if you think you might change your mind and Endure.
