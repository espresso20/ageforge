# Villagers

Villagers are your workforce. They consume food and generate resources when assigned. Idle villagers produce nothing — always keep them busy.

---

## Population mechanics

- Each villager costs **food** to recruit (price scales with current pop)
- Each villager drains **0.5 food/tick** regardless of assignment
- **Population cap** is determined by your Housing buildings (Huts, Houses, Manors…)
- **Max pop** shown in the status bar — recruit up to that number

```
recruit villager
```

---

## Villager types

### Worker 🪓
*Unlocked: Primitive Age*

The foundation of your workforce. Workers gather the three basic resources.

| Assignment | Resource gathered |
|---|---|
| `assign worker food <n>` | Food |
| `assign worker wood <n>` | Wood |
| `assign worker stone <n>` | Stone |

Workers benefit from most production technologies and building bonuses.

---

### Shaman 🌿
*Unlocked: Primitive Age*

Shamans generate knowledge, which fuels your research queue.

| Assignment | Resource |
|---|---|
| `assign shaman knowledge <n>` | Knowledge |

Shamans are essential — a civilization that falls behind on knowledge will stall at age transitions. Aim for at least 2–3 shamans before your first age advance.

---

### Scholar 📚
*Unlocked: Bronze Age (requires Library)*

Scholars generate knowledge significantly faster than shamans, and can also gather culture in later ages.

| Assignment | Resource |
|---|---|
| `assign scholar knowledge <n>` | Knowledge |
| `assign scholar culture <n>` | Culture (Renaissance Age+) |

Replace shaman assignments with scholars as you unlock them — they're roughly 2× more effective.

---

### Soldier ⚔️
*Unlocked: Iron Age (requires Barracks)*

Soldiers enable military expeditions and contribute to your defense rating.

| Assignment | Effect |
|---|---|
| `assign soldier military <n>` | Increases defense rating |

You need a minimum number of soldiers to launch each expedition. Soldiers cannot be assigned to resource gathering.

---

### Merchant 💹
*Unlocked: Medieval Age (requires Market × 3)*

Merchants boost your gold income and trade route effectiveness.

| Assignment | Resource |
|---|---|
| `assign merchant gold <n>` | Gold |
| `assign merchant culture <n>` | Culture |

Merchants scale especially well with the Banking and Mercantilism technologies.

---

### Engineer ⚙️
*Unlocked: Industrial Age (requires Factory)*

Engineers reduce building construction time and generate electricity in the Electric Age+.

| Assignment | Effect |
|---|---|
| `assign engineer build <n>` | Speeds up active build queue |
| `assign engineer electricity <n>` | Electricity (Electric Age+) |

---

### Scientist 🔬
*Unlocked: Information Age (requires Research Lab)*

Scientists dramatically accelerate knowledge generation in the late game.

| Assignment | Resource |
|---|---|
| `assign scientist knowledge <n>` | Knowledge (×5 rate multiplier) |

---

### AI Agent 🤖
*Unlocked: Digital Age (requires AI Lab)*

AI Agents autonomously generate data and can be assigned to any resource type at reduced efficiency compared to specialist humans.

| Assignment | Resource |
|---|---|
| `assign ai_agent data <n>` | Data |
| `assign ai_agent knowledge <n>` | Knowledge |

---

## Tips

- **Never leave villagers idle** — check the `Idle` count in the status bar
- **Shamans first** — early knowledge starvation is the most common mistake
- **Food balance** — every villager you recruit increases food drain. Always check `Food: -N/t` before recruiting
- **Soldiers on demand** — keep enough soldiers to run expeditions, but don't over-invest early
- **Merchants scale late** — in the Colonial/Industrial ages, a dozen merchants can outproduce a mine entirely

```
# Check who is idle
> e       (switch to Economy tab)

# Assign idle workers
> assign worker food 2
> assign shaman knowledge 1
```
