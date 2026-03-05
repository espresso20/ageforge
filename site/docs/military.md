# Military & Expeditions

The military system lets you send soldiers on expeditions for resources, gold, and knowledge. Higher-risk expeditions yield better loot — but failure means lost soldiers.

---

## Soldiers

Soldiers are workers in the **military domain**, unlocked in the **Iron Age** (requires Barracks).

```
recruit military       # recruit a military worker
assign barracks 5      # assign them to a barracks
```

Your **soldier count** determines which expeditions you can launch. Your **defense rating** reduces damage from raids and negative events.

---

## Expeditions

Only **one expedition** can be active at a time. Check the **F3: Military** tab for availability, soldier requirements, duration, and difficulty.

```
expedition <key>
```

Each expedition has:
- **Soldiers needed** — minimum to launch
- **Duration** — ticks until completion
- **Difficulty** — risk of failure (low/medium/high)
- **Loot** — resources awarded on success

---

## Expedition list

| Name | Key | Soldiers | Duration | Difficulty | Loot |
|---|---|---|---|---|---|
| Small Raid | `small_raid` | 3 | 20t | Low | Food, Wood |
| Ruins Delve | `ruins_delve` | 5 | 30t | Low | Stone, Knowledge |
| Bandit Hunt | `bandit_hunt` | 8 | 40t | Medium | Gold, Iron |
| Ancient Tomb | `ancient_tomb` | 10 | 50t | Medium | Knowledge, Gold |
| Orc Stronghold | `orc_stronghold` | 15 | 60t | Medium | Iron, Steel |
| Dragon's Lair | `dragon_lair` | 20 | 80t | High | Gold, Jewels |
| Lost City | `lost_city` | 25 | 100t | High | Knowledge, Culture |
| Undead Fortress | `undead_fortress` | 30 | 120t | High | Faith, Steel |
| Pirate Cove | `pirate_cove` | 20 | 60t | Medium | Gold, Spices |
| Rival Kingdom | `rival_kingdom` | 40 | 150t | High | Gold, Steel, Food |
| Orbital Debris | `orbital_debris` | 50 | 200t | High | Titanium, Data |
| Alien Outpost | `alien_outpost` | 60 | 250t | Very High | Dark Matter, Titanium |
| Rogue AI | `rogue_ai` | 70 | 300t | Very High | Data, Quantum Flux |
| Neutron Star | `neutron_star` | 80 | 350t | Extreme | Antimatter, Dark Matter |
| Void Rift | `void_rift` | 100 | 400t | Extreme | Quantum Flux, Antimatter |

---

## Military power & success chance

**Military power** is a combined stat from:
- Number of soldiers assigned
- Military technology bonuses (Military Tactics, Siege Warfare, Gunpowder, etc.)
- Prestige `military_power` upgrade
- Milestone and wonder bonuses

Higher military power increases your success chance on expeditions, especially on difficult ones. Check the difficulty bar in the F3 tab — if it reads above 60%, consider building up more soldiers or researching more military techs before launching.

---

## Defense rating

Your defense rating is your passive protection against:
- Bandit Raid events
- Pirate Attack events
- Rival kingdom aggression

More soldiers and higher military tech = higher defense rating. A high defense rating reduces resource losses from negative events.

---

## Tips

- **Research Navigation early** — it gives +30% expedition reward before the Colonial Age
- **Rocketry** (Atomic Age) doubles military power and gives +50% expedition rewards — transformational
- **Grand Lighthouse wonder** gives +80% expedition reward — combine with Rocketry for massive loot
- Don't launch high-difficulty expeditions without researching the relevant military tech first — the failure penalty costs more than skipping the expedition
- Expeditions stack with trade routes for gold — run a Crypto Market trade route while an Orbital Debris expedition is in progress
