package flavor

// catalog_eras.go holds the ERA-GATED half of the fragment banks: the lines whose
// imagery is the whole charm and which therefore must never leave their stretch
// of history. Carts and gate wardens narrate the feudal ages; sidings and the
// telegraph narrate the industrial ones; a raid on an orbital store is described
// in bays and transponders and nothing else.
//
// Nothing in this file is ever drawn by an ungated template. Each bank is named
// <moment>_core_<era> or <moment>_tail_c_<era>, and eraTemplates() in catalog.go
// is the only thing that references them — always with a matching Eras
// constraint. TestNoEraBleed checks both directions: that these fragments cannot
// reach a foreign age, and that no era-coded word has crept into the ungated
// banks in catalog.go.
//
// The five buckets, and the imagery each one owns:
//
//	ancient     primitive_age .. classical_age   fires, elders, spears, herds, hides
//	feudal      medieval_age .. colonial_age     carts, gates, bells, clerks, musters
//	industrial  industrial_age .. modern_age     depots, sidings, the telegraph, the yard
//	digital     information_age .. fusion_age    feeds, uplinks, drones, analysts
//	cosmic      space_age .. transcendent_age    holds, bays, hulls, relays, the dark
//
// Adding an era to eraBuckets means adding a _core_ and a _tail_c_ bank here for
// every Moment; the template expansion and the anchor set follow automatically.
//
// Voice is the same as everywhere else — dry, in-world, the joke on the world's
// logistics rather than on the player. An era bank is not licence to be florid.
var eraBanks = map[string][]string{

	// ========================================================================
	// EXPEDITION SUCCESS
	// ========================================================================

	"exp_success_core_ancient": {
		"The party comes back to the fires with a great deal more than the fires expected",
		"The elders are shown what was found and decline to look surprised",
		"The hunt goes well, and the telling of it goes considerably better",
		"What was carried back is set down in the middle of everything so nobody can miss it",
		"They come home ahead of the weather and heavier than they went out",
		"The marks cut for this one are cut deeper than the marks usually are",
		"The far country gave up rather more than it was asked for",
	},
	"exp_success_tail_c_ancient": {
		"the fires are built higher than they need to be",
		"somebody is already cutting the marks for it",
		"the elders are told last, as they always are",
		"the meat is shared out before anyone has finished counting it",
		"the whole camp finds a reason to walk past the pile",
	},

	"exp_success_core_feudal": {
		"The column straggles in at dusk, loaded and unbothered",
		"The return party is met at the gate by a clerk with a very long list",
		"What came back fills more carts than anyone budgeted for",
		"The party arrives home on schedule, which the quartermaster finds suspicious",
		"The work is done and the proof of it is stacked by the gate",
		"The wagons are unloaded in the market square while the whole town finds a reason to pass through it",
		"A clerk is sent out to count it and comes back for a longer roll of parchment",
	},
	"exp_success_tail_c_feudal": {
		"the quartermaster is already complaining about where to put it",
		"the gate is kept open until the last cart is through",
		"a crowd forms at the gate, mostly to see what is in the carts",
		"the bells are rung, briefly, by someone who was not asked to",
		"the storehouse doors are propped open for the afternoon",
	},

	"exp_success_core_industrial": {
		"The consignment reaches the depot a day early and blocks the platform",
		"The works take delivery of a great deal more than the paperwork allowed for",
		"The telegraph brings word ahead of the train, and the word is good",
		"The expedition returns under budget, and the office declines to believe the figures",
		"Everything is weighed, stamped, and shifted to the far end of the yard by evening",
		"The freight comes in heavy and the siding is not long enough to hold it",
		"The venture returns a profit and the shareholders hear of it before the workers do",
	},
	"exp_success_tail_c_industrial": {
		"the yard is blocked for the rest of the shift",
		"the foreman is asked to sign for it and asks for a longer form",
		"the figures are on the telegraph before the crates are open",
		"the siding is full and the next train is already due",
		"somebody proposes a photograph and is taken entirely seriously",
	},

	"exp_success_core_digital": {
		"The team comes back inside the window and the after-action file is mercifully short",
		"Everything the operation flagged as recoverable has been recovered",
		"The recovery is confirmed on three separate channels before anyone will believe it",
		"The take is larger than the projection, and the projection was already optimistic",
		"The uplink carries the good news home a clear hour ahead of the team",
		"The archive gains a very satisfying entry and a very long attachment",
		"The drones come back full and the analysts come back smug",
	},
	"exp_success_tail_c_digital": {
		"the after-action file is three lines long and thoroughly smug",
		"the numbers are on every screen in the building by mid-morning",
		"somebody puts it on the network before anyone senior has read it",
		"the analysts are asked to explain it and cannot",
		"the drones are still coming in when the celebrating starts",
	},

	"exp_success_core_cosmic": {
		"The crew comes back with full holds and unspent contingency",
		"The team returns on schedule and the manifest checks out, for once",
		"The return is quiet and the cargo seals are entirely intact",
		"The ship comes in heavy, which the dock crews can tell from the way she handles",
		"They come back with the bays full and the reactor barely warm",
		"Everything the far dark was holding is now sitting on your side of it",
		"The relay carries the tally in ahead of the hull, and the tally is a good one",
	},
	"exp_success_tail_c_cosmic": {
		"the bays stay open for half a watch",
		"the dock crews are already complaining about where to put it",
		"the manifest is rewritten twice on the approach",
		"the hull is still cooling when the arguing starts",
		"the relay carries three versions of it before she is tied off",
	},

	// ========================================================================
	// EXPEDITION FAILURE
	// ========================================================================

	"exp_fail_core_ancient": {
		"They come back to the fires with empty hands and a very long story",
		"The elders are told what happened and choose to say nothing at all",
		"The hunt fails, and the telling of it fails harder",
		"Whatever was out past the hills is still out past the hills",
		"The party comes home thin and quiet and a good deal earlier than anyone wanted",
		"No marks are cut for this one, and everybody notices",
		"The country beyond the ridge turns out to belong to somebody already",
	},
	"exp_fail_tail_c_ancient": {
		"the fires are not built up that night",
		"nobody cuts a mark for it",
		"the elders ask one question and then let it go",
		"the hills are given more credit for it than the enemy",
		"the party eats at the fire and is not made to talk",
	},

	"exp_fail_core_feudal": {
		"The column comes home the long way, quietly",
		"The expedition ends without ceremony and with a good many fewer carts",
		"The gate is opened for the party and closed again rather quickly",
		"A clerk meets them at the gate, writes four words, and goes back inside",
		"The wagons come back lighter than they went out, which is the wrong way round",
		"The bells are not rung, and everyone notices that they are not rung",
		"The quartermaster looks at what returned and asks nothing at all",
	},
	"exp_fail_tail_c_feudal": {
		"the gate is opened without any bells at all",
		"the crowd at the gate thins once the carts are counted",
		"a clerk is sent down and comes back with very little to write",
		"the quartermaster says nothing, loudly",
		"the carts are left standing where they stopped",
	},

	"exp_fail_core_industrial": {
		"The consignment arrives short and the paperwork arrives shorter",
		"The works take delivery of an apology and half a load",
		"The telegraph brings word ahead of the train, and the word is bad",
		"The venture returns over budget and under weight",
		"The freight is unloaded in twenty minutes by men who had expected a long day",
		"The figures are checked, rechecked, and filed where they will not be found",
		"The depot clears the platform early, there being nothing much to put on it",
	},
	"exp_fail_tail_c_industrial": {
		"the siding is cleared in a quarter of an hour",
		"the foreman signs for it without reading it",
		"the office writes the loss down as an operating expense",
		"the telegraph is used sparingly for a day or two",
		"the yard is quiet for the rest of the shift",
	},

	"exp_fail_core_digital": {
		"The mission closes amber and the after-action file runs to forty pages",
		"The recovery falls short of the projection, and the projection was already hedged",
		"The team comes back inside the window with nothing much to put in it",
		"Half of what was logged outbound is not logged inbound",
		"The debrief takes rather longer than the operation did",
		"The uplink goes quiet for two days and comes back apologetic",
		"The drones come home damaged and the analysts come home defensive",
	},
	"exp_fail_tail_c_digital": {
		"the after-action file is long and says nothing",
		"the projection is quietly revised and nobody is told",
		"the analysts are asked to explain it and explain it at length",
		"nothing goes on the network for a day or so",
		"the operations desk stops taking calls about it",
	},

	"exp_fail_core_cosmic": {
		"The crew comes home with the hull intact and very little inside it",
		"The manifest and the cargo do not agree, and the cargo wins",
		"The ship comes in light, which the dock crews can tell before she is tied off",
		"The mission returns degraded and badly under-provisioned",
		"The bays are opened, looked into, and closed again",
		"The far dark keeps most of what it was asked to give up",
		"The survey returns early, having found the vacuum uncooperative",
	},
	"exp_fail_tail_c_cosmic": {
		"the bays are shut again inside the hour",
		"the manifest is corrected downward on the approach",
		"the dock crews find nothing to do and do it slowly",
		"the relay carries one short sentence and no detail at all",
		"the crew comes down the ramp without hurrying",
	},

	// ========================================================================
	// ENCOUNTER — STANDOFF
	// ========================================================================

	"enc_standoff_core_ancient": {
		"Two patrols meet at a ford, take each other's measure, and go around",
		"A watch-fire is spotted, answered, and then left to burn by both sides",
		"An arrow is loosed by somebody nervous and lands in nothing at all",
		"Someone shouts across the water and nobody shouts back",
		"Both parties reach the same spring and take turns at it without a word",
		"Spears are held rather than raised, which takes rather more nerve",
		"The far fires are counted from the ridge and then left alone",
	},
	"enc_standoff_tail_c_ancient": {
		"the fires on both ridges burn all night and nothing crosses between them",
		"nobody reaches for a spear, which takes some doing",
		"the elders hear a considerably braver version of it",
		"the river is left to itself for the rest of the day",
		"two scouts drink from the same spring and never mention it",
	},

	"enc_standoff_core_feudal": {
		"The enemy pickets are counted from a ridge and then avoided entirely",
		"The two columns spend the day marching parallel and pretending not to",
		"Two banners are shown to one another across a valley and then put away",
		"A herald is very nearly sent, and then very nearly sent again",
		"The riders sight each other at the crossroads and both take the longer road",
		"Both companies make camp within sight of the other and neither posts a challenge",
		"The muster on both sides is called out, looked at, and sent home",
	},
	"enc_standoff_tail_c_feudal": {
		"the banners are furled before anyone can be accused of showing them",
		"the riders are back before supper with nothing at all to report",
		"the gate wardens hear three versions of it by nightfall",
		"both columns keep marching and neither breaks step",
		"the heralds are stood down without ever leaving the yard",
	},

	"enc_standoff_core_industrial": {
		"Two patrols meet on the same stretch of line and each waves the other through",
		"The observation post logs a sighting, a duration, and no incident",
		"Both sides send a telegram and neither sends a shell",
		"The enemy convoy is watched from a cutting until it is well out of sight",
		"Field glasses are raised on both sides of the wire and lowered again",
		"The two patrols pass on the same road and each writes the other up as unremarkable",
		"Somebody signals entirely the wrong thing and everybody agrees to ignore it",
	},
	"enc_standoff_tail_c_industrial": {
		"the incident is telegraphed twice and believed once",
		"the observation post logs it and goes back to the crossword",
		"the convoy is allowed to pass and is counted while it does",
		"both patrols are back at the depot before dark",
		"the wire is left exactly as unpleasant as it already was",
	},

	"enc_standoff_core_digital": {
		"Both sides paint each other, hold the lock, and then let it go",
		"The contact lasts four minutes on the feed and rather longer in the room",
		"Two drones circle one another politely until one of them runs low",
		"The enemy signal is acquired, catalogued, and left entirely alone",
		"Both operations rooms watch the same thing happen and neither acts on it",
		"The channel opens, stays open, and carries nothing whatsoever",
		"Somebody's targeting system is extremely keen and is talked out of it",
	},
	"enc_standoff_tail_c_digital": {
		"the feed is reviewed twice and shows nothing either time",
		"the drones come home charged and with no footage worth keeping",
		"the channel is closed politely by whoever opened it",
		"the operations desk logs it as contact, no engagement",
		"the analysts argue about it for longer than it lasted",
	},

	"enc_standoff_core_cosmic": {
		"Two hulls hold station a long way apart and let the moment pass",
		"The contact runs six hours at range and never once closes",
		"Both ships turn broadside, consider it, and turn away again",
		"The enemy transponder is read, recorded, and left unanswered",
		"Your crew and theirs share an approach lane and each pretends the other is weather",
		"Nobody powers anything up, which at that distance takes real discipline",
		"The two watches stare at the same patch of dark and find each other in it",
	},
	"enc_standoff_tail_c_cosmic": {
		"the transponder logs are filed without comment",
		"both hulls go quiet and stay that way for a watch",
		"the relay carries one line about it and no detail",
		"the crew stands down and nobody admits to having been ready",
		"the approach lane is left to whoever wants it",
	},

	// ========================================================================
	// ENCOUNTER — AT CAPACITY
	// ========================================================================

	"enc_cap_core_ancient": {
		"The gifts are laid out by the fire and there is nowhere left to put them",
		"The elders accept the courtesy and refuse the goods, which is the polite order",
		"There is no more room in the camp for anybody else's kindness",
		"The visitors are fed at the fire and sent back over the hills with their bundles",
		"The bundles are unloaded, looked at, and then loaded up again",
		"Every hide, pot and basket in the place is already spoken for",
		"The strangers are thanked at length and given nothing to carry home but salt",
	},
	"enc_cap_tail_c_ancient": {
		"the bundles never come off the poles",
		"the elders are relieved and say so",
		"the visitors are fed twice, which settles it",
		"the hides stay rolled and stacked where they were",
		"somebody suggests a bigger store-pit and is ignored",
	},

	"enc_cap_core_feudal": {
		"The steward refuses the crates on the grounds that the last lot are still unpacked",
		"The crates go back down the road under the same wax seals they arrived with",
		"Your court can carry no more obligations this season",
		"The envoys are received at the gate, fed, and walked back to the road",
		"Every shelf in the undercroft is spoken for and two of them are bowing",
		"The carts are turned round in the yard without ever being unhitched",
		"A larger storehouse is proposed at supper and forgotten well before morning",
	},
	"enc_cap_tail_c_feudal": {
		"the wax seals go home unbroken",
		"the steward is quietly and visibly relieved",
		"the carts are turned round before anyone can be offended",
		"the envoys are given supper and directions",
		"the gate is closed behind them with some relief",
	},

	"enc_cap_core_industrial": {
		"The consignment is refused at the depot and sent back up the line",
		"The warehouse is full, the annexe is full, and the yard is being used as a warehouse",
		"The delivery note is signed, struck through, and signed again by somebody senior",
		"The gifts arrive by rail and leave by rail on the same afternoon",
		"The company accepts the compliment and declines the freight",
		"Every square foot under a roof is spoken for, and the rain is not helping",
		"The office replies that it has no space, in writing, in triplicate",
	},
	"enc_cap_tail_c_industrial": {
		"the freight goes back up the line unopened",
		"the delivery note is filed under refused, with regret",
		"the depot manager is vindicated and insufferable",
		"the warehouse doors are not opened at all",
		"somebody asks the office for an annexe to the annexe",
	},

	"enc_cap_core_digital": {
		"The shipment is refused before the crates are off the pad",
		"Your registry is full, and has been full since the last lot arrived",
		"The offer arrives on three channels and is declined on all three",
		"The delegation is met, hosted, briefed, and quietly turned around",
		"Every allocation is spoken for and the overflow is itself overflowing",
		"The goodwill is logged, acknowledged, and left standing outside",
		"Somebody sends a very polite refusal and attaches the inventory to prove it",
	},
	"enc_cap_tail_c_digital": {
		"the refusal is drafted, reviewed, and sent inside the hour",
		"the delegation is given a tour instead of an answer",
		"the inventory is attached, which settles the argument",
		"the channels go quiet and stay courteous",
		"somebody files it under offers, declined, amicable",
	},

	"enc_cap_core_cosmic": {
		"The freighter is waved off before it has finished its approach",
		"Every hold, bay and locker is spoken for, and two of them twice",
		"The gifts stay in orbit because there is nowhere below to put them",
		"The delegation is met at the dock, thanked warmly, and not invited in",
		"The cargo seals are never broken and the ship never shuts her engines down",
		"Your stores are full to the bulkheads and the bulkheads have opinions",
		"The offer is refused across a distance that makes refusing it take a week",
	},
	"enc_cap_tail_c_cosmic": {
		"the freighter holds at the beacon and then leaves",
		"the bays are never opened at all",
		"the dock crews are relieved and entirely unsurprised",
		"the cargo seals go home exactly as they came",
		"the relay carries a courteous refusal at some expense",
	},

	// ========================================================================
	// WAR RAID
	// ========================================================================

	"war_raid_core_ancient": {
		"They come out of the trees at first light and are gone before the fires are up",
		"Two of the outlying camps are burned and one is simply emptied",
		"The raiders take the herd and leave the huts, which is the worse choice for you",
		"The drums go up along the river in entirely the wrong order",
		"Spears are found in the morning and the people who threw them are not",
		"They come down out of the hills, take the grain, and go back up the hills",
		"The elders count what is missing twice and get the same answer twice",
	},
	"war_raid_tail_c_ancient": {
		"the fires are rebuilt and the fences are not",
		"the elders count the herd and say nothing",
		"somebody follows the tracks to the river and stops there",
		"the drums are sounded again, later, for nothing",
		"the huts are rebuilt before the questions are finished",
	},

	"war_raid_core_feudal": {
		"They came at dawn, took what they came for, and were gone before the horns",
		"The attack is over by the time the militia is dressed",
		"The gate held, and nothing behind the gate did",
		"A raiding party crosses at the shallow ford and does not linger",
		"Smoke goes up on the eastern road and is answered much too late",
		"The border villages are counted and one of them is short",
		"The carts are away over the border before the bells have stopped",
	},
	"war_raid_tail_c_feudal": {
		"the militia arrives in time to look at the tracks",
		"the gate wardens are questioned and have very little to add",
		"the bells are rung after the fact, for the practice",
		"the carts are over the border before the muster is called",
		"a rider is sent for help and arrives after everything",
	},

	"war_raid_core_industrial": {
		"They hit the goods yard at four in the morning and were on the branch line by five",
		"The telegraph goes dead an hour before the raid and comes back an hour after",
		"A freight is stopped, emptied, and left standing exactly where it was stopped",
		"The works lose a night's output and a night watchman's good opinion",
		"They came in on the road with lorries and left on the road with lorries",
		"The alarm is raised by telephone to an office that closed at six",
		"The raiders take the copper, the payroll, and the good tools, in that order",
	},
	"war_raid_tail_c_industrial": {
		"the siding is empty and the paperwork says otherwise",
		"the telegraph is working perfectly by the time anyone tries it",
		"the depot is searched by men who know it is pointless",
		"the foreman is questioned and resents every minute of it",
		"the loss is entered in the ledger as shrinkage",
	},

	"war_raid_core_digital": {
		"The perimeter goes dark for nine minutes and nine minutes is plenty",
		"They were inside the fence before the feed had caught up with them",
		"The raid is over before the duty officer finishes the first call",
		"Somebody walked the drones into a holding pattern and then walked in",
		"The store is emptied by people who had the schedule and the door codes",
		"The alert goes off late, loudly, and to nobody in particular",
		"They took the stock and the records, so nobody can say what the stock was",
	},
	"war_raid_tail_c_digital": {
		"the feed is reviewed and shows nine minutes of nothing",
		"the door codes are changed, which everyone agrees is a little late",
		"the duty officer writes it up in the shortest form available",
		"the drones are recalled and have nothing to add",
		"the network is quiet about it for a suspiciously long while",
	},

	"war_raid_core_cosmic": {
		"They came in out of the dark with their transponders off and left the same way",
		"The outer station is emptied between one sweep and the next",
		"A hull comes over the horizon, takes the cargo, and is gone before the guns come round",
		"They were in and out of the bays inside four minutes",
		"The raid is over long before the message about it has finished travelling",
		"Somebody hit the orbital stores and knew exactly which bay to open",
		"The dust is still settling when the last of them clears the field",
	},
	"war_raid_tail_c_cosmic": {
		"the transponder logs are empty, which is itself the finding",
		"the bays are sealed several hours after they needed to be",
		"the relay carries the bad news at the speed it always does",
		"the dock crews sweep up and say nothing",
		"the manifest is corrected downward before the dust settles",
	},
}
