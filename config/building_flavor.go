package config

// building_flavor.go holds the cosmetic "Flavor" line for buildings — the
// humor/personality layer (pass 9a). These strings are ADDITIVE only: they
// never carry functional information (costs, rates, worker counts, effects),
// which all live in each building's Description. A building absent from this
// map simply renders no flavor line, so partial coverage is always safe.
//
// Voice: dry, occasionally absurdist, age-appropriate. Early ages read like
// caveman field notes; late ages drift toward existential dread. The joke is
// always on the game world, never on the player. Keep each line one sentence,
// plain text — no color tags or format directives (a test enforces this).

// buildingFlavor maps a building key to its flavor line. Keep entries grouped
// by age for readability; coverage is broad but intentionally not exhaustive.
var buildingFlavor = map[string]string{
	// ===== PRIMITIVE AGE — caveman field notes =====
	"hut":            "Four walls and the ambition of someday having five.",
	"stash":          "A hidden pile of supplies, hidden mostly from yourself.",
	"gathering_camp": "Where the bold venture out to bravely pick up things off the ground.",
	"wood_camp":      "It turns out trees are just wood that hasn't given up yet.",
	"story_circle":   "Where elders gather to definitely not make things up.",
	"shrine":         "A pile of nicer rocks than the rocks around it.",
	"sacred_grove":   "The first thing your people agreed not to set on fire.",

	// ===== STONE AGE =====
	"longhouse":       "Like a hut, but long enough that you can't see the other end's problems.",
	"forager_post":    "Professional gatherers. Same job as before, now with a title.",
	"stone_camp":      "We have decided rocks are a career and not just a hobby.",
	"stone_pit":       "A hole. We are extremely proud of this hole.",
	"woodcutter_camp": "The trees saw us coming. The trees were right.",
	"elders_hall":     "Where the oldest among you are kept dry and asked to remember things.",
	"standing_stones": "Heavy rocks, stood up on purpose. Don't ask how. Don't ask why.",
	"war_camp":        "Your civilization's solution to diplomacy.",
	"storage_pit":     "A bigger hole for the things that didn't fit in the first hole.",
	"great_monolith":  "One enormous stone, placed by many people who are no longer speaking.",

	// ===== BRONZE AGE =====
	"house":       "An upgrade from the hut, now with the radical innovation of a floor.",
	"farm":        "We stopped chasing food and convinced it to stay in one place.",
	"lumber_mill": "The trees have unionized, but we have a bigger saw.",
	"quarry":      "An industrial-scale hole. The hole has become a lifestyle.",
	"scriptorium": "Where we write things down so we can forget them with confidence.",
	"altar":       "Now with a roof, so the gods can hear you complain in comfort.",
	"barracks":    "Where we teach young people that the spear goes the other way.",
	"market":      "Two people discover that wanting different things can be profitable.",
	"smithy":      "We hit hot metal until it agrees to be a useful shape.",
	"warehouse":   "All your worldly goods, finally indoors and slightly judgmental.",
	"stonehenge":  "A calendar, a temple, or a very committed prank. History declines to say.",

	// ===== IRON AGE =====
	"townhouse":     "Walls thick enough that the neighbors are now a rumor.",
	"field_works":   "Farming, but organized by someone with a clipboard.",
	"granary":       "A fortress whose sole purpose is keeping the future fed.",
	"hunting_lodge": "Where men gather to discuss the deer that got away, in detail.",
	"ironworks":     "Bronze was a phase. We've moved on to something with edge.",
	"legion_fort":   "Discipline, formation, and the firm belief that walking in step wins wars.",
	"marble_quarry": "We dig up the prettiest rocks so future tourists have something to admire.",
	"smelter":       "Where rocks are persuaded, with heat, to give up the metal they were hiding.",
	"temple":        "A house for the gods, conveniently sized for collecting offerings.",
	"timber_yard":   "Logs, sorted by size, by people who take logs very seriously.",
	"trading_post":  "The frontier of commerce, where everyone is a little overcharged.",
	"agora":         "Public square, public debate, public refusal to reach any conclusion.",
	"colosseum":     "Mass entertainment, with the entertainment being mostly other people's misfortune.",

	// ===== CLASSICAL AGE — philosophy creeps in =====
	"villa":            "A house so refined it has rooms whose purpose no one can explain.",
	"estate_farm":      "Agriculture, now with enough surplus to support a man who only thinks.",
	"library":          "Where knowledge is stored, organized, and almost never returned on time.",
	"oracle_house":     "Predictions vague enough to be right no matter what happens.",
	"military_academy": "Where war becomes a subject with homework.",
	"merchant_quarter": "An entire district devoted to the noble art of marking things up.",
	"aqueduct":         "Water, persuaded to walk uphill against its better judgment.",
	"forge":            "Where the smith graduates from tools to genuine ambition.",
	"amphitheater":     "Drama, comedy, and the occasional tragedy that wasn't in the script.",
	"parthenon":        "Perfect proportions, built by people who refused to round anything off.",
	"cultural_obelisk": "A tall stone finger pointing at the sky, in case the sky forgot we exist.",
	"marble_works":     "Turning rough stone into things rich people pretend they understand.",
	"wood_workshop":    "Where carpentry stops being survival and starts being showing off.",

	// ===== MEDIEVAL AGE =====
	"manor":             "The lord's house, distinguished from yours by everything.",
	"keep":              "A vault disguised as a tower disguised as a threat.",
	"demesne":           "The lord's personal land, worked by people who'd rather not discuss it.",
	"sawmill":           "Water does the cutting now, which suits everyone except the river.",
	"stonemasons_guild": "A brotherhood of men who know secrets about rocks and won't share.",
	"monastery_library": "Monks copying books by hand, one illuminated mistake at a time.",
	"cathedral":         "We spent a century building upward to feel closer to heaven. Or taller than the next town.",
	"castle_keep":       "Stone, walls, and the comforting math that the attackers have to come uphill.",
	"guildhall":         "Where tradesmen gather to fix prices and call it standards.",
	"workshop":          "Where things get made by people who'll insist it's an art.",
	"ironmonger":        "Iron, refined past the point of need, purely out of pride.",
	"great_hall":        "A room large enough to host a feast and the grudges that follow it.",
	"great_library":     "Every book worth having, guarded by people who'd rather you didn't touch them.",

	// ===== RENAISSANCE AGE =====
	"art_studio":     "Where genius is produced on a deadline and a patron's budget.",
	"basilica":       "A church so grand that humility had to wait outside.",
	"coal_mine":      "We discovered that the past, compressed long enough, burns beautifully.",
	"estate":         "A home large enough to get lost in, which is rather the point.",
	"exchange":       "Where money is made by people who never touch the thing they're selling.",
	"fortress":       "A castle that read about cannons and took the news badly.",
	"foundry":        "Industrial-scale metalwork for an age that suddenly wants everything cast.",
	"iron_mine":      "Digging deeper for iron, because the easy iron ran out a century ago.",
	"market_garden":  "Farming optimized for the city's appetite and the gardener's patience.",
	"mill":           "Machinery that turns grain into flour and millers into the village rich.",
	"university":     "Where the young go to learn, and the old go to avoid retirement.",
	"sistine_chapel": "A ceiling so magnificent that everyone leaves with a sore neck and faith renewed.",

	// ===== COLONIAL AGE =====
	"settlement_block":        "Housing built fast, on the assumption that the frontier won't wait.",
	"plantation":              "Vast, profitable, and a chapter the history books handle carefully.",
	"coal_works":              "Coal, mined at a scale that would have horrified our grandparents and funds our grandchildren.",
	"natural_philosophy_hall": "Science before it admitted it was science, still wearing a powdered wig.",
	"mission":                 "A church at the edge of the map, converting whoever it finds first.",
	"fort":                    "A defensive position, planted firmly on land that recently belonged to someone else.",
	"port":                    "Where the world's goods arrive, get taxed, and pretend they enjoyed the trip.",
	"harbor":                  "Where ships are kept, repaired, and gently reminded the sea is winning.",
	"dockyard":                "Where ships are born, screaming, in a chorus of hammers.",
	"iron_works":              "Iron production at imperial scale, smoke included at no extra charge.",
	"concert_hall":            "Where the wealthy gather to be seen not talking during the music.",
	"embassy":                 "Diplomacy's home turf — where wars are politely postponed.",
	"grand_lighthouse":        "A tower of light warning sailors of the rocks, and of the harbor's prices.",

	// ===== INDUSTRIAL AGE — the smoke begins =====
	"tenement":           "Housing stacked as high as the landlord's conscience would allow.",
	"agricultural_works": "Farming with machines, because the fields stopped being romantic and started being inventory.",
	"steam_coal_plant":   "Coal in, steam out, and a sky that's never quite the right color again.",
	"steam_mine":         "Mining powered by steam, so the work goes faster and the lungs go slower.",
	"research_institute": "Where progress is invented on schedule and patented by lunch.",
	"church":             "Faith, industrialized — same comfort, now available to a larger congregation.",
	"military_base":      "War, given a permanent address and a supply budget.",
	"stock_exchange":     "A room where fortunes are made, lost, and shouted about simultaneously.",
	"iron_works_complex": "Iron production so vast it has its own weather.",
	"steel_mill":         "Iron, but with ambition and a higher melting point.",
	"coal_plant":         "We light the whole city by setting the past on fire.",
	"opera_house":        "Three hours of magnificent singing in a language no one in the audience speaks.",
	"grand_embassy":      "Diplomacy at scale — now several wars can be postponed at once.",
	"crystal_palace":     "An entire building made of glass and confidence, daring the weather to comment.",

	// ===== VICTORIAN AGE — propriety and repression =====
	"row_house":       "Identical homes in a tidy row, where individuality is frowned upon but well-drained.",
	"mechanized_farm": "Machines do the harvesting now; the horses have filed no complaint.",
	"oil_derrick":     "We struck a black, flammable future and decided to drink deeply.",
	"uranium_mine":    "We're digging up a rock that glows. Surely that's fine.",
	"academy":         "Higher learning for an age that's quite certain it's the cleverest yet.",
	"grand_cathedral": "A cathedral so vast that God Himself might struggle to find a seat.",
	"garrison":        "Soldiers, permanently stationed, permanently bored, permanently armed.",
	"bank":            "Where your money goes to make more money for someone else.",
	"steam_works":     "Steam-powered everything, because the age refuses to do anything by hand.",
	"bessemer_plant":  "Steel by the ton, cheap enough that the future is now affordable.",
	"steam_turbine":   "We spin steam into power and call it the height of refinement.",
	"grand_museum":    "Where the empire displays the treasures of nations it visited uninvited.",
	"eiffel_tower":    "A tower the locals called an eyesore until they learned to charge admission.",

	// ===== ELECTRIC AGE — lightning, domesticated =====
	"apartment_block":          "Vertical living for thousands, each convinced their neighbor is the loud one.",
	"industrial_farm":          "Agriculture as industry, where the only thing not mass-produced is the nostalgia.",
	"oil_field":                "An entire landscape rented to the pursuit of combustible wealth.",
	"nuclear_extraction_plant": "We extract the glowing rocks at scale now, with only mild signage.",
	"physics_laboratory":       "Where reality is poked, prodded, and occasionally apologized to.",
	"revival_hall":             "Faith, electrified — now with lights, sound, and a collection plate that hums.",
	"command_post":             "Where wars are managed by men with maps and very good posture.",
	"financial_district":       "An entire neighborhood that produces nothing and owns everything.",
	"power_station":            "The building that keeps the lights on and the bills coming.",
	"electric_arc_furnace":     "We melt metal with lightning now, because fire felt too quaint.",
	"power_generator":          "Spinning copper into electricity, the modern miracle nobody thinks about until it stops.",
	"radio_station":            "Voices through the air to millions, most of them selling something.",
	"hoover_dam":               "We told an entire river to stop, and astonishingly, it did.",

	// ===== ATOMIC AGE — the dread arrives =====
	"housing_project":          "Affordable housing, designed by optimists, maintained by no one.",
	"agricultural_complex":     "Food production at a scale that makes famine an administrative error.",
	"petroleum_refinery":       "Where crude oil is refined into the hundred things we can't live without anymore.",
	"uranium_processing_works": "We refine the glowing rocks into something far more concentrated. Sleep well.",
	"research_campus":          "An entire campus dedicated to discovering things we may regret.",
	"spiritual_center":         "Faith, for an age that has seen what it can build and is quietly terrified.",
	"bunker_complex":           "A home built underground, for when the surface becomes a poor lifestyle choice.",
	"corporate_hq":             "A tower where decisions are made about people who will never be allowed inside.",
	"nuclear_plant":            "We boil water with the fundamental forces of the universe. To make tea, mostly.",
	"nuclear_reactor":          "A small contained sun, kept in a box, watched very, very closely.",
	"cinema":                   "A dark room where strangers agree to believe the same lie for two hours.",
	"particle_accelerator":     "We smash the smallest things together to find smaller things to smash.",
	"monument_of_ages":         "A monument to every age that came before, including the ones we'd rather forget.",

	// ===== MODERN AGE — everything is connected, regrettably =====
	"tower_block":       "A spire of glass where the rent climbs faster than the elevator.",
	"agri_complex":      "Farming, fully industrialized, where the soil is now a spreadsheet.",
	"oil_platform":      "A city on stilts in the sea, drinking the last of the ancient sunlight.",
	"think_tank":        "A building full of clever people paid to be confidently uncertain.",
	"meditation_center": "Where a frantic civilization pays to be told to sit still and breathe.",
	"special_ops_hq":    "Where the wars nobody admits to are planned by people nobody can name.",
	"investment_firm":   "They turn money into more money, and occasionally into a cautionary tale.",
	"seaport":           "Global trade, containerized — the entire world reduced to stackable boxes.",
	"titanium_mine":     "We dig for the metal of jets and ambition, deeper than any sane person would.",
	"tv_studio":         "Where reality is filmed, edited, and improved until no one recognizes it.",
	"space_program":     "We strap people to a controlled explosion and call it exploration.",
	"nano_foundry":      "We build machines too small to see, which is either reassuring or not.",

	// ===== INFORMATION AGE =====
	"smart_complex":  "A building that knows when you're home, which seemed convenient at first.",
	"smart_farm":     "Crops monitored by sensors, harvested by robots, and Instagrammed by no one who works here.",
	"precision_mine": "Mining guided by satellite, accurate to the centimeter and indifferent to the view.",
	"innovation_hub": "An open-plan room where disruption is brainstormed over very expensive coffee.",
	"digital_temple": "Worship, migrated online — the soul now has a login.",
	"cyber_command":  "Where wars are fought with keyboards and the casualties are all in the cloud.",
	"venture_hub":    "Where money chases ideas, and the ideas chase money, and occasionally they collide.",
	"server_farm":    "A barn full of humming machines, fed electricity, producing nothing you can hold.",
	"media_center":   "Content, manufactured around the clock, for an audience that is never, ever full.",
	"global_network": "We connected every mind on the planet. The planet has notes.",

	// ===== DIGITAL AGE — civilization uploads =====
	"megaplex":             "A self-contained city in a single building, so you never have to go outside again.",
	"nano_farm":            "Food assembled atom by atom, indistinguishable from the real thing it replaced.",
	"bio_fabrication_lab":  "We print living tissue now, which the ethics board is still discussing.",
	"ai_research_lab":      "We're teaching machines to think, and politely not asking what about.",
	"cyber_shrine":         "A place of worship for an age that isn't sure what it believes, but believes it efficiently.",
	"drone_warfare_center": "War, conducted remotely, by operators who clock out and go home for dinner.",
	"crypto_exchange":      "Where imaginary money is traded for slightly different imaginary money.",
	"neural_grid":          "A network that thinks for itself, which we're choosing to find helpful.",
	"data_center":          "The cloud, which is just someone else's enormous, anxious basement.",
	"vr_studio":            "Where we build worlds more pleasant than this one and sell tickets back in.",
	"world_simulation":     "We simulated an entire world. It has not yet noticed. We think.",

	// ===== CYBERPUNK AGE — high tech, low life =====
	"arcology_pod":          "A whole life lived inside one megastructure, sky optional, sunlight extra.",
	"vat_farm":              "Meat grown in tanks, ethically sourced from absolutely nothing.",
	"nanobot_vat":           "A churning vat of tiny machines, building things while you sleep. Probably for you.",
	"dark_crystal_mine":     "We mine crystals that hum at frequencies the human ear was wise to avoid.",
	"neuro_research_center": "Where they study the brain by, increasingly, opening it.",
	"neon_sanctuary":        "Worship in a city that never sleeps, prays, or turns off its signage.",
	"combat_aug_center":     "Where soldiers trade body parts for advantages and the warranty is questionable.",
	"black_market":          "Everything illegal, available, and only slightly more expensive than the legal version.",
	"augmentation_foundry":  "Where the line between person and product gets a little blurrier each quarter.",
	"holographic_theater":   "Entertainment so immersive that some patrons forget to leave. We're working on it.",
	"cyber_hub":             "The beating data-heart of the city, where information is the only real currency.",
	"neon_citadel":          "A fortress of light and chrome, beautiful from a distance, expensive up close.",

	// ===== FUSION AGE =====
	"habitat_ring":          "A ring of homes spun for gravity, so down is wherever the engineers decided.",
	"bio_reactor_farm":      "Living systems engineered to produce food, energy, and faint unease.",
	"molecular_synthesizer": "We assemble matter from the bottom up, which makes 'making' a strong word.",
	"theoretical_institute": "Where the smartest people alive argue about things that may not exist.",
	"quantum_chapel":        "A place of worship where the congregation is, in a sense, both present and not.",
	"plasma_command":        "War, waged with bottled stars, by people very confident in their containment fields.",
	"energy_exchange":       "Where power itself is bought and sold, and the lights flicker with the market.",
	"fusion_reactor":        "We finally built a star in a box. The box is very, very expensive.",
	"exotic_matter_forge":   "We forge matter that shouldn't exist into shapes that shouldn't hold.",
	"fusion_reactor_array":  "Several captive stars, working in shifts, lighting a civilization that never dims.",
	"stellar_cradle":        "We are learning to grow stars. The hubris is, frankly, magnificent.",

	// ===== SPACE AGE — vast, silent, full of our garbage =====
	"orbital_habitat":        "Home, in orbit, where the view is spectacular and the rent is astronomical.",
	"hydroponic_bay":         "Crops grown without soil, in space, which the plants are taking surprisingly well.",
	"asteroid_crystal_mine":  "We mine asteroids now, having run out of planet to disappoint.",
	"deep_space_observatory": "We point our finest instruments at the void, and the void points back.",
	"orbital_sanctuary":      "A place to pray in orbit, closer to the heavens, technically.",
	"space_force_base":       "Military supremacy, extended upward, because two dimensions of conflict felt limiting.",
	"launch_complex":         "Where we leave the planet, repeatedly, at enormous expense, on purpose.",
	"orbital_refinery":       "Refining ore in zero gravity, where the slag has nowhere polite to go.",
	"solar_collector_array":  "We catch the raw sunlight before it even reaches the ground. The ground hasn't complained.",
	"zero_g_gallery":         "Art that floats, viewed by patrons who also float, judged by critics who somehow still float above it all.",
	"dyson_scaffold":         "The first scaffolding for a structure meant to swallow a star whole. Early days.",

	// ===== INTERSTELLAR AGE — finally, some peace =====
	"generation_ship":       "A ship so slow that the passengers who arrive were born aboard, to grandparents they buried in space.",
	"protein_synthesizer":   "Food, synthesized from raw elements, for crews who've forgotten what a tomato was.",
	"reality_matter_weaver": "We weave matter from the underlying fabric, and try not to pull a loose thread.",
	"stellar_core_drill":    "We drill into the hearts of stars, which is going about as well as it sounds.",
	"xenology_institute":    "The study of alien life, currently a very well-funded department with no samples.",
	"void_monastery":        "Monks who meditate on the empty space between stars, of which there is a great deal.",
	"galactic_trade_hub":    "Commerce spanning light-years, where the shipping delays are measured in lifetimes.",
	"warp_drive_plant":      "Where we build engines that fold space, so the universe finally gets out of our way.",
	"antimatter_forge":      "We manufacture the opposite of matter, carefully, in a facility everyone parks far from.",
	"cultural_beacon":       "Broadcasting our art across the stars, in the hope someone out there has taste.",
	"warp_nexus":            "A gateway that folds the galaxy together, sponsored by people who hate commuting.",

	// ===== GALACTIC AGE — too large to govern =====
	"dyson_sphere_habitat":     "A home built around an entire star, where the power bill is, finally, zero.",
	"matter_converter":         "We turn any matter into any other matter, which made the concept of 'value' deeply nervous.",
	"neutron_star_mine":        "We mine the densest matter in existence, one impossibly heavy teaspoon at a time.",
	"cosmic_research_station":  "Where we study the universe at the scale where it stops making sense.",
	"stellar_shrine":           "Worship for a people whose gods now seem, dimensionally speaking, like small fry.",
	"stellar_armada_hq":        "Command for a fleet so vast that its left flank and right flank keep different calendars.",
	"consciousness_upload_hub": "Where minds are copied into the network, and the original is asked to please step aside.",
	"civilization_archive":     "Everything we've ever been, backed up, in case the universe asks for a receipt.",
	"cosmic_beacon":            "A signal to the cosmos announcing we exist, which may have been premature.",

	// ===== QUANTUM AGE — existential dread, superposed =====
	"quantum_vault":        "Stores resources in a superposition of 'we have it' and 'oh no.'",
	"quantum_cultivator":   "Grows food across several probable timelines, then harvests the one that worked.",
	"reality_academy":      "Where students learn that the rules are negotiable, then negotiate badly.",
	"transcendence_hall":   "Where the faithful prepare to stop being, which the brochure calls 'graduation.'",
	"probability_war_room": "We don't fight wars anymore; we calculate which ones we've already won.",
	"probability_market":   "Trading in outcomes that haven't happened, to buyers who may not, strictly, exist.",
	"reality_forge":        "We hammer the laws of physics into more convenient shapes. The physics is filing a grievance.",
	"reality_processor":    "It computes by asking the universe directly. The universe is slow to respond.",
	"reality_art_engine":   "Generates beauty by editing reality itself, which the old painters would have called cheating.",
	"reality_anchor":       "The only thing keeping the local universe from drifting off mid-sentence. Don't lean on it.",
	"zero_point_generator": "We draw power from the emptiness of space, which had been saving it for something.",

	// ===== TRANSCENDENT AGE — the quiet ascension =====
	"singularity_core":       "The heart of everything you've become, humming at a frequency the universe finds presumptuous.",
	"transcendent_nexus":     "A home for beings who have outgrown the concept of homes, but kept one for nostalgia.",
	"omniversal_war_council": "Strategy across every universe at once, for conflicts that may be hypothetical, or may be next door.",
	"omniversal_bazaar":      "A marketplace spanning all realities, where someone, somewhere, is always getting a deal.",
	"singularity_engine":     "It does everything, everywhere, instantly. We've stopped asking how. We've stopped asking why.",

	// ===== STORAGE — one per age, the warehouse's existential journey =====
	"classical_vault":    "A reinforced vault for a civilization that finally has things worth stealing.",
	"renaissance_vault":  "Storage as architecture — even the warehouse wants to be admired now.",
	"colonial_warehouse": "Where the spoils of empire are stacked, inventoried, and quietly fought over.",
	"industrial_depot":   "A cathedral of crates, where surplus is worshipped and nothing is ever thrown away.",
	"victorian_vault":    "A vault built with the firm Victorian conviction that more is, in fact, more.",
	"electric_warehouse": "Storage with the lights on, so you can see exactly how much you've hoarded.",
	"atomic_vault":       "A bunker for your belongings, rated to outlast both you and the surface world.",
	"modern_depot":       "Logistics perfected — your goods now have a better tracking system than you do.",
	"info_vault":         "Where data is stored, indexed, and never, ever deleted, just in case.",
	"digital_archive":    "Everything you own, digitized and backed up, including things you forgot you owned.",
	"cyber_vault":        "Encrypted storage so secure that you may also lose access forever. Fair trade.",
	"fusion_vault":       "Storage powered by a captive star, because keeping the lights on the inventory mattered.",
	"orbital_depot":      "A warehouse in orbit, where your surplus circles the planet, judging it gently.",
	"stellar_vault":      "Storage spanning star systems, for an empire that has frankly lost track of its stuff.",
	"galactic_vault":     "A vault the size of a solar system, mostly empty, which is somehow worse.",

	// ===== MONUMENTS — the special wonders =====
	"grand_amphitheatre_monument": "A monument to spectacle, built so future ages know we knew how to have a good time.",
	"eternal_library_monument":    "A monument meant to outlast every book inside it, which is a grim sort of optimism.",

	// ===== LATE EXTRACTION / METALLURGY / ENERGY — running out of planet =====
	"deep_iron_mine":            "We've dug so deep the iron is starting to feel like a personal favor.",
	"nano_drill_complex":        "Microscopic drills, working in their billions, digging things you'll never see.",
	"exotic_mineral_extractor":  "We extract minerals that the periodic table is still nervous about.",
	"reality_excavator":         "We dig into reality itself and haul up whatever the universe left lying around.",
	"smart_refinery":            "A refinery that runs itself and only emails you when something's catastrophically wrong.",
	"quantum_organic_extractor": "Harvesting organic matter from probability, which the matter finds intrusive.",
	"cosmic_organic_works":      "Life's raw materials, refined at a scale that should require a license.",
	"reality_harvester":         "We reap the underlying stuff of existence, leaving the universe slightly threadbare.",
	"advanced_alloy_plant":      "Alloys engineered for an age that asks 'but can it survive a reactor breach?'",
	"titanium_smelter":          "We smelt titanium at industrial scale, for an era that builds nothing small.",
	"aerospace_foundry":         "Where we forge the metals that leave the ground and refuse to come back.",
	"nano_alloy_plant":          "Metal assembled atom by atom, stronger than anything that occurs by accident.",
	"dark_matter_refinery":      "We refine matter we can't see into things we can't explain. Profitable, though.",
	"stellar_metallurgy":        "Forging metal in the furnace of a living star, because the smithy felt cramped.",
	"quantum_metal_works":       "Metal that is, technically, several different shapes until you measure it.",
	"oil_refinery":              "The old black gold, refined into a thousand modern conveniences and one persistent regret.",
	"smart_energy_grid":         "Power that routes itself, balances itself, and never explains its decisions.",
	"quantum_battery_array":     "Stores energy in states that exist and don't, charging while you're not looking.",
	"dark_energy_tap":           "We tap the force pulling the universe apart, and bill it monthly.",
	"pulsar_tap":                "We siphon power from a spinning corpse of a star, which is metal in every sense.",
	"quasar_tap":                "Drawing power from the brightest objects in the universe, modestly.",

	// ===== LATE ENGINEERING / MILITARY / TRADE / CULTURE / NETWORKS =====
	"power_grid_hub":        "The nerve center of the grid, where a single bad day means a very dark city.",
	"smart_grid_node":       "A grid that thinks, balancing power so well you forget electricity was ever hard.",
	"dyson_assembly":        "The factory that builds the megastructure that swallows the sun. Big quarter ahead.",
	"fleet_command":         "Where admirals move fleets across light-years like very expensive chess pieces.",
	"asteroid_market":       "A bazaar carved into a tumbling rock, where the prices and the gravity both float.",
	"stellar_exchange":      "A stock market spanning star systems, where the closing bell takes a year to ring.",
	"neural_art_complex":    "Art generated directly from the artist's brain, no longer requiring talent or hands.",
	"quantum_server_farm":   "Servers computing in parallel realities, returning the answer from whichever finished first.",
	"orbital_data_relay":    "Bouncing the world's information off satellites, so the cloud now has an altitude.",
	"galactic_network_node": "A relay knitting the galaxy into one conversation, with truly heroic lag.",
	"harbor_authority":      "The bureaucracy of the sea, where every ship waits its turn and resents it.",
	"container_terminal":    "Global trade reduced to stackable boxes, sorted by cranes that never sleep or strike.",
	"logistics_hub":         "The brain of global shipping, routing everything to everywhere just in time, mostly.",
	"reality_fold":          "Housing folded into a pocket of space, so the address is technically a paradox.",
}

// applyBuildingFlavor stamps the cosmetic Flavor field onto each BuildingDef in
// place from the buildingFlavor map. Buildings with no map entry are left with
// an empty Flavor (rendered as no extra line). Called once at the BaseBuildings()
// chokepoint so every consumer inherits the text.
func applyBuildingFlavor(defs []BuildingDef) {
	for i := range defs {
		if f, ok := buildingFlavor[defs[i].Key]; ok {
			defs[i].Flavor = f
		}
	}
}
